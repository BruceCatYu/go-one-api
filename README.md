# O2M Relay
An openai api-compatible relay written in go.

## RUN
1. Create `config.yaml` in directory. Paste following content below and edit as needed.
```yaml
default_model: custom-seed
models:
  custom-seed:
    model: doubao-seed-1-6-250615
    provider: volc
    type: chat
  custom-qwen-turbo:
    model: qwen-turbo-2025-04-28
    provider: ali
    type: chat
providers:
  volc:
    api_key: sk-123
  ali:
    api_key: sk-123
  openai:
    api_key: sk-123
server:
  key: sk-456
  port: 8080
```
2. Start!
