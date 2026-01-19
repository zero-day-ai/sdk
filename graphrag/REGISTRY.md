# Node Type Registry

The Node Type Registry provides a centralized system for managing and validating identifying properties for all GraphRAG node types. It ensures that nodes have the minimum required properties before creation and forms the foundation for deterministic ID generation.

## Overview

The registry defines **identifying properties** for each canonical node type - the minimum set of properties that uniquely identify a node in the knowledge graph. These properties:

- Form a natural key for deduplication
- Enable deterministic ID generation (to be implemented)
- Ensure data consistency across the graph
- Provide validation before node creation

## Quick Start

```go
import "github.com/zero-day-ai/sdk/graphrag"

// Access the global registry
registry := graphrag.Registry()

// Check if a node type is registered
if registry.IsRegistered(graphrag.NodeTypeHost) {
    // Get its identifying properties
    props, _ := registry.GetIdentifyingProperties(graphrag.NodeTypeHost)
    // props = ["ip"]
}

// Validate properties before node creation
properties := map[string]any{
    graphrag.PropIP: "10.0.0.1",
}
missing, err := registry.ValidateProperties(graphrag.NodeTypeHost, properties)
if err != nil {
    // Handle missing properties
}
```

## Node Type Identifying Properties

### Asset Node Types

| Node Type | Identifying Properties | Example |
|-----------|------------------------|---------|
| `host` | `ip` | IP address uniquely identifies the host |
| `port` | `host_id`, `number`, `protocol` | Port on a specific host with protocol |
| `service` | `port_id`, `name` | Named service running on a port |
| `endpoint` | `service_id`, `url`, `method` | HTTP endpoint with method |
| `domain` | `name` | Domain name (e.g., example.com) |
| `subdomain` | `parent_domain`, `name` | Subdomain under a parent domain |
| `api` | `base_url` | API identified by base URL |
| `technology` | `name`, `version` | Technology name and version |
| `certificate` | `fingerprint` | Certificate fingerprint (SHA-256) |
| `cloud_asset` | `provider`, `resource_id` | Cloud resource in a provider |

### Finding Node Types

| Node Type | Identifying Properties | Example |
|-----------|------------------------|---------|
| `finding` | `mission_id`, `fingerprint` | Unique finding within a mission |
| `evidence` | `finding_id`, `type`, `fingerprint` | Evidence attached to a finding |
| `mitigation` | `finding_id`, `title` | Mitigation for a specific finding |

### Execution Node Types

| Node Type | Identifying Properties | Example |
|-----------|------------------------|---------|
| `mission` | `name`, `timestamp` | Mission with unique start time |
| `agent_run` | `mission_id`, `agent_name`, `run_number` | Agent execution within a mission |
| `tool_execution` | `agent_run_id`, `tool_name`, `sequence` | Tool call in an agent run |
| `llm_call` | `agent_run_id`, `sequence` | LLM API call in an agent run |

### Attack Node Types

| Node Type | Identifying Properties | Example |
|-----------|------------------------|---------|
| `technique` | `id` | MITRE ATT&CK or Arcanum technique ID |
| `tactic` | `id` | MITRE ATT&CK tactic ID |

### Intelligence Node Types

| Node Type | Identifying Properties | Example |
|-----------|------------------------|---------|
| `intelligence` | `mission_id`, `title`, `timestamp` | LLM-generated intelligence report |

## API Reference

### NodeTypeRegistry Interface

```go
type NodeTypeRegistry interface {
    // GetIdentifyingProperties returns the property names that uniquely identify a node type.
    GetIdentifyingProperties(nodeType string) ([]string, error)

    // IsRegistered checks if a node type exists in the registry.
    IsRegistered(nodeType string) bool

    // ValidateProperties checks if all identifying properties are present.
    // Returns missing property names if validation fails.
    ValidateProperties(nodeType string, properties map[string]any) ([]string, error)

    // AllNodeTypes returns a sorted list of all registered node type names.
    AllNodeTypes() []string
}
```

### Functions

```go
// Registry returns the global registry instance (lazily initialized)
func Registry() NodeTypeRegistry

// NewDefaultNodeTypeRegistry creates a new registry with all canonical types
func NewDefaultNodeTypeRegistry() *DefaultNodeTypeRegistry

// SetRegistry replaces the global registry (for testing)
func SetRegistry(registry NodeTypeRegistry)
```

## Usage Patterns

### Pattern 1: Validation Before Node Creation

```go
registry := graphrag.Registry()

// Prepare node properties
properties := map[string]any{
    graphrag.PropHostID:   "host-123",
    graphrag.PropNumber:   443,
    graphrag.PropProtocol: "tcp",
}

// Validate before creating
missing, err := registry.ValidateProperties(graphrag.NodeTypePort, properties)
if err != nil {
    return fmt.Errorf("cannot create port node: %w, missing: %v", err, missing)
}

// Safe to create node
node := graphrag.NewGraphNode(graphrag.NodeTypePort).
    WithProperties(properties)
```

### Pattern 2: Dynamic Property Discovery

```go
registry := graphrag.Registry()

// Discover required properties for a node type
props, err := registry.GetIdentifyingProperties(graphrag.NodeTypeService)
if err != nil {
    return err
}

fmt.Printf("Service requires: %v\n", props)
// Output: Service requires: [port_id name]
```

### Pattern 3: Listing All Node Types

```go
registry := graphrag.Registry()

// Get all canonical node types
types := registry.AllNodeTypes()

// Present to user or LLM
fmt.Println("Available node types:")
for _, t := range types {
    props, _ := registry.GetIdentifyingProperties(t)
    fmt.Printf("  %s: %v\n", t, props)
}
```

### Pattern 4: Custom Registry for Testing

```go
func TestMyFunction(t *testing.T) {
    // Create mock registry
    mockRegistry := &MockRegistry{
        types: map[string][]string{
            "test_type": {"prop1", "prop2"},
        },
    }

    // Temporarily replace global registry
    graphrag.SetRegistry(mockRegistry)
    defer graphrag.SetRegistry(graphrag.NewDefaultNodeTypeRegistry())

    // Run tests with mock registry
    // ...
}
```

## Error Handling

The registry defines two sentinel errors:

```go
// ErrNodeTypeNotRegistered - unknown node type
if errors.Is(err, graphrag.ErrNodeTypeNotRegistered) {
    // Handle unknown type
}

// ErrMissingIdentifyingProperties - validation failed
if errors.Is(err, graphrag.ErrMissingIdentifyingProperties) {
    // Handle missing properties
    fmt.Printf("Missing: %v\n", missing)
}
```

## Thread Safety

The registry is fully thread-safe:

- All read operations use `RLock()` for concurrent reads
- The global registry uses `sync.Once` for lazy initialization
- Property slices are copied on return to prevent external modification

## Future Enhancements

1. **Deterministic ID Generation**: Use identifying properties to generate consistent node IDs
   ```go
   id, err := registry.GenerateID(nodeType, properties)
   ```

2. **Property Type Validation**: Enforce property types beyond presence
   ```go
   registry.ValidatePropertyTypes(nodeType, properties)
   ```

3. **Custom Property Rules**: Support custom validation rules per node type
   ```go
   registry.RegisterValidator(nodeType, validatorFunc)
   ```

4. **Schema Export**: Generate JSON Schema for node types
   ```go
   schema, err := registry.ExportSchema(nodeType)
   ```

## Integration Points

### With Node Creation

```go
// In Gibson's graph storage layer
func (s *Storage) CreateNode(ctx context.Context, node *graphrag.GraphNode) error {
    // Validate before storing
    missing, err := graphrag.Registry().ValidateProperties(node.Type, node.Properties)
    if err != nil {
        return fmt.Errorf("invalid node: %w", err)
    }

    // Generate deterministic ID (future)
    // node.ID = generateID(node.Type, node.Properties)

    // Store in database
    return s.insert(ctx, node)
}
```

### With Agent SDK

```go
// Agents use validation in harness
func (h *Harness) StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error) {
    // Validate node properties
    if missing, err := graphrag.Registry().ValidateProperties(node.Type, node.Properties); err != nil {
        return "", fmt.Errorf("invalid %s node: %w (missing: %v)", node.Type, err, missing)
    }

    // Call Gibson storage
    return h.graphClient.StoreNode(ctx, &node)
}
```

### With LLM Tool Definitions

```go
// Generate tool schema from registry
func generateGraphToolSchema() schema.JSON {
    registry := graphrag.Registry()

    nodeTypes := registry.AllNodeTypes()
    typeSchemas := make(map[string]any)

    for _, nodeType := range nodeTypes {
        props, _ := registry.GetIdentifyingProperties(nodeType)
        typeSchemas[nodeType] = map[string]any{
            "required_properties": props,
        }
    }

    return schema.Object(typeSchemas)
}
```

## Best Practices

1. **Always validate before creating nodes**: Use `ValidateProperties()` to catch errors early

2. **Use constants for property names**: Use `graphrag.Prop*` constants instead of string literals

3. **Handle missing properties gracefully**: The `missing` slice tells you exactly what's needed

4. **Don't modify returned slices**: Property slices are copies, but avoid confusion

5. **Use the global registry**: `graphrag.Registry()` for production code, custom registries for tests

6. **Check IsRegistered first**: Before calling other methods, verify the type exists

7. **Sort node types for user display**: `AllNodeTypes()` returns sorted list for consistent UIs

## Testing

The registry includes comprehensive tests:

- Unit tests for all methods
- Validation tests for all 20 canonical node types
- Concurrency tests for thread safety
- Error handling tests for edge cases
- Example tests demonstrating usage patterns

Run tests:
```bash
cd opensource/sdk
make test
go test -v ./graphrag -run TestRegistry
go test -v ./graphrag -run Example
```

## See Also

- `taxonomy_generated.go` - All canonical node and relationship types
- `node.go` - GraphNode structure and creation
- `errors.go` - GraphRAG error types
- `CLAUDE.md` (SDK) - Complete SDK development guide
