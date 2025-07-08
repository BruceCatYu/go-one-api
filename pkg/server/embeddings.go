package server

import (
	"github.com/BruceCatYu/go-one-api/pkg/config"
	"github.com/BruceCatYu/go-one-api/pkg/providers"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
)

func embeddings(c *gin.Context) {
	body := openai.EmbeddingNewParams{}
	_ = c.BindJSON(&body)
	modelConfig := config.GetModelConfig(body.Model)
	client := providers.GetClient(modelConfig.Provider)
	if client == nil {
		c.JSON(400, gin.H{"error": "provider not found"})
		return
	}
	res, err := client.Embedding(c, body.Model, modelConfig.Model, &body)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}
