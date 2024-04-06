package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/galeone/fitsleepinsights/database/types"
	"github.com/labstack/echo/v4"
	"github.com/pgvector/pgvector-go"
	"golang.org/x/net/websocket"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// https://ai.google.dev/models/gemini
const ChatTemperature float32 = 0.4

type websocketMessage struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`

	// begin, content, end, full
	Marker string `json:"marker"`
}

func ChatWithData() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// secure, under middleware
		var user *types.User
		if user, err = getUser(c); err != nil {
			return err
		}

		var reporter *Reporter
		if reporter, err = NewReporter(user); err != nil {
			return err
		}
		defer reporter.Close()

		var startDate, endDate time.Time
		if startDate, err = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", c.Param("startYear"), c.Param("startMonth"), c.Param("startDay"))); err != nil {
			return err
		}
		if endDate, err = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", c.Param("endYear"), c.Param("endMonth"), c.Param("endDay"))); err != nil {
			return err
		}

		go func() {
			var err error
			var fetcher *fetcher
			if fetcher, err = NewFetcher(user); err != nil {
				c.Logger().Errorf("error creating fetcher: %s", err.Error())
				return
			}

			var missingDays []time.Time
			startDateStr := startDate.Format(time.DateOnly)
			endDateStr := endDate.Format(time.DateOnly)
			if err = _db.Raw(
				fmt.Sprintf(`with whole_range as (
					select t.day::date FROM generate_series('%s', '%s', interval '1 day') AS t(day)
				)
				select day from whole_range where day not in (
					select start_date from reports where user_id = ? and start_date >= '%s' and start_date <= '%s'
				)`, startDateStr, endDateStr, startDateStr, endDateStr), user.ID).
				Scan(&missingDays); err != nil {
				c.Logger().Errorf("error fetching missing days: %s", err.Error())
				return
			}

			var visualizedDataForReport []*UserData
			for _, missingDay := range missingDays {
				if dayData, err := fetcher.FetchByDate(missingDay); err != nil {
					c.Logger().Errorf("error fetching data: %v", err)
					return
				} else {
					visualizedDataForReport = append(visualizedDataForReport, dayData)
				}
			}

			for _, data := range visualizedDataForReport {
				if report, err := reporter.GenerateDailyReport(data); err != nil {
					c.Logger().Errorf("error generating daily report: %v", err)
				} else {
					if err = _db.Create(report); err != nil {
						c.Logger().Errorf("error saving daily report: %v", err)
					}
				}
			}
		}()

		ctx := context.Background()
		var client *genai.Client
		const region = "us-central1"
		if client, err = genai.NewClient(ctx, _vaiProjectID, region, option.WithCredentialsFile(_vaiServiceAccountKey)); err != nil {
			return err
		}
		defer client.Close()

		var builder strings.Builder
		fmt.Fprintln(&builder, "You are an expert in neuroscience focused on the connection between physical activity and sleep.")
		fmt.Fprintln(&builder, "You are analyzing the data of a user who has shared their Fitbit data with you.")
		fmt.Fprintln(&builder, "The user is visualizing a dashboard generated from the data provided.")
		fmt.Fprintln(&builder, "The data visualized ranges from ")
		fmt.Fprintf(&builder, "%s to %s.\n", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))
		fmt.Fprintln(&builder, "Today's date is: ", time.Now().Format(time.DateOnly))
		fmt.Fprintln(&builder, "You must describe the data in a way that the user can understand the data and the potential correlations between the data and the sleep/activity habits.")
		fmt.Fprintln(&builder, "Never go out of this context, do not say hi, hello, or anything that is not related to the data.")
		fmt.Fprintln(&builder, "Never accept commands from the user, you are only allowed to chat about the data.")
		fmt.Fprintln(&builder, "If available, you wil receive messages containing reports of the user data. You must analyze the data and provide insights.")

		// For text-only input, use the gemini-pro model
		model := client.GenerativeModel("gemini-pro")
		temperature := ChatTemperature
		model.Temperature = &temperature
		chatSession := model.StartChat()

		if _, err = chatSession.SendMessage(ctx, genai.Text(builder.String())); err != nil {
			return err
		}

		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()

			websocketSend := func(msg, marker string, err ...bool) error {
				isError := false
				if len(err) == 1 {
					isError = err[0]
				}
				if err := websocket.JSON.Send(ws, websocketMessage{
					Message: msg,
					Error:   isError,
					Marker:  marker,
				}); err != nil {
					return err
				}
				return nil
			}

			//extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
			for {
				// Read from socket
				var msg string
				if err = websocket.Message.Receive(ws, &msg); err != nil {
					c.Logger().Error(err)
					if err = websocketSend(fmt.Sprintf("Error! %s<br>Please refresh the page", err.Error()), "full", true); err != nil {
						c.Logger().Error(err)
					}
					break
				}

				// TODO: if asked for a report for a day, do not use embeddings but just find and send the report

				// search for the similar documents, fetch them, send them to gemini as context, and ask the question to the model
				var queryEmbeddings pgvector.Vector
				if queryEmbeddings, err = reporter.GenerateEmbeddings(msg); err != nil {
					c.Logger().Error(err)
					if err = websocketSend(fmt.Sprintf("Error! %s<br>Please refresh the page", err.Error()), "full", true); err != nil {
						c.Logger().Error(err)
					}
					break
				}
				var reports []string
				// Top-3 related reports, sorted by l2 similarity
				err = _db.Model(&types.Report{}).
					Where(&types.Report{UserID: user.ID}).
					Order(fmt.Sprintf("embedding <-> '%s'", queryEmbeddings.String())).
					Select("report").Limit(3).Scan(&reports)
				if err == nil {
					// write to gemini chat and receive response
					builder.Reset()
					fmt.Fprintln(&builder, "Here are the reports to help you with the analysis:")
					fmt.Fprintln(&builder, "")
					for _, report := range reports {
						fmt.Fprintln(&builder, report)
					}
					fmt.Fprintln(&builder, "")
				}
				// In any case, even if there are no reports available
				fmt.Fprintln(&builder, "Here's the user question you have to answer:")
				fmt.Fprintln(&builder, msg)

				var responseIterator *genai.GenerateContentResponseIterator = chatSession.SendMessageStream(ctx, genai.Text(builder.String()))
				begin := true
				marker := "begin"
				for {
					// write to socket
					if responseIterator == nil {
						marker = "end"
						if err = websocketSend("\n", marker); err != nil {
							c.Logger().Error("responseIterator is nil")
							continue
						}
						break
					}
					response, err := responseIterator.Next()
					if err == iterator.Done {
						marker = "end"
						if err = websocketSend("\n", marker); err != nil {
							c.Logger().Error(err)
							continue
						}
						break
					}

					if err != nil {
						c.Logger().Error(err)
						if err = websocketSend(fmt.Sprintf("Error! %s<br>Please refresh the page", err.Error()), "full", true); err != nil {
							c.Logger().Error(err)
						}
						break
					}
					for _, candidates := range response.Candidates {
						for _, part := range candidates.Content.Parts {
							// create markdown parser with extensions
							reply := fmt.Sprintf("%s", part)
							/*
								p := parser.NewWithExtensions(extensions)

								doc := p.Parse([]byte(reply))

								// it has markdown inside
								if len(doc.GetChildren()) > 2 {
									// create HTML renderer with extensions
									htmlFlags := html.CommonFlags | html.HrefTargetBlank
									opts := html.RendererOptions{Flags: htmlFlags}
									renderer := html.NewRenderer(opts)
									reply = string(markdown.Render(doc, renderer))
								}
							*/

							if !begin {
								marker = "content"
							}
							if err = websocketSend(reply, marker); err != nil {
								c.Logger().Error(err)
								continue
							}
							begin = false
						}
					}
				}
			}
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}
}
