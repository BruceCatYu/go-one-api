package providers

import (
	"context"
	"fmt"
	"io"

	"github.com/BruceCatYu/go-one-api/internal/utils"
	"github.com/BruceCatYu/go-one-api/pkg/config"
	"github.com/goccy/go-json"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	arkUtils "github.com/volcengine/volcengine-go-sdk/service/arkruntime/utils"
)

type Volc struct {
	client *arkruntime.Client
}
type VolcStream struct {
	stream *arkUtils.ChatCompletionStreamReader
	model  string
}

func (vs *VolcStream) NextChunk() (any, bool) {
	recv, err := vs.stream.Recv()
	if err == io.EOF {
		return nil, false
	}
	if err != nil {
		fmt.Println("Stream error:", err)
		return nil, false
	}
	if len(recv.Choices) > 0 {
		recv.Model = vs.model
		return recv, true
	}
	return nil, false
}

func (v *VolcStream) Close() {
	_ = v.stream.Close()
}

var volcClient *Volc

func GetVolc() *Volc {
	if volcClient == nil {
		volcClient = &Volc{client: arkruntime.NewClientWithApiKey(config.GetProviderConfig(ProviderVolc).ApiKey)}
	}
	return volcClient
}
func (v *Volc) FromOpenaiFormat(ctx context.Context, rawModel, modelId string, params map[string]json.RawMessage) (any, bool, error) {
	var err error
	req := &model.CreateChatCompletionRequest{
		Model: modelId,
	}

	// messages
	if messagesBytes, ok := params["messages"]; ok {
		if req.Messages, err = utils.BytesToObj[[]*model.ChatCompletionMessage](messagesBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal messages failed: %w", err)
		}
	} else {
		return nil, false, fmt.Errorf("messages is required")
	}
	// temperature
	if temperatureBytes, ok := params["temperature"]; ok {
		if req.Temperature, err = utils.BytesToObj[*float32](temperatureBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal temperature failed: %w", err)
		}
	}
	// top_p
	if topPBytes, ok := params["top_p"]; ok {
		if req.TopP, err = utils.BytesToObj[*float32](topPBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal top_p failed: %w", err)
		}
	}
	// parallel_tool_calls
	if parallelBytes, ok := params["parallel_tool_calls"]; ok {
		if req.ParallelToolCalls, err = utils.BytesToObj[*bool](parallelBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal parallel_tool_calls failed: %w", err)
		}
	}
	// stop
	if stopBytes, ok := params["stop"]; ok {
		if req.Stop, err = utils.BytesToObj[[]string](stopBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal stop failed: %w", err)
		}
	}
	// stream
	if streamBytes, ok := params["stream"]; ok {
		if req.Stream, err = utils.BytesToObj[*bool](streamBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal stream failed: %w", err)
		}
	}
	// stream options
	if streamOptionsBytes, ok := params["stream_options"]; ok {
		if req.StreamOptions, err = utils.BytesToObj[*model.StreamOptions](streamOptionsBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal stream_options failed: %w", err)
		}
	}
	// reasoning effort
	if reasoningEffortBytes, ok := params["reasoning_effort"]; ok {
		var reasoningEffort string
		if reasoningEffort, err = utils.BytesToObj[string](reasoningEffortBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal reasoning_effort failed: %w", err)
		}
		switch reasoningEffort {
		case "low":
			req.Thinking = &model.Thinking{Type: model.ThinkingTypeDisabled}
		case "medium":
			req.Thinking = &model.Thinking{Type: model.ThinkingTypeAuto}
		case "high":
			req.Thinking = &model.Thinking{Type: model.ThinkingTypeEnabled}
		default:
			return nil, false, fmt.Errorf("reasoning_effort unsupported: %s", reasoningEffort)
		}
	}
	// tools
	if toolsBytes, ok := params["tools"]; ok {
		if req.Tools, err = utils.BytesToObj[[]*model.Tool](toolsBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal tools failed: %w", err)
		}
	}
	// tool_choice
	if toolChoiceBytes, ok := params["tool_choice"]; ok {
		if req.ToolChoice, err = utils.BytesToObj[string](toolChoiceBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal tool_choice failed: %w", err)
		}
	}
	// structured output
	if responseFormatBytes, ok := params["response_format"]; ok {
		if req.ResponseFormat, err = utils.BytesToObj[*model.ResponseFormat](responseFormatBytes); err != nil {
			return nil, false, fmt.Errorf("unmarshal response_format failed: %w", err)
		}
	}

	if req.Stream != nil && *req.Stream {
		stream, err := v.client.CreateChatCompletionStream(ctx, req)
		return &VolcStream{stream: stream, model: rawModel}, true, err
	}
	resp, err := v.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, false, err
	}
	resp.Model = rawModel
	return resp, false, err
}

func (v *Volc) Embedding(ctx context.Context, rawModel, modelId string, params *openai.EmbeddingNewParams) (any, error) {
	req := model.EmbeddingRequestStrings{
		Model: modelId,
	}
	input := params.Input
	if !param.IsOmitted(input.OfString) {
		req.Input = []string{input.OfString.Value}
	} else if !param.IsOmitted(input.OfArrayOfStrings) {
		req.Input = input.OfArrayOfStrings
	}
	switch params.EncodingFormat {
	case openai.EmbeddingNewParamsEncodingFormatFloat:
		req.EncodingFormat = "float"
	case openai.EmbeddingNewParamsEncodingFormatBase64:
		req.EncodingFormat = "base64"
	}
	if !param.IsOmitted(params.User) {
		req.User = params.User.Value
	}
	if !param.IsOmitted(params.Dimensions) {
		req.Dimensions = int(params.Dimensions.Value)
	}
	resp, err := v.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}
	resp.Model = rawModel
	respBytes, _ := json.Marshal(resp)
	var respMap map[string]any
	_ = json.Unmarshal(respBytes, &respMap)
	delete(respMap, "HttpHeader")
	return respMap, nil
}
