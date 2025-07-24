package providers

import (
	"context"
	"fmt"
	"github.com/BruceCatYu/go-one-api/internal/utils"
	"github.com/BruceCatYu/go-one-api/pkg/config"
	"github.com/goccy/go-json"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"google.golang.org/genai"
)

type Gemini struct {
	client *openai.Client
}
type GeminiStream struct {
	stream *ssestream.Stream[openai.ChatCompletionChunk]
	model  string
}

func (gs *GeminiStream) NextChunk() (any, bool) {
	if !gs.stream.Next() {
		return nil, false
	}
	chunk := gs.stream.Current()
	if len(chunk.Choices) > 0 {
		chunk.Model = gs.model
		return chunk, true
	}
	return nil, false
}

func (gs *GeminiStream) Close() {
	_ = gs.stream.Close()
}

var geminiClient *Gemini

func GetGemini() *Gemini {
	if geminiClient == nil {
		client := openai.NewClient(
			option.WithAPIKey(config.GetProviderConfig(ProviderGemini).ApiKey),
			option.WithBaseURL("https://generativelanguage.googleapis.com/v1beta/openai/"),
		)
		geminiClient = &Gemini{
			client: &client,
		}
	}
	return geminiClient
}

type GoogleConfig map[string]*genai.ThinkingConfig

func (g *Gemini) FromOpenaiFormat(ctx context.Context, rawModel, modelId string, params map[string]json.RawMessage) (any, bool, error) {
	var (
		isStream *bool
		err      error
	)
	if streamBytes, ok := params["stream"]; ok {
		if isStream, err = utils.BytesToObj[*bool](streamBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal stream failed: %w", err)
		}
	}
	jsonBytes, _ := json.Marshal(params)
	var req openai.ChatCompletionNewParams
	if err = json.Unmarshal(jsonBytes, &req); err != nil {
		return nil, false, fmt.Errorf("unmarshal request failed: %w", err)
	}
	req.Model = modelId

	// reasoning effort
	if reasoningEffortBytes, ok := params["reasoning_effort"]; ok {
		var reasoningEffort string
		if reasoningEffort, err = utils.BytesToObj[string](reasoningEffortBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal reasoning_effort failed: %w", err)
		}
		thinkingConfig := &genai.ThinkingConfig{}
		switch reasoningEffort {
		case "low":
			budget := int32(0)
			thinkingConfig.ThinkingBudget = &budget
		case "medium":
			budget := int32(1024)
			thinkingConfig.ThinkingBudget = &budget
		case "high":
			budget := int32(-1)
			thinkingConfig.ThinkingBudget = &budget
		default:
			return nil, false, fmt.Errorf("reasoning_effort unsupported: %s", reasoningEffort)
		}
		googleMap := GoogleConfig{"thinking_config": thinkingConfig}
		req.SetExtraFields(map[string]any{"extra_body": googleMap})
	}

	if isStream != nil && *isStream {
		return &OpenaiStream{
			stream: g.client.Chat.Completions.NewStreaming(ctx, req),
			model:  rawModel,
		}, true, err
	}
	resp, err := g.client.Chat.Completions.New(ctx, req)
	if err != nil {
		return nil, false, err
	}
	resp.Model = rawModel
	return resp, false, nil
}

func (g *Gemini) Embedding(ctx context.Context, rawModel, modelId string, params *openai.EmbeddingNewParams) (any, error) {
	params.Model = modelId
	resp, err := g.client.Embeddings.New(ctx, *params)
	if err != nil {
		return nil, err
	}
	resp.Model = rawModel
	return resp, nil

}
