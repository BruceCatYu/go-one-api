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

type Openai struct {
	client *openai.Client
}
type OpenaiStream struct {
	stream *ssestream.Stream[openai.ChatCompletionChunk]
	model  string
}

func (os *OpenaiStream) NextChunk() (any, bool) {
	if !os.stream.Next() {
		return nil, false
	}
	chunk := os.stream.Current()
	if len(chunk.Choices) > 0 {
		chunk.Model = os.model
		return chunk, true
	}
	return nil, false
}
func (os *OpenaiStream) Close() {
	_ = os.stream.Close()
}

var openaiClient *Openai

func GetOpenai() *Openai {
	if openaiClient == nil {
		client := openai.NewClient(
			option.WithAPIKey(config.GetProviderConfig(ProviderOpenai).ApiKey),
		)
		openaiClient = &Openai{
			client: &client,
		}
	}
	return openaiClient
}
func (o *Openai) FromOpenaiFormat(ctx context.Context, rawModel, modelId string, params map[string]json.RawMessage) (any, bool, error) {
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
		return &OpenaiStream{
			stream: o.client.Chat.Completions.NewStreaming(ctx, req),
			model:  rawModel,
		}, true, err
	}
	resp, err := o.client.Chat.Completions.New(ctx, req)
	return resp, false, err
}
