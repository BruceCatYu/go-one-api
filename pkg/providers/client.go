package providers

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/openai/openai-go"
)

const (
	ProviderOpenai = "openai"
	ProviderVolc   = "volc"
	ProviderAli    = "ali"
	ProviderGemini = "gemini"
)

type Client interface {
	FromOpenaiFormat(context.Context, string, string, map[string]json.RawMessage) (any, bool, error)
	Embedding(context.Context, string, string, *openai.EmbeddingNewParams) (any, error)
}
type SseStream interface {
	NextChunk() (any, bool)
	Close()
}

func GetClient(provider string) Client {
	switch provider {
	case ProviderOpenai:
		return GetOpenai()
	case ProviderVolc:
		return GetVolc()
	case ProviderAli:
		return GetAli()
	case ProviderGemini:
		return GetGemini()
	default:
		return nil
	}
}
