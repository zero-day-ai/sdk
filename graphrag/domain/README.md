# GraphRAG Domain Types

This package provides strongly-typed domain objects for the Gibson GraphRAG knowledge graph. It replaces the old taxonomy mapping system with compile-time type safety and a simpler, more extensible architecture.

## Overview

### Purpose

The `domain` package enables agents and tools to create knowledge graph nodes with:

- **Type safety**: Compile-time checking for node types and properties
- **Simplicity**: No manual taxonomy mapping or JSON serialization
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

### Example Interface Implementation

```go
// Custom type for a Kubernetes Pod
type KubernetesPod struct {
    Namespace string
    Name      string
    Status    string
    Image     string
    NodeName  string
}

func (k *KubernetesPod) NodeType() string {
    return "k8s:pod"
}

func (k *KubernetesPod) IdentifyingProperties() map[string]any {
    return map[string]any{
        "namespace": k.Namespace,
        "name":      k.Name,
    }
}

func (k *KubernetesPod) Properties() map[string]any {
    return map[string]any{
        "namespace": k.Namespace,
        "name":      k.Name,
        "status":    k.Status,
        "image":     k.Image,
        "node_name": k.NodeName,
    }
}

func (k *KubernetesPod) ParentRef() *NodeRef {
    // Pods could reference a k8s:node parent
    return &NodeRef{
        NodeType: "k8s:node",
        Properties: map[string]any{
            "name": k.NodeName,
        },
    }
}

func (k *KubernetesPod) RelationshipType() string {
    return "RUNS_ON"
}
```

## Canonical Domain Types

### Assets: Network Infrastructure

#### Host
Root-level node representing a network host (IP address).

**Identifying Properties:** `ip`

**Parent:** None (root node)

```go
host := &domain.Host{
    IP:       "192.168.1.10",
    Hostname: "web-server.example.com",
    State:    "up",
    OS:       "Linux Ubuntu 22.04",
}
```

#### Port
Network port on a host.

**Identifying Properties:** `host_id`, `number`, `protocol`

**Parent:** Host (via `HAS_PORT` relationship)

```go
port := &domain.Port{
    HostID:   "192.168.1.10",
    Number:   443,
    Protocol: "tcp",
    State:    "open",
}
```

#### Service
Service running on a port.

**Identifying Properties:** `port_id`, `name`

**Parent:** Port (via `RUNS_SERVICE` relationship)

```go
service := &domain.Service{
    PortID:  "192.168.1.10:443:tcp",  // Format: {host_id}:{number}:{protocol}
    Name:    "https",
    Version: "nginx/1.18.0",
    Banner:  "nginx/1.18.0 (Ubuntu)",
}
```

#### Endpoint
HTTP/HTTPS endpoint on a service.

**Identifying Properties:** `service_id`, `url`, `method`

**Parent:** Service (via `HAS_ENDPOINT` relationship)

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

```go
domain := &domain.Domain{
    Name:         "example.com",
    Registrar:    "Cloudflare",
    RegisteredAt: "2010-01-15",
}
```

#### Subdomain
Subdomain under a root domain.

**Identifying Properties:** `name`

**Parent:** Domain (via `HAS_SUBDOMAIN` relationship)

```go
subdomain := &domain.Subdomain{
    Name:       "api.example.com",
    DomainName: "example.com",
}
```

### Assets: Technology Stack

#### Technology
Technology, framework, or software detected.

**Identifying Properties:** `name`, `version`

**Parent:** None (root node)

```go
tech := &domain.Technology{
    Name:     "nginx",
    Version:  "1.18.0",
    Category: "web-server",
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

### Execution: Agent and Tool Tracking

#### AgentRun
Single execution run of an agent within a mission.

**Identifying Properties:** `id`, `mission_id`, `agent_name`, `run_number`

**Parent:** None (root node in domain hierarchy, but linked to missions via PART_OF_MISSION)

```go
agentRun := &domain.AgentRun{
    ID:        "run-123",
    MissionID: "mission-456",
    AgentName: "network-recon",
    RunNumber: 1,
    StartTime: "2024-01-20T10:00:00Z",
    Status:    "running",
}
```

#### ToolExecution
Single execution of a tool within an agent run.

**Identifying Properties:** `id`, `agent_run_id`, `tool_name`, `sequence`

**Parent:** AgentRun (via `EXECUTED` relationship)

```go
toolExec := &domain.ToolExecution{
    ID:          "exec-789",
    AgentRunID:  "run-123",
    ToolName:    "nmap",
    Sequence:    1,
    StartTime:   "2024-01-20T10:05:00Z",
    EndTime:     "2024-01-20T10:10:00Z",
    Status:      "success",
    Duration:    300.5,
}
```

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

    // Create a service entity (also custom)
    service := domain.NewCustomEntity("k8s", "service").
        WithIDProps(map[string]any{
            "namespace": "default",
            "name":      "web-service",
        }).
        WithAllProps(map[string]any{
            "namespace":   "default",
            "name":        "web-service",
            "type":        "LoadBalancer",
            "cluster_ip":  "10.96.0.1",
            "external_ip": "203.0.113.5",
            "ports":       []int{80, 443},
        })

    result.Custom = append(result.Custom, service)

    return result
}
```

### Example: AWS Security Group

```go
// Create an AWS security group with parent VPC
securityGroup := domain.NewCustomEntity("aws", "security_group").
    WithIDProps(map[string]any{
        "id": "sg-0123456789abcdef0",
    }).
    WithAllProps(map[string]any{
        "id":          "sg-0123456789abcdef0",
        "name":        "web-server-sg",
        "description": "Security group for web servers",
        "vpc_id":      "vpc-abc123",
        "rules": map[string]any{
            "ingress": []map[string]any{
                {"port": 80, "protocol": "tcp", "cidr": "0.0.0.0/0"},
                {"port": 443, "protocol": "tcp", "cidr": "0.0.0.0/0"},
            },
        },
    }).
    WithParent(&domain.NodeRef{
        NodeType: "aws:vpc",
        Properties: map[string]any{
            "id": "vpc-abc123",
        },
    }, "BELONGS_TO")

result := domain.NewDiscoveryResult()
result.Custom = append(result.Custom, securityGroup)
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

## DiscoveryResult Container

The `DiscoveryResult` struct is the standard container for tool and agent output.

### Structure

```go
type DiscoveryResult struct {
    Hosts         []*Host
    Ports         []*Port
    Services      []*Service
    Endpoints     []*Endpoint
    Domains       []*Domain
    Subdomains    []*Subdomain
    Technologies  []*Technology
    Certificates  []*Certificate
    CloudAssets   []*CloudAsset
    APIs          []*API
    Custom        []GraphNode  // For CustomEntity and custom implementations
}
```

### Using DiscoveryResult in Tools

```go
import (
    "github.com/zero-day-ai/sdk/graphrag/domain"
)

func executePortScan(target string) (*domain.DiscoveryResult, error) {
    result := domain.NewDiscoveryResult()

    // Run nmap scan...

    // Add discovered host
    result.Hosts = append(result.Hosts, &domain.Host{
        IP:       "192.168.1.10",
        Hostname: "web-server",
        State:    "up",
    })

    // Add discovered ports
    result.Ports = append(result.Ports, &domain.Port{
        HostID:   "192.168.1.10",
        Number:   80,
        Protocol: "tcp",
        State:    "open",
    })

    result.Ports = append(result.Ports, &domain.Port{
        HostID:   "192.168.1.10",
        Number:   443,
        Protocol: "tcp",
        State:    "open",
    })

    // Add services
    result.Services = append(result.Services, &domain.Service{
        PortID:  "192.168.1.10:80:tcp",
        Name:    "http",
        Version: "nginx/1.18.0",
    })

    result.Services = append(result.Services, &domain.Service{
        PortID:  "192.168.1.10:443:tcp",
        Name:    "https",
        Version: "nginx/1.18.0",
    })

    return result, nil
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

1. **Root nodes** (no parent dependencies):
   - Hosts
   - Domains
   - Technologies
   - Certificates
   - CloudAssets
   - APIs

2. **First-level dependent nodes**:
   - Ports (depend on Hosts)
   - Subdomains (depend on Domains)

3. **Second-level dependent nodes**:
   - Services (depend on Ports)

4. **Leaf nodes**:
   - Endpoints (depend on Services)

5. **Custom nodes** (order preserved as added)

This ordering ensures the GraphRAG system can create nodes and relationships in a single pass without forward references.

## Migration Guide

### Before: Old Taxonomy Mapping Approach

```go
// Old approach: Manual JSON construction with taxonomy mapping
func createHost(ip, hostname string) map[string]any {
    return map[string]any{
        "type": "host",  // Manually look up taxonomy type
        "properties": map[string]any{
            "ip":       ip,
            "hostname": hostname,
        },
    }
}

func createPort(hostID string, number int, protocol string) map[string]any {
    return map[string]any{
        "type": "port",
        "properties": map[string]any{
            "host_id":  hostID,
            "number":   number,
            "protocol": protocol,
        },
        // Manually construct parent relationship
        "parent": map[string]any{
            "type": "host",
            "properties": map[string]any{"ip": hostID},
        },
        "relationship_type": "HAS_PORT",
    }
}

// Manual relationship creation, error-prone
nodes := []map[string]any{
    createHost("192.168.1.1", "web-server"),
    createPort("192.168.1.1", 80, "tcp"),
}
```

### After: New Domain Type Approach

```go
import "github.com/zero-day-ai/sdk/graphrag/domain"

// New approach: Type-safe domain objects
result := domain.NewDiscoveryResult()

// Strongly-typed host creation
result.Hosts = append(result.Hosts, &domain.Host{
    IP:       "192.168.1.1",
    Hostname: "web-server",
    State:    "up",
})

// Strongly-typed port creation with automatic parent relationship
result.Ports = append(result.Ports, &domain.Port{
    HostID:   "192.168.1.1",  // Automatic parent lookup
    Number:   80,
    Protocol: "tcp",
    State:    "open",
})

// Relationships are automatic via ParentRef() and RelationshipType()
// No manual JSON construction needed
```

### Step-by-Step Migration for Existing Tools

#### Step 1: Replace Manual JSON with DiscoveryResult

**Before:**
```go
func executeTool(input map[string]any) (map[string]any, error) {
    nodes := []map[string]any{}

    // ... scan logic ...

    nodes = append(nodes, map[string]any{
        "type": "host",
        "properties": map[string]any{
            "ip": "192.168.1.1",
        },
    })

    return map[string]any{"nodes": nodes}, nil
}
```

**After:**
```go
import "github.com/zero-day-ai/sdk/graphrag/domain"

func executeTool(input map[string]any) (*domain.DiscoveryResult, error) {
    result := domain.NewDiscoveryResult()

    // ... scan logic ...

    result.Hosts = append(result.Hosts, &domain.Host{
        IP:    "192.168.1.1",
        State: "up",
    })

    return result, nil
}
```

#### Step 2: Replace Type Strings with Domain Types

**Before:**
```go
// Hard-coded type strings, no validation
nodeType := "host"
if someCondition {
    nodeType = "service"
}
```

**After:**
```go
// Type-safe domain types
if someCondition {
    result.Hosts = append(result.Hosts, host)
} else {
    result.Services = append(result.Services, service)
}
```

#### Step 3: Replace Manual Relationships with ParentRef

**Before:**
```go
port := map[string]any{
    "type": "port",
    "properties": map[string]any{
        "host_id": hostID,
        "number":  80,
    },
    "parent": map[string]any{
        "type": "host",
        "properties": map[string]any{"ip": hostID},
    },
    "relationship_type": "HAS_PORT",
}
```

**After:**
```go
// ParentRef() is automatic - just set HostID
port := &domain.Port{
    HostID:   hostID,  // Parent relationship is automatic
    Number:   80,
    Protocol: "tcp",
}
result.Ports = append(result.Ports, port)
```

#### Step 4: Migrate Custom Types to CustomEntity

**Before:**
```go
// Custom type with manual JSON
customNode := map[string]any{
    "type": "k8s:pod",
    "properties": map[string]any{
        "namespace": "default",
        "name":      "web-server",
    },
}
```

**After:**
```go
// CustomEntity with fluent builder
pod := domain.NewCustomEntity("k8s", "pod").
    WithIDProps(map[string]any{
        "namespace": "default",
        "name":      "web-server",
    }).
    WithAllProps(map[string]any{
        "namespace": "default",
        "name":      "web-server",
        "status":    "Running",
    })

result.Custom = append(result.Custom, pod)
```

## Complete Examples

### Example 1: Tool that Discovers Hosts and Ports

```go
package main

import (
    "context"
    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/graphrag/domain"
    "github.com/zero-day-ai/sdk/tool"
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

    // Return the discovery result
    // The GraphRAG system will automatically create nodes and relationships
    return map[string]any{
        "discovery": result,
    }, nil
}
```

### Example 2: Agent with Custom Type

```go
package main

import (
    "context"
    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/graphrag/domain"
)

func main() {
    k8sAgent, err := sdk.NewAgent(
        sdk.WithName("k8s-scanner"),
        sdk.WithVersion("1.0.0"),
        sdk.WithDescription("Scans Kubernetes clusters for security issues"),
        sdk.WithExecuteFunc(executeK8sScan),
    )
    if err != nil {
        panic(err)
    }
    _ = k8sAgent
}

func executeK8sScan(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    result := domain.NewDiscoveryResult()

    // Discover Kubernetes nodes (worker machines)
    nodeNames := []string{"worker-01", "worker-02", "worker-03"}
    for _, nodeName := range nodeNames {
        node := domain.NewCustomEntity("k8s", "node").
            WithIDProps(map[string]any{
                "name": nodeName,
            }).
            WithAllProps(map[string]any{
                "name":     nodeName,
                "status":   "Ready",
                "version":  "v1.28.0",
                "capacity": map[string]any{"cpu": "8", "memory": "32Gi"},
            })

        result.Custom = append(result.Custom, node)
    }

    // Discover pods running on nodes
    pods := []struct {
        namespace string
        name      string
        nodeName  string
        image     string
    }{
        {"default", "web-server-abc123", "worker-01", "nginx:1.21"},
        {"default", "api-server-def456", "worker-02", "golang:1.21"},
        {"kube-system", "coredns-ghi789", "worker-03", "coredns:1.10"},
    }

    for _, podInfo := range pods {
        pod := domain.NewCustomEntity("k8s", "pod").
            WithIDProps(map[string]any{
                "namespace": podInfo.namespace,
                "name":      podInfo.name,
            }).
            WithAllProps(map[string]any{
                "namespace": podInfo.namespace,
                "name":      podInfo.name,
                "status":    "Running",
                "image":     podInfo.image,
                "phase":     "Running",
            }).
            WithParent(&domain.NodeRef{
                NodeType: "k8s:node",
                Properties: map[string]any{
                    "name": podInfo.nodeName,
                },
            }, "RUNS_ON")

        result.Custom = append(result.Custom, pod)
    }

    // Discover Kubernetes services
    service := domain.NewCustomEntity("k8s", "service").
        WithIDProps(map[string]any{
            "namespace": "default",
            "name":      "web-service",
        }).
        WithAllProps(map[string]any{
            "namespace":   "default",
            "name":        "web-service",
            "type":        "LoadBalancer",
            "cluster_ip":  "10.96.0.1",
            "external_ip": "203.0.113.5",
            "ports":       []int{80, 443},
        })

    result.Custom = append(result.Custom, service)

    // Store discoveries in GraphRAG
    // (In real implementation, this would be done via harness)

    return agent.Result{
        Status: agent.StatusSuccess,
        Output: map[string]any{
            "nodes_discovered": result.NodeCount(),
            "pods":             len(pods),
            "nodes":            len(nodeNames),
        },
    }, nil
}
```

## See Also

- `../taxonomy_generated.go` - Generated taxonomy constants (NodeType*, RelType*, Prop*)
- `../../agent/harness.go` - Agent harness interface for storing nodes
- `../loader.go` - GraphRAG loader that processes DiscoveryResult
- `../README.md` - GraphRAG system overview
