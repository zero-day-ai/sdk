package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/schema"
)

func main() {
	// Create a custom data analyzer plugin using the plugin builder
	// Plugins provide multiple related methods that can be invoked
	// This example demonstrates text analysis capabilities
	cfg := plugin.NewConfig()

	// Set basic plugin metadata
	cfg.SetName("data-analyzer")
	cfg.SetVersion("1.0.0")
	cfg.SetDescription("Analyze data patterns and extract insights from text")

	// Add method 1: Analyze text for patterns
	cfg.AddMethodWithDesc(
		"analyze",
		"Analyze text to find patterns, keywords, and structure",
		analyzeHandler,
		// Input schema: expects a "data" field containing text
		schema.Object(map[string]schema.JSON{
			"data": schema.StringWithDesc("Text data to analyze"),
			"options": schema.Object(map[string]schema.JSON{
				"case_sensitive": schema.JSON{
					Type:        "boolean",
					Description: "Whether pattern matching should be case sensitive",
					Default:     false,
				},
				"max_patterns": schema.JSON{
					Type:        "integer",
					Description: "Maximum number of patterns to return",
					Default:     10,
				},
			}),
		}, "data"),
		// Output schema: returns patterns and statistics
		schema.Object(map[string]schema.JSON{
			"patterns": schema.Array(schema.String()),
			"word_count": schema.JSON{
				Type:        "integer",
				Description: "Total number of words",
			},
			"unique_words": schema.JSON{
				Type:        "integer",
				Description: "Number of unique words",
			},
			"avg_word_length": schema.JSON{
				Type:        "number",
				Description: "Average word length",
			},
		}),
	)

	// Add method 2: Summarize text
	cfg.AddMethodWithDesc(
		"summarize",
		"Generate a concise summary of the input text",
		summarizeHandler,
		// Input schema
		schema.Object(map[string]schema.JSON{
			"text": schema.StringWithDesc("Text to summarize"),
			"max_length": schema.JSON{
				Type:        "integer",
				Description: "Maximum summary length in characters",
				Default:     200,
			},
		}, "text"),
		// Output schema
		schema.Object(map[string]schema.JSON{
			"summary": schema.StringWithDesc("Generated summary"),
			"compression_ratio": schema.JSON{
				Type:        "number",
				Description: "Ratio of summary length to original length",
			},
		}),
	)

	// Add method 3: Extract entities
	cfg.AddMethodWithDesc(
		"extract_entities",
		"Extract named entities like emails, URLs, and numbers from text",
		extractEntitiesHandler,
		// Input schema
		schema.Object(map[string]schema.JSON{
			"text": schema.StringWithDesc("Text to extract entities from"),
		}, "text"),
		// Output schema
		schema.Object(map[string]schema.JSON{
			"emails": schema.Array(schema.String()),
			"urls":   schema.Array(schema.String()),
			"numbers": schema.Array(schema.String()),
		}),
	)

	// Set optional initialization function
	// This is called once when the plugin is loaded
	cfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
		fmt.Println("Plugin initializing...")
		// In a real plugin, you might:
		// - Load models or resources
		// - Connect to databases
		// - Validate configuration
		return nil
	})

	// Set optional shutdown function
	// This is called when the plugin is being unloaded
	cfg.SetShutdownFunc(func(ctx context.Context) error {
		fmt.Println("Plugin shutting down...")
		// In a real plugin, you might:
		// - Close connections
		// - Save state
		// - Release resources
		return nil
	})

	// Build the plugin
	dataPlugin, err := plugin.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create plugin: %v", err)
	}

	fmt.Printf("Plugin created successfully!\n")
	fmt.Printf("  Name: %s\n", dataPlugin.Name())
	fmt.Printf("  Version: %s\n", dataPlugin.Version())
	fmt.Printf("  Description: %s\n", dataPlugin.Description())
	fmt.Printf("  Methods: %d\n", len(dataPlugin.Methods()))

	// List available methods
	fmt.Println("\nAvailable methods:")
	for _, method := range dataPlugin.Methods() {
		fmt.Printf("  - %s: %s\n", method.Name, method.Description)
	}

	// Initialize the plugin
	fmt.Println("\n--- Initializing Plugin ---")
	if err := dataPlugin.Initialize(context.Background(), nil); err != nil {
		log.Fatalf("Failed to initialize plugin: %v", err)
	}

	// Test method 1: Analyze
	fmt.Println("\n--- Test 1: Analyze Text ---")
	analyzeResult, err := dataPlugin.Query(context.Background(), "analyze", map[string]any{
		"data": "The quick brown fox jumps over the lazy dog. The dog was very lazy indeed.",
		"options": map[string]any{
			"case_sensitive": false,
			"max_patterns":   5,
		},
	})
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		result := analyzeResult.(map[string]any)
		fmt.Printf("Word count: %v\n", result["word_count"])
		fmt.Printf("Unique words: %v\n", result["unique_words"])
		fmt.Printf("Average word length: %.2f\n", result["avg_word_length"])
		fmt.Printf("Patterns: %v\n", result["patterns"])
	}

	// Test method 2: Summarize
	fmt.Println("\n--- Test 2: Summarize Text ---")
	summarizeResult, err := dataPlugin.Query(context.Background(), "summarize", map[string]any{
		"text": "The Gibson Framework is a security testing platform for AI systems. " +
			"It provides tools and agents to discover vulnerabilities in LLM applications. " +
			"The framework supports multiple testing techniques including prompt injection, " +
			"jailbreaking, and data exfiltration detection.",
		"max_length": 100,
	})
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		result := summarizeResult.(map[string]any)
		fmt.Printf("Summary: %s\n", result["summary"])
		fmt.Printf("Compression ratio: %.2f\n", result["compression_ratio"])
	}

	// Test method 3: Extract entities
	fmt.Println("\n--- Test 3: Extract Entities ---")
	extractResult, err := dataPlugin.Query(context.Background(), "extract_entities", map[string]any{
		"text": "Contact us at support@example.com or visit https://github.com/zero-day-ai. " +
			"Our phone number is 555-1234 and the issue number is #42.",
	})
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		result := extractResult.(map[string]any)
		fmt.Printf("Emails: %v\n", result["emails"])
		fmt.Printf("URLs: %v\n", result["urls"])
		fmt.Printf("Numbers: %v\n", result["numbers"])
	}

	// Check health status
	fmt.Println("\n--- Health Check ---")
	health := dataPlugin.Health(context.Background())
	fmt.Printf("Status: %s\n", health.Status)
	fmt.Printf("Message: %s\n", health.Message)

	// Shutdown the plugin
	fmt.Println("\n--- Shutting Down Plugin ---")
	if err := dataPlugin.Shutdown(context.Background()); err != nil {
		log.Printf("Shutdown failed: %v", err)
	}

	// Optionally serve the plugin as a gRPC service
	// This allows remote invocation of plugin methods
	//
	// Note: ServePlugin is not yet implemented in the SDK
	// Uncomment the following lines when serving is available:
	//
	// fmt.Println("\nStarting plugin server on port 50053...")
	// if err := sdk.ServePlugin(dataPlugin, sdk.WithPort(50053)); err != nil {
	//     log.Fatalf("Failed to serve plugin: %v", err)
	// }
}

// analyzeHandler implements text analysis functionality
func analyzeHandler(ctx context.Context, params map[string]any) (any, error) {
	// Extract input parameters
	data, _ := params["data"].(string)

	// Extract options with defaults
	caseSensitive := false
	maxPatterns := 10
	if options, ok := params["options"].(map[string]any); ok {
		if cs, ok := options["case_sensitive"].(bool); ok {
			caseSensitive = cs
		}
		if mp, ok := options["max_patterns"].(float64); ok {
			maxPatterns = int(mp)
		} else if mp, ok := options["max_patterns"].(int); ok {
			maxPatterns = mp
		}
	}

	// Convert to lowercase if not case sensitive
	text := data
	if !caseSensitive {
		text = strings.ToLower(text)
	}

	// Split into words
	words := strings.Fields(text)
	wordCount := len(words)

	// Count unique words and calculate average length
	uniqueWords := make(map[string]bool)
	totalLength := 0
	for _, word := range words {
		uniqueWords[word] = true
		totalLength += len(word)
	}

	avgWordLength := 0.0
	if wordCount > 0 {
		avgWordLength = float64(totalLength) / float64(wordCount)
	}

	// Find repeated patterns (words that appear more than once)
	wordFreq := make(map[string]int)
	for _, word := range words {
		wordFreq[word]++
	}

	patterns := []string{}
	for word, count := range wordFreq {
		if count > 1 && len(patterns) < maxPatterns {
			patterns = append(patterns, fmt.Sprintf("%s (x%d)", word, count))
		}
	}

	return map[string]any{
		"patterns":        patterns,
		"word_count":      wordCount,
		"unique_words":    len(uniqueWords),
		"avg_word_length": avgWordLength,
	}, nil
}

// summarizeHandler implements text summarization
func summarizeHandler(ctx context.Context, params map[string]any) (any, error) {
	// Extract input parameters
	text, _ := params["text"].(string)
	maxLength := 200
	if ml, ok := params["max_length"].(float64); ok {
		maxLength = int(ml)
	} else if ml, ok := params["max_length"].(int); ok {
		maxLength = ml
	}

	// Simple summarization: take first sentence or truncate to max length
	sentences := strings.Split(text, ".")
	summary := text

	if len(sentences) > 0 && len(sentences[0]) > 0 {
		summary = strings.TrimSpace(sentences[0]) + "."
	}

	// Truncate if still too long
	if len(summary) > maxLength {
		summary = summary[:maxLength-3] + "..."
	}

	compressionRatio := float64(len(summary)) / float64(len(text))

	return map[string]any{
		"summary":           summary,
		"compression_ratio": compressionRatio,
	}, nil
}

// extractEntitiesHandler extracts entities like emails, URLs, and numbers
func extractEntitiesHandler(ctx context.Context, params map[string]any) (any, error) {
	// Extract input
	text, _ := params["text"].(string)

	// Regular expressions for entity extraction
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	numberRegex := regexp.MustCompile(`\b\d+\b`)

	// Extract entities
	emails := emailRegex.FindAllString(text, -1)
	urls := urlRegex.FindAllString(text, -1)
	numbers := numberRegex.FindAllString(text, -1)

	// Ensure empty arrays instead of nil
	if emails == nil {
		emails = []string{}
	}
	if urls == nil {
		urls = []string{}
	}
	if numbers == nil {
		numbers = []string{}
	}

	return map[string]any{
		"emails":  emails,
		"urls":    urls,
		"numbers": numbers,
	}, nil
}
