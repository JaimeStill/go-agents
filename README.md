# Go Agents

A platform and model agnostic Go agent primitive library.

## Status: Pre-Release (v0.2.0)

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
- **Configuration Pass-Through**: Simple option pass-through without runtime merging for predictable behavior
- **Structured Content Support**: Vision protocol handles multimodal content with image options, tools protocol returns structured tool calls
- **Enhanced Development Tools**: Command-line testing infrastructure with comprehensive protocol examples
- **Human-Readable Configuration**: Duration strings ("24s", "1m") and clean JSON configuration
- **Thread-Safe Operations**: Proper connection pooling, streaming support (chat, vision, tools), and concurrent request handling
- **Mock Implementations**: Complete mock package for testing agent-based systems

## Development Status

### Current Phase: v0.2.0 - Protocol-Centric Architecture

**Completed**: Major architectural refactor to protocol-specific request types. Separates protocol input data (images, tools, input text) from model configuration options (temperature, max_tokens). All protocols validated with Ollama and Azure.

**Active Focus**: The system is stable and ready for use. Architecture simplified with clear boundaries between protocol data and configuration.

**Architecture Highlights**: Protocol-specific request types implementing ProtocolRequest interface, OpenAI format as default standard, option pass-through without runtime merging, vision and tools protocols properly handling structured data.

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
      "provider": {
        "name": "ollama",
        "base_url": "http://localhost:11434",
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
      },
      "timeout": "24s",
      "max_retries": 3,
      "retry_backoff_base": "1s",
      "connection_pool_size": 10,
      "connection_timeout": "9s"
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
      "provider": {
        "name": "azure",
        "base_url": "https://go-agents-platform.openai.azure.com/openai",
        "model": {
          "name": "o3-mini",
          "capabilities": {
            "chat": {
              "max_completion_tokens": 4096
            }
          }
        },
        "options": {
          "deployment": "o3-mini",
          "api_version": "2025-01-01-preview",
          "auth_type": "api_key"
        }
      },
      "timeout": "24s",
      "max_retries": 3,
      "retry_backoff_base": "1s",
      "connection_pool_size": 10,
      "connection_timeout": "9s"
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
    "name": "azure-key-agent",
    "system_prompt": "You are an expert software architect specializing in cloud native systems design",
    "client": {
      "provider": {
        "name": "azure",
        "base_url": "https://go-agents-platform.openai.azure.com/openai",
        "model": {
          "name": "o3-mini",
          "capabilities": {
            "chat": {
              "max_completion_tokens": 4096
            }
          }
        },
        "options": {
          "deployment": "o3-mini",
          "api_version": "2025-01-01-preview",
          "auth_type": "bearer"
        }
      },
      "timeout": "24s",
      "max_retries": 3,
      "retry_backoff_base": "1s",
      "connection_pool_size": 10,
      "connection_timeout": "9s"
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
      "provider": {
        "name": "ollama",
        "base_url": "http://localhost:11434",
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
              "detail": "auto"
            }
          }
        }
      },
      "timeout": "24s",
      "max_retries": 3,
      "retry_backoff_base": "1s",
      "connection_pool_size": 10,
      "connection_timeout": "9s"
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
      "provider": {
        "name": "ollama",
        "base_url": "http://localhost:11434",
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
              "detail": "auto"
            }
          }
        }
      },
      "timeout": "24s",
      "max_retries": 3,
      "retry_backoff_base": "1s",
      "connection_pool_size": 10,
      "connection_timeout": "9s"
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
      "provider": {
        "name": "ollama",
        "base_url": "http://localhost:11434",
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
      },
      "timeout": "24s",
      "max_retries": 3,
      "retry_backoff_base": "1s",
      "connection_pool_size": 10,
      "connection_timeout": "9s"
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
      "provider": {
        "name": "ollama",
        "base_url": "http://localhost:11434",
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
      },
      "timeout": "24s",
      "max_retries": 3,
      "retry_backoff_base": "1s",
      "connection_pool_size": 10,
      "connection_timeout": "9s"
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
      "provider": {
        "name": "ollama",
        "base_url": "http://localhost:11434",
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
      },
      "timeout": "24s",
      "max_retries": 3,
      "retry_backoff_base": "1s",
      "connection_pool_size": 10,
      "connection_timeout": "9s"
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
      "provider": {
        "name": "ollama",
        "base_url": "http://localhost:11434",
        "model": {
          "name": "embeddinggemma:300m",
          "capabilities": {
            "embeddings": {
              "dimensions": 768
            }
          }
        }
      },
      "timeout": "24s",
      "max_retries": 3,
      "retry_backoff_base": "1s",
      "connection_pool_size": 10,
      "connection_timeout": "6s"
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

Agent configurations use hierarchical JSON with client-based structure:

```json
{
  "name": "research-assistant",
  "system_prompt": "You are a helpful research assistant focused on providing accurate and comprehensive information",
  "client": {
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434",
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
    },
    "timeout": "24s",
    "max_retries": 3,
    "retry_backoff_base": "1s",
    "connection_pool_size": 10,
    "connection_timeout": "9s"
  }
}
```

For Azure AI Foundry with reasoning models:

```json
{
  "name": "azure-assistant",
  "system_prompt": "You are a thoughtful AI assistant that provides detailed analysis and reasoning",
  "client": {
    "provider": {
      "name": "azure",
      "base_url": "https://go-agents-platform.openai.azure.com/openai",
      "model": {
        "name": "o3-mini",
        "capabilities": {
          "chat": {
            "max_completion_tokens": 4096
          }
        }
      },
      "options": {
        "deployment": "o3-mini",
        "api_version": "2025-01-01-preview",
        "auth_type": "api_key"
      }
    },
    "timeout": "24s",
    "max_retries": 3,
    "retry_backoff_base": "1s",
    "connection_pool_size": 10,
    "connection_timeout": "9s"
  }
}
```

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
go test ./tests/types/... -v
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
go doc github.com/JaimeStill/go-agents/pkg/types.Protocol
go doc github.com/JaimeStill/go-agents/pkg/client.Client
```

**View all available packages:**
```bash
go doc github.com/JaimeStill/go-agents/pkg/config
go doc github.com/JaimeStill/go-agents/pkg/types
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
