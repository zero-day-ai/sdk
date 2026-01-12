# Agent Developer Guide: GraphRAG Taxonomy

This guide explains how agents interact with the GraphRAG knowledge graph system using the standardized taxonomy for node types, relationships, and attack techniques.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Node Types](#node-types)
- [Relationship Types](#relationship-types)
- [Attack Techniques](#attack-techniques)
- [Using the Taxonomy in Agents](#using-the-taxonomy-in-agents)
- [Creating Nodes](#creating-nodes)
- [Creating Relationships](#creating-relationships)
- [Building Attack Chains](#building-attack-chains)
- [Querying the Graph](#querying-the-graph)
- [Custom Types](#custom-types)
- [Best Practices](#best-practices)
- [Complete Example](#complete-example)

---

## Overview

The GraphRAG taxonomy provides a **standardized vocabulary** for agents to describe:

- **Assets**: Domains, hosts, services, endpoints discovered during reconnaissance
- **Findings**: Vulnerabilities, evidence, and mitigations
- **Execution**: Missions, agent runs, tool executions, LLM calls
- **Attacks**: Techniques and tactics from MITRE ATT&CK and Arcanum PI

All taxonomy constants are **embedded in the Gibson binary** and exposed to agents via the SDK's `graphrag` package. The taxonomy version is tied to the Gibson release version.

### Why Use Canonical Types?

1. **Consistency**: All agents use the same vocabulary
2. **Queryability**: Standard types enable cross-mission queries
3. **Interoperability**: Agents can understand each other's findings
4. **Reporting**: Consistent types enable standardized reports

---

## Quick Start

```go
import "github.com/zero-day-ai/sdk/graphrag"

// Create a domain node using canonical type
domain := graphrag.NewGraphNode(graphrag.NodeTypeDomain)
domain.SetProperty(graphrag.PropName, "example.com")

// Create a finding
finding := graphrag.NewGraphNode(graphrag.NodeTypeFinding)
finding.SetProperty(graphrag.PropTitle, "SQL Injection")
finding.SetProperty(graphrag.PropSeverity, "critical")

// Create a relationship
rel := graphrag.NewRelationship(finding.ID, domain.ID, graphrag.RelTypeAffects)

// Reference a MITRE technique
finding.SetProperty("technique", graphrag.TechniqueT1190) // Exploit Public-Facing Application
```

---

## Node Types

The taxonomy defines **19 canonical node types** across 4 categories:

### Asset Nodes

| Constant | Value | Description |
|----------|-------|-------------|
| `NodeTypeDomain` | `"domain"` | Root domain entity (e.g., `example.com`) |
| `NodeTypeSubdomain` | `"subdomain"` | Subdomain under a root domain |
| `NodeTypeHost` | `"host"` | IP address or hostname |
| `NodeTypePort` | `"port"` | Network port on a host |
| `NodeTypeService` | `"service"` | Service running on a port |
| `NodeTypeEndpoint` | `"endpoint"` | Web endpoint or URL |
| `NodeTypeApi` | `"api"` | Web API with multiple endpoints |
| `NodeTypeTechnology` | `"technology"` | Software/framework detected |
| `NodeTypeCloudAsset` | `"cloud_asset"` | Cloud infrastructure resource |
| `NodeTypeCertificate` | `"certificate"` | TLS/SSL certificate |

### Finding Nodes

| Constant | Value | Description |
|----------|-------|-------------|
| `NodeTypeFinding` | `"finding"` | Security vulnerability or issue |
| `NodeTypeEvidence` | `"evidence"` | Supporting evidence for a finding |
| `NodeTypeMitigation` | `"mitigation"` | Remediation measure |

### Execution Nodes

| Constant | Value | Description |
|----------|-------|-------------|
| `NodeTypeMission` | `"mission"` | Top-level security assessment |
| `NodeTypeAgentRun` | `"agent_run"` | Single agent execution within a mission |
| `NodeTypeToolExecution` | `"tool_execution"` | Tool invocation |
| `NodeTypeLlmCall` | `"llm_call"` | LLM API call |

### Attack Nodes

| Constant | Value | Description |
|----------|-------|-------------|
| `NodeTypeTechnique` | `"technique"` | Attack technique (MITRE/Arcanum) |
| `NodeTypeTactic` | `"tactic"` | High-level adversary goal |

---

## Relationship Types

The taxonomy defines **20 canonical relationship types** across 3 categories:

### Asset Hierarchy

| Constant | Value | From → To |
|----------|-------|-----------|
| `RelTypeHasSubdomain` | `"HAS_SUBDOMAIN"` | Domain → Subdomain |
| `RelTypeResolvesTo` | `"RESOLVES_TO"` | Domain/Subdomain → Host |
| `RelTypeHasPort` | `"HAS_PORT"` | Host → Port |
| `RelTypeRunsService` | `"RUNS_SERVICE"` | Port → Service |
| `RelTypeHasEndpoint` | `"HAS_ENDPOINT"` | Service → Endpoint |
| `RelTypeUsesTechnology` | `"USES_TECHNOLOGY"` | Service/Endpoint → Technology |
| `RelTypeServesCertificate` | `"SERVES_CERTIFICATE"` | Host → Certificate |
| `RelTypeHosts` | `"HOSTS"` | CloudAsset → Host/Service |

### Finding Links

| Constant | Value | From → To |
|----------|-------|-----------|
| `RelTypeAffects` | `"AFFECTS"` | Finding → Asset |
| `RelTypeHasEvidence` | `"HAS_EVIDENCE"` | Finding → Evidence |
| `RelTypeUsesTechnique` | `"USES_TECHNIQUE"` | Finding → Technique |
| `RelTypeExploits` | `"EXPLOITS"` | Finding → Technology |
| `RelTypeMitigates` | `"MITIGATES"` | Mitigation → Finding |
| `RelTypeLeadsTo` | `"LEADS_TO"` | Finding → Finding (attack chains) |
| `RelTypeSimilarTo` | `"SIMILAR_TO"` | Finding ↔ Finding (bidirectional) |

### Execution Context

| Constant | Value | From → To |
|----------|-------|-----------|
| `RelTypePartOf` | `"PART_OF"` | AgentRun → Mission |
| `RelTypeExecutedBy` | `"EXECUTED_BY"` | ToolExecution → AgentRun |
| `RelTypeDiscovered` | `"DISCOVERED"` | AgentRun → Asset |
| `RelTypeProduced` | `"PRODUCED"` | ToolExecution → Finding |
| `RelTypeMadeCall` | `"MADE_CALL"` | AgentRun → LlmCall |

---

## Attack Techniques

### MITRE ATT&CK Enterprise

The taxonomy includes 24 common MITRE ATT&CK techniques:

```go
// Reconnaissance & Initial Access
graphrag.TechniqueT1190  // Exploit Public-Facing Application
graphrag.TechniqueT1566  // Phishing

// Execution
graphrag.TechniqueT1059  // Command and Scripting Interpreter
graphrag.TechniqueT1203  // Exploitation for Client Execution

// Persistence
graphrag.TechniqueT1053  // Scheduled Task/Job
graphrag.TechniqueT1078  // Valid Accounts

// Privilege Escalation
graphrag.TechniqueT1068  // Exploitation for Privilege Escalation
graphrag.TechniqueT1548  // Abuse Elevation Control Mechanism

// Credential Access
graphrag.TechniqueT1003  // OS Credential Dumping
graphrag.TechniqueT1110  // Brute Force

// Discovery
graphrag.TechniqueT1046  // Network Service Discovery
graphrag.TechniqueT1087  // Account Discovery

// Lateral Movement
graphrag.TechniqueT1021  // Remote Services
graphrag.TechniqueT1570  // Lateral Tool Transfer

// Collection
graphrag.TechniqueT1005  // Data from Local System

// Command and Control
graphrag.TechniqueT1071  // Application Layer Protocol
graphrag.TechniqueT1573  // Encrypted Channel

// Exfiltration
graphrag.TechniqueT1041  // Exfiltration Over C2 Channel

// Impact
graphrag.TechniqueT1486  // Data Encrypted for Impact
graphrag.TechniqueT1499  // Endpoint Denial of Service

// Defense Evasion
graphrag.TechniqueT1027  // Obfuscated Files or Information
graphrag.TechniqueT1070  // Indicator Removal
```

### Arcanum Prompt Injection Taxonomy

For LLM/AI-specific attacks, use the Arcanum PI techniques:

```go
// Evasion Techniques
graphrag.TechniqueARCE001  // Encoding Obfuscation
graphrag.TechniqueARCE002  // Language Switching
graphrag.TechniqueARCE003  // Token Smuggling
graphrag.TechniqueARCE004  // Synonym Substitution

// Intent Categories
graphrag.TechniqueARCI001  // System Manipulation
graphrag.TechniqueARCI002  // Information Extraction
graphrag.TechniqueARCI003  // Jailbreaking

// Attack Techniques
graphrag.TechniqueARCT001  // Framing
graphrag.TechniqueARCT002  // Narrative Smuggling
graphrag.TechniqueARCT003  // Instruction Hierarchy Exploitation
graphrag.TechniqueARCT004  // Payload Splitting
graphrag.TechniqueARCT005  // Delimiter Confusion
graphrag.TechniqueARCT006  // Prompt Injection via External Content

// Vectors
graphrag.TechniqueARCV001  // Direct User Input
graphrag.TechniqueARCV002  // Retrieval Augmentation Poisoning
graphrag.TechniqueARCV003  // Tool/Function Call Injection
graphrag.TechniqueARCV004  // Image-based Injection
```

---

## Using the Taxonomy in Agents

### Importing the Package

```go
import "github.com/zero-day-ai/sdk/graphrag"
```

### Checking Taxonomy Version

```go
// Get the taxonomy version (matches Gibson version)
version := graphrag.TaxonomyVersion  // e.g., "0.15.0"
```

### Validating Types at Runtime

```go
// Check if a type is canonical
tax := graphrag.Taxonomy()
if tax != nil {
    if tax.IsCanonicalNodeType("domain") {
        // It's a standard type
    }

    if tax.IsCanonicalRelationType("AFFECTS") {
        // It's a standard relationship
    }
}
```

### Using Validation Helpers

```go
// These log warnings for non-canonical types but don't block
graphrag.ValidateAndWarnNodeType("my_custom_type")
graphrag.ValidateAndWarnRelationType("CUSTOM_REL")

// Create nodes with automatic validation
node := graphrag.NewNodeWithValidation("custom_type")  // Logs warning

// Create relationships with automatic validation
rel := graphrag.NewRelationshipWithValidation(fromID, toID, "CUSTOM_REL")  // Logs warning
```

---

## Creating Nodes

### Basic Node Creation

```go
// Using canonical constants (recommended)
domain := graphrag.NewGraphNode(graphrag.NodeTypeDomain)
domain.SetProperty(graphrag.PropName, "example.com")
domain.SetProperty("registrar", "GoDaddy")

// With validation (logs warning if non-canonical)
custom := graphrag.NewNodeWithValidation("custom_asset_type")
```

### Asset Discovery Example

```go
func discoverAssets(ctx context.Context, h agent.Harness, target string) error {
    // Create domain node
    domain := graphrag.NewGraphNode(graphrag.NodeTypeDomain)
    domain.SetProperty(graphrag.PropName, target)

    // Discover subdomains
    subdomains := discoverSubdomains(target)
    for _, sub := range subdomains {
        subdomain := graphrag.NewGraphNode(graphrag.NodeTypeSubdomain)
        subdomain.SetProperty(graphrag.PropName, sub)

        // Link subdomain to domain
        rel := graphrag.NewRelationship(domain.ID, subdomain.ID, graphrag.RelTypeHasSubdomain)

        // Store in GraphRAG
        h.Memory().GraphRAG().AddNode(ctx, domain)
        h.Memory().GraphRAG().AddNode(ctx, subdomain)
        h.Memory().GraphRAG().AddRelationship(ctx, rel)
    }

    return nil
}
```

### Finding Creation Example

```go
func reportVulnerability(ctx context.Context, h agent.Harness, vuln VulnInfo) error {
    // Create finding node
    finding := graphrag.NewGraphNode(graphrag.NodeTypeFinding)
    finding.SetProperty(graphrag.PropTitle, vuln.Title)
    finding.SetProperty(graphrag.PropDescription, vuln.Description)
    finding.SetProperty(graphrag.PropSeverity, vuln.Severity)
    finding.SetProperty(graphrag.PropConfidence, vuln.Confidence)

    // Create evidence node
    evidence := graphrag.NewGraphNode(graphrag.NodeTypeEvidence)
    evidence.SetProperty("type", "http_response")
    evidence.SetProperty("content", vuln.Response)

    // Link finding to evidence
    hasEvidence := graphrag.NewRelationship(finding.ID, evidence.ID, graphrag.RelTypeHasEvidence)

    // Link finding to affected asset
    affects := graphrag.NewRelationship(finding.ID, vuln.AssetID, graphrag.RelTypeAffects)

    // Link finding to technique
    usesTechnique := graphrag.NewRelationship(finding.ID, vuln.TechniqueID, graphrag.RelTypeUsesTechnique)

    // Store everything
    gr := h.Memory().GraphRAG()
    gr.AddNode(ctx, finding)
    gr.AddNode(ctx, evidence)
    gr.AddRelationship(ctx, hasEvidence)
    gr.AddRelationship(ctx, affects)
    gr.AddRelationship(ctx, usesTechnique)

    return nil
}
```

---

## Creating Relationships

### Basic Relationship Creation

```go
// Using canonical constants
rel := graphrag.NewRelationship(fromID, toID, graphrag.RelTypeAffects)

// With properties
rel.SetProperty("discovered_at", time.Now().Unix())
rel.SetProperty("confidence", 0.95)
```

### Building Asset Hierarchies

```go
// Domain → Subdomain → Host → Port → Service → Endpoint
//
// example.com
//     └── api.example.com
//             └── 192.168.1.100
//                     └── :443
//                             └── nginx/1.19
//                                     └── /api/v1/users

func buildAssetHierarchy(ctx context.Context, gr graphrag.Store) {
    domain := graphrag.NewGraphNode(graphrag.NodeTypeDomain)
    domain.SetProperty(graphrag.PropName, "example.com")

    subdomain := graphrag.NewGraphNode(graphrag.NodeTypeSubdomain)
    subdomain.SetProperty(graphrag.PropName, "api.example.com")

    host := graphrag.NewGraphNode(graphrag.NodeTypeHost)
    host.SetProperty(graphrag.PropIP, "192.168.1.100")

    port := graphrag.NewGraphNode(graphrag.NodeTypePort)
    port.SetProperty(graphrag.PropPort, 443)
    port.SetProperty(graphrag.PropProtocol, "tcp")

    service := graphrag.NewGraphNode(graphrag.NodeTypeService)
    service.SetProperty(graphrag.PropName, "nginx")
    service.SetProperty("version", "1.19")

    endpoint := graphrag.NewGraphNode(graphrag.NodeTypeEndpoint)
    endpoint.SetProperty(graphrag.PropURL, "/api/v1/users")
    endpoint.SetProperty(graphrag.PropMethod, "GET")

    // Create hierarchy relationships
    gr.AddRelationship(ctx, graphrag.NewRelationship(domain.ID, subdomain.ID, graphrag.RelTypeHasSubdomain))
    gr.AddRelationship(ctx, graphrag.NewRelationship(subdomain.ID, host.ID, graphrag.RelTypeResolvesTo))
    gr.AddRelationship(ctx, graphrag.NewRelationship(host.ID, port.ID, graphrag.RelTypeHasPort))
    gr.AddRelationship(ctx, graphrag.NewRelationship(port.ID, service.ID, graphrag.RelTypeRunsService))
    gr.AddRelationship(ctx, graphrag.NewRelationship(service.ID, endpoint.ID, graphrag.RelTypeHasEndpoint))
}
```

---

## Building Attack Chains

Use the `LEADS_TO` relationship to model multi-step attacks:

```go
func buildAttackChain(ctx context.Context, gr graphrag.Store, findings []Finding) {
    // Finding 1: Information Disclosure → Finding 2: Auth Bypass → Finding 3: RCE

    for i := 0; i < len(findings)-1; i++ {
        current := findings[i]
        next := findings[i+1]

        // Create LEADS_TO relationship with sequence number
        rel := graphrag.NewRelationship(current.ID, next.ID, graphrag.RelTypeLeadsTo)
        rel.SetProperty("sequence", i+1)
        rel.SetProperty("description", fmt.Sprintf("Step %d enables step %d", i+1, i+2))

        gr.AddRelationship(ctx, rel)
    }
}

// Example: SQL Injection → Credential Dump → Lateral Movement
func documentExploitChain(ctx context.Context, h agent.Harness) {
    gr := h.Memory().GraphRAG()

    // Step 1: SQL Injection
    sqli := graphrag.NewGraphNode(graphrag.NodeTypeFinding)
    sqli.SetProperty(graphrag.PropTitle, "SQL Injection")
    sqli.SetProperty(graphrag.PropSeverity, "high")

    // Step 2: Credential Dump
    creds := graphrag.NewGraphNode(graphrag.NodeTypeFinding)
    creds.SetProperty(graphrag.PropTitle, "Credential Extraction")
    creds.SetProperty(graphrag.PropSeverity, "critical")

    // Step 3: Lateral Movement
    lateral := graphrag.NewGraphNode(graphrag.NodeTypeFinding)
    lateral.SetProperty(graphrag.PropTitle, "Lateral Movement to Admin Server")
    lateral.SetProperty(graphrag.PropSeverity, "critical")

    // Chain them together
    chain1 := graphrag.NewRelationship(sqli.ID, creds.ID, graphrag.RelTypeLeadsTo)
    chain1.SetProperty("sequence", 1)

    chain2 := graphrag.NewRelationship(creds.ID, lateral.ID, graphrag.RelTypeLeadsTo)
    chain2.SetProperty("sequence", 2)

    // Link to MITRE techniques
    gr.AddRelationship(ctx, graphrag.NewRelationship(sqli.ID, graphrag.TechniqueT1190, graphrag.RelTypeUsesTechnique))
    gr.AddRelationship(ctx, graphrag.NewRelationship(creds.ID, graphrag.TechniqueT1003, graphrag.RelTypeUsesTechnique))
    gr.AddRelationship(ctx, graphrag.NewRelationship(lateral.ID, graphrag.TechniqueT1021, graphrag.RelTypeUsesTechnique))
}
```

---

## Querying the Graph

### Query Scopes

GraphRAG queries can be scoped to different levels:

```go
// Current run only
results, _ := gr.Query(ctx, query, graphrag.ScopeCurrentRun)

// Same mission (all runs)
results, _ := gr.Query(ctx, query, graphrag.ScopeSameMission)

// All historical data
results, _ := gr.Query(ctx, query, graphrag.ScopeAll)
```

### Finding Related Nodes

```go
// Find all subdomains of a domain
func getSubdomains(ctx context.Context, gr graphrag.Store, domainID string) ([]graphrag.Node, error) {
    query := fmt.Sprintf(`
        MATCH (d)-[:%s]->(s:%s)
        WHERE d.id = $domainID
        RETURN s
    `, graphrag.RelTypeHasSubdomain, graphrag.NodeTypeSubdomain)

    return gr.Query(ctx, query, graphrag.ScopeSameMission, map[string]any{
        "domainID": domainID,
    })
}

// Find all findings affecting a specific host
func getFindingsForHost(ctx context.Context, gr graphrag.Store, hostID string) ([]graphrag.Node, error) {
    query := fmt.Sprintf(`
        MATCH (f:%s)-[:%s]->(h:%s)
        WHERE h.id = $hostID
        RETURN f
    `, graphrag.NodeTypeFinding, graphrag.RelTypeAffects, graphrag.NodeTypeHost)

    return gr.Query(ctx, query, graphrag.ScopeSameMission, map[string]any{
        "hostID": hostID,
    })
}

// Find attack chains
func getAttackChains(ctx context.Context, gr graphrag.Store) ([][]graphrag.Node, error) {
    query := fmt.Sprintf(`
        MATCH path = (start:%s)-[:%s*]->(end:%s)
        WHERE NOT ()-[:%s]->(start)
        RETURN path
    `, graphrag.NodeTypeFinding, graphrag.RelTypeLeadsTo, graphrag.NodeTypeFinding, graphrag.RelTypeLeadsTo)

    return gr.QueryPaths(ctx, query, graphrag.ScopeSameMission)
}
```

---

## Custom Types

You can use custom node and relationship types, but canonical types are preferred:

```go
// Custom types work but generate warnings
customNode := graphrag.NewNodeWithValidation("kubernetes_pod")
// WARNING: Node type 'kubernetes_pod' is not in the canonical taxonomy

// Better: Request the type be added to the taxonomy
// Or use the closest canonical type with additional properties
host := graphrag.NewGraphNode(graphrag.NodeTypeHost)
host.SetProperty("host_type", "kubernetes_pod")
host.SetProperty("pod_name", "nginx-12345")
host.SetProperty("namespace", "default")
```

### Extending the Taxonomy

To add custom types to your Gibson deployment:

1. Create a custom taxonomy YAML file
2. Run Gibson with `--taxonomy-path /path/to/custom.yaml`
3. Custom types are merged with bundled types (additive only)

```yaml
# custom-taxonomy.yaml
version: "0.15.0+custom"
nodes:
  - id: kubernetes_pod
    name: Kubernetes Pod
    category: assets
    description: A Kubernetes pod resource
    properties:
      - name: pod_name
        type: string
        required: true
      - name: namespace
        type: string
        required: true
```

---

## Best Practices

### 1. Always Use Canonical Types

```go
// Good - uses canonical constant
node := graphrag.NewGraphNode(graphrag.NodeTypeDomain)

// Avoid - hardcoded string
node := graphrag.NewGraphNode("domain")

// Bad - custom type without good reason
node := graphrag.NewGraphNode("my_domain_type")
```

### 2. Use Property Constants

```go
// Good
node.SetProperty(graphrag.PropName, "example.com")
node.SetProperty(graphrag.PropSeverity, "high")

// Avoid - hardcoded strings
node.SetProperty("name", "example.com")
```

### 3. Link Findings to Techniques

```go
// Always associate findings with attack techniques
finding := graphrag.NewGraphNode(graphrag.NodeTypeFinding)
finding.SetProperty(graphrag.PropTitle, "Prompt Injection")

// Link to Arcanum PI technique
gr.AddRelationship(ctx, graphrag.NewRelationship(
    finding.ID,
    graphrag.TechniqueARCT001,  // Framing
    graphrag.RelTypeUsesTechnique,
))
```

### 4. Build Complete Hierarchies

```go
// Don't create orphan nodes - always link to parents
subdomain := graphrag.NewGraphNode(graphrag.NodeTypeSubdomain)
// Always link to parent domain
gr.AddRelationship(ctx, graphrag.NewRelationship(
    domainID, subdomain.ID, graphrag.RelTypeHasSubdomain,
))
```

### 5. Include Evidence

```go
// Always attach evidence to findings
finding := graphrag.NewGraphNode(graphrag.NodeTypeFinding)
evidence := graphrag.NewGraphNode(graphrag.NodeTypeEvidence)
evidence.SetProperty("type", "http_request")
evidence.SetProperty("content", requestDump)

gr.AddRelationship(ctx, graphrag.NewRelationship(
    finding.ID, evidence.ID, graphrag.RelTypeHasEvidence,
))
```

---

## Dynamic/LLM-Driven Usage

Agents don't need to hardcode taxonomy knowledge - they can query it at runtime and let the LLM structure data dynamically.

### Getting the Taxonomy as a Prompt

```go
import "github.com/zero-day-ai/sdk/graphrag"

func executeWithDynamicTaxonomy(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    // Get taxonomy as a prompt-friendly string
    taxonomyPrompt := graphrag.TaxonomyPrompt()

    // Let LLM decide how to structure findings
    resp, _ := h.Complete(ctx, "primary", []llm.Message{
        {Role: "system", Content: `You are a security agent. Use this taxonomy to structure your findings:

` + taxonomyPrompt + `

When you discover assets or vulnerabilities, output JSON like:
{
  "nodes": [
    {"type": "domain", "properties": {"name": "example.com"}},
    {"type": "finding", "properties": {"title": "XSS", "severity": "high"}}
  ],
  "relationships": [
    {"from": 1, "to": 0, "type": "AFFECTS"}
  ]
}
`},
        {Role: "user", Content: task.Goal},
    })

    // Parse LLM output and save to GraphRAG
    var graph struct {
        Nodes         []map[string]any `json:"nodes"`
        Relationships []struct {
            From int    `json:"from"`
            To   int    `json:"to"`
            Type string `json:"type"`
        } `json:"relationships"`
    }
    json.Unmarshal([]byte(resp.Content), &graph)

    gr := h.Memory().GraphRAG()
    nodeIDs := make([]string, len(graph.Nodes))

    // Create nodes
    for i, n := range graph.Nodes {
        nodeType := n["type"].(string)
        props := n["properties"].(map[string]any)

        node := graphrag.NewNodeWithValidation(nodeType)
        for k, v := range props {
            node.SetProperty(k, v)
        }
        gr.AddNode(ctx, node)
        nodeIDs[i] = node.ID
    }

    // Create relationships
    for _, r := range graph.Relationships {
        rel := graphrag.NewRelationshipWithValidation(
            nodeIDs[r.From],
            nodeIDs[r.To],
            r.Type,
        )
        gr.AddRelationship(ctx, rel)
    }

    return agent.NewSuccessResult("Structured data saved to graph"), nil
}
```

### Getting the Taxonomy as JSON

For tool use or structured responses:

```go
// Get taxonomy as JSON for tool responses
taxonomyJSON := graphrag.TaxonomyJSON()

// Returns:
// {
//   "version": "0.15.0",
//   "node_types": [
//     {"type": "domain", "name": "Domain", "category": "assets", "description": "..."},
//     ...
//   ],
//   "relationship_types": [
//     {"type": "HAS_SUBDOMAIN", "from_types": ["domain"], "to_types": ["subdomain"], ...},
//     ...
//   ],
//   "techniques": [
//     {"id": "T1190", "name": "Exploit Public-Facing Application", "taxonomy": "mitre"},
//     {"id": "ARC-T001", "name": "Framing", "taxonomy": "arcanum"},
//     ...
//   ]
// }
```

### Full Introspection

For detailed taxonomy access:

```go
intro := graphrag.TaxonomyIntrospect()
if intro != nil {
    // List all node types
    nodeTypes := intro.NodeTypes()  // []string{"domain", "subdomain", ...}

    // Get detailed info about a specific type
    info := intro.NodeTypeInfo("finding")
    // info.Type = "finding"
    // info.Category = "findings"
    // info.Description = "Security finding, vulnerability, or security issue..."
    // info.Properties = [{Name: "title", Type: "string", Required: true}, ...]

    // List all techniques
    mitreTechniques := intro.TechniqueIDs("mitre")
    arcanumTechniques := intro.TechniqueIDs("arcanum")

    // Get technique details
    techInfo := intro.TechniqueInfo("T1190")
    // techInfo.ID = "T1190"
    // techInfo.Name = "Exploit Public-Facing Application"
    // techInfo.Taxonomy = "mitre"
}
```

### Example: Fully Dynamic Recon Agent

This agent has zero hardcoded taxonomy knowledge - everything comes from runtime introspection:

```go
func executeDynamicRecon(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    target := h.Target().URL
    gr := h.Memory().GraphRAG()

    // Run tools and collect raw data
    dnsResult, _ := h.CallTool(ctx, "dns-lookup", map[string]any{"domain": target})
    httpResult, _ := h.CallTool(ctx, "http-request", map[string]any{"url": target})

    // Get taxonomy prompt
    taxonomyPrompt := graphrag.TaxonomyPrompt()

    // Let LLM analyze and structure
    systemPrompt := fmt.Sprintf(`You are a security reconnaissance agent.

%s

Analyze the tool results and structure them according to the taxonomy above.
Output valid JSON with "nodes" and "relationships" arrays.
Use the exact type names from the taxonomy.
`, taxonomyPrompt)

    resp, err := h.Complete(ctx, "primary", []llm.Message{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: fmt.Sprintf(`Target: %s

DNS Results:
%v

HTTP Results:
%v

Structure these findings using the GraphRAG taxonomy.`, target, dnsResult, httpResult)},
    })
    if err != nil {
        return agent.NewFailedResult(err), err
    }

    // Parse and save (same as above)
    // ...

    return agent.NewSuccessResult(resp.Content), nil
}
```

This approach lets you:
- Write generic agents that adapt to taxonomy changes
- Let LLMs decide the best way to structure data
- Avoid maintaining hardcoded constants in agent code
- Support custom taxonomies without code changes

---

## Complete Example

Here's a complete agent that uses the taxonomy to document a security assessment:

```go
package main

import (
    "context"
    "fmt"

    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/graphrag"
    "github.com/zero-day-ai/sdk/llm"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    reconAgent, _ := sdk.NewAgent(
        sdk.WithName("recon-agent"),
        sdk.WithVersion("1.0.0"),
        sdk.WithDescription("Web reconnaissance agent with GraphRAG integration"),

        sdk.WithLLMSlot("primary", llm.SlotRequirements{
            MinContextWindow: 100000,
            RequiredFeatures: []string{"tool_use"},
        }),

        sdk.WithTools("http-request", "dns-lookup", "subdomain-enum"),
        sdk.WithExecuteFunc(executeRecon),
    )

    serve.Agent(reconAgent, serve.WithPort(50051))
}

func executeRecon(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    target := h.Target().URL
    gr := h.Memory().GraphRAG()
    logger := h.Logger()

    // Create mission context in graph
    mission := graphrag.NewGraphNode(graphrag.NodeTypeMission)
    mission.SetProperty(graphrag.PropName, task.Goal)
    mission.SetProperty("target", target)
    gr.AddNode(ctx, mission)

    // Create agent run node
    agentRun := graphrag.NewGraphNode(graphrag.NodeTypeAgentRun)
    agentRun.SetProperty("agent", h.Mission().AgentName)
    agentRun.SetProperty("started_at", time.Now().Unix())
    gr.AddNode(ctx, agentRun)
    gr.AddRelationship(ctx, graphrag.NewRelationship(agentRun.ID, mission.ID, graphrag.RelTypePartOf))

    // Create root domain
    domain := graphrag.NewGraphNode(graphrag.NodeTypeDomain)
    domain.SetProperty(graphrag.PropName, target)
    gr.AddNode(ctx, domain)
    gr.AddRelationship(ctx, graphrag.NewRelationship(agentRun.ID, domain.ID, graphrag.RelTypeDiscovered))

    // Enumerate subdomains
    logger.Info("enumerating subdomains", "target", target)
    subdomainsResult, err := h.CallTool(ctx, "subdomain-enum", map[string]any{
        "domain": target,
    })
    if err != nil {
        return agent.NewFailedResult(err), err
    }

    // Record tool execution
    toolExec := graphrag.NewGraphNode(graphrag.NodeTypeToolExecution)
    toolExec.SetProperty("tool", "subdomain-enum")
    toolExec.SetProperty("input", target)
    gr.AddNode(ctx, toolExec)
    gr.AddRelationship(ctx, graphrag.NewRelationship(toolExec.ID, agentRun.ID, graphrag.RelTypeExecutedBy))

    // Process discovered subdomains
    subdomains := subdomainsResult["subdomains"].([]string)
    for _, sub := range subdomains {
        subdomain := graphrag.NewGraphNode(graphrag.NodeTypeSubdomain)
        subdomain.SetProperty(graphrag.PropName, sub)
        gr.AddNode(ctx, subdomain)
        gr.AddRelationship(ctx, graphrag.NewRelationship(domain.ID, subdomain.ID, graphrag.RelTypeHasSubdomain))
        gr.AddRelationship(ctx, graphrag.NewRelationship(agentRun.ID, subdomain.ID, graphrag.RelTypeDiscovered))

        // DNS lookup for each subdomain
        dnsResult, err := h.CallTool(ctx, "dns-lookup", map[string]any{"domain": sub})
        if err != nil {
            continue
        }

        if ip, ok := dnsResult["ip"].(string); ok {
            host := graphrag.NewGraphNode(graphrag.NodeTypeHost)
            host.SetProperty(graphrag.PropIP, ip)
            gr.AddNode(ctx, host)
            gr.AddRelationship(ctx, graphrag.NewRelationship(subdomain.ID, host.ID, graphrag.RelTypeResolvesTo))

            // Check for common ports
            for _, port := range []int{80, 443, 8080, 8443} {
                httpResult, err := h.CallTool(ctx, "http-request", map[string]any{
                    "url":    fmt.Sprintf("https://%s:%d", ip, port),
                    "method": "GET",
                })
                if err != nil {
                    continue
                }

                if httpResult["status_code"].(int) == 200 {
                    // Create port and service nodes
                    portNode := graphrag.NewGraphNode(graphrag.NodeTypePort)
                    portNode.SetProperty(graphrag.PropPort, port)
                    portNode.SetProperty(graphrag.PropProtocol, "tcp")
                    gr.AddNode(ctx, portNode)
                    gr.AddRelationship(ctx, graphrag.NewRelationship(host.ID, portNode.ID, graphrag.RelTypeHasPort))

                    // Check for information disclosure
                    headers := httpResult["headers"].(map[string]string)
                    if server, ok := headers["Server"]; ok {
                        // Create technology node
                        tech := graphrag.NewGraphNode(graphrag.NodeTypeTechnology)
                        tech.SetProperty(graphrag.PropName, server)
                        gr.AddNode(ctx, tech)

                        // Check if it's a vulnerable version
                        if isVulnerableVersion(server) {
                            finding := graphrag.NewGraphNode(graphrag.NodeTypeFinding)
                            finding.SetProperty(graphrag.PropTitle, "Server Version Disclosure")
                            finding.SetProperty(graphrag.PropDescription,
                                fmt.Sprintf("Server header reveals potentially vulnerable version: %s", server))
                            finding.SetProperty(graphrag.PropSeverity, "medium")
                            finding.SetProperty(graphrag.PropConfidence, 85)
                            gr.AddNode(ctx, finding)

                            // Link finding to affected asset
                            gr.AddRelationship(ctx, graphrag.NewRelationship(
                                finding.ID, host.ID, graphrag.RelTypeAffects))

                            // Link to MITRE technique
                            gr.AddRelationship(ctx, graphrag.NewRelationship(
                                finding.ID, graphrag.TechniqueT1046, graphrag.RelTypeUsesTechnique))

                            // Create evidence
                            evidence := graphrag.NewGraphNode(graphrag.NodeTypeEvidence)
                            evidence.SetProperty("type", "http_header")
                            evidence.SetProperty("header", "Server")
                            evidence.SetProperty("value", server)
                            gr.AddNode(ctx, evidence)
                            gr.AddRelationship(ctx, graphrag.NewRelationship(
                                finding.ID, evidence.ID, graphrag.RelTypeHasEvidence))

                            // Tool execution produced the finding
                            gr.AddRelationship(ctx, graphrag.NewRelationship(
                                toolExec.ID, finding.ID, graphrag.RelTypeProduced))
                        }
                    }
                }
            }
        }
    }

    logger.Info("reconnaissance complete",
        "subdomains_found", len(subdomains),
        "target", target)

    return agent.NewSuccessResult(fmt.Sprintf("Discovered %d subdomains for %s", len(subdomains), target)), nil
}

func isVulnerableVersion(server string) bool {
    // Check against known vulnerable versions
    // Implementation omitted for brevity
    return false
}
```

---

## Summary

The GraphRAG taxonomy provides:

1. **19 Node Types**: Assets, findings, execution context, and attack concepts
2. **20 Relationship Types**: Asset hierarchy, finding links, and execution context
3. **43 Attack Techniques**: MITRE ATT&CK Enterprise + Arcanum Prompt Injection
4. **Type Validation**: Warnings for non-canonical types (doesn't block)
5. **Property Constants**: Standardized property names

Always prefer canonical types for consistency across agents and missions. The taxonomy is versioned with Gibson releases, ensuring all agents in a deployment share the same vocabulary.

For the complete list of types and their properties, see the [YAML taxonomy files](https://github.com/zero-day-ai/gibson/tree/main/internal/graphrag/taxonomy).
