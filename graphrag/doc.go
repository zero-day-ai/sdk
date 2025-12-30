// Package graphrag provides Graph-based Retrieval-Augmented Generation (GraphRAG) capabilities
// for the Gibson AI Security Testing Framework.
//
// This package enables agents to store, query, and traverse structured knowledge graphs that
// combine semantic embeddings with graph relationships. GraphRAG extends traditional RAG by
// adding graph-based context propagation, relationship-aware retrieval, and multi-hop reasoning.
//
// # Core Capabilities
//
// GraphRAG provides several key capabilities:
//   - Vector similarity search with configurable embedding models
//   - Graph traversal with relationship-aware context propagation
//   - Hybrid scoring combining semantic and structural relevance
//   - Node type filtering for domain-specific queries
//   - Multi-hop reasoning through graph relationships
//   - Mission-scoped knowledge isolation
//   - Batch operations for efficient bulk storage
//   - Bidirectional relationship support
//
// # Core Types
//
// The package provides the following types:
//
//   - GraphNode: Represents a node in the knowledge graph with properties and content
//   - Query: Fluent builder for creating GraphRAG queries with filtering options
//   - Relationship: Represents connections between nodes with optional properties
//   - Batch: Collection of nodes and relationships for bulk operations
//   - TraversalOptions: Configuration for graph traversal operations
//   - Result: Query results with hybrid scoring (vector + graph scores)
//   - TraversalResult: Results from graph traversal with path information
//
// Domain-specific types for security operations:
//
//   - AttackPattern: MITRE ATT&CK patterns with similarity scoring
//   - FindingNode: Security findings with severity and confidence scores
//   - AttackChain: Sequences of attack steps forming attack chains
//   - AttackStep: Individual steps in an attack chain
//
// # Creating and Storing Nodes
//
// Create nodes using the fluent GraphNode builder:
//
//	// Create a simple node
//	node := graphrag.NewGraphNode("finding").
//	    WithID("finding-123").
//	    WithContent("SQL injection vulnerability in login endpoint").
//	    WithProperty("severity", "high").
//	    WithProperty("confidence", 0.95)
//
//	// Validate before use
//	if err := node.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a technique node
//	technique := graphrag.NewGraphNode("technique").
//	    WithID("T1190").
//	    WithContent("Exploit Public-Facing Application").
//	    WithProperty("tactic", "Initial Access").
//	    WithProperty("platform", "Web")
//
// The MissionID and AgentName fields are auto-populated by the Gibson harness.
//
// # Query Operations
//
// Create queries using the fluent Query builder:
//
//	query := graphrag.NewQuery("What vulnerabilities were found in the authentication module?").
//	    WithTopK(5).
//	    WithMaxHops(2).
//	    WithMinScore(0.7).
//	    WithNodeTypes("finding", "technique").
//	    WithMission("mission-123")
//
//	// Or use pre-computed embeddings
//	query := graphrag.NewQueryFromEmbedding(embedding).
//	    WithWeights(0.6, 0.4)  // 60% vector, 40% graph
//
//	// Always validate queries before execution
//	if err := query.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
// Query parameters:
//   - TopK: Number of results to return (default: 10)
//   - MaxHops: Maximum graph traversal depth (default: 3)
//   - MinScore: Minimum similarity threshold 0.0-1.0 (default: 0.7)
//   - NodeTypes: Filter by specific node types (optional)
//   - MissionID: Filter by mission context (optional)
//   - VectorWeight: Weight for semantic similarity (default: 0.6)
//   - GraphWeight: Weight for graph structure (default: 0.4)
//
// The weights must sum to 1.0 for proper hybrid scoring.
//
// # Relationship Management
//
// Create and manage graph relationships:
//
//	// Simple unidirectional relationship
//	rel := graphrag.NewRelationship(
//	    "finding-123",
//	    "technique-T1190",
//	    "USES_TECHNIQUE",
//	).WithProperty("confidence", 0.95)
//
//	// Bidirectional relationship
//	rel := graphrag.NewRelationship(
//	    "finding-123",
//	    "finding-456",
//	    "SIMILAR_TO",
//	).WithProperty("similarity", 0.87).
//	  WithBidirectional(true)
//
//	// Validate before use
//	if err := rel.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
// Common relationship types:
//   - ELICITED: Technique successfully triggered behavior
//   - PART_OF: Component or hierarchical relationships
//   - SIMILAR_TO: Semantic similarity relationships
//   - USES_TECHNIQUE: Finding uses specific technique
//   - TARGETS: Technique targets specific component
//   - REFERENCES: Cross-reference relationships
//   - EXPLOITS: Finding exploits a vulnerability
//   - MITIGATES: Control mitigates a risk
//
// # Batch Operations
//
// Use Batch for efficient bulk storage of nodes and relationships:
//
//	batch := graphrag.NewBatch().
//	    AddNode(*finding1).
//	    AddNode(*finding2).
//	    AddNode(*technique).
//	    AddRelationship(*rel1).
//	    AddRelationship(*rel2)
//
//	// Submit batch to GraphRAG storage
//	// (actual storage operations are handled by the Gibson framework)
//
// Batch operations are more efficient than individual operations when
// creating multiple nodes and relationships simultaneously.
//
// # Graph Traversal
//
// Configure graph traversal with TraversalOptions:
//
//	opts := graphrag.NewTraversalOptions().
//	    WithMaxDepth(3).
//	    WithRelationshipTypes([]string{"USES_TECHNIQUE", "SIMILAR_TO"}).
//	    WithNodeTypes([]string{"finding", "technique"}).
//	    WithDirection("both")
//
// Traversal directions:
//   - "outgoing": Follow relationships from source to target (default)
//   - "incoming": Follow relationships from target to source
//   - "both": Follow relationships in both directions
//
// # Node Types
//
// Common node types in GraphRAG:
//   - finding: Security findings and vulnerabilities
//   - technique: Attack techniques and methods (MITRE ATT&CK)
//   - target: System components being tested
//   - agent: Agent executions and actions
//   - conversation: LLM conversation threads
//   - tool: Tool invocations and results
//   - evidence: Evidence artifacts and data
//   - mitigation: Security controls and mitigations
//   - asset: System assets and resources
//
// # Hybrid Scoring
//
// GraphRAG combines two scoring dimensions:
//
//  1. Vector Score: Semantic similarity via embeddings
//     - Measures conceptual relevance
//     - Uses cosine similarity on embedding vectors
//
//  2. Graph Score: Structural relevance via relationships
//     - Measures connectedness and relationship strength
//     - Incorporates relationship types and properties
//     - Uses multi-hop propagation with decay
//
// Control the balance with WithWeights(vector, graph):
//
//	// Emphasize semantic similarity
//	query.WithWeights(0.8, 0.2)
//
//	// Emphasize graph structure
//	query.WithWeights(0.3, 0.7)
//
//	// Balanced (default)
//	query.WithWeights(0.6, 0.4)
//
// # Error Handling
//
// GraphRAG operations return specific sentinel errors:
//
//	result, err := client.Query(ctx, query)
//	if errors.Is(err, graphrag.ErrGraphRAGNotEnabled) {
//	    // GraphRAG is not configured in the framework
//	}
//	if errors.Is(err, graphrag.ErrNodeNotFound) {
//	    // Referenced node does not exist
//	}
//	if errors.Is(err, graphrag.ErrInvalidQuery) {
//	    // Query validation failed
//	}
//	if errors.Is(err, graphrag.ErrEmbeddingFailed) {
//	    // Embedding generation failed
//	}
//	if errors.Is(err, graphrag.ErrQueryTimeout) {
//	    // Query exceeded timeout
//	}
//	if errors.Is(err, graphrag.ErrStorageFailed) {
//	    // Storage operation failed
//	}
//	if errors.Is(err, graphrag.ErrRelationshipFailed) {
//	    // Relationship operation failed
//	}
//
// Always use errors.Is() for error checking to properly handle wrapped errors.
//
// # Mission Isolation
//
// GraphRAG supports mission-scoped knowledge graphs:
//
//	// Store finding in mission context
//	node := graphrag.NewGraphNode("finding").
//	    WithID("finding-123").
//	    WithContent("SQL injection vulnerability found")
//	// MissionID is auto-populated by the Gibson harness
//
//	// Query within mission context
//	query := graphrag.NewQuery("similar findings").
//	    WithMission("mission-123")
//
// This ensures knowledge isolation between different testing missions
// while allowing cross-mission analysis when needed.
//
// # Common Patterns for Security Agents
//
// ## Storing Attack Data
//
// Store attack attempts and findings:
//
//	// Create finding node
//	finding := graphrag.NewGraphNode("finding").
//	    WithContent("SQL injection in /api/login parameter 'username'").
//	    WithProperty("severity", "critical").
//	    WithProperty("cvss_score", 9.8).
//	    WithProperty("endpoint", "/api/login").
//	    WithProperty("parameter", "username")
//
//	// Create technique node
//	technique := graphrag.NewGraphNode("technique").
//	    WithID("T1190").
//	    WithContent("Exploit Public-Facing Application").
//	    WithProperty("tactic", "Initial Access")
//
//	// Link finding to technique
//	rel := graphrag.NewRelationship(
//	    finding.ID,
//	    "T1190",
//	    "USES_TECHNIQUE",
//	).WithProperty("confidence", 0.95)
//
// ## Querying for Similar Patterns
//
// Find similar attack patterns:
//
//	// Query for similar findings
//	query := graphrag.NewQuery("SQL injection vulnerabilities").
//	    WithTopK(10).
//	    WithNodeTypes("finding").
//	    WithMinScore(0.75).
//	    WithMission("current-mission-id")
//
//	// Query for related techniques
//	query := graphrag.NewQuery("privilege escalation techniques").
//	    WithNodeTypes("technique").
//	    WithMaxHops(2).
//	    WithWeights(0.7, 0.3)  // Emphasize semantic similarity
//
// ## Building Attack Chains
//
// Create attack chain from multiple findings:
//
//	// Create findings
//	recon := graphrag.NewGraphNode("finding").
//	    WithContent("Port scan detected open SSH port")
//
//	bruteforce := graphrag.NewGraphNode("finding").
//	    WithContent("SSH bruteforce successful with weak credentials")
//
//	privesc := graphrag.NewGraphNode("finding").
//	    WithContent("Privilege escalation via sudo misconfiguration")
//
//	// Link findings in sequence
//	rel1 := graphrag.NewRelationship(recon.ID, bruteforce.ID, "LEADS_TO").
//	    WithProperty("sequence", 1)
//
//	rel2 := graphrag.NewRelationship(bruteforce.ID, privesc.ID, "LEADS_TO").
//	    WithProperty("sequence", 2)
//
// ## Traversing the Knowledge Graph
//
// Explore relationships from a starting node:
//
//	opts := graphrag.NewTraversalOptions().
//	    WithMaxDepth(3).
//	    WithRelationshipTypes([]string{"USES_TECHNIQUE", "LEADS_TO"}).
//	    WithDirection("both")
//
//	// Traverse from a finding to discover attack chains and techniques
//	// (traversal execution is handled by the Gibson framework)
//
// # Integration with Gibson
//
// GraphRAG integrates seamlessly with Gibson agents:
//
//	func (a *MyAgent) Execute(ctx context.Context, req sdk.ExecuteRequest) error {
//	    // Store finding in graph
//	    finding := graphrag.NewGraphNode("finding").
//	        WithContent("XSS vulnerability in search parameter").
//	        WithProperty("severity", "medium").
//	        WithProperty("url", req.Target.URL)
//
//	    if err := finding.Validate(); err != nil {
//	        return fmt.Errorf("invalid node: %w", err)
//	    }
//
//	    // Query related findings
//	    query := graphrag.NewQuery("XSS vulnerabilities in search").
//	        WithMission(req.Mission.ID).
//	        WithNodeTypes("finding").
//	        WithTopK(5)
//
//	    if err := query.Validate(); err != nil {
//	        return fmt.Errorf("invalid query: %w", err)
//	    }
//
//	    // Use results to inform decision making
//	    // (actual query execution is handled by the Gibson framework)
//
//	    return nil
//	}
//
// # Performance Considerations
//
// GraphRAG operations are optimized for different query patterns:
//
//   - Vector search: O(n) with approximate nearest neighbor index
//   - Graph traversal: O(d * b^k) where d=degree, b=branching, k=MaxHops
//   - Hybrid scoring: Computed in parallel with result streaming
//
// Best practices:
//   - Use appropriate TopK values (typically 5-20)
//   - Limit MaxHops for large graphs (typically 2-3)
//   - Set MinScore to filter low-quality matches
//   - Use NodeTypes filtering to reduce search space
//   - Cache frequently accessed embeddings
//   - Use Batch for bulk operations
//   - Validate nodes, queries, and relationships before use
//   - Handle errors with errors.Is() for sentinel errors
//
// # Concurrency Safety
//
// All GraphRAG operations are safe for concurrent use by multiple goroutines.
// The underlying storage and indexing layers handle concurrent reads and writes
// with appropriate locking and isolation guarantees.
package graphrag
