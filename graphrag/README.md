# GraphRAG Package - Deterministic ID System

This package provides a deterministic, content-addressable ID generation system for GraphRAG knowledge graph nodes and relationships. The system guarantees that the same logical entity always receives the same ID, enabling idempotent graph operations and reliable relationship references.

## Table of Contents

- [Overview](#overview)
- [Why Deterministic IDs?](#why-deterministic-ids)
- [Quick Start](#quick-start)
- [Node Type Registry](#node-type-registry)
- [ID Generator](#id-generator)
- [Taxonomy Mapping Format](#taxonomy-mapping-format)
- [Migration Guide](#migration-guide)
- [Examples](#examples)
- [Testing](#testing)

## Overview

The deterministic ID system consists of three core components:

1. **NodeTypeRegistry** - Defines identifying properties for each node type
2. **ID Generator** - Creates content-addressable IDs from node type and properties
3. **Taxonomy Mappings** - Declarative mappings from tool output to graph entities

### System Guarantees

- **Deterministic**: Same inputs always produce the same ID
- **Collision-Resistant**: Different inputs produce different IDs (SHA-256 based)
- **Human-Readable**: IDs include node type prefix (e.g., `host:a1b2c3d4`)
- **Stable**: IDs remain consistent across agent runs and missions
- **Fail-Fast**: Missing properties cause explicit errors, never silent failures

## Why Deterministic IDs?

### The Problem with Templates

Previously, IDs were generated using string templates:

```yaml
# OLD FORMAT (DEPRECATED)
id_template: "host:{{.ip}}"
relationships:
  - type: HAS_PORT
    from_template: "host:{{.ip}}"           # Fragile string manipulation
    to_template: "port:{{.ip}}:{{.port}}"   # Easy to misalign
```

This approach had critical flaws:

- **Silent Failures**: Typos in templates created invalid IDs silently
- **Relationship Misalignment**: From/to templates could reference non-existent nodes
- **No Validation**: Missing properties were only discovered at storage time
- **Inconsistent Formatting**: Different agents could format the same ID differently

### The Solution: Content-Addressable IDs

The new system uses **identifying properties** defined in the registry:

```go
// A "host" node is uniquely identified by its "ip" property
registry.GetIdentifyingProperties("host") // Returns: ["ip"]

// Generate ID from properties
generator.Generate("host", map[string]any{"ip": "192.168.1.1"})
// Returns: "host:a1b2c3d4e5f6" (deterministic hash)
```

This guarantees:

- Properties define IDs, not string templates
- Missing properties cause immediate errors
- Same properties always produce same ID
- Relationships can only reference valid nodes

## Quick Start

### Using the Registry

```go
import "github.com/zero-day-ai/sdk/graphrag"

// Get the global registry
registry := graphrag.Registry()

// Check identifying properties for a node type
props, err := registry.GetIdentifyingProperties("host")
if err != nil {
    log.Fatalf("Unknown node type: %v", err)
}
// props = ["ip"]

// Validate properties before creating a node
properties := map[string]any{
    "ip":       "192.168.1.1",
    "hostname": "web-server",
}
missing, err := registry.ValidateProperties("host", properties)
if err != nil {
    log.Fatalf("Missing properties: %v", missing)
}
// No error - all identifying properties present
```

### Generating IDs

```go
import (
    "github.com/zero-day-ai/sdk/graphrag"
    "github.com/zero-day-ai/sdk/graphrag/id"
)

// Create generator with registry
generator := id.NewGenerator(graphrag.Registry())

// Generate ID for a host
hostID, err := generator.Generate("host", map[string]any{
    "ip": "192.168.1.1",
})
if err != nil {
    log.Fatalf("ID generation failed: %v", err)
}
// hostID = "host:a1b2c3d4e5f6" (deterministic)

// Generate ID for a port (compound key)
portID, err := generator.Generate("port", map[string]any{
    "host_id":     hostID,
    "number":      443,
    "protocol":    "tcp",
})
// portID = "port:x7y8z9w0v1u2"

// Same inputs always produce same output
portID2, _ := generator.Generate("port", map[string]any{
    "host_id":  hostID,
    "number":   443,
    "protocol": "tcp",
})
// portID == portID2 (guaranteed)
```

## Node Type Registry

The `NodeTypeRegistry` defines which properties uniquely identify each node type. This is the single source of truth for ID generation.

### Interface

```go
type NodeTypeRegistry interface {
    // GetIdentifyingProperties returns the property names that uniquely
    // identify a node of the given type.
    GetIdentifyingProperties(nodeType string) ([]string, error)

    // IsRegistered checks if a node type exists in the registry.
    IsRegistered(nodeType string) bool

    // ValidateProperties checks if all identifying properties are present.
    // Returns missing property names if validation fails.
    ValidateProperties(nodeType string, properties map[string]any) ([]string, error)

    // AllNodeTypes returns a sorted list of all registered node types.
    AllNodeTypes() []string
}
```

### Registered Node Types

The default registry includes all canonical node types from the GraphRAG taxonomy:

#### Asset Node Types

| Node Type | Identifying Properties | Example |
|-----------|----------------------|---------|
| `host` | `[ip]` | Host identified by IP address |
| `port` | `[host_id, number, protocol]` | Port on a specific host |
| `service` | `[port_id, name]` | Service running on a port |
| `endpoint` | `[service_id, url, method]` | HTTP endpoint on a service |
| `domain` | `[name]` | Domain name |
| `subdomain` | `[parent_domain, name]` | Subdomain under a domain |
| `api` | `[base_url]` | API base URL |
| `technology` | `[name, version]` | Detected technology stack |
| `certificate` | `[fingerprint]` | TLS certificate fingerprint |
| `cloud_asset` | `[provider, resource_id]` | Cloud resource identifier |

#### Finding Node Types

| Node Type | Identifying Properties | Example |
|-----------|----------------------|---------|
| `finding` | `[mission_id, fingerprint]` | Security finding in a mission |
| `evidence` | `[finding_id, type, fingerprint]` | Evidence for a finding |
| `mitigation` | `[finding_id, title]` | Mitigation for a finding |

#### Execution Node Types

| Node Type | Identifying Properties | Example |
|-----------|----------------------|---------|
| `mission` | `[name, timestamp]` | Mission execution |
| `agent_run` | `[mission_id, agent_name, run_number]` | Agent execution in a mission |
| `tool_execution` | `[agent_run_id, tool_name, sequence]` | Tool invocation |
| `llm_call` | `[agent_run_id, sequence]` | LLM API call |

#### Attack Node Types

| Node Type | Identifying Properties | Example |
|-----------|----------------------|---------|
| `technique` | `[id]` | Attack technique (MITRE/Arcanum) |
| `tactic` | `[id]` | Attack tactic |

#### Intelligence Node Types

| Node Type | Identifying Properties | Example |
|-----------|----------------------|---------|
| `intelligence` | `[mission_id, title, timestamp]` | Intelligence insight |

### Using the Registry

```go
// Check if a node type is registered
if !registry.IsRegistered("unknown_type") {
    log.Println("Node type not found")
}

// List all registered types
allTypes := registry.AllNodeTypes()
// allTypes = ["agent_run", "api", "certificate", "cloud_asset", ...]

// Validate before creating node
properties := map[string]any{
    "number":   443,
    "protocol": "tcp",
    // Missing: host_id
}
missing, err := registry.ValidateProperties("port", properties)
if errors.Is(err, graphrag.ErrMissingIdentifyingProperties) {
    log.Printf("Cannot create port: missing %v", missing)
    // Output: Cannot create port: missing [host_id]
}
```

## ID Generator

The `Generator` interface creates deterministic IDs from node type and identifying properties.

### Interface

```go
package id

type Generator interface {
    // Generate creates a deterministic ID from node type and properties.
    // Returns an error if:
    //   - The node type is not registered
    //   - Required identifying properties are missing
    Generate(nodeType string, properties map[string]any) (string, error)
}
```

### Implementation Details

The `DeterministicGenerator` uses SHA-256 hashing:

1. Get identifying properties from registry for the node type
2. Validate all identifying properties are present
3. Build canonical string: `nodeType:prop1=val1|prop2=val2` (sorted keys)
4. Normalize values (lowercase strings, format numbers)
5. SHA-256 hash the canonical string
6. Base64url encode first 12 bytes (96 bits)
7. Return `{nodeType}:{encoded}`

### ID Format

```
Format: {node_type}:{base64url_hash}
Length: ~20-25 characters
Example: host:a1b2c3d4e5f6

Components:
- node_type: Canonical type from registry
- separator: Colon (:)
- hash: Base64url-encoded first 12 bytes of SHA-256 (no padding)
```

### Value Normalization

To ensure deterministic IDs, property values are normalized:

| Type | Normalization |
|------|---------------|
| `string` | Lowercase and trimmed whitespace |
| `int`, `int64`, etc. | Formatted as `%d` |
| `float32`, `float64` | Formatted as `%.6f` (6 decimals) |
| `bool` | `"true"` or `"false"` |
| `nil` | `"null"` |
| Complex types | JSON marshaled |

### Examples

```go
generator := id.NewGenerator(graphrag.Registry())

// Simple key (single property)
id1, _ := generator.Generate("host", map[string]any{"ip": "192.168.1.1"})
id2, _ := generator.Generate("host", map[string]any{"ip": "192.168.1.1"})
// id1 == id2 (guaranteed)

// Compound key (multiple properties)
id3, _ := generator.Generate("port", map[string]any{
    "host_id":  "host:abc123",
    "number":   443,
    "protocol": "tcp",
})

// Property order doesn't matter
id4, _ := generator.Generate("port", map[string]any{
    "protocol": "tcp",        // Different order
    "number":   443,
    "host_id":  "host:abc123",
})
// id3 == id4 (properties are sorted internally)

// Case normalization
id5, _ := generator.Generate("domain", map[string]any{"name": "Example.COM"})
id6, _ := generator.Generate("domain", map[string]any{"name": "example.com"})
// id5 == id6 (strings normalized to lowercase)
```

### Error Handling

```go
// Unknown node type
_, err := generator.Generate("unknown_type", map[string]any{})
if errors.Is(err, graphrag.ErrNodeTypeNotRegistered) {
    log.Println("Node type not in registry")
}

// Missing identifying properties
_, err = generator.Generate("host", map[string]any{
    "hostname": "web-server", // Has hostname but missing "ip"
})
if errors.Is(err, graphrag.ErrMissingIdentifyingProperties) {
    log.Println("Missing required property: ip")
}
```

## Taxonomy Mapping Format

Taxonomy mappings define how tool output maps to graph entities. The new format uses **identifying properties** and **node references** instead of string templates.

### New Format (Current)

```yaml
taxonomy:
  node_type: port

  # Identifying properties: map property names to JSONPath expressions
  identifying_properties:
    host_id: "$.host_ip"       # Will be used to generate parent host ID
    number: "$.port"
    protocol: "$.protocol"

  # Additional properties (not used for ID generation)
  properties:
    - source: "$.state"
      target: "state"
    - source: "$.service"
      target: "service_name"

  # Relationships use node references, not templates
  relationships:
    - type: HAS_PORT
      from:
        type: host              # Reference parent host
        properties:
          ip: "$.host_ip"       # Properties to generate host ID
      to:
        type: self              # Reference current port node
```

### Old Format (Deprecated)

```yaml
# DO NOT USE - This format is deprecated and will be removed
taxonomy:
  node_type: port
  id_template: "port:{{.host_ip}}:{{.port}}"  # REMOVED

  properties:
    - source: "$.port"
      target: "port_number"

  relationships:
    - type: HAS_PORT
      from_template: "host:{{.host_ip}}"           # REMOVED
      to_template: "port:{{.host_ip}}:{{.port}}"   # REMOVED
```

### Key Differences

| Aspect | Old Format | New Format |
|--------|-----------|-----------|
| Node ID | `id_template` with Go templates | Generated from `identifying_properties` |
| Relationships | `from_template`, `to_template` strings | `NodeReference` with type and properties |
| Validation | Runtime (at storage) | Compile-time (at mapping parse) |
| Type Safety | None (strings) | Full (typed references) |
| Errors | Silent failures | Explicit errors with context |

## Migration Guide

### Step 1: Convert `id_template` to `identifying_properties`

**Before:**
```yaml
node_type: host
id_template: "host:{{.ip}}"
```

**After:**
```yaml
node_type: host
identifying_properties:
  ip: "$.ip"  # JSONPath to extract identifying property
```

**Before:**
```yaml
node_type: port
id_template: "port:{{.host_id}}:{{.port_number}}:{{.protocol}}"
```

**After:**
```yaml
node_type: port
identifying_properties:
  host_id: "$.host_id"
  number: "$.port_number"
  protocol: "$.protocol"
```

### Step 2: Convert Relationship Templates to Node References

**Before:**
```yaml
relationships:
  - type: HAS_PORT
    from_template: "host:{{.host_ip}}"
    to_template: "port:{{.host_ip}}:{{.port}}:{{.protocol}}"
```

**After:**
```yaml
relationships:
  - type: HAS_PORT
    from:
      type: host
      properties:
        ip: "$.host_ip"
    to:
      type: self  # References the current port node
```

### Step 3: Use "self" for Current Node

**Before:**
```yaml
relationships:
  - type: RUNS_SERVICE
    from_template: "port:{{.host_id}}:{{.port_number}}:{{.protocol}}"
    to_template: "service:{{.port_id}}:{{.service_name}}"
```

**After:**
```yaml
relationships:
  - type: RUNS_SERVICE
    from:
      type: self  # Current port node
    to:
      type: service
      properties:
        port_id: "$.port_id"
        name: "$.service_name"
```

### Complete Migration Example

**Before (nmap port scan output):**
```yaml
- node_type: port
  id_template: "port:{{.host}}:{{.portid}}:{{.protocol}}"
  path: "$.ports[*]"
  properties:
    - source: "$.portid"
      target: "port_number"
    - source: "$.protocol"
      target: "protocol"
    - source: "$.state.state"
      target: "state"
  relationships:
    - type: HAS_PORT
      from_template: "host:{{.host}}"
      to_template: "port:{{.host}}:{{.portid}}:{{.protocol}}"
```

**After:**
```yaml
- node_type: port
  path: "$.ports[*]"

  identifying_properties:
    host_id: "$.host"
    number: "$.portid"
    protocol: "$.protocol"

  properties:
    - source: "$.state.state"
      target: "state"

  relationships:
    - type: HAS_PORT
      from:
        type: host
        properties:
          ip: "$.host"
      to:
        type: self
```

## Examples

### Example 1: Simple Host Node

```go
import (
    "github.com/zero-day-ai/sdk/graphrag"
    "github.com/zero-day-ai/sdk/graphrag/id"
)

func createHostNode(ipAddress string) (*graphrag.GraphNode, error) {
    registry := graphrag.Registry()
    generator := id.NewGenerator(registry)

    // Generate deterministic ID
    properties := map[string]any{"ip": ipAddress}
    hostID, err := generator.Generate("host", properties)
    if err != nil {
        return nil, fmt.Errorf("failed to generate host ID: %w", err)
    }

    // Create node with generated ID
    node := graphrag.NewGraphNode("host").
        WithID(hostID).
        WithProperty("ip", ipAddress).
        WithProperty("hostname", "web-server.example.com")

    return node, nil
}
```

### Example 2: Port with Parent Relationship

```go
func createPortWithRelationship(hostIP string, portNumber int, protocol string) (*graphrag.Batch, error) {
    registry := graphrag.Registry()
    generator := id.NewGenerator(registry)

    // Generate host ID
    hostID, err := generator.Generate("host", map[string]any{"ip": hostIP})
    if err != nil {
        return nil, err
    }

    // Generate port ID
    portID, err := generator.Generate("port", map[string]any{
        "host_id":  hostID,
        "number":   portNumber,
        "protocol": protocol,
    })
    if err != nil {
        return nil, err
    }

    // Create batch with both nodes and relationship
    batch := graphrag.NewBatch()

    // Host node
    host := graphrag.NewGraphNode("host").
        WithID(hostID).
        WithProperty("ip", hostIP)
    batch.Nodes = append(batch.Nodes, *host)

    // Port node
    port := graphrag.NewGraphNode("port").
        WithID(portID).
        WithProperty("host_id", hostID).
        WithProperty("number", portNumber).
        WithProperty("protocol", protocol)
    batch.Nodes = append(batch.Nodes, *port)

    // Relationship (validated - both nodes in batch)
    rel := graphrag.NewRelationship(hostID, portID, graphrag.RelTypeHasPort)
    batch.Relationships = append(batch.Relationships, *rel)

    return batch, nil
}
```

### Example 3: Extracting from Tool Output

```go
func extractFromNmapOutput(nmapJSON []byte) (*graphrag.Batch, error) {
    registry := graphrag.Registry()
    generator := id.NewGenerator(registry)
    batch := graphrag.NewBatch()

    // Parse nmap output
    var output struct {
        Host  string `json:"host"`
        Ports []struct {
            Port     int    `json:"portid"`
            Protocol string `json:"protocol"`
            State    string `json:"state"`
        } `json:"ports"`
    }
    if err := json.Unmarshal(nmapJSON, &output); err != nil {
        return nil, err
    }

    // Generate host ID
    hostID, err := generator.Generate("host", map[string]any{
        "ip": output.Host,
    })
    if err != nil {
        return nil, err
    }

    // Create host node
    host := graphrag.NewGraphNode("host").
        WithID(hostID).
        WithProperty("ip", output.Host)
    batch.Nodes = append(batch.Nodes, *host)

    // Create port nodes and relationships
    for _, p := range output.Ports {
        portID, err := generator.Generate("port", map[string]any{
            "host_id":  hostID,
            "number":   p.Port,
            "protocol": p.Protocol,
        })
        if err != nil {
            return nil, err
        }

        port := graphrag.NewGraphNode("port").
            WithID(portID).
            WithProperty("host_id", hostID).
            WithProperty("number", p.Port).
            WithProperty("protocol", p.Protocol).
            WithProperty("state", p.State)
        batch.Nodes = append(batch.Nodes, *port)

        rel := graphrag.NewRelationship(hostID, portID, graphrag.RelTypeHasPort)
        batch.Relationships = append(batch.Relationships, *rel)
    }

    return batch, nil
}
```

## Testing

### Unit Testing ID Generation

```go
func TestDeterministicIDs(t *testing.T) {
    registry := graphrag.NewDefaultNodeTypeRegistry()
    generator := id.NewGenerator(registry)

    // Test 1: Same inputs produce same ID
    id1, err := generator.Generate("host", map[string]any{"ip": "10.0.0.1"})
    require.NoError(t, err)

    id2, err := generator.Generate("host", map[string]any{"ip": "10.0.0.1"})
    require.NoError(t, err)

    assert.Equal(t, id1, id2, "Same inputs must produce same ID")

    // Test 2: Different inputs produce different IDs
    id3, err := generator.Generate("host", map[string]any{"ip": "10.0.0.2"})
    require.NoError(t, err)

    assert.NotEqual(t, id1, id3, "Different inputs must produce different IDs")

    // Test 3: Property order doesn't matter
    id4, err := generator.Generate("port", map[string]any{
        "host_id":  "host:abc",
        "number":   443,
        "protocol": "tcp",
    })
    require.NoError(t, err)

    id5, err := generator.Generate("port", map[string]any{
        "protocol": "tcp",
        "number":   443,
        "host_id":  "host:abc",
    })
    require.NoError(t, err)

    assert.Equal(t, id4, id5, "Property order must not affect ID")
}
```

### Testing Registry

```go
func TestRegistry(t *testing.T) {
    registry := graphrag.NewDefaultNodeTypeRegistry()

    // Test 1: All standard types registered
    assert.True(t, registry.IsRegistered("host"))
    assert.True(t, registry.IsRegistered("port"))
    assert.True(t, registry.IsRegistered("finding"))

    // Test 2: Unknown types return error
    _, err := registry.GetIdentifyingProperties("unknown_type")
    assert.ErrorIs(t, err, graphrag.ErrNodeTypeNotRegistered)

    // Test 3: Validation catches missing properties
    _, err = registry.ValidateProperties("host", map[string]any{
        "hostname": "server", // Missing "ip"
    })
    assert.ErrorIs(t, err, graphrag.ErrMissingIdentifyingProperties)

    // Test 4: Validation passes with all properties
    _, err = registry.ValidateProperties("host", map[string]any{
        "ip":       "10.0.0.1",
        "hostname": "server", // Extra properties OK
    })
    assert.NoError(t, err)
}
```

### Integration Testing

```go
func TestEndToEndExtraction(t *testing.T) {
    // Setup
    registry := graphrag.NewDefaultNodeTypeRegistry()
    generator := id.NewGenerator(registry)

    // Simulate nmap output
    nmapOutput := `{"host": "10.0.0.1", "ports": [{"portid": 443, "protocol": "tcp"}]}`

    // Extract to batch
    batch, err := extractFromNmapOutput([]byte(nmapOutput))
    require.NoError(t, err)

    // Verify nodes
    assert.Len(t, batch.Nodes, 2) // host + port

    hostNode := batch.Nodes[0]
    assert.Equal(t, "host", hostNode.Type)
    assert.Contains(t, hostNode.ID, "host:")

    portNode := batch.Nodes[1]
    assert.Equal(t, "port", portNode.Type)
    assert.Contains(t, portNode.ID, "port:")

    // Verify relationship
    assert.Len(t, batch.Relationships, 1)
    rel := batch.Relationships[0]
    assert.Equal(t, hostNode.ID, rel.FromID)
    assert.Equal(t, portNode.ID, rel.ToID)
    assert.Equal(t, "HAS_PORT", rel.Type)

    // Test idempotency: re-extract same output
    batch2, err := extractFromNmapOutput([]byte(nmapOutput))
    require.NoError(t, err)

    assert.Equal(t, batch.Nodes[0].ID, batch2.Nodes[0].ID, "Host ID must be stable")
    assert.Equal(t, batch.Nodes[1].ID, batch2.Nodes[1].ID, "Port ID must be stable")
}
```

## Best Practices

1. **Always use the registry** - Never hardcode property lists
2. **Validate before creating** - Use `ValidateProperties` to catch errors early
3. **Use identifying properties only** - Don't include extra properties in ID generation
4. **Normalize inputs** - The generator handles normalization, but consistent inputs help
5. **Test idempotency** - Verify same inputs produce same IDs in tests
6. **Handle errors explicitly** - Never ignore ID generation errors

## See Also

- **`registry.go`** - NodeTypeRegistry implementation
- **`id/generator.go`** - Deterministic ID generator
- **`taxonomy_generated.go`** - Canonical node and relationship types
- **`node.go`** - GraphNode and Relationship types
- **`../schema/taxonomy.go`** - Taxonomy mapping structures
