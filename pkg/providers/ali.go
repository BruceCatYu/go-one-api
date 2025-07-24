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
)

type Ali struct {
	client *openai.Client
}
type AliStream struct {
	stream *ssestream.Stream[openai.ChatCompletionChunk]
	model  string
}

func (as *AliStream) NextChunk() (any, bool) {
	if !as.stream.Next() {
		return nil, false
	}
	chunk := as.stream.Current()
	if len(chunk.Choices) > 0 {
		chunk.Model = as.model
		chunk.Choices[0].Delta.Role = "assistant"
		return chunk, true
	}
	return nil, false
}
func (as *AliStream) Close() {
	_ = as.stream.Close()
}

var aliClient *Ali

func GetAli() *Ali {
	if aliClient == nil {
		client := openai.NewClient(
			option.WithAPIKey(config.GetProviderConfig(ProviderAli).ApiKey),
			option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1/"),
		)
		aliClient = &Ali{
			client: &client,
		}
	}
	return aliClient
}
func (a *Ali) FromOpenaiFormat(ctx context.Context, rawModel, modelId string, params map[string]json.RawMessage) (any, bool, error) {
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

	if isStream != nil && *isStream {
		return &AliStream{
			stream: a.client.Chat.Completions.NewStreaming(ctx, req),
			model:  rawModel,
		}, true, err
	}
	resp, err := a.client.Chat.Completions.New(ctx, req)
	if err != nil {
		return nil, false, err
	}
	resp.Model = rawModel
	return resp, false, err
}

func (a *Ali) Embedding(ctx context.Context, rawModel, modelId string, params *openai.EmbeddingNewParams) (any, error) {
	params.Model = modelId
	resp, err := a.client.Embeddings.New(ctx, *params)
	if err != nil {
		return nil, err
	}
	resp.Model = rawModel
	return resp, nil
}
