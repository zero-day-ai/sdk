# Gibson Tool Development Guide

Complete reference for building security tools that integrate with the Gibson framework.

## Table of Contents

1. [Overview](#overview)
2. [Tool Interface](#tool-interface)
3. [Protocol Buffers](#protocol-buffers)
4. [Building a Tool](#building-a-tool)
5. [Automatic Graph Storage](#automatic-graph-storage)
6. [Health Checks](#health-checks)
7. [Error Handling](#error-handling)
8. [MITRE ATT&CK Mappings](#mitre-attck-mappings)
9. [Testing Tools](#testing-tools)
10. [Serving Tools](#serving-tools)
11. [Complete Examples](#complete-examples)
12. [Best Practices](#best-practices)

---

## Overview

Gibson tools are **atomic, stateless operations** that wrap security utilities and provide structured, LLM-consumable I/O. Key characteristics:

- **Protocol Buffer I/O** - Type-safe, schema-validated input/output
- **Stateless** - No persistent state between executions
- **Health Monitoring** - Report dependency status
- **MITRE Mappings** - Technique/tactic categorization
- **gRPC Distribution** - Serve tools over the network

```
┌─────────────────────────────────────────────────────────────────┐
│                          AGENT                                   │
│                             │                                    │
│                             ▼                                    │
│              harness.CallToolProto(ctx, "nmap", req, resp)      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       TOOL REGISTRY                              │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │  nmap   │ │  httpx  │ │ nuclei  │ │ sslyze  │ │   ...   │   │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        YOUR TOOL                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              ExecuteProto(ctx, input) output             │    │
│  │                          │                               │    │
│  │    ┌───────────────────────────────────────────────┐    │    │
│  │    │  Proto Request → Execute Binary → Proto Response │    │    │
│  │    └───────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### Tools vs Plugins

| Aspect | Tools | Plugins |
|--------|-------|---------|
| **Purpose** | Atomic operations (scan, probe, analyze) | Stateful services (APIs, databases) |
| **I/O** | Protocol Buffers (type-safe) | JSON maps (flexible) |
| **State** | Stateless | Maintains state |
| **Lifecycle** | No init/shutdown | Initialize/Shutdown |
| **Examples** | nmap, httpx, nuclei, sslyze | Shodan API, scope parser, vector DB |

---

## Tool Interface

Every tool must implement the `Tool` interface:

```go
type Tool interface {
    // Identity & Metadata
    Name() string                              // Unique identifier (e.g., "nmap")
    Version() string                           // Semantic version (e.g., "1.0.0")
    Description() string                       // Human-readable description
    Tags() []string                            // Categorization (e.g., "discovery", "network")

    // Proto Message Types
    InputMessageType() string                  // Full proto message name (e.g., "gibson.tools.nmap.ScanRequest")
    OutputMessageType() string                 // Full proto message name (e.g., "gibson.tools.nmap.ScanResponse")

    // Execution
    ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error)

    // Health
    Health(ctx context.Context) types.HealthStatus
}
```

### Tool Descriptor

Tools expose metadata through descriptors:

```go
type Descriptor struct {
    Name              string   // Tool identifier
    Version           string   // Semantic version
    Description       string   // What the tool does
    Tags              []string // Categories for filtering
    InputMessageType  string   // Proto input type
    OutputMessageType string   // Proto output type
}
```

---

## Protocol Buffers

Tools use Protocol Buffers for type-safe I/O. This ensures:

- **Schema Validation** - Invalid inputs rejected automatically
- **Type Safety** - No runtime type errors
- **Language Agnostic** - Works across Go, Python, etc.
- **LLM Friendly** - Clear schemas for AI understanding
- **Versioning** - Proto3 supports backward compatibility

### Defining Proto Messages

Create `proto/tool.proto`:

```protobuf
syntax = "proto3";

package gibson.tools.nmap;

option go_package = "github.com/zero-day-ai/tools/nmap/proto";

// Input message
message ScanRequest {
    // Target specification
    string target = 1;                    // IP, hostname, CIDR, or range
    repeated string targets = 2;          // Multiple targets

    // Port specification
    string ports = 3;                     // Port range (e.g., "22,80,443" or "1-1000")
    bool top_ports = 4;                   // Use top 1000 ports
    int32 top_ports_count = 5;            // Custom top N ports

    // Scan type
    ScanType scan_type = 6;

    // Timing
    TimingTemplate timing = 7;
    int32 timeout_seconds = 8;            // Per-host timeout
    int32 max_retries = 9;

    // Output options
    bool service_detection = 10;          // -sV
    bool os_detection = 11;               // -O
    bool script_scan = 12;                // -sC
    repeated string scripts = 13;         // Specific NSE scripts

    // Advanced
    repeated string extra_flags = 14;     // Raw nmap flags
}

enum ScanType {
    SCAN_TYPE_UNSPECIFIED = 0;
    SCAN_TYPE_SYN = 1;          // -sS (default, requires root)
    SCAN_TYPE_CONNECT = 2;      // -sT (TCP connect)
    SCAN_TYPE_UDP = 3;          // -sU
    SCAN_TYPE_ACK = 4;          // -sA
    SCAN_TYPE_PING = 5;         // -sn (host discovery only)
}

enum TimingTemplate {
    TIMING_UNSPECIFIED = 0;
    TIMING_PARANOID = 1;        // -T0
    TIMING_SNEAKY = 2;          // -T1
    TIMING_POLITE = 3;          // -T2
    TIMING_NORMAL = 4;          // -T3 (default)
    TIMING_AGGRESSIVE = 5;      // -T4
    TIMING_INSANE = 6;          // -T5
}

// Output message
message ScanResponse {
    // Results
    repeated Host hosts = 1;
    ScanStats stats = 2;

    // Raw output
    string raw_output = 3;
    string xml_output = 4;

    // Errors
    repeated string warnings = 5;
    string error = 6;
}

message Host {
    string address = 1;
    string hostname = 2;
    HostStatus status = 3;
    repeated Port ports = 4;
    OSMatch os = 5;
    map<string, string> metadata = 6;
}

enum HostStatus {
    HOST_STATUS_UNSPECIFIED = 0;
    HOST_STATUS_UP = 1;
    HOST_STATUS_DOWN = 2;
    HOST_STATUS_UNKNOWN = 3;
}

message Port {
    int32 port = 1;
    string protocol = 2;          // tcp, udp
    PortState state = 3;
    Service service = 4;
    repeated ScriptResult scripts = 5;
}

enum PortState {
    PORT_STATE_UNSPECIFIED = 0;
    PORT_STATE_OPEN = 1;
    PORT_STATE_CLOSED = 2;
    PORT_STATE_FILTERED = 3;
    PORT_STATE_OPEN_FILTERED = 4;
}

message Service {
    string name = 1;              // e.g., "ssh", "http"
    string product = 2;           // e.g., "OpenSSH"
    string version = 3;           // e.g., "8.2p1"
    string extra_info = 4;
    repeated string cpe = 5;      // CPE identifiers
}

message ScriptResult {
    string name = 1;              // Script name
    string output = 2;            // Script output
    map<string, string> elements = 3;
}

message OSMatch {
    string name = 1;
    int32 accuracy = 2;
    repeated string cpe = 3;
}

message ScanStats {
    int32 hosts_up = 1;
    int32 hosts_down = 2;
    int32 hosts_total = 3;
    double elapsed_seconds = 4;
}
```

### Generating Go Code

```bash
# Install protoc and Go plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative proto/tool.proto
```

---

## Building a Tool

### Step 1: Project Structure

```
mytool/
├── proto/
│   └── tool.proto           # Proto definitions
│   └── tool.pb.go           # Generated Go code
├── tool.go                  # Tool implementation
├── parser.go                # Output parsing
├── main.go                  # Entry point
├── component.yaml           # Component metadata
├── go.mod
├── go.sum
└── Makefile
```

### Step 2: Implement the Tool Interface

```go
// tool.go
package mytool

import (
    "context"
    "fmt"
    "os/exec"
    "strings"

    pb "github.com/zero-day-ai/tools/mytool/proto"
    "github.com/zero-day-ai/sdk/health"
    "github.com/zero-day-ai/sdk/types"
    "google.golang.org/protobuf/proto"
)

type Tool struct {
    binaryPath string
    minVersion string
}

func New() *Tool {
    return &Tool{
        binaryPath: "mytool",      // Binary name in PATH
        minVersion: "2.0.0",       // Minimum required version
    }
}

// ═══════════════════════════════════════════════════════════════
// IDENTITY & METADATA
// ═══════════════════════════════════════════════════════════════

func (t *Tool) Name() string {
    return "mytool"
}

func (t *Tool) Version() string {
    return "1.0.0"
}

func (t *Tool) Description() string {
    return "Security scanning tool for vulnerability detection"
}

func (t *Tool) Tags() []string {
    return []string{
        "scanning",
        "vulnerability",
        "reconnaissance",
    }
}

// ═══════════════════════════════════════════════════════════════
// PROTO MESSAGE TYPES
// ═══════════════════════════════════════════════════════════════

func (t *Tool) InputMessageType() string {
    return "gibson.tools.mytool.ScanRequest"
}

func (t *Tool) OutputMessageType() string {
    return "gibson.tools.mytool.ScanResponse"
}

// ═══════════════════════════════════════════════════════════════
// EXECUTION
// ═══════════════════════════════════════════════════════════════

func (t *Tool) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
    // Type assertion
    req, ok := input.(*pb.ScanRequest)
    if !ok {
        return nil, fmt.Errorf("invalid input type: expected *ScanRequest, got %T", input)
    }

    // Validate input
    if err := t.validateRequest(req); err != nil {
        return &pb.ScanResponse{
            Error: fmt.Sprintf("validation error: %v", err),
        }, nil
    }

    // Build command arguments
    args := t.buildArgs(req)

    // Execute with context (respects cancellation/timeout)
    cmd := exec.CommandContext(ctx, t.binaryPath, args...)

    // Capture output
    output, err := cmd.CombinedOutput()
    if err != nil {
        // Check if context was cancelled
        if ctx.Err() != nil {
            return &pb.ScanResponse{
                Error: fmt.Sprintf("execution cancelled: %v", ctx.Err()),
            }, nil
        }

        // Non-zero exit might still have useful output
        if exitErr, ok := err.(*exec.ExitError); ok {
            return &pb.ScanResponse{
                RawOutput: string(output),
                Error:     fmt.Sprintf("exit code %d: %v", exitErr.ExitCode(), err),
                Warnings:  []string{"Tool exited with non-zero status"},
            }, nil
        }

        return &pb.ScanResponse{
            Error: fmt.Sprintf("execution failed: %v", err),
        }, nil
    }

    // Parse output into structured response
    response, err := t.parseOutput(string(output), req)
    if err != nil {
        return &pb.ScanResponse{
            RawOutput: string(output),
            Error:     fmt.Sprintf("parse error: %v", err),
            Warnings:  []string{"Output parsing failed, raw output available"},
        }, nil
    }

    response.RawOutput = string(output)
    return response, nil
}

func (t *Tool) validateRequest(req *pb.ScanRequest) error {
    if req.Target == "" && len(req.Targets) == 0 {
        return fmt.Errorf("target or targets required")
    }
    return nil
}

func (t *Tool) buildArgs(req *pb.ScanRequest) []string {
    var args []string

    // Add targets
    if req.Target != "" {
        args = append(args, req.Target)
    }
    args = append(args, req.Targets...)

    // Add options based on request
    if req.Ports != "" {
        args = append(args, "-p", req.Ports)
    }

    if req.TimeoutSeconds > 0 {
        args = append(args, "--timeout", fmt.Sprintf("%d", req.TimeoutSeconds))
    }

    // Add extra flags
    args = append(args, req.ExtraFlags...)

    return args
}

func (t *Tool) parseOutput(output string, req *pb.ScanRequest) (*pb.ScanResponse, error) {
    // Parse tool-specific output format
    // This is tool-specific - parse JSON, XML, or text output

    response := &pb.ScanResponse{
        Results: []*pb.Result{},
    }

    // Example: Parse JSON output
    // json.Unmarshal([]byte(output), &results)

    // Example: Parse line-by-line text output
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if result := t.parseLine(line); result != nil {
            response.Results = append(response.Results, result)
        }
    }

    return response, nil
}

func (t *Tool) parseLine(line string) *pb.Result {
    // Tool-specific line parsing
    return nil
}

// ═══════════════════════════════════════════════════════════════
// HEALTH CHECK
// ═══════════════════════════════════════════════════════════════

func (t *Tool) Health(ctx context.Context) types.HealthStatus {
    // Check binary exists
    binaryCheck := health.BinaryCheck(t.binaryPath)
    if binaryCheck.IsUnhealthy() {
        return binaryCheck
    }

    // Check version meets minimum
    versionCheck := health.BinaryVersionCheck(t.binaryPath, t.minVersion, "--version")
    if versionCheck.IsUnhealthy() {
        return types.NewDegradedStatus(
            fmt.Sprintf("version below minimum %s", t.minVersion),
            map[string]any{"minimum": t.minVersion},
        )
    }

    return types.NewHealthyStatus("ready")
}
```

### Step 3: Create Entry Point

```go
// main.go
package main

import (
    "github.com/zero-day-ai/tools/mytool"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    tool := mytool.New()

    // Serve via gRPC
    serve.Tool(tool, serve.WithPort(50052))
}
```

### Step 4: Component Metadata

```yaml
# component.yaml
name: mytool
version: 1.0.0
type: tool
description: Security scanning tool for vulnerability detection

tags:
  - scanning
  - vulnerability
  - reconnaissance

# MITRE ATT&CK mapping
mitre_attack:
  tactics:
    - TA0043  # Reconnaissance
  techniques:
    - T1595   # Active Scanning
    - T1595.002  # Vulnerability Scanning

# Proto definitions
proto:
  input: gibson.tools.mytool.ScanRequest
  output: gibson.tools.mytool.ScanResponse

# Dependencies
dependencies:
  binaries:
    - name: mytool
      version: ">=2.0.0"
      install: |
        # Installation instructions
        apt-get install mytool
        # or
        go install github.com/example/mytool@latest
  system: []

# Build configuration
build:
  command: make build
  artifacts:
    - bin/mytool-wrapper
```

### Step 5: Makefile

```makefile
.PHONY: all build test clean proto install

BINARY_NAME=mytool-wrapper
VERSION=$(shell git describe --tags --always --dirty)

all: proto build

proto:
	protoc --go_out=. --go_opt=paths=source_relative proto/tool.proto

build:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/$(BINARY_NAME) .

test:
	go test -v ./...

test-race:
	go test -race -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/ coverage.out coverage.html

install: build
	gibson tool install ./

lint:
	golangci-lint run

fmt:
	go fmt ./...
```

---

## Automatic Graph Storage

Gibson automatically persists discovered entities to the GraphRAG knowledge graph. Tool developers simply populate discovery data in responses - no graph knowledge required.

### How It Works

When tools return discovery data, Gibson:

1. Detects the `discovery` field in tool responses (via reflection)
2. Extracts entities (hosts, ports, services, etc.)
3. Creates graph nodes automatically
4. Infers parent-child relationships based on foreign keys
5. Stores everything in Neo4j

Tool developers just populate standard proto fields - the framework handles the rest.

### The Old Way (Manual Graph Management)

Before automatic graph storage, tools had to manually manage graph nodes:

```go
// BEFORE: Tools manually created graph nodes
func (t *NmapTool) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
    req := input.(*pb.ScanRequest)

    // Run scan...
    hosts, ports := runNmap(req)

    // Manual graph management (complex!)
    graphClient := getGraphClient()
    for _, host := range hosts {
        // Manually create host node
        hostNode := createHostNode(host.IP, host.Hostname)
        graphClient.CreateNode(ctx, hostNode)

        for _, port := range host.Ports {
            // Manually create port node
            portNode := createPortNode(port.Number, port.Protocol)
            graphClient.CreateNode(ctx, portNode)

            // Manually create relationship
            graphClient.CreateRelationship(ctx, portNode, hostNode, "RUNS_ON")
        }
    }

    return &pb.ScanResponse{
        Hosts: hosts,
        Ports: ports,
    }, nil
}
```

Problems with the old approach:
- Tools needed graph database knowledge
- Duplicate code across tools
- Error-prone relationship management
- Tight coupling to graph implementation

### The New Way (Automatic Graph Storage)

Now, tools just populate discovery data:

```go
// AFTER: Tools just populate discovery data
func (t *NmapTool) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
    req := input.(*pb.ScanRequest)

    // Run scan...
    hosts, ports := runNmap(req)

    // Create discovery result (proto field 100)
    discovery := &graphragpb.DiscoveryResult{
        Hosts: []*graphragpb.Host{
            {
                Ip:       "192.168.1.100",
                Hostname: "server01.example.com",
                State:    "up",
                Os:       "Linux 5.10",
            },
        },
        Ports: []*graphragpb.Port{
            {
                HostId:   "192.168.1.100",  // Foreign key to Host
                Number:   80,
                Protocol: "tcp",
                State:    "open",
            },
        },
        Services: []*graphragpb.Service{
            {
                PortId:  "192.168.1.100:80:tcp",  // Foreign key to Port
                Name:    "http",
                Version: "nginx 1.18.0",
            },
        },
    }

    // Return response with discovery field
    return &pb.ScanResponse{
        Hosts:     hosts,     // Tool-specific format
        Ports:     ports,     // Tool-specific format
        Discovery: discovery, // Standard discovery format (field 100)
    }, nil
}
```

Benefits of the new approach:
- No graph knowledge needed
- Consistent across all tools
- Relationships inferred automatically
- Maintainable and simple

### Adding Discovery to Your Tool

#### Step 1: Import Discovery Proto

```go
import (
    graphragpb "github.com/zero-day-ai/sdk/api/gen/graphragpb"
)
```

#### Step 2: Define Discovery Field in Proto

Add field 100 to your tool's response message:

```protobuf
syntax = "proto3";

package gibson.tools.mytool;

import "graphrag.proto";  // Import discovery types

message MyToolResponse {
    // Tool-specific fields (1-99)
    repeated Host hosts = 1;
    repeated Port ports = 2;
    string raw_output = 3;

    // Discovery data (field 100 is RESERVED)
    gibson.graphrag.DiscoveryResult discovery = 100;
}
```

#### Step 3: Populate Discovery in Tool Code

```go
func (t *MyTool) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
    req := input.(*pb.MyToolRequest)

    // Execute tool logic
    results := executeTool(req)

    // Build discovery result
    discovery := &graphragpb.DiscoveryResult{}

    // Add discovered hosts
    for _, host := range results.Hosts {
        discovery.Hosts = append(discovery.Hosts, &graphragpb.Host{
            Ip:       host.Address,
            Hostname: host.Name,
            State:    host.Status,
            Os:       host.OperatingSystem,
        })
    }

    // Add discovered ports
    for _, port := range results.Ports {
        discovery.Ports = append(discovery.Ports, &graphragpb.Port{
            HostId:   port.HostAddress,  // Links to parent host
            Number:   int32(port.Number),
            Protocol: port.Protocol,
            State:    port.State,
        })
    }

    // Return with discovery field populated
    return &pb.MyToolResponse{
        Hosts:     results.Hosts,  // Tool-specific format
        Ports:     results.Ports,  // Tool-specific format
        Discovery: discovery,      // Standard format
    }, nil
}
```

### Automatic Relationship Inference

Gibson creates parent-child relationships automatically based on foreign key patterns:

| Child Entity | Parent Entity | Foreign Key Field | Relationship |
|--------------|---------------|-------------------|--------------|
| Port | Host | `host_id` | `RUNS_ON` |
| Service | Port | `port_id` | `RUNS_ON` |
| Endpoint | Service | `service_id` | `EXPOSED_BY` |
| Subdomain | Domain | `domain_id` | `SUBDOMAIN_OF` |
| K8sNamespace | K8sCluster | `cluster_id` | `BELONGS_TO` |
| K8sPod | K8sNamespace | `namespace_id` | `BELONGS_TO` |
| K8sContainer | K8sPod | `pod_id` | `RUNS_IN` |

Example:

```go
discovery := &graphragpb.DiscoveryResult{
    Hosts: []*graphragpb.Host{
        {Ip: "192.168.1.100", Hostname: "web01"},
    },
    Ports: []*graphragpb.Port{
        {
            HostId:   "192.168.1.100",  // Automatic relationship created!
            Number:   443,
            Protocol: "tcp",
        },
    },
}
```

Creates graph:
```
(Host:192.168.1.100) <-[:RUNS_ON]- (Port:192.168.1.100:443:tcp)
```

### Explicit Relationships (Field 100)

Override automatic inference by declaring explicit relationships:

```go
discovery := &graphragpb.DiscoveryResult{
    // Standard entities
    Hosts: []*graphragpb.Host{
        {Ip: "192.168.1.100", Hostname: "web01"},
    },
    Vulnerabilities: []*graphragpb.Vulnerability{
        {Id: "CVE-2024-1234", CvssScore: 9.8},
    },

    // Explicit relationship (field 100)
    ExplicitRelationships: []*graphragpb.ExplicitRelationship{
        {
            FromType: graphragpb.NodeType_NODE_TYPE_VULNERABILITY,
            FromId:   map[string]string{"id": "CVE-2024-1234"},
            ToType:   graphragpb.NodeType_NODE_TYPE_HOST,
            ToId:     map[string]string{"ip": "192.168.1.100"},
            RelationshipType: graphragpb.RelationType_RELATION_TYPE_AFFECTS,
            Properties: map[string]string{
                "severity":    "critical",
                "exploitable": "true",
                "cvss":        "9.8",
            },
        },
    },
}
```

Creates graph:
```
(Vulnerability:CVE-2024-1234) -[:AFFECTS {severity:"critical"}]-> (Host:192.168.1.100)
```

### Custom Nodes (Field 200)

Create entities not in the standard taxonomy:

```go
discovery := &graphragpb.DiscoveryResult{
    // Standard entities
    Hosts: []*graphragpb.Host{
        {Ip: "192.168.1.100"},
    },

    // Custom node (field 200)
    CustomNodes: []*graphragpb.CustomNode{
        {
            NodeType: "WAFRule",
            IdProperties: map[string]string{
                "rule_id":  "12345",
                "waf_name": "cloudflare",
            },
            Properties: map[string]string{
                "pattern": "^/admin/.*",
                "action":  "block",
                "enabled": "true",
            },
            ParentType: graphragpb.NodeType_NODE_TYPE_HOST,
            ParentId:   map[string]string{"ip": "192.168.1.100"},
            RelationshipType: graphragpb.RelationType_RELATION_TYPE_PROTECTED_BY_SG,
        },
    },
}
```

Creates graph:
```
(WAFRule:12345) -[:PROTECTED_BY_SG]-> (Host:192.168.1.100)
```

### Available Entity Types

Discovery supports 30+ entity types across multiple categories:

#### Asset Discovery
- `Hosts` - IP addresses, hostnames, OS info
- `Ports` - Port numbers, protocols, states
- `Services` - Service names, versions, banners
- `Endpoints` - Web URLs, HTTP methods, responses
- `Domains` - Root domains, registrars
- `Subdomains` - Subdomains with parent links
- `Technologies` - Software, frameworks, libraries
- `Certificates` - TLS/SSL certificates

#### Network Infrastructure
- `Networks` - Subnets, CIDR ranges
- `IpRanges` - IP address ranges

#### Cloud Resources
- `CloudAccounts` - AWS, GCP, Azure accounts
- `CloudAssets` - VPCs, instances, buckets
- `CloudRegions` - Cloud provider regions

#### Kubernetes Resources
- `K8sClusters` - Kubernetes clusters
- `K8sNamespaces` - Namespaces with labels
- `K8sPods` - Pods with phase and IP
- `K8sContainers` - Containers within pods
- `K8sServices` - Services with ports
- `K8sIngresses` - Ingress rules

#### Security Findings
- `Vulnerabilities` - CVEs with CVSS scores
- `Credentials` - Username/password pairs
- `Secrets` - API keys, tokens

See `api/proto/graphrag.proto` for complete definitions.

### Field Numbering Convention

To ensure consistency across all tools:

| Field Range | Purpose | Example |
|-------------|---------|---------|
| 1-99 | Tool-specific response fields | `hosts`, `ports`, `raw_output` |
| 100 | **Discovery data (RESERVED)** | `gibson.graphrag.DiscoveryResult discovery = 100;` |

**IMPORTANT**: Always use field 100 for discovery data. This convention is enforced across the Gibson ecosystem.

### Complete Example: Web Scanner

```go
package webscanner

import (
    "context"
    "net/http"

    graphragpb "github.com/zero-day-ai/sdk/api/gen/graphragpb"
    pb "github.com/myorg/webscanner/proto"
    "google.golang.org/protobuf/proto"
)

func (t *WebScanner) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
    req := input.(*pb.ScanRequest)

    // Scan web application
    results := scanWebApp(req.Url)

    // Build discovery result
    discovery := &graphragpb.DiscoveryResult{
        Hosts: []*graphragpb.Host{
            {
                Ip:       results.ServerIP,
                Hostname: results.Hostname,
                State:    "up",
            },
        },
        Endpoints: []*graphragpb.Endpoint{
            {
                Url:        "https://example.com/api/users",
                Method:     "GET",
                StatusCode: 200,
                HostId:     results.ServerIP,
            },
            {
                Url:        "https://example.com/api/admin",
                Method:     "GET",
                StatusCode: 403,
                HostId:     results.ServerIP,
            },
        },
        Technologies: []*graphragpb.Technology{
            {
                Name:     "nginx",
                Version:  "1.18.0",
                Category: "web-server",
            },
            {
                Name:     "React",
                Version:  "18.2.0",
                Category: "frontend-framework",
            },
        },
    }

    // If vulnerabilities found, add them with explicit relationships
    if results.HasVulnerabilities {
        discovery.Vulnerabilities = []*graphragpb.Vulnerability{
            {
                Id:          "CVE-2024-1234",
                Name:        "Nginx Buffer Overflow",
                CvssScore:   7.5,
                Description: "Buffer overflow in nginx < 1.19.0",
            },
        }

        // Link vulnerability to technology
        discovery.ExplicitRelationships = []*graphragpb.ExplicitRelationship{
            {
                FromType: graphragpb.NodeType_NODE_TYPE_VULNERABILITY,
                FromId:   map[string]string{"id": "CVE-2024-1234"},
                ToType:   graphragpb.NodeType_NODE_TYPE_TECHNOLOGY,
                ToId:     map[string]string{"name": "nginx", "version": "1.18.0"},
                RelationshipType: graphragpb.RelationType_RELATION_TYPE_AFFECTS,
            },
        }
    }

    return &pb.ScanResponse{
        Url:       req.Url,
        Findings:  results.Findings,
        Discovery: discovery,  // Gibson handles the rest!
    }, nil
}
```

### Testing Discovery Integration

```go
func TestWebScanner_Discovery(t *testing.T) {
    scanner := New()

    req := &pb.ScanRequest{
        Url: "https://example.com",
    }

    resp, err := scanner.ExecuteProto(context.Background(), req)
    require.NoError(t, err)

    scanResp := resp.(*pb.ScanResponse)

    // Verify discovery data populated
    require.NotNil(t, scanResp.Discovery)
    assert.NotEmpty(t, scanResp.Discovery.Hosts)
    assert.NotEmpty(t, scanResp.Discovery.Endpoints)
    assert.NotEmpty(t, scanResp.Discovery.Technologies)

    // Verify host details
    host := scanResp.Discovery.Hosts[0]
    assert.NotEmpty(t, host.Ip)
    assert.NotEmpty(t, host.Hostname)

    // Verify endpoints have proper host_id references
    for _, endpoint := range scanResp.Discovery.Endpoints {
        assert.Equal(t, host.Ip, endpoint.HostId)
    }
}
```

### Migration Guide

To add automatic graph storage to existing tools:

1. **Import discovery proto**:
   ```go
   import graphragpb "github.com/zero-day-ai/sdk/api/gen/graphragpb"
   ```

2. **Add discovery field to response proto**:
   ```protobuf
   import "graphrag.proto";

   message MyToolResponse {
       // Existing fields...
       gibson.graphrag.DiscoveryResult discovery = 100;
   }
   ```

3. **Regenerate proto code**:
   ```bash
   make proto
   ```

4. **Populate discovery in tool**:
   ```go
   discovery := &graphragpb.DiscoveryResult{
       Hosts: convertToDiscoveryHosts(toolResults.Hosts),
       Ports: convertToDiscoveryPorts(toolResults.Ports),
   }

   return &pb.MyToolResponse{
       // Existing fields...
       Discovery: discovery,
   }
   ```

5. **Remove manual graph code** (if any):
   - Delete graph client initialization
   - Delete manual node creation
   - Delete manual relationship creation

### Pro Tips

1. **Use consistent IDs**: Use the same ID format across entities (e.g., IP addresses for `host_id`)

2. **Populate all relationships**: Always set foreign key fields (`host_id`, `port_id`, etc.)

3. **Test locally**: Use Gibson's local mode to verify graph creation before deployment

4. **Check graph in Neo4j**: Query the graph to verify relationships:
   ```cypher
   MATCH (h:Host)-[:RUNS_ON]-(p:Port)-[:RUNS_ON]-(s:Service)
   WHERE h.ip = "192.168.1.100"
   RETURN h, p, s
   ```

5. **Use explicit relationships for non-hierarchical links**: When entities don't have a parent-child relationship, use `ExplicitRelationships`

---

## Health Checks

Tools must report their health status, particularly checking for required dependencies.

### Health Check Utilities

```go
import "github.com/zero-day-ai/sdk/health"

// Check if binary exists in PATH
health.BinaryCheck("nmap")  // Returns HealthStatus

// Check binary version meets minimum
health.BinaryVersionCheck("nmap", "7.80", "--version")

// Check network connectivity
health.NetworkCheck(ctx, "api.shodan.io", 443)

// Check file exists and is readable
health.FileCheck("/etc/nmap/scripts")

// Combine multiple checks
health.Combine(
    health.BinaryCheck("nmap"),
    health.BinaryCheck("masscan"),
    health.NetworkCheck(ctx, "localhost", 5432),
)
```

### Custom Health Checks

```go
func (t *Tool) Health(ctx context.Context) types.HealthStatus {
    var checks []types.HealthStatus

    // Binary check
    checks = append(checks, health.BinaryCheck(t.binaryPath))

    // Version check
    version, err := t.getVersion()
    if err != nil {
        checks = append(checks, types.NewUnhealthyStatus(
            "failed to get version",
            map[string]any{"error": err.Error()},
        ))
    } else if !t.isVersionCompatible(version) {
        checks = append(checks, types.NewDegradedStatus(
            fmt.Sprintf("version %s below minimum %s", version, t.minVersion),
            map[string]any{"current": version, "minimum": t.minVersion},
        ))
    }

    // Database check (if needed)
    if t.requiresDB {
        checks = append(checks, t.checkDatabase(ctx))
    }

    // API key check
    if t.requiresAPIKey {
        if os.Getenv("MYTOOL_API_KEY") == "" {
            checks = append(checks, types.NewDegradedStatus(
                "API key not configured",
                map[string]any{"env_var": "MYTOOL_API_KEY"},
            ))
        }
    }

    // Combine all checks
    combined := health.Combine(checks...)

    // Add tool-specific details
    if combined.IsHealthy() {
        combined.Details = map[string]any{
            "version":    version,
            "binary":     t.binaryPath,
            "features":   t.getFeatures(),
        }
    }

    return combined
}
```

### Health Status Types

```go
// Healthy - fully operational
types.NewHealthyStatus("all systems operational")

// Degraded - functional with limitations
types.NewDegradedStatus("running with reduced features", map[string]any{
    "reason":   "API key missing",
    "affected": []string{"cloud scanning", "CVE lookup"},
})

// Unhealthy - cannot function
types.NewUnhealthyStatus("binary not found", map[string]any{
    "binary": "nmap",
    "path":   "/usr/bin/nmap",
})

// Check status
status.IsHealthy()   // true if Status == "healthy"
status.IsDegraded()  // true if Status == "degraded"
status.IsUnhealthy() // true if Status == "unhealthy"
```

---

## Error Handling

Tools should handle errors gracefully and return structured error information.

### Error Response Pattern

```go
func (t *Tool) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
    req := input.(*pb.ScanRequest)

    // Validation errors - return in response, not as error
    if err := t.validate(req); err != nil {
        return &pb.ScanResponse{
            Success: false,
            Error: &pb.Error{
                Code:    "VALIDATION_ERROR",
                Message: err.Error(),
                Details: map[string]string{
                    "field": "target",
                },
            },
        }, nil  // Note: nil error, structured error in response
    }

    // Execution errors
    output, err := t.execute(ctx, req)
    if err != nil {
        // Context cancellation
        if ctx.Err() == context.Canceled {
            return &pb.ScanResponse{
                Success: false,
                Error: &pb.Error{
                    Code:    "CANCELLED",
                    Message: "operation cancelled by user",
                },
            }, nil
        }

        // Context timeout
        if ctx.Err() == context.DeadlineExceeded {
            return &pb.ScanResponse{
                Success: false,
                Error: &pb.Error{
                    Code:    "TIMEOUT",
                    Message: "operation timed out",
                },
            }, nil
        }

        // Execution error
        return &pb.ScanResponse{
            Success: false,
            Error: &pb.Error{
                Code:    "EXECUTION_ERROR",
                Message: err.Error(),
            },
            RawOutput: t.lastOutput,  // Include partial output
        }, nil
    }

    return output, nil
}
```

### Error Codes

Define standard error codes in your proto:

```protobuf
message Error {
    string code = 1;           // Machine-readable code
    string message = 2;        // Human-readable message
    map<string, string> details = 3;
}

// Common error codes:
// - VALIDATION_ERROR: Invalid input
// - NOT_FOUND: Target not found/reachable
// - TIMEOUT: Operation timed out
// - CANCELLED: Operation cancelled
// - PERMISSION_DENIED: Insufficient permissions
// - RATE_LIMITED: Rate limit exceeded
// - DEPENDENCY_ERROR: Required dependency unavailable
// - PARSE_ERROR: Output parsing failed
// - EXECUTION_ERROR: Generic execution failure
```

### Warnings vs Errors

```go
// Use warnings for non-fatal issues
response := &pb.ScanResponse{
    Success: true,
    Results: results,
    Warnings: []string{
        "Some hosts were unreachable",
        "Script execution partially failed",
    },
}

// Use errors for fatal issues
response := &pb.ScanResponse{
    Success: false,
    Error: &pb.Error{
        Code:    "EXECUTION_ERROR",
        Message: "nmap binary not found",
    },
}
```

---

## MITRE ATT&CK Mappings

Map tools to MITRE ATT&CK/ATLAS techniques for classification.

### Component Metadata

```yaml
# component.yaml
mitre_attack:
  tactics:
    - TA0043  # Reconnaissance
    - TA0007  # Discovery
  techniques:
    - T1595           # Active Scanning
    - T1595.001       # Scanning IP Blocks
    - T1595.002       # Vulnerability Scanning
    - T1046           # Network Service Discovery

# For AI/ML tools, use ATLAS
mitre_atlas:
  tactics:
    - AML.TA0002      # ML Attack Staging
  techniques:
    - AML.T0051       # LLM Prompt Injection
    - AML.T0054       # LLM Jailbreak
```

### Common Mappings

| Tool Type | Tactics | Techniques |
|-----------|---------|------------|
| Port Scanner | TA0043 (Recon) | T1595 (Active Scanning) |
| Vulnerability Scanner | TA0043 | T1595.002 (Vuln Scanning) |
| Web Crawler | TA0043 | T1595.003 (Wordlist Scanning) |
| DNS Tools | TA0043 | T1596.001 (DNS/Passive DNS) |
| SSL Analyzer | TA0043 | T1595 (Active Scanning) |
| Credential Tools | TA0006 (Cred Access) | T1110 (Brute Force) |

---

## Testing Tools

### Unit Tests

```go
// tool_test.go
package mytool

import (
    "context"
    "testing"
    "time"

    pb "github.com/zero-day-ai/tools/mytool/proto"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestTool_ExecuteProto_ValidInput(t *testing.T) {
    tool := New()

    req := &pb.ScanRequest{
        Target: "localhost",
        Ports:  "80,443",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    resp, err := tool.ExecuteProto(ctx, req)
    require.NoError(t, err)

    scanResp := resp.(*pb.ScanResponse)
    assert.Empty(t, scanResp.Error)
    assert.NotEmpty(t, scanResp.RawOutput)
}

func TestTool_ExecuteProto_InvalidInput(t *testing.T) {
    tool := New()

    req := &pb.ScanRequest{
        // Missing target
        Ports: "80",
    }

    resp, err := tool.ExecuteProto(context.Background(), req)
    require.NoError(t, err)  // Error in response, not returned

    scanResp := resp.(*pb.ScanResponse)
    assert.Contains(t, scanResp.Error.Message, "target")
}

func TestTool_ExecuteProto_ContextCancellation(t *testing.T) {
    tool := New()

    req := &pb.ScanRequest{
        Target: "10.0.0.0/8",  // Large scan
    }

    ctx, cancel := context.WithCancel(context.Background())

    // Cancel immediately
    cancel()

    resp, err := tool.ExecuteProto(ctx, req)
    require.NoError(t, err)

    scanResp := resp.(*pb.ScanResponse)
    assert.Equal(t, "CANCELLED", scanResp.Error.Code)
}

func TestTool_ExecuteProto_Timeout(t *testing.T) {
    tool := New()

    req := &pb.ScanRequest{
        Target: "10.0.0.0/8",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()

    resp, err := tool.ExecuteProto(ctx, req)
    require.NoError(t, err)

    scanResp := resp.(*pb.ScanResponse)
    assert.Equal(t, "TIMEOUT", scanResp.Error.Code)
}

func TestTool_Health(t *testing.T) {
    tool := New()

    status := tool.Health(context.Background())

    // May be unhealthy if binary not installed in test env
    if status.IsHealthy() {
        assert.NotEmpty(t, status.Details["version"])
    }
}

func TestTool_BuildArgs(t *testing.T) {
    tool := New()

    tests := []struct {
        name     string
        req      *pb.ScanRequest
        expected []string
    }{
        {
            name: "basic target",
            req: &pb.ScanRequest{
                Target: "192.168.1.1",
            },
            expected: []string{"192.168.1.1"},
        },
        {
            name: "with ports",
            req: &pb.ScanRequest{
                Target: "192.168.1.1",
                Ports:  "22,80,443",
            },
            expected: []string{"192.168.1.1", "-p", "22,80,443"},
        },
        {
            name: "multiple targets",
            req: &pb.ScanRequest{
                Targets: []string{"192.168.1.1", "192.168.1.2"},
            },
            expected: []string{"192.168.1.1", "192.168.1.2"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            args := tool.buildArgs(tt.req)
            assert.Equal(t, tt.expected, args)
        })
    }
}
```

### Integration Tests

```go
// +build integration

package mytool

import (
    "context"
    "os"
    "testing"
    "time"

    pb "github.com/zero-day-ai/tools/mytool/proto"
    "github.com/stretchr/testify/require"
)

func TestTool_Integration(t *testing.T) {
    if os.Getenv("INTEGRATION_TEST") == "" {
        t.Skip("Skipping integration test")
    }

    tool := New()

    // Ensure healthy
    status := tool.Health(context.Background())
    require.True(t, status.IsHealthy(), "tool must be healthy: %s", status.Message)

    // Real scan
    req := &pb.ScanRequest{
        Target: "scanme.nmap.org",  // Nmap's test target
        Ports:  "22,80,443",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    resp, err := tool.ExecuteProto(ctx, req)
    require.NoError(t, err)

    scanResp := resp.(*pb.ScanResponse)
    require.Empty(t, scanResp.Error)
    require.NotEmpty(t, scanResp.Results)
}
```

---

## Serving Tools

### gRPC Server

```go
package main

import (
    "github.com/zero-day-ai/tools/mytool"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    tool := mytool.New()

    // Basic serving
    serve.Tool(tool)

    // With options
    serve.Tool(tool,
        serve.WithPort(50052),
        serve.WithTLS("cert.pem", "key.pem"),
        serve.WithHealthEndpoint("/health"),
        serve.WithGracefulTimeout(30*time.Second),
    )
}
```

### Server Configuration

```go
cfg := &serve.Config{
    Port:            50052,
    HealthEndpoint:  "/health",
    GracefulTimeout: 30 * time.Second,
    TLSCertFile:     "/path/to/cert.pem",
    TLSKeyFile:      "/path/to/key.pem",
    LocalMode:       "/tmp/mytool.sock",  // Unix socket (optional)
    AdvertiseAddr:   "mytool.local:50052", // For service discovery
}

server, err := serve.NewServer(cfg)
if err != nil {
    log.Fatal(err)
}

// Register tool
server.RegisterTool(tool)

// Start serving (blocks)
if err := server.Serve(context.Background()); err != nil {
    log.Fatal(err)
}
```

### Graceful Shutdown

The server handles SIGINT/SIGTERM automatically:

```go
// Server listens for signals and gracefully shuts down
// - Stops accepting new requests
// - Waits for in-flight requests to complete (up to GracefulTimeout)
// - Closes connections
```

---

## Complete Examples

### Nmap Wrapper

```go
package nmap

import (
    "context"
    "encoding/xml"
    "fmt"
    "os/exec"
    "strings"

    pb "github.com/zero-day-ai/tools/nmap/proto"
    "github.com/zero-day-ai/sdk/health"
    "github.com/zero-day-ai/sdk/types"
    "google.golang.org/protobuf/proto"
)

type Tool struct {
    binaryPath string
}

func New() *Tool {
    return &Tool{binaryPath: "nmap"}
}

func (t *Tool) Name() string        { return "nmap" }
func (t *Tool) Version() string     { return "1.0.0" }
func (t *Tool) Description() string { return "Network exploration and security auditing tool" }
func (t *Tool) Tags() []string      { return []string{"discovery", "network", "scanning", "reconnaissance"} }

func (t *Tool) InputMessageType() string  { return "gibson.tools.nmap.ScanRequest" }
func (t *Tool) OutputMessageType() string { return "gibson.tools.nmap.ScanResponse" }

func (t *Tool) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
    req, ok := input.(*pb.ScanRequest)
    if !ok {
        return nil, fmt.Errorf("expected *ScanRequest, got %T", input)
    }

    // Build arguments
    args := []string{"-oX", "-"}  // XML output to stdout

    // Add target
    if req.Target != "" {
        args = append(args, req.Target)
    }
    args = append(args, req.Targets...)

    // Add ports
    if req.Ports != "" {
        args = append(args, "-p", req.Ports)
    } else if req.TopPorts {
        count := req.TopPortsCount
        if count == 0 {
            count = 1000
        }
        args = append(args, "--top-ports", fmt.Sprintf("%d", count))
    }

    // Scan type
    switch req.ScanType {
    case pb.ScanType_SCAN_TYPE_SYN:
        args = append(args, "-sS")
    case pb.ScanType_SCAN_TYPE_CONNECT:
        args = append(args, "-sT")
    case pb.ScanType_SCAN_TYPE_UDP:
        args = append(args, "-sU")
    case pb.ScanType_SCAN_TYPE_PING:
        args = append(args, "-sn")
    }

    // Timing
    if req.Timing != pb.TimingTemplate_TIMING_UNSPECIFIED {
        args = append(args, fmt.Sprintf("-T%d", req.Timing-1))
    }

    // Service detection
    if req.ServiceDetection {
        args = append(args, "-sV")
    }

    // OS detection
    if req.OsDetection {
        args = append(args, "-O")
    }

    // Script scan
    if req.ScriptScan {
        args = append(args, "-sC")
    }
    for _, script := range req.Scripts {
        args = append(args, "--script", script)
    }

    // Extra flags
    args = append(args, req.ExtraFlags...)

    // Execute
    cmd := exec.CommandContext(ctx, t.binaryPath, args...)
    output, err := cmd.CombinedOutput()

    response := &pb.ScanResponse{
        RawOutput: string(output),
    }

    if err != nil {
        if ctx.Err() != nil {
            response.Error = ctx.Err().Error()
            return response, nil
        }
        response.Error = err.Error()
        response.Warnings = append(response.Warnings, "nmap exited with error")
    }

    // Parse XML output
    if err := t.parseXML(output, response); err != nil {
        response.Warnings = append(response.Warnings, fmt.Sprintf("XML parse error: %v", err))
    }

    return response, nil
}

func (t *Tool) parseXML(data []byte, response *pb.ScanResponse) error {
    var nmapRun struct {
        Hosts []struct {
            Address struct {
                Addr string `xml:"addr,attr"`
            } `xml:"address"`
            Hostnames struct {
                Hostname []struct {
                    Name string `xml:"name,attr"`
                } `xml:"hostname"`
            } `xml:"hostnames"`
            Status struct {
                State string `xml:"state,attr"`
            } `xml:"status"`
            Ports struct {
                Port []struct {
                    Protocol string `xml:"protocol,attr"`
                    PortID   int32  `xml:"portid,attr"`
                    State    struct {
                        State string `xml:"state,attr"`
                    } `xml:"state"`
                    Service struct {
                        Name    string `xml:"name,attr"`
                        Product string `xml:"product,attr"`
                        Version string `xml:"version,attr"`
                    } `xml:"service"`
                } `xml:"port"`
            } `xml:"ports"`
        } `xml:"host"`
        RunStats struct {
            Hosts struct {
                Up    int32 `xml:"up,attr"`
                Down  int32 `xml:"down,attr"`
                Total int32 `xml:"total,attr"`
            } `xml:"hosts"`
            Finished struct {
                Elapsed float64 `xml:"elapsed,attr"`
            } `xml:"finished"`
        } `xml:"runstats"`
    }

    if err := xml.Unmarshal(data, &nmapRun); err != nil {
        return err
    }

    for _, h := range nmapRun.Hosts {
        host := &pb.Host{
            Address: h.Address.Addr,
        }

        if len(h.Hostnames.Hostname) > 0 {
            host.Hostname = h.Hostnames.Hostname[0].Name
        }

        switch h.Status.State {
        case "up":
            host.Status = pb.HostStatus_HOST_STATUS_UP
        case "down":
            host.Status = pb.HostStatus_HOST_STATUS_DOWN
        default:
            host.Status = pb.HostStatus_HOST_STATUS_UNKNOWN
        }

        for _, p := range h.Ports.Port {
            port := &pb.Port{
                Port:     p.PortID,
                Protocol: p.Protocol,
                Service: &pb.Service{
                    Name:    p.Service.Name,
                    Product: p.Service.Product,
                    Version: p.Service.Version,
                },
            }

            switch p.State.State {
            case "open":
                port.State = pb.PortState_PORT_STATE_OPEN
            case "closed":
                port.State = pb.PortState_PORT_STATE_CLOSED
            case "filtered":
                port.State = pb.PortState_PORT_STATE_FILTERED
            }

            host.Ports = append(host.Ports, port)
        }

        response.Hosts = append(response.Hosts, host)
    }

    response.Stats = &pb.ScanStats{
        HostsUp:        nmapRun.RunStats.Hosts.Up,
        HostsDown:      nmapRun.RunStats.Hosts.Down,
        HostsTotal:     nmapRun.RunStats.Hosts.Total,
        ElapsedSeconds: nmapRun.RunStats.Finished.Elapsed,
    }

    return nil
}

func (t *Tool) Health(ctx context.Context) types.HealthStatus {
    check := health.BinaryVersionCheck(t.binaryPath, "7.80", "--version")
    if check.IsUnhealthy() {
        return check
    }

    return types.NewHealthyStatus("nmap ready")
}
```

### HTTP Request Tool

```go
package http

import (
    "context"
    "crypto/tls"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"

    pb "github.com/zero-day-ai/tools/http/proto"
    "github.com/zero-day-ai/sdk/types"
    "google.golang.org/protobuf/proto"
)

type Tool struct {
    client *http.Client
}

func New() *Tool {
    return &Tool{
        client: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    InsecureSkipVerify: true,  // For security testing
                },
            },
            CheckRedirect: func(req *http.Request, via []*http.Request) error {
                if len(via) >= 10 {
                    return fmt.Errorf("too many redirects")
                }
                return nil
            },
        },
    }
}

func (t *Tool) Name() string        { return "http" }
func (t *Tool) Version() string     { return "1.0.0" }
func (t *Tool) Description() string { return "HTTP request tool for web testing" }
func (t *Tool) Tags() []string      { return []string{"http", "web", "request"} }

func (t *Tool) InputMessageType() string  { return "gibson.tools.http.Request" }
func (t *Tool) OutputMessageType() string { return "gibson.tools.http.Response" }

func (t *Tool) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
    req, ok := input.(*pb.Request)
    if !ok {
        return nil, fmt.Errorf("expected *Request, got %T", input)
    }

    // Build HTTP request
    httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.Url, strings.NewReader(req.Body))
    if err != nil {
        return &pb.Response{
            Error: &pb.Error{
                Code:    "REQUEST_BUILD_ERROR",
                Message: err.Error(),
            },
        }, nil
    }

    // Add headers
    for key, value := range req.Headers {
        httpReq.Header.Set(key, value)
    }

    // Set timeout if specified
    if req.TimeoutSeconds > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutSeconds)*time.Second)
        defer cancel()
        httpReq = httpReq.WithContext(ctx)
    }

    // Execute request
    start := time.Now()
    httpResp, err := t.client.Do(httpReq)
    elapsed := time.Since(start)

    if err != nil {
        return &pb.Response{
            Error: &pb.Error{
                Code:    "REQUEST_ERROR",
                Message: err.Error(),
            },
            ElapsedMs: elapsed.Milliseconds(),
        }, nil
    }
    defer httpResp.Body.Close()

    // Read body
    body, err := io.ReadAll(httpResp.Body)
    if err != nil {
        return &pb.Response{
            StatusCode: int32(httpResp.StatusCode),
            Error: &pb.Error{
                Code:    "BODY_READ_ERROR",
                Message: err.Error(),
            },
            ElapsedMs: elapsed.Milliseconds(),
        }, nil
    }

    // Build response
    response := &pb.Response{
        StatusCode: int32(httpResp.StatusCode),
        Headers:    make(map[string]string),
        Body:       string(body),
        BodyBytes:  body,
        ElapsedMs:  elapsed.Milliseconds(),
    }

    for key, values := range httpResp.Header {
        response.Headers[key] = strings.Join(values, ", ")
    }

    // Extract cookies
    for _, cookie := range httpResp.Cookies() {
        response.Cookies = append(response.Cookies, &pb.Cookie{
            Name:     cookie.Name,
            Value:    cookie.Value,
            Domain:   cookie.Domain,
            Path:     cookie.Path,
            Secure:   cookie.Secure,
            HttpOnly: cookie.HttpOnly,
        })
    }

    return response, nil
}

func (t *Tool) Health(ctx context.Context) types.HealthStatus {
    // HTTP tool is always healthy (no external dependencies)
    return types.NewHealthyStatus("ready")
}
```

---

## Best Practices

### 1. Always Validate Input

```go
func (t *Tool) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
    req := input.(*pb.Request)

    // Validate required fields
    if req.Target == "" {
        return &pb.Response{
            Error: &pb.Error{Code: "VALIDATION_ERROR", Message: "target required"},
        }, nil
    }

    // Validate ranges
    if req.Timeout < 0 || req.Timeout > 3600 {
        return &pb.Response{
            Error: &pb.Error{Code: "VALIDATION_ERROR", Message: "timeout must be 0-3600"},
        }, nil
    }

    // Continue execution...
}
```

### 2. Respect Context

```go
// Always use context for cancellation and timeout
cmd := exec.CommandContext(ctx, binary, args...)

// Check context in long operations
for i := 0; i < len(targets); i++ {
    select {
    case <-ctx.Done():
        return partialResults, nil
    default:
        result := scan(targets[i])
        results = append(results, result)
    }
}
```

### 3. Return Structured Errors

```go
// Don't return Go errors for tool failures
// BAD:
return nil, fmt.Errorf("scan failed: %v", err)

// GOOD:
return &pb.Response{
    Error: &pb.Error{
        Code:    "SCAN_FAILED",
        Message: err.Error(),
        Details: map[string]string{"target": req.Target},
    },
    RawOutput: partialOutput,  // Include what we got
}, nil
```

### 4. Include Raw Output

```go
// Always include raw output for debugging
response := &pb.Response{
    Results:   parsedResults,
    RawOutput: rawToolOutput,  // Original tool output
    XmlOutput: xmlOutput,      // If applicable
}
```

### 5. Implement Comprehensive Health Checks

```go
func (t *Tool) Health(ctx context.Context) types.HealthStatus {
    // Check ALL dependencies
    checks := []types.HealthStatus{
        health.BinaryCheck(t.binaryPath),
        health.BinaryVersionCheck(t.binaryPath, t.minVersion, "--version"),
    }

    // Add optional dependency checks
    if t.usesDatabase {
        checks = append(checks, t.checkDB(ctx))
    }

    return health.Combine(checks...)
}
```

### 6. Use Appropriate Timeouts

```go
// Set reasonable defaults
const (
    DefaultTimeout = 30 * time.Second
    MaxTimeout     = 10 * time.Minute
)

// Honor request timeout
timeout := DefaultTimeout
if req.TimeoutSeconds > 0 {
    timeout = time.Duration(req.TimeoutSeconds) * time.Second
    if timeout > MaxTimeout {
        timeout = MaxTimeout
    }
}

ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()
```

### 7. Parse Output Robustly

```go
func (t *Tool) parseOutput(output string) (*pb.Results, error) {
    // Try JSON first
    var jsonResults Results
    if err := json.Unmarshal([]byte(output), &jsonResults); err == nil {
        return convertJSON(jsonResults), nil
    }

    // Try XML
    var xmlResults XMLResults
    if err := xml.Unmarshal([]byte(output), &xmlResults); err == nil {
        return convertXML(xmlResults), nil
    }

    // Fall back to line parsing
    return parseLines(output), nil
}
```

### 8. Document Proto Messages

```protobuf
// Request to scan a target for vulnerabilities
message ScanRequest {
    // The target to scan. Can be:
    // - IP address: "192.168.1.1"
    // - Hostname: "example.com"
    // - CIDR range: "192.168.1.0/24"
    // - IP range: "192.168.1.1-100"
    string target = 1;

    // Port specification. Examples:
    // - Single port: "80"
    // - Range: "1-1000"
    // - List: "22,80,443,8080"
    // - Combined: "22,80,100-200"
    // Default: top 1000 ports
    string ports = 2;
}
```

### 9. Test Edge Cases

```go
func TestTool_EdgeCases(t *testing.T) {
    tool := New()

    tests := []struct {
        name string
        req  *pb.Request
        want string  // Expected error code
    }{
        {"empty target", &pb.Request{}, "VALIDATION_ERROR"},
        {"invalid port", &pb.Request{Target: "x", Ports: "abc"}, "VALIDATION_ERROR"},
        {"huge timeout", &pb.Request{Target: "x", Timeout: 999999}, "VALIDATION_ERROR"},
        {"special chars", &pb.Request{Target: "'; DROP TABLE"}, "VALIDATION_ERROR"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, _ := tool.ExecuteProto(context.Background(), tt.req)
            r := resp.(*pb.Response)
            assert.Equal(t, tt.want, r.Error.Code)
        })
    }
}
```

### 10. Security Considerations

```go
// Sanitize inputs to prevent command injection
func sanitizeTarget(target string) (string, error) {
    // Reject shell metacharacters
    dangerous := []string{";", "|", "&", "$", "`", "(", ")", "{", "}", "<", ">", "\\", "'", "\""}
    for _, char := range dangerous {
        if strings.Contains(target, char) {
            return "", fmt.Errorf("invalid character in target: %s", char)
        }
    }
    return target, nil
}

// Use exec.Command properly (not shell)
// GOOD:
cmd := exec.Command("nmap", "-p", ports, target)

// BAD (shell injection risk):
cmd := exec.Command("sh", "-c", fmt.Sprintf("nmap -p %s %s", ports, target))
```
