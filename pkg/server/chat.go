package server

import (
	"io"

	"github.com/BruceCatYu/go-one-api/internal/utils"
	"github.com/BruceCatYu/go-one-api/pkg/config"
	"github.com/BruceCatYu/go-one-api/pkg/providers"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
)

func chatCompletions(c *gin.Context) {
	body := map[string]json.RawMessage{}
	_ = c.BindJSON(&body)
	var model string
	if err := json.Unmarshal(body["model"], &model); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	modelConfig := config.GetModelConfig(model)
	client := providers.GetClient(modelConfig.Provider)
	if client == nil {
		c.JSON(500, gin.H{"error": "provider not supported"})
		return
	}
	resp, isStream, err := client.FromOpenaiFormat(c, model, modelConfig.Model, body)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if !isStream {
		c.JSON(200, resp)
		return
	}
	stream := resp.(providers.SseStream)
	defer stream.Close()
	c.Stream(func(w io.Writer) bool {
		chunk, ok := stream.NextChunk()
		if !ok {
			c.SSEvent("", " [DONE]")
			return false
		}
		data, _ := utils.MarshalToSseData(chunk)
		c.SSEvent("", data)
		return true
	})
}
