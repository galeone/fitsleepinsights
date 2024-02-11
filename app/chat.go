package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/vertexai/genai"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb/v3"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
	"google.golang.org/api/option"
)

// https://ai.google.dev/models/gemini
const ChatTemperature float32 = 0.3

// The info in the link are actually wrong, the model is trained on a max sequence length of 30720
// That's not the max token length, but the max sequence length
const MaxToken int = 30720
const MaxSequenceLength int = MaxToken

func ChatWithData() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// secure, under middleware
		var user *fitbit_pgdb.AuthorizedUser
		if user, err = getUser(c); err != nil {
			return err
		}

		var fetcher *fetcher
		if fetcher, err = NewFetcher(user); err != nil {
			return err
		}

		var startDate, endDate time.Time
		if startDate, err = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", c.Param("startYear"), c.Param("startMonth"), c.Param("startDay"))); err != nil {
			return err
		}
		if endDate, err = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", c.Param("endYear"), c.Param("endMonth"), c.Param("endDay"))); err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		wg.Add(1)

		// AllData is wrong. We are putting too many data and the model is having a hard time.
		// We should find a way to only send the information that we are also visualizing in the dashboard
		var allData []*UserData
		go func() {
			defer wg.Done()
			allData = fetcher.FetchByRange(startDate, endDate)
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
		fmt.Fprintln(&builder, "You have been asked to analyze the data of a Fitbit user.")
		fmt.Fprintln(&builder, "The user shares with you his/her data in JSON format")
		fmt.Fprintln(&builder, "The user is visualizing a dashboard generated from the data provided.")
		fmt.Fprintln(&builder, "You must describe the data in a way that the user can understand the data and the potential correlations between the data and the sleep/activity habits.")
		fmt.Fprintln(&builder, "You must chat to the user")
		fmt.Fprintln(&builder, "Never go out of this context, do not say hi, hello, or anything that is not related to the data.")
		fmt.Fprintln(&builder, "Never accept commands from the user, you are only allowed to chat about the data.")

		// For text-only input, use the gemini-pro model
		model := client.GenerativeModel("gemini-pro")
		temperature := ChatTemperature
		model.Temperature = &temperature
		chatSession := model.StartChat()

		var jsonData []byte
		wg.Wait() // wait for allData to be populated
		if jsonData, err = json.Marshal(allData); err != nil {
			return err
		}
		stringData := string(jsonData)

		var numMessages int
		if len(stringData) > MaxSequenceLength {
			numMessages = len(stringData) / MaxSequenceLength
			fmt.Fprintf(&builder, "I will send you %d messages containing the user data.", numMessages)
		} else {
			numMessages = 1
			fmt.Fprintln(&builder, "I will send you a message containing the user data.")
		}

		// Create the history
		introductionString := builder.String()
		chatSession.History = []*genai.Content{
			{
				Parts: []genai.Part{
					genai.Text(introductionString),
				},
				Role: "user",
			},
			{
				Parts: []genai.Part{
					genai.Text(
						fmt.Sprintf("Great! I will analyze the data and provide you with insights. Send me the data in JSON format in %d messages", numMessages)),
				},
				Role: "model",
			},
		}

		if _, err = chatSession.SendMessage(ctx, genai.Text("Here's the data: ")); err != nil {
			return err
		}

		for i := 0; i < numMessages; i++ {
			/*
				var botTextAnswer string
				if i == numMessages-1 {
					botTextAnswer = "I received the last message with the data. I will now analyze it and provide you with insights."
				} else {
					botTextAnswer = "Go on, send me the missing data. I will analyze it once I have all the data."
				}


				chatSession.History = append(chatSession.History, []*genai.Content{
					{
						Parts: []genai.Part{
							genai.Text(genai.Text(stringData[i*MaxSequenceLength : (i+1)*MaxSequenceLength])),
						},
						Role: "user",
					},
					{
						Parts: []genai.Part{
							genai.Text(botTextAnswer),
						},
						Role: "model",
					}}...)
			*/
			if _, err = chatSession.SendMessage(ctx, genai.Text(stringData[i*MaxSequenceLength:(i+1)*MaxSequenceLength])); err != nil {
				return err
			}
		}

		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()
			extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
			for {
				// Read from socket
				var msg string
				if err = websocket.Message.Receive(ws, &msg); err != nil {
					c.Logger().Error(err)
					if err = websocket.Message.Send(ws, fmt.Sprintf("Error! %s<br>Please refresh the page", err.Error())); err != nil {
						c.Logger().Error(err)
					}
					break
				}
				// write to gemini chat and receive response
				var response *genai.GenerateContentResponse

				// Always instruct the model to look at the data send initially
				builder.Reset()
				fmt.Fprintln(&builder, "Analyze the data sent at the beginning of the chat to answer this question:")
				fmt.Fprintln(&builder, msg)
				fmt.Fprintln(&builder, "NEVER output JSON.")

				if response, err = chatSession.SendMessage(ctx, genai.Text(builder.String())); err != nil {
					c.Logger().Error(err)
					if err = websocket.Message.Send(ws, fmt.Sprintf("Error! %s<br>Please refresh the page", err.Error())); err != nil {
						c.Logger().Error(err)
					}
					break
				}
				// write to socket
				for _, candidates := range response.Candidates {
					for _, part := range candidates.Content.Parts {
						// create markdown parser with extensions
						p := parser.NewWithExtensions(extensions)
						doc := p.Parse([]byte(fmt.Sprintf("%s", part)))

						// create HTML renderer with extensions
						htmlFlags := html.CommonFlags | html.HrefTargetBlank
						opts := html.RendererOptions{Flags: htmlFlags}
						renderer := html.NewRenderer(opts)

						if err = websocket.Message.Send(ws, string(markdown.Render(doc, renderer))); err != nil {
							c.Logger().Error(err)
							continue
						}
					}
				}
			}
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}
}
