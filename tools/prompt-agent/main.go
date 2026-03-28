package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/format"
)

func main() {
	var (
		configFile   = flag.String("config", "config.json", "Configuration file to use")
		protocol     = flag.String("protocol", "chat", "Protocol to use (chat, vision, tools, embeddings)")
		prompt       = flag.String("prompt", "", "Prompt to send to the agent")
		systemPrompt = flag.String("system-prompt", "", "System prompt (overrides config)")
		token        = flag.String("token", "", "Authentication token (overrides config)")
		stream       = flag.Bool("stream", false, "Enable streaming responses")

		images    = flag.String("images", "", "Comma-separated image URLs/paths (for vision)")
		toolsFile = flag.String("tools-file", "", "JSON file containing tool definitions (for tools)")
	)
	flag.Parse()

	if *prompt == "" {
		log.Fatal("Error: -prompt flag is required")
	}

	cfg, err := config.LoadAgentConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *token != "" {
		if cfg.Provider.Options == nil {
			cfg.Provider.Options = make(map[string]any)
		}
		cfg.Provider.Options["token"] = *token
	}

	if *systemPrompt != "" {
		cfg.SystemPrompt = *systemPrompt
	}

	a, err := agent.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Client.Timeout.ToDuration())
	defer cancel()

	switch *protocol {
	case "chat":
		if *stream {
			executeChatStream(ctx, a, *prompt)
		} else {
			executeChat(ctx, a, *prompt)
		}
	case "vision":
		if *images == "" {
			log.Fatal("Error: -images flag is required for vision protocol")
		}
		imageList := strings.Split(*images, ",")
		for i, img := range imageList {
			imageList[i] = strings.TrimSpace(img)
		}
		preparedImages := prepareImages(imageList)
		if *stream {
			executeVisionStream(ctx, a, *prompt, preparedImages)
		} else {
			executeVision(ctx, a, *prompt, preparedImages)
		}
	case "tools":
		if *toolsFile == "" {
			log.Fatal("Error: -tools-file flag is required for tools protocol")
		}
		toolList := loadTools(*toolsFile)
		executeTools(ctx, a, *prompt, toolList)
	case "embeddings":
		executeEmbeddings(ctx, a, *prompt)
	default:
		log.Fatalf("Unknown protocol: %s", *protocol)
	}
}

func executeChat(
	ctx context.Context,
	agent agent.Agent,
	prompt string,
) {
	response, err := agent.Chat(ctx, prompt)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}
	fmt.Printf("Response: %s\n", response.Text())
	if response.Usage != nil {
		fmt.Printf(
			"Tokens: %d prompt + %d completions = %d total",
			response.Usage.InputTokens,
			response.Usage.OutputTokens,
			response.Usage.TotalTokens,
		)
	}
}

func executeChatStream(
	ctx context.Context,
	agent agent.Agent,
	prompt string,
) {
	stream, err := agent.ChatStream(ctx, prompt)
	if err != nil {
		log.Fatalf("ChatStream failed: %v", err)
	}

	for chunk := range stream {
		if chunk.Error != nil {
			log.Fatalf("Stream error: %v", chunk.Error)
		}
		fmt.Print(chunk.Text())
	}
	fmt.Println()
}

func executeVision(
	ctx context.Context,
	agent agent.Agent,
	prompt string,
	images []format.Image,
) {
	response, err := agent.Vision(ctx, prompt, images)
	if err != nil {
		log.Fatalf("Vision failed: %v", err)
	}
	fmt.Printf("Vision response: %s\n", response.Text())
	if response.Usage != nil {
		fmt.Printf(
			"Tokens: %d prompt + %d completion = %d total\n",
			response.Usage.InputTokens,
			response.Usage.OutputTokens,
			response.Usage.TotalTokens,
		)
	}
}

func executeVisionStream(
	ctx context.Context,
	agent agent.Agent,
	prompt string,
	images []format.Image,
) {
	stream, err := agent.VisionStream(ctx, prompt, images)
	if err != nil {
		log.Fatalf("VisionStream failed: %v", err)
	}

	for chunk := range stream {
		if chunk.Error != nil {
			log.Fatalf("Stream error: %v", chunk.Error)
		}

		fmt.Print(chunk.Text())
	}

	fmt.Println()
}

func executeTools(
	ctx context.Context,
	agent agent.Agent,
	prompt string,
	tools []agent.Tool,
) {
	response, err := agent.Tools(ctx, prompt, tools)
	if err != nil {
		log.Fatalf("Tools failed: %v", err)
	}

	if text := response.Text(); text != "" {
		fmt.Printf("Response: %s\n", text)
	}

	toolCalls := response.ToolCalls()
	if len(toolCalls) > 0 {
		fmt.Printf("\nTool Calls:\n")
		for _, tc := range toolCalls {
			argsJSON, _ := json.Marshal(tc.Input)
			fmt.Printf("  - %s(%s)\n", tc.Name, string(argsJSON))
		}
	}

	if response.Usage != nil {
		fmt.Printf("\nTokens: %d prompt + %d completion = %d total\n",
			response.Usage.InputTokens,
			response.Usage.OutputTokens,
			response.Usage.TotalTokens,
		)
	}
}

func executeEmbeddings(
	ctx context.Context,
	agent agent.Agent,
	input string,
) {
	response, err := agent.Embed(ctx, input)
	if err != nil {
		log.Fatalf("Embeddings failed: %v", err)
	}

	fmt.Printf("Input: %q\n\n", input)
	fmt.Printf("Generated %d embedding(s):\n\n", len(response.Embeddings))

	for i, embedding := range response.Embeddings {
		fmt.Printf("Embedding [%d]:\n", i)
		fmt.Printf("  Dimensions: %d\n", len(embedding))

		if len(embedding) > 0 {
			previewCount := 5
			if len(embedding) <= previewCount*2 {
				fmt.Printf("  Values: [")
				for j, val := range embedding {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%.6f", val)
				}
				fmt.Printf("]\n")
			} else {
				fmt.Printf("  Values: [")
				for j := range previewCount {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%.6f", embedding[j])
				}
				fmt.Printf(", ..., ")
				start := len(embedding) - previewCount
				for j := start; j < len(embedding); j++ {
					if j > start {
						fmt.Printf(", ")
					}
					fmt.Printf("%.6f", embedding[j])
				}
				fmt.Printf("]\n")
			}

			var sum, min, max float64
			min = embedding[0]
			max = embedding[0]

			for _, val := range embedding {
				sum += val
				if val < min {
					min = val
				}
				if val > max {
					max = val
				}
			}

			mean := sum / float64(len(embedding))
			fmt.Printf("  Statistics: min=%.6f, max=%.6f, mean=%.6f\n", min, max, mean)
		}

		fmt.Println()
	}

	if response.Usage != nil {
		fmt.Printf("Token Usage: %d total\n", response.Usage.TotalTokens)
	}
}

func loadTools(filename string) []agent.Tool {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read tools file: %v", err)
	}

	var tools []agent.Tool
	if err := json.Unmarshal(data, &tools); err != nil {
		log.Fatalf("Failed to parse tools file: %v", err)
	}

	return tools
}

func prepareImages(imageList []string) []format.Image {
	prepared := make([]format.Image, len(imageList))
	for i, img := range imageList {
		if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
			// Download and encode remote images (some providers only support base64)
			data, err := downloadImage(img)
			if err != nil {
				log.Fatalf("Failed to download image %s: %v", img, err)
			}

			// Detect MIME type from downloaded content
			mimeType := http.DetectContentType(data)

			// Validate it's an image
			if !strings.HasPrefix(mimeType, "image/") {
				log.Fatalf("URL %s is not an image (detected type: %s)", img, mimeType)
			}

			prepared[i] = format.Image{
				Data:   data,
				Format: strings.TrimPrefix(mimeType, "image/"),
			}
		} else {
			// Expand home directory if needed
			if strings.HasPrefix(img, "~/") {
				home, err := os.UserHomeDir()
				if err != nil {
					log.Fatalf("Failed to get home directory: %v", err)
				}
				img = strings.Replace(img, "~", home, 1)
			}

			// Local file, read and encode
			data, err := os.ReadFile(img)
			if err != nil {
				log.Fatalf("Failed to read image %s: %v", img, err)
			}

			// Detect MIME type from content
			mimeType := http.DetectContentType(data)

			// Validate it's an image
			if !strings.HasPrefix(mimeType, "image/") {
				log.Fatalf("File %s is not an image (detected type: %s)", img, mimeType)
			}

			prepared[i] = format.Image{
				Data:   data,
				Format: strings.TrimPrefix(mimeType, "image/"),
			}
		}
	}
	return prepared
}

func downloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}
