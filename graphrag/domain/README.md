# GraphRAG Domain Types

This package provides strongly-typed domain objects for the Gibson GraphRAG knowledge graph. It replaces the old taxonomy mapping system with compile-time type safety and a simpler, more extensible architecture.

## Overview

### Purpose

The `domain` package enables agents and tools to create knowledge graph nodes with:

- **Type safety**: Compile-time checking for node types and properties
- **JSON serialization**: All types have proper `json` tags for clean serialization/deserialization
- **Simplicity**: No manual taxonomy mapping or property construction
- **Extensibility**: CustomEntity for agent-specific types (e.g., "k8s:pod", "aws:security_group")
- **Hierarchical relationships**: Automatic parent-child relationship creation

### Benefits Over Old System

**Before (Old Taxonomy Mapping):**
```go
// Manual JSON construction, no type safety, error-prone
node := map[string]any{
    "type": "host",
    "properties": map[string]any{
        "ip": "192.168.1.1",
        "hostname": "web-server",
    },
}
// Need to manually create relationships, track parent IDs, etc.
```

**After (Domain Types):**
```go
// Strongly-typed, compiler-verified, automatic relationships
host := &domain.Host{
    IP:       "192.168.1.1",
    Hostname: "web-server",
    State:    "up",
}
// Parent relationships are automatic via ParentRef()
```

### JSON Serialization

All domain types have proper JSON tags for consistent serialization. This enables:
- Clean JSON output from tools with lowercase snake_case keys
- Automatic deserialization by Gibson's callback service
- Interoperability with external systems

```go
// Domain types serialize to clean JSON:
host := &domain.Host{IP: "192.168.1.1", Hostname: "web-server"}
// Serializes to: {"ip": "192.168.1.1", "hostname": "web-server"}

port := &domain.Port{HostID: "192.168.1.1", Number: 80, Protocol: "tcp"}
// Serializes to: {"host_id": "192.168.1.1", "number": 80, "protocol": "tcp"}
```

## Tool Output Format

**IMPORTANT**: Tools must return a `DiscoveryResult` under the `discovery_result` key in their output:

```go
func executeTool(ctx context.Context, input map[string]any) (map[string]any, error) {
    result := domain.NewDiscoveryResult()

    // Populate discovery result...
    result.Hosts = append(result.Hosts, &domain.Host{
        IP:    "192.168.1.1",
        State: "up",
    })

    // Return with discovery_result key - Gibson extracts this automatically
    return map[string]any{
        "discovery_result": result,
        "metadata": map[string]any{
            "scan_time": "5s",
        },
    }, nil
}
```

Gibson's callback service automatically:
1. Extracts the `discovery_result` key from tool output
2. Deserializes it into a `domain.DiscoveryResult` struct
3. Loads all nodes and relationships into the Neo4j knowledge graph

## GraphNode Interface

All domain types implement the `GraphNode` interface:

```go
type GraphNode interface {
    // NodeType returns the canonical node type from the taxonomy
    // (e.g., "host", "port", "service")
    NodeType() string

    // IdentifyingProperties returns properties that uniquely identify this node
    // Used for deduplication and node lookups
    IdentifyingProperties() map[string]any

    // Properties returns all properties to set on the node
    // Includes both identifying and descriptive properties
    Properties() map[string]any

    // ParentRef returns reference to parent node for hierarchical relationships
    // Returns nil for root nodes
    ParentRef() *NodeRef

    // RelationshipType returns the relationship type to parent node
    // Returns empty string for root nodes
    RelationshipType() string
}
```

### When to Implement vs Use Existing Types

**Use existing canonical types** (Host, Port, Service, etc.) for standard network reconnaissance and web testing.

**Implement GraphNode** only when creating custom agent-specific types. For most cases, use `CustomEntity` instead of implementing the interface directly.

## Canonical Domain Types

### Assets: Network Infrastructure

#### Host
Root-level node representing a network host (IP address).

**Identifying Properties:** `ip`

**Parent:** None (root node)

**JSON Fields:** `ip`, `hostname`, `state`, `os`

```go
host := &domain.Host{
    IP:       "192.168.1.10",
    Hostname: "web-server.example.com",
    State:    "up",
    OS:       "Linux Ubuntu 22.04",
}
// JSON: {"ip":"192.168.1.10","hostname":"web-server.example.com","state":"up","os":"Linux Ubuntu 22.04"}
```

#### Port
Network port on a host.

**Identifying Properties:** `host_id`, `number`, `protocol`

**Parent:** Host (via `HAS_PORT` relationship)

**JSON Fields:** `host_id`, `number`, `protocol`, `state`

```go
port := &domain.Port{
    HostID:   "192.168.1.10",
    Number:   443,
    Protocol: "tcp",
    State:    "open",
}
// JSON: {"host_id":"192.168.1.10","number":443,"protocol":"tcp","state":"open"}
```

#### Service
Service running on a port.

**Identifying Properties:** `port_id`, `name`

**Parent:** Port (via `RUNS_SERVICE` relationship)

**JSON Fields:** `port_id`, `name`, `version`, `banner`

```go
service := &domain.Service{
    PortID:  "192.168.1.10:443:tcp",  // Format: {host_id}:{number}:{protocol}
    Name:    "https",
    Version: "nginx/1.18.0",
    Banner:  "nginx/1.18.0 (Ubuntu)",
}
// JSON: {"port_id":"192.168.1.10:443:tcp","name":"https","version":"nginx/1.18.0","banner":"nginx/1.18.0 (Ubuntu)"}
```

#### Endpoint
HTTP/HTTPS endpoint on a service.

**Identifying Properties:** `service_id`, `url`, `method`

**Parent:** Service (via `HAS_ENDPOINT` relationship)

**JSON Fields:** `service_id`, `url`, `method`, `status_code`, `headers`, `response_time`, `content_type`, `content_length`

```go
endpoint := &domain.Endpoint{
    ServiceID:     "192.168.1.10:443:tcp:https",
    URL:           "/api/users",
    Method:        "GET",
    StatusCode:    200,
    ContentType:   "application/json",
    ResponseTime:  45,
}
```

### Assets: DNS

#### Domain
Root domain name.

**Identifying Properties:** `name`

**Parent:** None (root node)

**JSON Fields:** `name`, `registrar`, `created_at`, `expires_at`, `nameservers`, `status`

```go
d := &domain.Domain{
    Name:      "example.com",
    Registrar: "Cloudflare",
    CreatedAt: "2010-01-15",
    ExpiresAt: "2025-01-15",
    Status:    "active",
}
```

#### Subdomain
Subdomain under a root domain.

**Identifying Properties:** `parent_domain`, `name`

**Parent:** Domain (via `HAS_SUBDOMAIN` relationship)

**JSON Fields:** `parent_domain`, `name`, `record_type`, `record_value`, `ttl`, `status`

```go
subdomain := &domain.Subdomain{
    ParentDomain: "example.com",
    Name:         "api.example.com",
    RecordType:   "A",
    RecordValue:  "192.168.1.10",
    TTL:          300,
    Status:       "active",
}
```

### Assets: Technology Stack

#### Technology
Technology, framework, or software detected.

**Identifying Properties:** `name`, `version`

**Parent:** None (root node)

**JSON Fields:** `name`, `version`, `category`, `vendor`, `cpe`, `license`, `eol`

```go
tech := &domain.Technology{
    Name:     "nginx",
    Version:  "1.18.0",
    Category: "web-server",
    Vendor:   "Nginx Inc.",
    CPE:      "cpe:2.3:a:nginx:nginx:1.18.0:*:*:*:*:*:*:*",
}
```

#### Certificate
TLS/SSL certificate.

**Identifying Properties:** `serial_number`

**Parent:** None (root node)

```go
cert := &domain.Certificate{
    SerialNumber: "03:5d:e2:52:9a:b8:47:d9",
    Subject:      "CN=example.com",
    Issuer:       "C=US, O=Let's Encrypt",
    NotBefore:    "2024-01-01T00:00:00Z",
    NotAfter:     "2024-04-01T00:00:00Z",
}
```

### Assets: Cloud Infrastructure

#### CloudAsset
Cloud infrastructure resource (AWS, Azure, GCP).

**Identifying Properties:** `provider`, `region`, `resource_id`

**Parent:** None (root node)

```go
cloud := &domain.CloudAsset{
    Provider:     "aws",
    Region:       "us-east-1",
    ResourceID:   "i-0123456789abcdef0",
    ResourceType: "ec2-instance",
    Name:         "web-server-01",
    Status:       "running",
}
```

### Assets: Web APIs

#### API
Web API service with multiple endpoints.

**Identifying Properties:** `base_url`

**Parent:** None (root node)

```go
api := &domain.API{
    BaseURL:     "https://api.example.com",
    Name:        "Example API",
    Version:     "v1",
    Description: "RESTful API for example service",
    AuthType:    "bearer",
    SwaggerURL:  "https://api.example.com/swagger.json",
}
```

## DiscoveryResult Container

The `DiscoveryResult` struct is the standard container for tool and agent output. It contains over 150 typed slices covering all domain categories.

### Core Categories

| Category | Types | Examples |
|----------|-------|----------|
| Network Infrastructure | 10 | Host, Port, Service, Endpoint |
| DNS | 3 | Domain, Subdomain, DNSRecord |
| Technology | 2 | Technology, Certificate |
| Cloud | 20 | CloudAsset, CloudVPC, CloudInstance |
| Kubernetes | 22 | K8sPod, K8sDeployment, K8sService |
| Identity/Access | 16 | User, Role, Credential, APIKey |
| AI/LLM | 16 | LLM, Prompt, Guardrail |
| MCP | 9 | MCPServer, MCPTool, MCPResource |
| RAG | 11 | VectorStore, Document, Retriever |
| Data | 16 | Database, Table, Queue |
| Web/API | 16 | APIEndpoint, Parameter, Form |
| Attack | 2 | Tactic, Technique |

### Using DiscoveryResult in Tools

```go
import (
    "github.com/zero-day-ai/sdk/graphrag/domain"
)

func executePortScan(ctx context.Context, input map[string]any) (map[string]any, error) {
    target := input["target"].(string)
    result := domain.NewDiscoveryResult()

    // Add discovered host
    result.Hosts = append(result.Hosts, &domain.Host{
        IP:       target,
        Hostname: "web-server",
        State:    "up",
    })

    // Add discovered ports
    result.Ports = append(result.Ports, &domain.Port{
        HostID:   target,
        Number:   80,
        Protocol: "tcp",
        State:    "open",
    })

    result.Ports = append(result.Ports, &domain.Port{
        HostID:   target,
        Number:   443,
        Protocol: "tcp",
        State:    "open",
    })

    // Add services
    result.Services = append(result.Services, &domain.Service{
        PortID:  target + ":80:tcp",
        Name:    "http",
        Version: "nginx/1.18.0",
    })

    // Return with discovery_result key
    return map[string]any{
        "discovery_result": result,
    }, nil
}
```

### AllNodes() Method

The `AllNodes()` method returns all discovered nodes as a flattened slice in dependency order:

```go
result := &domain.DiscoveryResult{
    Hosts: []*domain.Host{...},
    Ports: []*domain.Port{...},
    Services: []*domain.Service{...},
}

// Get all nodes in dependency order
nodes := result.AllNodes()
// Returns: [Host, Port, Port, Service, Service]
//          ^---- parents first, children after
```

### Node Ordering (Dependency Order)

`AllNodes()` returns nodes in this order to ensure parents exist before children:

1. **Cloud foundation**: Accounts, regions, VPCs, subnets, security groups
2. **Network infrastructure**: Networks, zones, firewalls, routers
3. **K8s foundation**: Clusters, namespaces
4. **Compute resources**: Hosts, cloud instances, containers
5. **K8s workloads**: Pods, deployments, services
6. **Network details**: Interfaces, DNS, load balancers, ports
7. **Services and domains**: Services, domains, subdomains, technologies
8. **Web/API resources**: APIs, endpoints, parameters, forms
9. **AI/LLM infrastructure**: Models, deployments, embeddings
10. **AI agents and workflows**: Agents, chains, crews
11. **MCP servers and resources**
12. **RAG systems**: Vector stores, documents, pipelines
13. **Data resources**: Databases, tables, queues
14. **Identity and access**: Users, groups, roles, credentials
15. **Security findings and attack techniques**
16. **Custom nodes** (order preserved as added)

This ordering ensures the GraphRAG system can create nodes and relationships in a single pass without forward references.

## CustomEntity for Extensibility

Use `CustomEntity` when your agent needs domain-specific types not covered by canonical types.

### When to Use CustomEntity

- Kubernetes resources (pods, services, deployments)
- AWS resources beyond generic CloudAsset (security groups, IAM roles, Lambda functions)
- Application-specific entities (databases, message queues, caches)
- Security-specific entities (vulnerabilities, exploits, credentials)

### Namespace Conventions

Use namespace prefixes to organize custom types:

| Namespace | Purpose | Examples |
|-----------|---------|----------|
| `k8s:` | Kubernetes resources | `k8s:pod`, `k8s:service`, `k8s:deployment` |
| `aws:` | AWS-specific resources | `aws:security_group`, `aws:iam_role`, `aws:lambda` |
| `azure:` | Azure-specific resources | `azure:resource_group`, `azure:storage_account` |
| `gcp:` | GCP-specific resources | `gcp:compute_instance`, `gcp:storage_bucket` |
| `vuln:` | Vulnerability types | `vuln:cve`, `vuln:exploit`, `vuln:weakness` |
| `cred:` | Credential types | `cred:api_key`, `cred:ssh_key`, `cred:password` |
| `custom:` | Agent-specific types | `custom:my_agent_type` |

### Full Example: Kubernetes Agent

```go
package main

import (
    "github.com/zero-day-ai/sdk/graphrag/domain"
)

// Discover Kubernetes pods and create custom entities
func discoverKubernetesPods() *domain.DiscoveryResult {
    result := domain.NewDiscoveryResult()

    // Create a Kubernetes node entity (root)
    node := domain.NewCustomEntity("k8s", "node").
        WithIDProps(map[string]any{
            "name": "worker-node-01",
        }).
        WithAllProps(map[string]any{
            "name":        "worker-node-01",
            "status":      "Ready",
            "version":     "v1.28.0",
            "capacity":    map[string]any{"cpu": "4", "memory": "16Gi"},
            "os":          "Ubuntu 22.04",
        })

    result.Custom = append(result.Custom, node)

    // Create a pod entity with parent relationship to node
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
            "phase":     "Running",
            "restarts":  0,
        }).
        WithParent(&domain.NodeRef{
            NodeType: "k8s:node",
            Properties: map[string]any{
                "name": "worker-node-01",
            },
        }, "RUNS_ON")

    result.Custom = append(result.Custom, pod)

    return result
}
```

### How Custom Types Coexist with Canonical Types

Custom types and canonical types work seamlessly together:

```go
result := domain.NewDiscoveryResult()

// Canonical types: Standard network assets
result.Hosts = append(result.Hosts, &domain.Host{
    IP:       "10.0.1.50",
    Hostname: "k8s-node-01",
    State:    "up",
})

result.Ports = append(result.Ports, &domain.Port{
    HostID:   "10.0.1.50",
    Number:   6443,
    Protocol: "tcp",
    State:    "open",
})

// Custom types: Kubernetes-specific entities
result.Custom = append(result.Custom, domain.NewCustomEntity("k8s", "node").
    WithIDProps(map[string]any{
        "name": "k8s-node-01",
    }).
    WithAllProps(map[string]any{
        "name":    "k8s-node-01",
        "ip":      "10.0.1.50",
        "version": "v1.28.0",
    }))

// The graph will contain both canonical and custom nodes:
// - Host "10.0.1.50" (canonical)
// - Port "10.0.1.50:6443:tcp" (canonical)
// - k8s:node "k8s-node-01" (custom)
```

## Complete Example: Port Scanner Tool

```go
package main

import (
    "context"
    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/graphrag/domain"
)

func main() {
    portScanner, err := sdk.NewTool(
        sdk.WithToolName("port-scanner"),
        sdk.WithToolVersion("1.0.0"),
        sdk.WithToolDescription("Scans ports on a target host"),
        sdk.WithExecuteHandler(executePortScan),
    )
    if err != nil {
        panic(err)
    }
    _ = portScanner
}

func executePortScan(ctx context.Context, input map[string]any) (map[string]any, error) {
    target := input["target"].(string)

    // Create discovery result container
    result := domain.NewDiscoveryResult()

    // Discover host
    result.Hosts = append(result.Hosts, &domain.Host{
        IP:       target,
        Hostname: "web-server.example.com",
        State:    "up",
        OS:       "Linux",
    })

    // Discover open ports
    openPorts := []int{22, 80, 443, 3306}
    for _, portNum := range openPorts {
        result.Ports = append(result.Ports, &domain.Port{
            HostID:   target,
            Number:   portNum,
            Protocol: "tcp",
            State:    "open",
        })
    }

    // Discover services on ports
    result.Services = append(result.Services, &domain.Service{
        PortID:  target + ":80:tcp",
        Name:    "http",
        Version: "nginx/1.18.0",
        Banner:  "nginx/1.18.0 (Ubuntu)",
    })

    result.Services = append(result.Services, &domain.Service{
        PortID:  target + ":443:tcp",
        Name:    "https",
        Version: "nginx/1.18.0",
    })

    result.Services = append(result.Services, &domain.Service{
        PortID:  target + ":3306:tcp",
        Name:    "mysql",
        Version: "8.0.35",
    })

    // IMPORTANT: Return discovery_result key for Gibson to process
    return map[string]any{
        "discovery_result": result,
        "metadata": map[string]any{
            "scan_duration": "2.5s",
            "ports_scanned": 1000,
        },
    }, nil
}
```

## See Also

- `../taxonomy_generated.go` - Generated taxonomy constants (NodeType*, RelType*, Prop*)
- `../../agent/harness.go` - Agent harness interface for storing nodes
- `../loader.go` - GraphRAG loader that processes DiscoveryResult
- `../README.md` - GraphRAG system overview
