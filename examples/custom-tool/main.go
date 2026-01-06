package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	sdk "github.com/zero-day-ai/sdk"
	"github.com/zero-day-ai/sdk/schema"
)

func main() {
	// Create a custom HTTP request tool using SDK options
	// This tool demonstrates how to build reusable, schema-validated components
	httpTool, err := sdk.NewTool(
		// Basic tool metadata
		sdk.WithToolName("http-request"),
		sdk.WithToolVersion("1.0.0"),
		sdk.WithToolDescription("Make HTTP requests to test endpoints and analyze responses"),

		// Tags help with tool discovery and categorization
		sdk.WithToolTags("http", "network", "testing"),

		// Define the input schema using the schema builder
		// This validates inputs and generates documentation
		sdk.WithInputSchema(schema.Object(map[string]schema.JSON{
			"url": schema.StringWithDesc("Target URL to request"),
			"method": schema.JSON{
				Type:        "string",
				Description: "HTTP method to use",
				Enum:        []any{"GET", "POST", "PUT", "DELETE", "PATCH"},
				Default:     "GET",
			},
			"headers": schema.JSON{
				Type:        "object",
				Description: "Optional HTTP headers to include",
				Properties: map[string]schema.JSON{
					"*": schema.String(),
				},
			},
			"body": schema.StringWithDesc("Optional request body (for POST, PUT, PATCH)"),
			"timeout": schema.JSON{
				Type:        "integer",
				Description: "Request timeout in seconds",
				Default:     30,
				Minimum:     ptrFloat64(1),
				Maximum:     ptrFloat64(300),
			},
		}, "url", "method")), // url and method are required fields

		// Define the output schema
		// This documents what the tool returns
		sdk.WithOutputSchema(schema.Object(map[string]schema.JSON{
			"status": schema.JSON{
				Type:        "integer",
				Description: "HTTP status code",
			},
			"headers": schema.JSON{
				Type:        "object",
				Description: "Response headers",
			},
			"body": schema.StringWithDesc("Response body"),
			"duration_ms": schema.JSON{
				Type:        "integer",
				Description: "Request duration in milliseconds",
			},
		})),

		// Set the execution handler - this is the core tool logic
		sdk.WithExecuteHandler(executeHTTP),
	)
	if err != nil {
		log.Fatalf("Failed to create tool: %v", err)
	}

	fmt.Printf("Tool created successfully!\n")
	fmt.Printf("  Name: %s\n", httpTool.Name())
	fmt.Printf("  Version: %s\n", httpTool.Version())
	fmt.Printf("  Description: %s\n", httpTool.Description())
	fmt.Printf("  Tags: %v\n", httpTool.Tags())

	// Demonstrate tool usage with a test execution
	fmt.Println("\n--- Test Execution 1: GET Request ---")
	result1, err := httpTool.Execute(context.Background(), map[string]any{
		"url":    "https://httpbin.org/get",
		"method": "GET",
	})
	if err != nil {
		log.Printf("Execution failed: %v", err)
	} else {
		fmt.Printf("Status: %v\n", result1["status"])
		fmt.Printf("Duration: %vms\n", result1["duration_ms"])
		fmt.Printf("Body (truncated): %.200s...\n", result1["body"])
	}

	// Example with headers and timeout
	fmt.Println("\n--- Test Execution 2: GET with Custom Headers ---")
	result2, err := httpTool.Execute(context.Background(), map[string]any{
		"url":    "https://httpbin.org/headers",
		"method": "GET",
		"headers": map[string]any{
			"User-Agent":      "Gibson-SDK/1.0",
			"X-Custom-Header": "test-value",
		},
		"timeout": 10,
	})
	if err != nil {
		log.Printf("Execution failed: %v", err)
	} else {
		fmt.Printf("Status: %v\n", result2["status"])
		fmt.Printf("Duration: %vms\n", result2["duration_ms"])
	}

	// Example with POST request and body
	fmt.Println("\n--- Test Execution 3: POST with Body ---")
	result3, err := httpTool.Execute(context.Background(), map[string]any{
		"url":    "https://httpbin.org/post",
		"method": "POST",
		"headers": map[string]any{
			"Content-Type": "application/json",
		},
		"body": `{"test": "data", "from": "Gibson SDK"}`,
	})
	if err != nil {
		log.Printf("Execution failed: %v", err)
	} else {
		fmt.Printf("Status: %v\n", result3["status"])
		fmt.Printf("Duration: %vms\n", result3["duration_ms"])
	}

	// Check tool health
	fmt.Println("\n--- Health Check ---")
	health := httpTool.Health(context.Background())
	fmt.Printf("Status: %s\n", health.Status)
	if health.Message != "" {
		fmt.Printf("Message: %s\n", health.Message)
	}

	// Optionally serve the tool as a gRPC service
	// This allows agents to invoke the tool remotely
	//
	// Note: ServeTool is not yet implemented in the SDK
	// Uncomment the following lines when serving is available:
	//
	// fmt.Println("\nStarting tool server on port 50052...")
	// if err := sdk.ServeTool(httpTool, sdk.WithPort(50052)); err != nil {
	//     log.Fatalf("Failed to serve tool: %v", err)
	// }
}

// executeHTTP is the core tool implementation that makes HTTP requests.
// It receives validated input according to the InputSchema and returns
// output conforming to the OutputSchema.
func executeHTTP(ctx context.Context, input map[string]any) (map[string]any, error) {
	// Extract and validate inputs
	// The SDK has already validated against the schema, but we still need type assertions
	url, _ := input["url"].(string)
	method, _ := input["method"].(string)
	if method == "" {
		method = "GET"
	}

	// Extract optional parameters with defaults
	timeout := 30
	if t, ok := input["timeout"].(float64); ok {
		timeout = int(t)
	} else if t, ok := input["timeout"].(int); ok {
		timeout = t
	}

	// Extract optional body
	var body io.Reader
	if bodyStr, ok := input["body"].(string); ok && bodyStr != "" {
		body = strings.NewReader(bodyStr)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add custom headers if provided
	if headers, ok := input["headers"].(map[string]any); ok {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}
	}

	// Execute request and measure duration
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer sdk.CloseWithLog(resp.Body, nil, "HTTP response body")

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert response headers to map
	responseHeaders := make(map[string]any)
	for key, values := range resp.Header {
		if len(values) == 1 {
			responseHeaders[key] = values[0]
		} else {
			responseHeaders[key] = values
		}
	}

	// Return structured output matching OutputSchema
	return map[string]any{
		"status":      resp.StatusCode,
		"headers":     responseHeaders,
		"body":        string(bodyBytes),
		"duration_ms": duration.Milliseconds(),
	}, nil
}

// ptrFloat64 is a helper to create a pointer to a float64 value
// This is needed for the schema Min/Max fields which are pointers
func ptrFloat64(f float64) *float64 {
	return &f
}
