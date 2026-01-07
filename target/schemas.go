package target

import (
	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/types"
)

// Built-in target schemas for common target types.
// These schemas define the connection parameters required for different attack surfaces.

var (
	// HTTPAPISchema defines connection parameters for HTTP API targets.
	// This is the most common target type for LLM APIs and web services.
	HTTPAPISchema = types.TargetSchema{
		Type:        "http_api",
		Version:     "1.0",
		Description: "HTTP API endpoint for testing web services and LLM APIs",
		Schema: schema.Object(map[string]schema.JSON{
			"url": schema.JSON{
				Type:        "string",
				Description: "Target endpoint URL",
				Format:      "uri",
			},
			"method": schema.JSON{
				Type:        "string",
				Description: "HTTP method to use",
				Enum:        []any{"GET", "POST", "PUT", "DELETE", "PATCH"},
				Default:     "POST",
			},
			"headers": schema.JSON{
				Type:        "object",
				Description: "HTTP headers to include in requests",
			},
			"timeout": schema.JSON{
				Type:        "integer",
				Description: "Request timeout in seconds",
				Minimum:     intPtr(1),
				Default:     30,
			},
		}, "url"),
	}

	// LLMChatSchema defines connection parameters for conversational LLM interfaces.
	// Used for testing ChatGPT-style chat interfaces.
	LLMChatSchema = types.TargetSchema{
		Type:        "llm_chat",
		Version:     "1.0",
		Description: "Conversational LLM interface (e.g., ChatGPT, Claude chat)",
		Schema: schema.Object(map[string]schema.JSON{
			"url": schema.JSON{
				Type:        "string",
				Description: "Chat interface URL or API endpoint",
				Format:      "uri",
			},
			"model": schema.JSON{
				Type:        "string",
				Description: "Model identifier (e.g., gpt-4, claude-3-opus)",
			},
			"headers": schema.JSON{
				Type:        "object",
				Description: "HTTP headers for authentication and configuration",
			},
			"provider": schema.JSON{
				Type:        "string",
				Description: "LLM provider (e.g., openai, anthropic, custom)",
			},
			"system_prompt": schema.JSON{
				Type:        "string",
				Description: "System prompt or instructions for the LLM",
			},
		}, "url"),
	}

	// LLMAPISchema defines connection parameters for programmatic LLM API endpoints.
	// Similar to HTTPAPISchema but with LLM-specific fields.
	LLMAPISchema = types.TargetSchema{
		Type:        "llm_api",
		Version:     "1.0",
		Description: "Programmatic LLM API endpoint for direct model access",
		Schema: schema.Object(map[string]schema.JSON{
			"url": schema.JSON{
				Type:        "string",
				Description: "API endpoint URL",
				Format:      "uri",
			},
			"method": schema.JSON{
				Type:        "string",
				Description: "HTTP method",
				Enum:        []any{"GET", "POST"},
				Default:     "POST",
			},
			"headers": schema.JSON{
				Type:        "object",
				Description: "HTTP headers including API keys",
			},
			"model": schema.JSON{
				Type:        "string",
				Description: "Model identifier",
			},
			"provider": schema.JSON{
				Type:        "string",
				Description: "LLM provider",
			},
			"timeout": schema.JSON{
				Type:        "integer",
				Description: "Request timeout in seconds",
				Minimum:     intPtr(1),
				Default:     30,
			},
		}, "url"),
	}

	// KubernetesSchema defines connection parameters for Kubernetes cluster targets.
	// Used by k8skiller and other Kubernetes-focused agents.
	KubernetesSchema = types.TargetSchema{
		Type:        "kubernetes",
		Version:     "1.0",
		Description: "Kubernetes cluster for testing container security and orchestration",
		Schema: schema.Object(map[string]schema.JSON{
			"cluster": schema.JSON{
				Type:        "string",
				Description: "Cluster name or kubeconfig context",
			},
			"namespace": schema.JSON{
				Type:        "string",
				Description: "Kubernetes namespace to target",
				Default:     "default",
			},
			"kubeconfig": schema.JSON{
				Type:        "string",
				Description: "Path to kubeconfig file (optional if using in-cluster config)",
			},
			"api_server": schema.JSON{
				Type:        "string",
				Description: "Kubernetes API server URL (e.g., https://api.cluster.example.com:6443)",
				Format:      "uri",
			},
		}, "cluster"),
	}

	// SmartContractSchema defines connection parameters for blockchain smart contract targets.
	// Used for testing AI-powered smart contracts and blockchain oracles.
	SmartContractSchema = types.TargetSchema{
		Type:        "smart_contract",
		Version:     "1.0",
		Description: "Blockchain smart contract for testing decentralized AI systems",
		Schema: schema.Object(map[string]schema.JSON{
			"chain": schema.JSON{
				Type:        "string",
				Description: "Blockchain network",
				Enum:        []any{"ethereum", "polygon", "arbitrum", "base", "solana", "optimism"},
			},
			"address": schema.JSON{
				Type:        "string",
				Description: "Contract address",
				Pattern:     "^0x[a-fA-F0-9]{40}$",
			},
			"rpc_url": schema.JSON{
				Type:        "string",
				Description: "RPC endpoint URL for blockchain interaction",
				Format:      "uri",
			},
			"abi": schema.JSON{
				Type:        "string",
				Description: "Contract ABI (Application Binary Interface) as JSON string",
			},
		}, "chain", "address"),
	}
)

// GetBuiltinSchema returns a built-in target schema by type name.
// Returns nil if the type is not recognized.
//
// Example:
//
//	schema := target.GetBuiltinSchema("kubernetes")
//	if schema == nil {
//		log.Fatal("unknown target type")
//	}
func GetBuiltinSchema(typeName string) *types.TargetSchema {
	switch typeName {
	case "http_api":
		return &HTTPAPISchema
	case "llm_chat":
		return &LLMChatSchema
	case "llm_api":
		return &LLMAPISchema
	case "kubernetes":
		return &KubernetesSchema
	case "smart_contract":
		return &SmartContractSchema
	default:
		return nil
	}
}

// ListBuiltinSchemas returns a list of all built-in target schema type names.
// This is useful for CLI help text and documentation.
//
// Example:
//
//	types := target.ListBuiltinSchemas()
//	fmt.Printf("Supported target types: %s\n", strings.Join(types, ", "))
func ListBuiltinSchemas() []string {
	return []string{
		"http_api",
		"llm_chat",
		"llm_api",
		"kubernetes",
		"smart_contract",
	}
}

// intPtr returns a pointer to an int.
// Helper function for setting integer constraints in JSON Schema.
func intPtr(i int) *float64 {
	f := float64(i)
	return &f
}
