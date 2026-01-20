# GraphRAG Package

This package provides the GraphRAG (Graph Retrieval-Augmented Generation) system for the Gibson security testing framework. It includes strongly-typed domain objects, deterministic ID generation, and a registry for node type definitions.

**Version:** 0.20.0

## Table of Contents

- [Overview](#overview)
- [Domain Types (Recommended)](#domain-types-recommended)
- [Quick Start](#quick-start)
- [Node Type Registry](#node-type-registry)
- [ID Generator](#id-generator)
- [Custom Types](#custom-types)
- [Migration from Taxonomy Mappings](#migration-from-taxonomy-mappings)
- [Examples](#examples)
- [Testing](#testing)

## Overview

The GraphRAG package provides:

1. **Domain Types** (`domain/`) - Strongly-typed Go structs for graph entities (Host, Port, Service, etc.)
2. **NodeTypeRegistry** - Defines identifying properties for each node type
3. **ID Generator** - Creates deterministic, content-addressable IDs
4. **CustomEntity** - Base type for agent-specific custom node types

### Architecture (v0.20.0)

```
SDK GraphRAG Package
├── domain/           # Strongly-typed domain objects (NEW)
│   ├── interfaces.go # GraphNode interface
│   ├── host.go       # Host domain type
│   ├── port.go       # Port domain type
│   ├── service.go    # Service domain type
│   ├── custom.go     # CustomEntity for extensibility
│   └── discovery.go  # DiscoveryResult container
├── id/               # Deterministic ID generation
│   └── generator.go
├── registry.go       # NodeTypeRegistry
└── taxonomy_generated.go  # Node/relationship type constants
```

### System Guarantees

- **Type-Safe**: Domain types provide compile-time checking
- **Deterministic**: Same inputs always produce the same ID
- **Collision-Resistant**: Different inputs produce different IDs (SHA-256 based)
- **Automatic Relationships**: Parent references create relationships automatically
- **Fail-Fast**: Missing properties cause explicit errors, never silent failures

## Domain Types (Recommended)

The recommended way to work with GraphRAG is through strongly-typed domain objects. These provide compile-time safety and automatic relationship creation.

### GraphNode Interface

All domain types implement the `GraphNode` interface:

```go
type GraphNode interface {
    // NodeType returns the canonical node type (e.g., "host", "port")
    NodeType() string

    // IdentifyingProperties returns properties that uniquely identify this node
    IdentifyingProperties() map[string]any

    // Properties returns all properties to set on the node
    Properties() map[string]any

    // ParentRef returns reference to parent node for relationship creation
    ParentRef() *NodeRef

    // RelationshipType returns the relationship type to parent
    RelationshipType() string
}
```

### Using Domain Types

```go
import "github.com/zero-day-ai/sdk/graphrag/domain"

// Create a DiscoveryResult to hold discoveries
result := domain.NewDiscoveryResult()

// Add hosts (root nodes - no parent)
result.Hosts = append(result.Hosts, &domain.Host{
    IP:       "192.168.1.10",
    Hostname: "web-server.example.com",
    State:    "up",
})

// Add ports (child of Host - parent relationship automatic)
result.Ports = append(result.Ports, &domain.Port{
    HostID:   "192.168.1.10",
    Number:   443,
    Protocol: "tcp",
    State:    "open",
})

// Add services (child of Port)
result.Services = append(result.Services, &domain.Service{
    PortID:  "192.168.1.10:443:tcp",
    Name:    "https",
    Version: "nginx/1.18.0",
})

// Get all nodes in dependency order (parents before children)
nodes := result.AllNodes()
```

### Available Domain Types

| Type | Parent | Identifying Properties |
|------|--------|----------------------|
| `Host` | None | `ip` |
| `Port` | Host | `host_id`, `number`, `protocol` |
| `Service` | Port | `port_id`, `name` |
| `Endpoint` | Service | `service_id`, `url`, `method` |
| `Domain` | None | `name` |
| `Subdomain` | Domain | `name`, `domain_name` |
| `Technology` | None | `name`, `version` |
| `Certificate` | None | `serial_number` |
| `CloudAsset` | None | `provider`, `region`, `resource_id` |
| `API` | None | `base_url` |
| `AgentRun` | None | `id`, `mission_id`, `agent_name`, `run_number` |
| `ToolExecution` | AgentRun | `id`, `agent_run_id`, `tool_name`, `sequence` |

### Why Domain Types?

**Before (Old Taxonomy Mapping):**
```go
// Manual JSON construction, no type safety, error-prone
node := map[string]any{
    "type": "host",
    "properties": map[string]any{
        "ip": "192.168.1.1",
    },
}
// Relationships created manually, IDs could misalign
```

**After (Domain Types):**
```go
// Strongly-typed, compiler-verified, automatic relationships
host := &domain.Host{
    IP:       "192.168.1.1",
    Hostname: "web-server",
}
// ParentRef() and RelationshipType() handle relationships automatically
```

## Why Deterministic IDs?

IDs are generated from identifying properties using SHA-256 hashing. This ensures:

- **Idempotent Operations**: Same entity always gets same ID
- **MERGE-Safe**: Neo4j MERGE operations work correctly
- **Stable References**: Relationships reference correct nodes
- **Debug-Friendly**: IDs include type prefix (`host:a1b2c3d4`)

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

## Custom Types

Use `CustomEntity` for agent-specific types not covered by canonical domain types:

```go
import "github.com/zero-day-ai/sdk/graphrag/domain"

// Create a Kubernetes pod (custom type)
pod := domain.NewCustomEntity("k8s", "pod").
    WithIDProps(map[string]any{
        "namespace": "default",
        "name":      "web-server-abc123",
    }).
    WithAllProps(map[string]any{
        "namespace": "default",
        "name":      "web-server-abc123",
        "status":    "Running",
        "image":     "nginx:1.21",
    }).
    WithParent(&domain.NodeRef{
        NodeType: "k8s:node",
        Properties: map[string]any{
            "name": "worker-node-01",
        },
    }, "RUNS_ON")

result.Custom = append(result.Custom, pod)
```

### Namespace Conventions

| Namespace | Purpose | Examples |
|-----------|---------|----------|
| `k8s:` | Kubernetes resources | `k8s:pod`, `k8s:service`, `k8s:deployment` |
| `aws:` | AWS-specific resources | `aws:security_group`, `aws:iam_role` |
| `azure:` | Azure-specific resources | `azure:resource_group` |
| `gcp:` | GCP-specific resources | `gcp:compute_instance` |
| `vuln:` | Vulnerability types | `vuln:cve`, `vuln:exploit` |
| `custom:` | Agent-specific types | `custom:my_type` |

### When to Use CustomEntity vs Canonical Types

| Scenario | Recommendation |
|----------|----------------|
| Standard network/web assets | Use canonical types (`Host`, `Port`, `Service`) |
| Kubernetes resources | Use `CustomEntity` with `k8s:` namespace |
| Cloud provider specific | Use `CustomEntity` with `aws:`/`azure:`/`gcp:` namespace |
| Used by multiple agents | Consider adding canonical type to SDK |

## Migration from Taxonomy Mappings

The old taxonomy mapping system (YAML-based JSONPath extraction) has been **removed** in v0.20.0. Tools must now return `DiscoveryResult` with domain types.

### What Was Removed

The following files were deleted:
- `gibson/internal/graphrag/engine/taxonomy_engine.go`
- `gibson/internal/graphrag/engine/jsonpath.go`
- `gibson/internal/graphrag/engine/template.go`
- `gibson/internal/graphrag/taxonomy/` directory
- `sdk/schema/taxonomy.go`

### Migration Steps

**1. Replace raw JSON output with DiscoveryResult:**

```go
// OLD (Removed)
func execute(input map[string]any) (map[string]any, error) {
    return map[string]any{
        "hosts": []map[string]any{
            {"ip": "192.168.1.1"},
        },
    }, nil
}

// NEW
func execute(input map[string]any) (*domain.DiscoveryResult, error) {
    result := domain.NewDiscoveryResult()
    result.Hosts = append(result.Hosts, &domain.Host{
        IP: "192.168.1.1",
    })
    return result, nil
}
```

**2. Remove schema.TaxonomyMapping definitions:**

```go
// OLD (Removed) - No longer needed
hostSchema := schema.Object(...).WithTaxonomy(schema.TaxonomyMapping{
    NodeType: "host",
    IdentifyingProperties: map[string]string{"ip": "ip"},
    // ...
})

// NEW - Domain types are self-describing
host := &domain.Host{IP: "192.168.1.1"}
// NodeType(), IdentifyingProperties(), etc. are built-in methods
```

**3. Remove `--schema` flag handlers:**

Tools no longer need to export taxonomy schemas. The domain types define everything.

## Examples

### Example 1: Simple Port Scanner Tool

```go
import "github.com/zero-day-ai/sdk/graphrag/domain"

func executePortScan(input map[string]any) (*domain.DiscoveryResult, error) {
    target := input["target"].(string)
    result := domain.NewDiscoveryResult()

    // Discover host
    result.Hosts = append(result.Hosts, &domain.Host{
        IP:       target,
        Hostname: "web-server.example.com",
        State:    "up",
        OS:       "Linux",
    })

    // Discover ports (relationships are automatic via HostID)
    openPorts := []int{22, 80, 443, 3306}
    for _, portNum := range openPorts {
        result.Ports = append(result.Ports, &domain.Port{
            HostID:   target,
            Number:   portNum,
            Protocol: "tcp",
            State:    "open",
        })
    }

    // Discover services
    result.Services = append(result.Services, &domain.Service{
        PortID:  target + ":443:tcp",
        Name:    "https",
        Version: "nginx/1.18.0",
    })

    return result, nil
}
```

### Example 2: Kubernetes Scanner Agent

```go
import "github.com/zero-day-ai/sdk/graphrag/domain"

func scanKubernetesCluster() *domain.DiscoveryResult {
    result := domain.NewDiscoveryResult()

    // Mix canonical and custom types
    // Canonical: Discovered hosts
    result.Hosts = append(result.Hosts, &domain.Host{
        IP:       "10.0.1.50",
        Hostname: "k8s-node-01",
    })

    // Custom: Kubernetes nodes
    k8sNode := domain.NewCustomEntity("k8s", "node").
        WithIDProps(map[string]any{"name": "k8s-node-01"}).
        WithAllProps(map[string]any{
            "name":     "k8s-node-01",
            "status":   "Ready",
            "version":  "v1.28.0",
        })
    result.Custom = append(result.Custom, k8sNode)

    // Custom: Kubernetes pods with parent relationship
    pod := domain.NewCustomEntity("k8s", "pod").
        WithIDProps(map[string]any{
            "namespace": "default",
            "name":      "web-server-abc123",
        }).
        WithAllProps(map[string]any{
            "namespace": "default",
            "name":      "web-server-abc123",
            "status":    "Running",
            "image":     "nginx:1.21",
        }).
        WithParent(&domain.NodeRef{
            NodeType:   "k8s:node",
            Properties: map[string]any{"name": "k8s-node-01"},
        }, "RUNS_ON")
    result.Custom = append(result.Custom, pod)

    return result
}
```

### Example 3: Processing Tool Output with Domain Types

```go
import "github.com/zero-day-ai/sdk/graphrag/domain"

// NmapOutput represents parsed nmap output
type NmapOutput struct {
    Host  string     `json:"host"`
    Ports []NmapPort `json:"ports"`
}

type NmapPort struct {
    Port     int    `json:"portid"`
    Protocol string `json:"protocol"`
    State    string `json:"state"`
    Service  string `json:"service"`
    Version  string `json:"version"`
}

func processNmapOutput(output NmapOutput) *domain.DiscoveryResult {
    result := domain.NewDiscoveryResult()

    // Create host
    result.Hosts = append(result.Hosts, &domain.Host{
        IP:    output.Host,
        State: "up",
    })

    // Create ports and services
    for _, p := range output.Ports {
        result.Ports = append(result.Ports, &domain.Port{
            HostID:   output.Host,
            Number:   p.Port,
            Protocol: p.Protocol,
            State:    p.State,
        })

        if p.Service != "" {
            result.Services = append(result.Services, &domain.Service{
                PortID:  fmt.Sprintf("%s:%d:%s", output.Host, p.Port, p.Protocol),
                Name:    p.Service,
                Version: p.Version,
            })
        }
    }

    return result
}
```

## Testing

### Testing Domain Types

```go
import (
    "testing"
    "github.com/zero-day-ai/sdk/graphrag/domain"
    "github.com/stretchr/testify/assert"
)

func TestHostDomainType(t *testing.T) {
    host := &domain.Host{
        IP:       "192.168.1.1",
        Hostname: "web-server",
        State:    "up",
    }

    // Test GraphNode interface implementation
    assert.Equal(t, "host", host.NodeType())
    assert.Equal(t, map[string]any{"ip": "192.168.1.1"}, host.IdentifyingProperties())
    assert.Nil(t, host.ParentRef(), "Host should have no parent")
}

func TestPortDomainType(t *testing.T) {
    port := &domain.Port{
        HostID:   "192.168.1.1",
        Number:   443,
        Protocol: "tcp",
        State:    "open",
    }

    // Test GraphNode interface
    assert.Equal(t, "port", port.NodeType())
    assert.Equal(t, map[string]any{
        "host_id":  "192.168.1.1",
        "number":   443,
        "protocol": "tcp",
    }, port.IdentifyingProperties())

    // Test parent relationship
    parent := port.ParentRef()
    assert.NotNil(t, parent)
    assert.Equal(t, "host", parent.NodeType)
    assert.Equal(t, "192.168.1.1", parent.Properties["ip"])
    assert.Equal(t, "HAS_PORT", port.RelationshipType())
}

func TestDiscoveryResult(t *testing.T) {
    result := domain.NewDiscoveryResult()

    result.Hosts = append(result.Hosts, &domain.Host{IP: "10.0.0.1"})
    result.Ports = append(result.Ports, &domain.Port{
        HostID:   "10.0.0.1",
        Number:   22,
        Protocol: "tcp",
    })

    // AllNodes returns in dependency order
    nodes := result.AllNodes()
    assert.Len(t, nodes, 2)
    assert.Equal(t, "host", nodes[0].NodeType())  // Parent first
    assert.Equal(t, "port", nodes[1].NodeType())  // Child second
}

func TestCustomEntity(t *testing.T) {
    entity := domain.NewCustomEntity("k8s", "pod").
        WithIDProps(map[string]any{"name": "web-pod"}).
        WithAllProps(map[string]any{
            "name":   "web-pod",
            "status": "Running",
        })

    assert.Equal(t, "k8s:pod", entity.NodeType())
    assert.Equal(t, map[string]any{"name": "web-pod"}, entity.IdentifyingProperties())
}
```

### Testing ID Generation

```go
func TestDeterministicIDs(t *testing.T) {
    registry := graphrag.Registry()
    generator := id.NewGenerator(registry)

    // Same inputs produce same ID
    id1, _ := generator.Generate("host", map[string]any{"ip": "10.0.0.1"})
    id2, _ := generator.Generate("host", map[string]any{"ip": "10.0.0.1"})
    assert.Equal(t, id1, id2)

    // Different inputs produce different IDs
    id3, _ := generator.Generate("host", map[string]any{"ip": "10.0.0.2"})
    assert.NotEqual(t, id1, id3)

    // Property order doesn't matter
    id4, _ := generator.Generate("port", map[string]any{
        "host_id": "h1", "number": 443, "protocol": "tcp",
    })
    id5, _ := generator.Generate("port", map[string]any{
        "protocol": "tcp", "number": 443, "host_id": "h1",
    })
    assert.Equal(t, id4, id5)
}
```

## Best Practices

1. **Use domain types** - Prefer `Host`, `Port`, `Service` over raw maps
2. **Use DiscoveryResult** - Container ensures proper dependency ordering
3. **Use CustomEntity for custom types** - Namespace with `k8s:`, `aws:`, etc.
4. **Test GraphNode methods** - Verify interface implementation
5. **Validate identifying properties** - All required properties must be set
6. **Check parent relationships** - Ensure parent nodes are created first

## See Also

- **`domain/README.md`** - Detailed domain types documentation
- **`domain/interfaces.go`** - GraphNode interface definition
- **`registry.go`** - NodeTypeRegistry implementation
- **`id/generator.go`** - Deterministic ID generator
- **`taxonomy_generated.go`** - Node and relationship type constants
