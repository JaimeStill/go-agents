# Go Agents

A platform and model agnostic Go agent primitive library.

## Status: Pre-Release (v0.3.0)

**go-agents** is currently in pre-release development. The API may change between minor versions until v1.0.0 is released.

## Documentation

- **[ARCHITECTURE.md](./ARCHITECTURE.md)**: Technical specifications and implementation details
- **[PROJECT.md](./PROJECT.md)**: Project roadmap, scope, and future enhancements
- **[_context/](./_context/)**: Implementation guides and development summaries

## Current Implementation

The package provides a complete multi-protocol LLM integration system with a protocol-centric architecture:

- **Protocol-Specific Request Types**: Dedicated request types (ChatRequest, VisionRequest, ToolsRequest, EmbeddingsRequest) with protocol-appropriate fields
- **Complete Protocol Support**: All four core protocols (chat, vision, tools, embeddings) fully operational with protocol-specific response types
- **Multi-Provider Support**: Working Ollama and Azure AI Foundry providers with authentication (API keys, Entra ID)
- **OpenAI Format Standard**: Tools wrapped in OpenAI format by default, vision images embedded in message content
- **Configuration Option Merging**: Model configurations provide baseline defaults, runtime options override per request
- **Structured Content Support**: Vision protocol handles multimodal content with vision-specific options, tools protocol returns structured tool calls
- **Enhanced Development Tools**: Command-line testing infrastructure with comprehensive protocol examples
- **Human-Readable Configuration**: Duration strings ("24s", "1m") and clean JSON configuration
- **Thread-Safe Operations**: Proper connection pooling, streaming support (chat, vision, tools), and concurrent request handling
- **Mock Implementations**: Complete mock package for testing agent-based systems

## Development Status

### Current Phase: v0.3.0 - Flattened Configuration Architecture

**Completed**: Configuration hierarchy flattened with `provider` and `model` as peers with `client` at the AgentConfig level. Broke `pkg/types/` into focused packages: `pkg/protocol/` for protocol types and message structures, `pkg/response/` for response parsing. All protocols validated with Ollama and Azure.

**Active Focus**: The system is stable and ready for use. Architecture simplified with clear separation between protocol, response, and configuration concerns.

**Architecture Highlights**: Flattened configuration structure, dedicated protocol and response packages, provider-owned request marshaling, clean separation between HTTP client settings and LLM configuration.

## Getting Started

### Prerequisites

- Go 1.25 or later
- For Ollama: at a minimum, Docker. If you have an NVIDIA GPU, you'll want to install the `nvidia-container-toolkit` package and [configure Docker for CDI to work with the NVIDIA Container Toolkit](https://github.com/JaimeStill/LinuxNotes/tree/main/omarchy#configure-docker-for-cdi-to-work-with-nvidia-container-toolkit). If you just have Docker, you can still run and fallback to CPU processing, but your performance will be noticeably bad.
- For Azure: you will need to have the Azure CLI installed and authenticated to a tenant where you have [created the necessary infrastructure](./scripts/azure/README.md) to connect to a deployed Azure OpenAI model.

### Deploy an Azure OpenAI Service Model

```sh
. scripts/azure/components/cognitive-services-deployment.sh \
  --model-format "OpenAI" \
  --model-name "gpt-4o" \
  --model-version "2024-11-20" \
  --deployment-name "gpt-4o" \
  --sku "Standard" \
  --sku-capacity 10 \
  --name "GoAgentsCognitiveService" \
  --resource-group "GoAgentsResourceGroup"
```

### Basic Usage

The `tools/prompt-agent/` utility provides command-line testing of provider implementations. See the [tools/prompt-agent/README](./tools/prompt-agent/README.md) for comprehensive documentation.

#### Test with Ollama (local models)

You will first need to ensure you startup the Docker container:

```sh
docker compose up -d
```

This will automatically pull down the `llama3.2:3b` model and store it in a volume pointed to `~/.ollama`. Once the model is downloaded, you are good to prompt.

```sh
# Test with Ollama (local)
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -prompt "In 300 words or less, describe the Go programming language" \
  -stream # optional, removing will process all at once and return
```

<details>
  <summary>Configuration</summary>

  ```json
  {
    "name": "ollama-agent",
    "system_prompt": "You are an expert software architect specializing in cloud native systems design",
    "client": {
      "timeout": "24s",
      "retry": {
        "max_retries": 3,
        "initial_backoff": "1s",
        "max_backoff": "30s",
        "backoff_multiplier": 2.0,
        "jitter": true
      },
      "connection_pool_size": 10,
      "connection_timeout": "9s"
    },
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434"
    },
    "model": {
      "name": "llama3.2:3b",
      "capabilities": {
        "chat": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "top_p": 0.95
        },
        "tools": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "tool_choice": "auto"
        }
      }
    }
  }
  ```

</details>

##### Output

Go (also known as Golang) is a statically typed, compiled, designed-for-concurrency programming language developed by Google. Its primary design goals include simplicity, performance, reliability, and ease of use.

**Key Features:**

1. **Concurrency**: Go's concurrency model uses lightweight goroutines, which can run concurrently without the need for explicit thread management.
2. **Simple syntax**: Go's syntax is designed to be concise and easy to read, with a focus on simplicity over complexity.
3. **Statically typed**: Go is statically typed, which means type errors are caught at compile-time rather than runtime.
4. **Compiled language**: Go code is compiled into machine code, making it faster and more efficient than interpreted languages.

**Language Design Philosophy:**

1. **Consistency**: Go aims to be consistent in its behavior and syntax, making it easier for developers to learn and use.
2. **Composability**: Go encourages modular programming through the use of packages and interfaces.
3. **Error handling**: Go's error handling system is designed to be explicit and easy to use.

**Use Cases:**

1. **Cloud native applications**: Go is well-suited for building cloud-native applications due to its concurrency model, performance, and simplicity.
2. **Networking and distributed systems**: Go's design makes it an excellent choice for building networking and distributed systems.
3. **Microservices architecture**: Go's modular programming model and package-based design make it a popular choice for microservices architecture.

Overall, Go is a modern language that balances simplicity, performance, and concurrency, making it an attractive choice for building scalable, maintainable cloud native applications.

#### Test with Azure API Key

```sh
# Capture the Azure Foundry API key
AZURE_API_KEY=$(. scripts/azure/utilities/get-foundry-key.sh)

# Test with Azure AI Foundry (API key)
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure.json \
  -token $AZURE_API_KEY \
  -prompt "In 300 words or less, describe Kubernetes" \
  -stream
```

<details>
  <summary>Configuration</summary>

  ```json
  {
    "name": "azure-key-agent",
    "system_prompt": "You are an expert software architect specializing in cloud native systems design",
    "client": {
      "timeout": "24s",
      "retry": {
        "max_retries": 3,
        "initial_backoff": "1s",
        "max_backoff": "30s",
        "backoff_multiplier": 2.0,
        "jitter": true
      },
      "connection_pool_size": 10,
      "connection_timeout": "9s"
    },
    "provider": {
      "name": "azure",
      "base_url": "https://go-agents-platform.openai.azure.com/openai",
      "options": {
        "deployment": "o3-mini",
        "api_version": "2025-01-01-preview",
        "auth_type": "api_key"
      }
    },
    "model": {
      "name": "o3-mini",
      "capabilities": {
        "chat": {
          "max_completion_tokens": 4096
        }
      }
    }
  }
  ```

</details>

##### Output

Kubernetes, often abbreviated as K8s, is an open-source container orchestration platform designed to automate the deployment, scaling, and management of containerized applications. It abstracts the underlying infrastructure by organizing containers into the smallest deployable units called pods, which run on nodes (servers) grouped into clusters.

At its core, Kubernetes uses a declarative configuration model, allowing developers to define the desired state of their applications through configuration files. The system then continuously works to ensure that the actual state matches the desired one, providing self-healing capabilities such as automatic restarts, rescheduling of failed containers, and load balancing.

Key features include automated rollouts and rollbacks, horizontal scaling, service discovery, and management of persistent storage. Kubernetes also supports advanced networking policies and security configurations, making it a robust platform for managing microservices architectures. Its API-driven design enables seamless integrations with other cloud-native tools and services, fostering a vibrant ecosystem of extensions and custom controllers.

By decoupling application logic from infrastructure concerns, Kubernetes provides a consistent environment across different deployment landscapes—whether on public clouds, on-premises data centers, or hybrid environments. This flexibility, along with its community-driven evolution and support from major cloud providers, has made Kubernetes the de facto standard for orchestrating containerized applications in modern cloud-native environments.

#### Test with Azure Entra Auth

```sh
# Capture an Bearer token
AZURE_TOKEN=$(. scripts/azure/utilities/get-foundry-token.sh)

# Test with Azure AI Foundry (Entra ID)
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure-entra.json \
  -token $AZURE_TOKEN \
  -prompt "In 300 words or less, describe OAuth and OIDC" \
  -stream
```


<details>
  <summary>Configuration</summary>

  ```json
  {
    "name": "azure-entra-agent",
    "system_prompt": "You are an expert software architect specializing in cloud native systems design",
    "client": {
      "timeout": "24s",
      "retry": {
        "max_retries": 3,
        "initial_backoff": "1s",
        "max_backoff": "30s",
        "backoff_multiplier": 2.0,
        "jitter": true
      },
      "connection_pool_size": 10,
      "connection_timeout": "9s"
    },
    "provider": {
      "name": "azure",
      "base_url": "https://go-agents-platform.openai.azure.com/openai",
      "options": {
        "deployment": "o3-mini",
        "api_version": "2025-01-01-preview",
        "auth_type": "bearer"
      }
    },
    "model": {
      "name": "o3-mini",
      "capabilities": {
        "chat": {
          "max_completion_tokens": 4096
        }
      }
    }
  }
  ```

</details>

##### Output

OAuth (Open Authorization) is an open standard for delegated authorization. It enables third-party applications to access user resources on a service without requiring users to share their credentials. Instead, the user grants a permission token (access token) that defines what resources the application can access, and for how long. OAuth focuses solely on resource authorization, not user identity verification.

OIDC (OpenID Connect) builds on OAuth 2.0 by introducing an additional layer for user authentication. While OAuth provides secure authorization for resource access, OIDC adds the means to verify a user's identity. It does this through an ID token—a JSON Web Token (JWT) that carries information about the user and the authentication event. OIDC simplifies user login and enables applications to obtain basic user profile information, ensuring that the user is who they claim to be.

In summary, OAuth is primarily used to grant limited access to user data without exposing login credentials, making it ideal for authorizing actions like posting on social media or accessing personal data. In contrast, OIDC is perfect for scenarios where both authentication (verifying the user's identity) and authorization (granting permission to access resources) are needed. Together, they allow modern applications to securely manage access and provide a streamlined user experience by reducing the need for additional credentials.

See [scripts/azure/README.md](./scripts/azure/README.md) for full documentation on Azure scripts.

#### Vision Protocol (Local Image)

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.gemma.json \
  -protocol vision \
  -images ~/Pictures/wallpapers/monks-journey.jpg \
  -prompt "Provide a comprehensive description of this image" \
  -stream
```

<details>
  <summary>Configuration</summary>

  ```json
  {
    "name": "vision-agent",
    "client": {
      "timeout": "24s",
      "retry": {
        "max_retries": 3,
        "initial_backoff": "1s",
        "max_backoff": "30s",
        "backoff_multiplier": 2.0,
        "jitter": true
      },
      "connection_pool_size": 10,
      "connection_timeout": "9s"
    },
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434"
    },
    "model": {
      "name": "gemma3:4b",
      "capabilities": {
        "chat": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "top_p": 0.95
        },
        "vision": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "vision_options": {
            "detail": "auto"
          }
        }
      }
    }
  }
  ```

</details>

![monks-journey](https://w.wallhaven.cc/full/39/wallhaven-396dp9.jpg)

##### Output

Here's a detailed description of the image:

**Overall Impression:**

The image is a striking and surreal digital painting that evokes a sense of ancient mystery, serenity, and perhaps a touch of melancholy. It’s highly stylized with dramatic lighting and a focus on texture.

**Key Elements:**

* **Statue of Buddha:** The focal point is a massive, weathered statue of the Buddha. It's depicted in a traditional pose, with one arm raised in a gesture of blessing. The statue is covered in moss and lichen, giving it an aged, almost ethereal quality.  The details of the face are soft, suggesting time and the elements have worn it smooth.
* **Waterfall:** A powerful waterfall dominates the background. The water cascades down with impressive force, creating a dramatic spray and a sheet of light that illuminates the statue. The water is rendered with a sense of motion and volume.
* **Figure:** A small, solitary figure – a person dressed in a bright orange garment – stands on the statue’s hand. They appear tiny in comparison to the immense scale of the statue and the waterfall, emphasizing the theme of humility or contemplation.
* **Birds:** Several birds, rendered in white, are flying around the statue and the waterfall, adding a touch of life and movement to the scene.

**Color and Lighting:**

* **Dominant Colors:** The color palette is dominated by cool tones – greens, blues, and grays. This contributes to the sense of age, serenity, and perhaps a slight sadness.
* **Lighting:** The lighting is dramatic, with a strong light source coming from the waterfall, creating a bright, almost holy glow around the statue. This highlights the textures and adds a sense of depth and scale.

**Style and Mood:**

* **Digital Painting Style:** The image has a highly detailed, almost painterly digital painting style.  The use of texture and light gives it a realistic yet fantastical quality.
* **Mood:**  The overall mood is contemplative and slightly melancholic. It suggests themes of peace, time, and the impermanence of things. It feels like a place of quiet reflection and ancient wisdom.

Do you want me to focus on a specific aspect of the image, such as the symbolism or the artistic techniques used?

#### Vision Protocol (Web URL)

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.gemma.json \
  -protocol vision \
  -images https://ollama.com/public/ollama.png \
  -prompt "Provide a comprehensive description of this image" \
  -stream
```

<details>
  <summary>Configuration</summary>

  ```json
  {
    "name": "vision-agent",
    "client": {
      "timeout": "24s",
      "retry": {
        "max_retries": 3,
        "initial_backoff": "1s",
        "max_backoff": "30s",
        "backoff_multiplier": 2.0,
        "jitter": true
      },
      "connection_pool_size": 10,
      "connection_timeout": "9s"
    },
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434"
    },
    "model": {
      "name": "gemma3:4b",
      "capabilities": {
        "chat": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "top_p": 0.95
        },
        "vision": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "vision_options": {
            "detail": "auto"
          }
        }
      }
    }
  }
  ```

</details>

![ollama](https://ollama.com/public/ollama.png)

##### Output

Here's a comprehensive description of the image:

**Overall Impression:**

The image is a simple, cartoon-style illustration of a llama. It's rendered in black lines on a white background. The style is minimalistic and cute, with a focus on basic shapes.

**Specific Details:**

*   **Subject:** The image depicts a llama.
*   **Style:** The illustration is drawn in a flat, cartoon style. It doesn't have shading or detailed textures.
*   **Shape and Lines:** The llama’s body is indicated by a series of curved lines. It has large, upright ears, two large, circular eyes, and a small, rounded nose.
*   **Color:** The image is monochromatic – entirely black for the outlines and the eyes.
*   **Background:** The background is pure white.

**Overall Aesthetic:** The image has a friendly and approachable feel due to its simple design and cute character.

Do you want me to analyze any specific aspect of the image in more detail?

#### Tools Protocol (Weather)

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -protocol tools \
  -tools-file tools/prompt-agent/tools.json \
  -prompt "What's the weather like in Dallas, TX?"
```

<details>
  <summary>Configuration</summary>

  ```json
  {
    "name": "ollama-agent",
    "system_prompt": "You are an expert software architect specializing in cloud native systems design",
    "client": {
      "timeout": "24s",
      "retry": {
        "max_retries": 3,
        "initial_backoff": "1s",
        "max_backoff": "30s",
        "backoff_multiplier": 2.0,
        "jitter": true
      },
      "connection_pool_size": 10,
      "connection_timeout": "9s"
    },
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434"
    },
    "model": {
      "name": "llama3.2:3b",
      "capabilities": {
        "chat": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "top_p": 0.95
        },
        "tools": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "tool_choice": "auto"
        }
      }
    }
  }
  ```

</details>

##### Output

```
Tool Calls:
  - get_weather({"location":"Dallas, TX"})

Tokens: 224 prompt + 19 completion = 243 total
```

#### Tools Protocol (Calculate)

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -protocol tools \
  -tools-file tools/prompt-agent/tools.json \
  -prompt "Calculate 15 * 234 + 567"
```

<details>
  <summary>Configuration</summary>

  ```json
  {
    "name": "ollama-agent",
    "system_prompt": "You are an expert software architect specializing in cloud native systems design",
    "client": {
      "timeout": "24s",
      "retry": {
        "max_retries": 3,
        "initial_backoff": "1s",
        "max_backoff": "30s",
        "backoff_multiplier": 2.0,
        "jitter": true
      },
      "connection_pool_size": 10,
      "connection_timeout": "9s"
    },
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434"
    },
    "model": {
      "name": "llama3.2:3b",
      "capabilities": {
        "chat": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "top_p": 0.95
        },
        "tools": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "tool_choice": "auto"
        }
      }
    }
  }
  ```

</details>

##### Output

```
Tool Calls:
  - calculate({"expression":"15*234+567"})

Tokens: 221 prompt + 20 completion = 241 total
```

#### Tools Protocol (Multiple)

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -protocol tools \
  -tools-file tools/prompt-agent/tools.json \
  -prompt "Calculate the square root of pi, then get the weather in Dallas, TX"
```

<details>
  <summary>Configuration</summary>

  ```json
  {
    "name": "ollama-agent",
    "system_prompt": "You are an expert software architect specializing in cloud native systems design",
    "client": {
      "timeout": "24s",
      "retry": {
        "max_retries": 3,
        "initial_backoff": "1s",
        "max_backoff": "30s",
        "backoff_multiplier": 2.0,
        "jitter": true
      },
      "connection_pool_size": 10,
      "connection_timeout": "9s"
    },
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434"
    },
    "model": {
      "name": "llama3.2:3b",
      "capabilities": {
        "chat": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "top_p": 0.95
        },
        "tools": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "tool_choice": "auto"
        }
      }
    }
  }
  ```

</details>

##### Output

```
Tool Calls:
  - calculate({"expression":"sqrt(pi)"})
  - get_weather({"location":"Dallas, TX"})

Tokens: 229 prompt + 37 completion = 266 total
```

#### Embeddings Protocol

```sh
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.embedding.json \
  -protocol embeddings \
  -prompt "The quick brown fox jumps over the lazy dog"
```

<details>
  <summary>Configuration</summary>

  ```json
  {
    "name": "embeddings-agent",
    "client": {
      "timeout": "24s",
      "retry": {
        "max_retries": 3,
        "initial_backoff": "1s",
        "max_backoff": "30s",
        "backoff_multiplier": 2.0,
        "jitter": true
      },
      "connection_pool_size": 10,
      "connection_timeout": "6s"
    },
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434"
    },
    "model": {
      "name": "embeddinggemma:300m",
      "capabilities": {
        "embeddings": {
          "dimensions": 768
        }
      }
    }
  }
  ```

</details>

##### Output

```
Input: "The quick brown fox jumps over the lazy dog"

Generated 1 embedding(s):

Embedding [0]:
  Dimensions: 768
  Values: [-0.163660, 0.000575, 0.048880, -0.016126, -0.029346, ..., -0.009430, -0.012544, 0.006529, -0.025449, -0.004286]
  Statistics: min=-0.174384, max=0.257713, mean=0.000025
  Index: 0

Token Usage: 9 total
```

### Configuration

Agent configurations use flat JSON structure with `client`, `provider`, and `model` as peer fields. The library supports configuration option merging where model-configured options provide baseline values that can be overridden at runtime.

**Important**: Configuration options do not define defaults - they specify what options to send to the model. Each model has its own inherent defaults determined by the model implementation (e.g., GPT-4's default temperature is 1.0, Llama models may differ). The go-agents library passes configured options through to the model without implementing model-specific defaults.

#### Configuration Structure

**Required Fields:**
- `name` - Human-readable agent identifier
- `provider.name` - Provider platform name ("ollama", "azure")
- `provider.base_url` - Provider API endpoint base URL
- `model.name` - Model name on the provider platform (e.g., "llama3.2:3b" on Ollama, "gpt-4o" on Azure)
- `model.capabilities` - At least one protocol must be specified (can be empty object `{}` to use model defaults)

**Optional Fields:**
- `system_prompt` - System instructions injected into message arrays
- `provider.options` - Provider-specific configuration (e.g., Azure deployment name, API version, auth type)
- `client.timeout` - Overall request timeout including retries (default: "30s")
- `client.retry` - Retry configuration object:
  - `max_retries` - Maximum retry attempts (default: 3)
  - `initial_backoff` - Initial backoff duration (default: "1s")
  - `max_backoff` - Maximum backoff duration (default: "30s")
  - `backoff_multiplier` - Backoff multiplier for exponential backoff (default: 2.0)
  - `jitter` - Add randomization to backoff delays (default: true)
- `client.connection_pool_size` - HTTP connection pool size (default: 10)
- `client.connection_timeout` - Connection establishment timeout (default: "10s")

**Retry Behavior**: The client automatically retries transient failures (HTTP 429, 502, 503, 504, network errors, DNS errors) using exponential backoff with optional jitter. Backoff delay = `initial_backoff * (backoff_multiplier ^ attempt)`, capped at `max_backoff`. Jitter randomizes delays by ±25% to prevent thundering herd. Non-retryable errors (context cancellation, HTTP 4xx except 429) fail immediately.

#### Protocol Capabilities

The `capabilities` map configures protocol-specific options. At least one protocol must be specified for the agent to execute prompts. Each protocol can be configured with options or as an empty object `{}` to use the model's inherent defaults.

**Chat Protocol:**
```json
"chat": {
  "max_tokens": 4096,
  "temperature": 0.7,
  "top_p": 0.95
}
```

Common options: `max_tokens`, `temperature`, `top_p`, `frequency_penalty`, `presence_penalty`

Or use model defaults:
```json
"chat": {}
```

**Vision Protocol:**
```json
"vision": {
  "max_tokens": 4096,
  "temperature": 0.7,
  "vision_options": {
    "detail": "high"
  }
}
```

Common options: Same as chat, plus `vision_options` nested map for protocol-specific parameters like `detail` ("low", "high", "auto")

**Tools Protocol:**
```json
"tools": {
  "max_tokens": 4096,
  "temperature": 0.7,
  "tool_choice": "auto"
}
```

Common options: Same as chat, plus `tool_choice` ("auto", "required", specific tool name)

**Embeddings Protocol:**
```json
"embeddings": {
  "dimensions": 768
}
```

Common options: `dimensions` (output vector dimensions)

#### Option Merging Behavior

Agent methods merge configured options with runtime options:

1. **Baseline values**: Model's configured protocol options from configuration file
2. **Runtime override**: Options passed to agent method calls override matching keys
3. **Model name**: Automatically added to ensure correct routing

Example:
```go
// Configuration provides baseline values
"chat": {"temperature": 0.7, "max_tokens": 4096}

// Runtime call overrides temperature
agent.Chat(ctx, "prompt", map[string]any{"temperature": 0.9})

// Final options sent to model
{"temperature": 0.9, "max_tokens": 4096, "model": "llama3.2:3b"}
```

#### Complete Configuration Examples

**Multi-Protocol Agent (Ollama Platform, Llama Model):**
```json
{
  "name": "multi-protocol-agent",
  "system_prompt": "You are a helpful AI assistant",
  "client": {
    "timeout": "24s",
    "retry": {
      "max_retries": 3,
      "initial_backoff": "1s",
      "max_backoff": "30s",
      "backoff_multiplier": 2.0,
      "jitter": true
    },
    "connection_pool_size": 10,
    "connection_timeout": "9s"
  },
  "provider": {
    "name": "ollama",
    "base_url": "http://localhost:11434"
  },
  "model": {
    "name": "llama3.2:3b",
    "capabilities": {
      "chat": {
        "max_tokens": 4096,
        "temperature": 0.7,
        "top_p": 0.95
      },
      "vision": {
        "max_tokens": 4096,
        "temperature": 0.7,
        "vision_options": {
          "detail": "auto"
        }
      },
      "tools": {
        "max_tokens": 4096,
        "temperature": 0.7,
        "tool_choice": "auto"
      }
    }
  }
}
```

**Minimal Chat Agent (Using Model Defaults):**
```json
{
  "name": "minimal-chat-agent",
  "provider": {
    "name": "ollama",
    "base_url": "http://localhost:11434"
  },
  "model": {
    "name": "llama3.2:3b",
    "capabilities": {
      "chat": {}
    }
  }
}
```

**Embeddings Agent (Ollama Platform, Gemma Embedding Model):**
```json
{
  "name": "embeddings-agent",
  "client": {
    "timeout": "24s",
    "connection_timeout": "6s"
  },
  "provider": {
    "name": "ollama",
    "base_url": "http://localhost:11434"
  },
  "model": {
    "name": "embeddinggemma:300m",
    "capabilities": {
      "embeddings": {
        "dimensions": 768
      }
    }
  }
}
```

**Azure OpenAI Platform with Reasoning Model:**
```json
{
  "name": "azure-reasoning-agent",
  "system_prompt": "You are a thoughtful AI assistant that provides detailed analysis",
  "client": {
    "timeout": "24s"
  },
  "provider": {
    "name": "azure",
    "base_url": "https://go-agents-platform.openai.azure.com/openai",
    "options": {
      "deployment": "o3-mini",
      "api_version": "2025-01-01-preview",
      "auth_type": "api_key"
    }
  },
  "model": {
    "name": "o3-mini",
    "capabilities": {
      "chat": {
        "max_completion_tokens": 4096
      }
    }
  }
}
```

Note: Reasoning models (o1, o3-mini) use `max_completion_tokens` and don't support `temperature` or `top_p`.

## Development

### Running Tests

The library includes comprehensive unit tests organized in the `tests/` directory. All tests use black-box testing with the `package_test` suffix.

**Run all tests:**
```bash
go test ./tests/... -v
```

**Run tests for a specific package:**
```bash
go test ./tests/config/... -v
go test ./tests/protocol/... -v
go test ./tests/response/... -v
go test ./tests/providers/... -v
go test ./tests/client/... -v
go test ./tests/agent/... -v
go test ./tests/mock/... -v
```

**Generate coverage report:**
```bash
# Generate coverage for all packages
go test ./tests/... -coverprofile=coverage.out -coverpkg=./pkg/...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

**Current Coverage**: 54.8% overall (working towards 80% minimum requirement)

### Testing Your Code

The `pkg/mock` package provides mock implementations for testing code that depends on go-agents without requiring live LLM services.

**Mock Types Available**:
- `MockAgent` - Complete agent interface implementation
- `MockClient` - Transport client with configurable responses
- `MockProvider` - Provider with endpoint mapping
- `MockModel` - Model with protocol support configuration
- `MockCapability` - Capability with validation and processing

**Quick Example**:
```go
package mypackage_test

import (
    "context"
    "testing"

    "github.com/JaimeStill/go-agents/pkg/agent"
    "github.com/JaimeStill/go-agents/pkg/mock"
)

func TestMyOrchestrator(t *testing.T) {
    // Create mock agent with predetermined response
    mockAgent := mock.NewSimpleChatAgent(
        "test-agent",
        "Test response from mock agent",
    )

    // Use mock in your code
    orchestrator := NewOrchestrator(mockAgent)
    result, err := orchestrator.Process(context.Background(), "test input")

    // Assert behavior
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    // ... additional assertions
}
```

**Helper Constructors for Common Scenarios**:
- `NewSimpleChatAgent(id, response)` - Basic chat responses
- `NewStreamingChatAgent(id, chunks)` - Streaming chat
- `NewToolsAgent(id, toolCalls)` - Tool calling
- `NewEmbeddingsAgent(id, embedding)` - Embeddings generation
- `NewMultiProtocolAgent(id)` - Multi-protocol support
- `NewFailingAgent(id, err)` - Error handling testing

See `pkg/mock` package documentation for complete API details.

### Viewing Documentation

All packages include comprehensive godoc documentation.

**View package documentation:**
```bash
# View main package overview
go doc github.com/JaimeStill/go-agents/pkg/agent

# View specific type documentation
go doc github.com/JaimeStill/go-agents/pkg/agent.Agent
go doc github.com/JaimeStill/go-agents/pkg/protocol.Protocol
go doc github.com/JaimeStill/go-agents/pkg/client.Client
```

**View all available packages:**
```bash
go doc github.com/JaimeStill/go-agents/pkg/config
go doc github.com/JaimeStill/go-agents/pkg/protocol
go doc github.com/JaimeStill/go-agents/pkg/response
go doc github.com/JaimeStill/go-agents/pkg/model
go doc github.com/JaimeStill/go-agents/pkg/request
go doc github.com/JaimeStill/go-agents/pkg/providers
go doc github.com/JaimeStill/go-agents/pkg/client
go doc github.com/JaimeStill/go-agents/pkg/agent
go doc github.com/JaimeStill/go-agents/pkg/mock
```

**Start local documentation server:**
```bash
# Install godoc if not already installed
go install golang.org/x/tools/cmd/godoc@latest

# Start documentation server
godoc -http=:6060

# Visit: http://localhost:6060/pkg/github.com/JaimeStill/go-agents/
```

### Testing Strategy

For detailed information on the testing approach, patterns, and coverage requirements, see the **Testing Strategy** section in [ARCHITECTURE.md](./ARCHITECTURE.md#testing-strategy).

**Key Points**:
- Tests organized in separate `tests/` directory
- Black-box testing using `package_test` suffix
- Table-driven test patterns
- HTTP mocking with `httptest.Server`
- 80% minimum coverage requirement (54.8% achieved, working towards goal)
