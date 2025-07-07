package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultModel string              `yaml:"default_model"` // Model to use if not specified or falls back to
	Models       map[string]Model    `yaml:"models"`  // custom model name
	Providers    map[string]Provider `yaml:"providers"`
	Server       Server              `yaml:"server"`
}

type Model struct {
	Model    string `yaml:"model"` // real model id
	Provider string `yaml:"provider"` // provider name
	Type     string `yaml:"type"` // model type, "chat" for now
}

type Provider struct {
	ApiKey string `yaml:"api_key"` // API key for the provider
}

type Server struct {
	Key  string `yaml:"key"` // Server key for authentication
	Port string `yaml:"port"` // Server port to listen on
}

var cfg = &Config{}

func init() {
	// Load configuration from config file
	file, err := os.Open("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("failed to open config file: %v", err))
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		panic(fmt.Sprintf("failed to decode config file: %v", err))
	}
}

func GetModelConfig(model string) *Model {
	if m, ok := cfg.Models[model]; ok {
		return &m
	}
	m := cfg.Models[cfg.DefaultModel]
	return &m
}
func GetProviderConfig(api string) *Provider{
	if a, ok := cfg.Providers[api]; ok {
		return &a
	}
	return nil
}
func GetServerConfig() *Server {
	return &cfg.Server
}
