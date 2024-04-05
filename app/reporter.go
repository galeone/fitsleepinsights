package app

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	vai "cloud.google.com/go/aiplatform/apiv1beta1"
	vaipb "cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb"
	"cloud.google.com/go/vertexai/genai"
	"github.com/galeone/fitsleepinsights/database/types"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	//go:embed templates/daily_report.md
	dailyReportTemplate string
)

type Reporter struct {
	user             *types.User
	predictionClient *vai.PredictionClient
	genaiClient      *genai.Client
	ctx              context.Context
}

// NewReporter creates a new Reporter
func NewReporter(user *types.User) (*Reporter, error) {
	ctx := context.Background()

	var predictionClient *vai.PredictionClient
	var err error
	if predictionClient, err = vai.NewPredictionClient(ctx, option.WithEndpoint(_vaiEndpoint)); err != nil {
		return nil, err
	}

	var genaiClient *genai.Client
	const region = "us-central1"
	if genaiClient, err = genai.NewClient(ctx, _vaiProjectID, region, option.WithCredentialsFile(_vaiServiceAccountKey)); err != nil {
		return nil, err
	}

	return &Reporter{user: user, predictionClient: predictionClient, genaiClient: genaiClient, ctx: ctx}, nil
}

// Close closes the client
func (r *Reporter) Close() {
	r.predictionClient.Close()
	r.genaiClient.Close()
}

// GenerateEmbeddings uses VertexAI to generate embeddings for a given prompt
func (r *Reporter) GenerateEmbeddings(prompt string) (embeddings pgvector.Vector, err error) {
	// Instances: the prompt
	var promptValue *structpb.Value
	if promptValue, err = structpb.NewValue(map[string]interface{}{"content": prompt}); err != nil {
		return
	}

	// PredictRequest: create the model prediction request
	// autoTruncate: false
	// https://cloud.google.com/vertex-ai/generative-ai/docs/embeddings/get-text-embeddings#generative-ai-get-text-embedding-go
	var autoTruncate *structpb.Value
	if autoTruncate, err = structpb.NewValue(map[string]interface{}{"autoTruncate": false}); err != nil {
		return
	}

	req := &vaipb.PredictRequest{
		Endpoint:   _vaiEmbeddingsEndpoint,
		Instances:  []*structpb.Value{promptValue},
		Parameters: autoTruncate,
	}

	// PredictResponse: receive the response from the model
	var resp *vaipb.PredictResponse
	if resp, err = r.predictionClient.Predict(r.ctx, req); err != nil {
		return
	}

	// Extract the embeddings from the response
	mapResponse, ok := resp.Predictions[0].GetStructValue().AsMap()["embeddings"].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("error extracting embeddings")
		return
	}
	values, ok := mapResponse["values"].([]interface{})
	if !ok {
		err = fmt.Errorf("error extracting embeddings")
		return
	}
	rawEmbeddings := make([]float32, len(values))
	for i, v := range values {
		rawEmbeddings[i] = float32(v.(float64))
	}
	// dim: 768
	embeddings = pgvector.NewVector(rawEmbeddings)
	return
}

// GenerateDailyReport generates a daily report for the given user
func (r *Reporter) GenerateDailyReport(data *UserData) (report *types.Report, err error) {
	gemini := r.genaiClient.GenerativeModel("gemini-pro")
	temperature := ChatTemperature
	gemini.Temperature = &temperature

	var builder strings.Builder
	fmt.Fprintln(&builder, "This is a markdown template you have to fill with the data I will provide you in the next message.")
	fmt.Fprintf(&builder, "```\n%s```\n\n", dailyReportTemplate)
	fmt.Fprintln(&builder, "You can find the sections to fill highlighted with \"[LLM to ...]\" with instructions on how to fill the section.")
	fmt.Fprintln(&builder, "I will send you the data in JSON format in the next message.")
	introductionString := builder.String()

	chatSession := gemini.StartChat()
	chatSession.History = []*genai.Content{
		{
			Parts: []genai.Part{
				genai.Text(introductionString),
			},
			Role: "user",
		},
		{
			Parts: []genai.Part{
				genai.Text("Send me the data in JSON format. I will fill the template you provided using this data"),
			},
			Role: "model",
		},
	}

	var jsonData []byte
	if jsonData, err = json.Marshal(data); err != nil {
		return nil, err
	}

	var response *genai.GenerateContentResponse
	if response, err = chatSession.SendMessage(r.ctx, genai.Text(string(jsonData))); err != nil {
		return nil, err
	}
	report = &types.Report{
		StartDate:  data.Date,
		EndDate:    data.Date,
		ReportType: "daily",
		UserID:     r.user.ID,
	}
	for _, candidates := range response.Candidates {
		for _, part := range candidates.Content.Parts {
			report.Report += fmt.Sprintf("%s\n", part)
		}
	}

	if report.Embedding, err = r.GenerateEmbeddings(report.Report); err != nil {
		return nil, err
	}

	return report, nil
}
