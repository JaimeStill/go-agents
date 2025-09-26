# Prompt Agent Tool Commands

## Chat

### Ollama

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -prompt "In 300 words or less, describe the Go programming language" \
  -stream # optional, removing will process all at once and return
```

### Azure API Key

```sh
AZURE_API_KEY=$(. scripts/azure/utilities/get-foundry-key.sh)

go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure.json \
  -token $AZURE_API_KEY \
  -prompt "In 300 words or less, describe Kubernetes" \
  -stream
```

### Azure Entra Token

```sh
AZURE_TOKEN=$(. scripts/azure/utilities/get-foundry-token.sh)

go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure-entra.json \
  -token $AZURE_TOKEN \
  -prompt "In 300 words or less, describe OAuth and OIDC" \
  -stream
```

## Embeddings

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.embedding.json \
  -protocol embeddings \
  -prompt "The quick brown fox jumps over the lazy dog"
```

## Vision

### Local File

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.gemma.json \
  -protocol vision \
  -images ~/Pictures/wallpapers/monks-journey.jpg \
  -prompt "Provide a comprehensive description of this image" \
  -stream
```

### Web URL

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.gemma.json \
  -protocol vision \
  -images https://ollama.com/public/ollama.png \
  -prompt "Provide a comprehensive description of this image" \
  -stream
```
