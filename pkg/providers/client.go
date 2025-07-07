package providers

import (
	"context"

	"github.com/goccy/go-json"
)

const (
	ProviderOpenai = "openai"
	ProviderVolc   = "volc"
	ProviderAli    = "ali"
)

type Client interface {
	FromOpenaiFormat(context.Context, string, string, map[string]json.RawMessage) (any, bool, error)
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
	default:
		return nil
	}
}
