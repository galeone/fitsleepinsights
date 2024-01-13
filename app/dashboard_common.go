package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"google.golang.org/api/option"
)

type CalendarType int

const (
	verticalOffset   int = 120
	cellSize         int = 15
	marginLeft       int = 60
	marginRight      int = 30
	minCalendarWidth int = cellSize * 6 * 3
	maxCalendarWidth int = cellSize * 52

	msToMin float64 = 1.0 / (1000.0 * 60.0)
)

const (
	WeeklyCalendar CalendarType = iota
	MonthlyCalendar
	YearlyCalendar
)

func nextMultipleOfTen(n int) int {
	// https://stackoverflow.com/a/2403917/2891324
	return ((n + 9) / 10) * 10
}

func twoDecimals(n float64) float64 {
	return math.Round(n*100) / 100
}

func min2ddhhmm(totalMin float64) string {
	days := int64(totalMin / 1440)
	hours := int64(totalMin/60) % 24
	minutes := int64(totalMin) % 60

	if days == 0 && hours == 0 {
		return fmt.Sprintf("%02dm", minutes)
	} else if days == 0 {
		return fmt.Sprintf("%02dh %02dm", hours, minutes)
	}

	return fmt.Sprintf("%dd %02dh %02dm", days, hours, minutes)
}

// globalChartSettings returns the global settings for the calendar charts
// these are the settings that all the calendars should have in common
func globalChartSettings(calendarType CalendarType, numYears int) charts.GlobalOpts {
	var contentWidth int
	switch calendarType {
	case WeeklyCalendar, MonthlyCalendar:
		contentWidth = minCalendarWidth
	case YearlyCalendar:
		contentWidth = maxCalendarWidth
	}
	return charts.WithInitializationOpts(opts.Initialization{
		Theme:  "dark",
		Height: fmt.Sprintf("%dpx", verticalOffset+numYears*(verticalOffset+30)),
		Width:  fmt.Sprintf("%dpx", contentWidth+marginLeft+marginRight),
	})
}

func globalCalendarSettings(calendarType CalendarType, id, year int, coveredMonthsPerYear map[int]map[int]bool, firstActivityDate time.Time) *opts.Calendar {
	var calendarRange []string = make([]string, 0, 2)
	var orient string = "horizontal"

	// Depending on the number of months covered by the data, we define a different range
	// in order to create a calendar without too many empty cells
	months := make([]int, 0, len(coveredMonthsPerYear[year]))
	for k := range coveredMonthsPerYear[year] {
		months = append(months, k)
	}
	sort.Ints(months)

	if calendarType == MonthlyCalendar {
		calendarRange = append(calendarRange, fmt.Sprintf("%d-%02d", year, months[0]))
		if months[0] == 12 {
			calendarRange = append(calendarRange, fmt.Sprintf("%d-%02d", year+1, 1))
		} else {
			calendarRange = append(calendarRange, fmt.Sprintf("%d-%02d", year, months[0]+1))
		}
	} else if calendarType == YearlyCalendar {
		calendarRange = append(calendarRange, fmt.Sprintf("%d", year))
	} else if calendarType == WeeklyCalendar {
		// Weekly calendar: get an activity date, extract the first day of the week and use it as the starting point
		weekStartDay := GetStartDayOfWeek(firstActivityDate)
		calendarRange = append(calendarRange, weekStartDay.Format(time.DateOnly))
		calendarRange = append(calendarRange, weekStartDay.AddDate(0, 0, 7).Format(time.DateOnly))
	}

	return &opts.Calendar{
		Orient: orient,
		Silent: false,
		Range:  calendarRange,
		Top:    fmt.Sprintf("%d", verticalOffset+id*(verticalOffset+30)),
		Left:   "center", //fmt.Sprintf("%d", marginLeft),
		// Right:    "30", keeping this commented allows us to have cell of the same sizes
		CellSize: fmt.Sprintf("%d", cellSize),
		ItemStyle: &opts.ItemStyle{
			BorderWidth: 0.5,
		},
		YearLabel: &opts.CalendarLabel{
			Show: true,
		},
		DayLabel: &opts.CalendarLabel{
			Show:  true,
			Color: "white",
		},
		MonthLabel: &opts.CalendarLabel{
			Show:  true,
			Color: "white",
		},
	}
}

func globalVisualMapSettings(maxValue int, visualMapType string) opts.VisualMap {
	return opts.VisualMap{
		Type:       visualMapType,
		Calculable: true,
		Max:        float32(nextMultipleOfTen(maxValue)),
		Show:       true,
		Orient:     "horizontal",
		Left:       "center",
		Top:        "30",
		TextStyle: &opts.TextStyle{
			Color: "white",
		},
	}
}

func globalTitleSettings(title string) opts.Title {
	return opts.Title{
		Title: title,
		Top:   "5",
		Left:  "center",
	}
}

func globalLegendSettings() opts.Legend {
	return opts.Legend{
		Show: true,
		Type: "scroll",
		Top:  "30",
		Left: "center",
	}
}

func describeChartContent(chart *charts.BaseConfiguration, chartType string, additionalPrompts ...string) (string, error) {
	var builder strings.Builder
	fmt.Fprintln(&builder, "You are an expert in neuroscience focused on the connection between physical activity and sleep.")
	fmt.Fprintln(&builder, "You are asked to describe a chart whose content is related to your domain of expertise.")
	fmt.Fprintf(&builder, "This is the data used to generate the chart titled %s\n", chart.Title.Title)
	fmt.Fprintf(&builder, "The data is in the format of a %s", chartType)

	fmt.Fprintln(&builder, "Here's the data in JSON format")
	for _, series := range chart.MultiSeries {
		seriesInfo := map[string]interface{}{
			"name": series.Name,
			"data": series.Data,
		}
		if jsonData, err := json.Marshal(seriesInfo); err != nil {
			return "", err
		} else {
			fmt.Fprintf(&builder, "%s\n", jsonData)
		}
	}

	for _, additionalPrompt := range additionalPrompts {
		fmt.Fprintf(&builder, "%s\n", additionalPrompt)
	}

	fmt.Fprintln(&builder, "Add to the description hints and insights about the data.")
	fmt.Fprintln(&builder, "Talk directly to the user that generated the data.")
	fmt.Fprintln(&builder, "The description must be at most 200 words long.")
	fmt.Fprintln(&builder, "The description must be in Markdown.")
	fmt.Fprintln(&builder, "Do not say hi, hello, or anything that is not related to the data.")
	fmt.Fprintln(&builder, "Do not talk about the chart colors, focus only on the data and potential correlations with the sleep/activity habits.")

	description := builder.String()

	ctx := context.Background()
	var client *genai.Client
	var err error
	const region = "us-central1"
	if client, err = genai.NewClient(ctx, _vaiProjectID, region, option.WithCredentialsFile(_vaiServiceAccountKey)); err != nil {
		return "", err
	}
	defer client.Close()

	// For text-only input, use the gemini-pro model
	model := client.GenerativeModel("gemini-pro")
	var resp *genai.GenerateContentResponse
	if resp, err = model.GenerateContent(ctx, genai.Text(description)); err != nil {
		return "", err
	}
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned")
	}

	return fmt.Sprint(resp.Candidates[0].Content.Parts[0]), nil
}
