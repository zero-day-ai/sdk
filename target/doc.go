// Package target provides built-in target schemas for common attack surfaces.
//
// Target schemas define the connection parameters required to interact with
// different types of systems under test. This package includes pre-defined
// schemas for HTTP APIs, LLM interfaces, Kubernetes clusters, and smart contracts.
//
// # Built-in Schemas
//
// The following target schemas are provided:
//   - http_api: HTTP API endpoints and web services
//   - llm_chat: Conversational LLM interfaces (ChatGPT, Claude)
//   - llm_api: Programmatic LLM API endpoints
//   - kubernetes: Kubernetes cluster targets
//   - smart_contract: Blockchain smart contracts
//
// # Usage
//
// Agents can reference built-in schemas directly:
//
//	import "github.com/zero-day-ai/sdk/target"
//
//	func (a *MyAgent) TargetSchemas() []types.TargetSchema {
//		return []types.TargetSchema{target.HTTPAPISchema}
//	}
//
// Or use the lookup function:
//
//	schema := target.GetBuiltinSchema("kubernetes")
//	if schema == nil {
//		return fmt.Errorf("unknown target type")
//	}
//
// # Custom Schemas
//
// Agents can also define custom target schemas:
//
//	customSchema := types.TargetSchema{
//		Type:        "custom_protocol",
//		Version:     "1.0",
//		Description: "My custom protocol",
//		Schema: schema.Object(map[string]schema.JSON{
//			"host": schema.StringWithDesc("Server hostname"),
//			"port": schema.Int(),
//		}, "host"),
//	}
package target
