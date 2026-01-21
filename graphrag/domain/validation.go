package domain

import (
	"fmt"

	"github.com/zero-day-ai/sdk/graphrag"
)

// RequiresParent maps node types to whether they require a parent relationship.
// This is used by ValidateGraphNode to enforce structural relationships in the
// knowledge graph during mission-scoped storage.
//
// Root nodes (false): Can be stored directly under MissionRun, no parent required
// Child nodes (true): Must declare a parent via BelongsTo() before storage
//
// Design rationale:
//   - Root nodes represent top-level entities discovered during a mission (hosts, domains, cloud accounts, findings)
//   - Child nodes represent hierarchical entities that semantically belong to a parent (port belongs to host)
//   - This prevents orphaned nodes and ensures graph structure integrity
var RequiresParent = map[string]bool{
	// ========== Root Nodes (No Parent Required) ==========
	// These nodes represent top-level entities discovered during missions
	// and are automatically attached to the current MissionRun

	// Core Execution Types - Root level entities
	graphrag.NodeTypeMission:       false, // Top-level mission entity
	graphrag.NodeTypeAgentRun:      false, // Execution runs are root-level
	graphrag.NodeTypeToolExecution: false, // Tool executions are root-level
	graphrag.NodeTypeLlmCall:       false, // LLM calls are root-level
	graphrag.NodeTypeIntelligence:  false, // Intelligence reports are root-level

	// Asset Discovery Types - Root nodes
	graphrag.NodeTypeDomain:      false, // Root domain (e.g., example.com)
	graphrag.NodeTypeHost:        false, // IP addresses are root discovery entities
	graphrag.NodeTypeCertificate: false, // Certificates can be discovered independently
	graphrag.NodeTypeTechnology:  false, // Technologies can be discovered independently
	graphrag.NodeTypeDNSRecord:   false, // DNS records can be discovered independently

	// Network Infrastructure Types - Root nodes
	graphrag.NodeTypeFirewall:         false, // Firewalls are root infrastructure
	graphrag.NodeTypeRouter:           false, // Routers are root infrastructure
	graphrag.NodeTypeLoadBalancer:     false, // Load balancers are root infrastructure
	graphrag.NodeTypeProxy:            false, // Proxies are root infrastructure
	graphrag.NodeTypeVPN:              false, // VPNs are root infrastructure
	graphrag.NodeTypeNetwork:          false, // Networks are root infrastructure
	graphrag.NodeTypeVLAN:             false, // VLANs are root infrastructure
	graphrag.NodeTypeNetworkZone:      false, // Network zones are root security boundaries
	graphrag.NodeTypeNetworkACL:       false, // ACLs are root security policies
	graphrag.NodeTypeNATGateway:       false, // NAT gateways are root infrastructure
	graphrag.NodeTypeBGPPeer:          false, // BGP peers are root routing entities
	graphrag.NodeTypeNetworkInterface: false, // Network interfaces can be discovered independently

	// Web/API Types - Root nodes
	graphrag.NodeTypeApi: false, // APIs are root-level services

	// Identity & Access Types - Root nodes
	graphrag.NodeTypeUser:             false, // Users are root identities
	graphrag.NodeTypeGroup:            false, // Groups are root organizational units
	graphrag.NodeTypeRole:             false, // Roles are root access control entities
	graphrag.NodeTypePolicy:           false, // Policies are root security controls
	graphrag.NodeTypeCredential:       false, // Credentials can be discovered independently
	graphrag.NodeTypeAPIKey:           false, // API keys can be discovered independently
	graphrag.NodeTypeToken:            false, // Tokens can be discovered independently
	graphrag.NodeTypeOAuthClient:      false, // OAuth clients are root entities
	graphrag.NodeTypeSAMLProvider:     false, // SAML providers are root identity sources
	graphrag.NodeTypeIdentityProvider: false, // Identity providers are root auth sources
	graphrag.NodeTypeServiceAccount:   false, // Service accounts are root identities
	graphrag.NodeTypeAccessKey:        false, // Access keys can be discovered independently
	graphrag.NodeTypeMFADevice:        false, // MFA devices can be discovered independently
	graphrag.NodeTypeSession:          false, // Sessions can be discovered independently

	// AI/LLM Types - Root nodes
	graphrag.NodeTypeLLM:               false, // LLMs are root model entities
	graphrag.NodeTypeLLMDeployment:     false, // LLM deployments are root infrastructure
	graphrag.NodeTypePrompt:            false, // Prompts can be discovered independently
	graphrag.NodeTypeSystemPrompt:      false, // System prompts can be discovered independently
	graphrag.NodeTypeGuardrail:         false, // Guardrails are root safety controls
	graphrag.NodeTypeContentFilter:     false, // Content filters are root safety controls
	graphrag.NodeTypeEmbeddingModel:    false, // Embedding models are root entities
	graphrag.NodeTypeModelRegistry:     false, // Model registries are root catalogs
	graphrag.NodeTypeInferenceEndpoint: false, // Inference endpoints are root services
	graphrag.NodeTypeDataset:           false, // Datasets can be discovered independently

	// AI Agent Types - Root nodes
	graphrag.NodeTypeAIAgent:       false, // AI agents are root entities
	graphrag.NodeTypeAgentConfig:   false, // Agent configs can be discovered independently
	graphrag.NodeTypeAgentMemory:   false, // Agent memory stores are root entities
	graphrag.NodeTypeAgentTool:     false, // Agent tools are root capabilities
	graphrag.NodeTypeChain:         false, // Chains are root workflow entities
	graphrag.NodeTypeWorkflow:      false, // Workflows are root automation entities
	graphrag.NodeTypeCrew:          false, // Crews are root multi-agent entities
	graphrag.NodeTypeAgentTask:     false, // Agent tasks can be discovered independently
	graphrag.NodeTypeAgentRole:     false, // Agent roles are root definitions
	graphrag.NodeTypeAgentLoop:     false, // Agent loops are root control structures
	graphrag.NodeTypeAgentArtifact: false, // Agent artifacts can be discovered independently

	// MCP Types - Root nodes
	graphrag.NodeTypeMCPServer:     false, // MCP servers are root services
	graphrag.NodeTypeMCPClient:     false, // MCP clients are root connections
	graphrag.NodeTypeMCPTransport:  false, // MCP transports are root mechanisms
	graphrag.NodeTypeMCPCapability: false, // MCP capabilities can be discovered independently

	// RAG Types - Root nodes
	graphrag.NodeTypeVectorStore:      false, // Vector stores are root data stores
	graphrag.NodeTypeVectorIndex:      false, // Vector indices are root structures
	graphrag.NodeTypeDocument:         false, // Documents are root knowledge entities
	graphrag.NodeTypeKnowledgeBase:    false, // Knowledge bases are root catalogs
	graphrag.NodeTypeRetriever:        false, // Retrievers are root components
	graphrag.NodeTypeRAGPipeline:      false, // RAG pipelines are root workflows
	graphrag.NodeTypeReranker:         false, // Rerankers are root components
	graphrag.NodeTypeChunkingStrategy: false, // Chunking strategies are root configurations

	// Data Types - Root nodes
	graphrag.NodeTypeDatabase:      false, // Databases are root data stores
	graphrag.NodeTypeStorageBucket: false, // Storage buckets are root storage
	graphrag.NodeTypeQueue:         false, // Queues are root messaging systems
	graphrag.NodeTypeTopic:         false, // Topics are root pub/sub entities
	graphrag.NodeTypeStream:        false, // Streams are root data flows
	graphrag.NodeTypeCache:         false, // Caches are root performance layers
	graphrag.NodeTypeSchema:        false, // Schemas can be discovered independently
	graphrag.NodeTypeDataPipeline:  false, // Data pipelines are root ETL workflows

	// Container Types - Root nodes
	graphrag.NodeTypeContainer:         false, // Containers can be discovered independently
	graphrag.NodeTypeContainerImage:    false, // Container images are root artifacts
	graphrag.NodeTypeContainerRegistry: false, // Container registries are root services
	graphrag.NodeTypeDockerfile:        false, // Dockerfiles can be discovered independently

	// Kubernetes Types - Root nodes
	graphrag.NodeTypeK8sCluster:        false, // K8s clusters are root infrastructure
	graphrag.NodeTypeK8sNamespace:      false, // Namespaces are root organizational units
	graphrag.NodeTypeK8sConfigMap:      false, // ConfigMaps can be discovered independently
	graphrag.NodeTypeK8sSecret:         false, // Secrets can be discovered independently
	graphrag.NodeTypeK8sPV:             false, // PersistentVolumes are root storage
	graphrag.NodeTypeK8sServiceAccount: false, // ServiceAccounts can be discovered independently
	graphrag.NodeTypeK8sRole:           false, // Roles can be discovered independently
	graphrag.NodeTypeK8sClusterRole:    false, // ClusterRoles are root RBAC entities
	graphrag.NodeTypeK8sNetworkPolicy:  false, // NetworkPolicies can be discovered independently
	graphrag.NodeTypeK8sLimitRange:     false, // LimitRanges can be discovered independently
	graphrag.NodeTypeK8sResourceQuota:  false, // ResourceQuotas can be discovered independently

	// Cloud Provider Types - Root nodes
	graphrag.NodeTypeCloudAsset:         false, // Cloud assets can be discovered independently
	graphrag.NodeTypeCloudAccount:       false, // Cloud accounts are root organizational entities
	graphrag.NodeTypeCloudVPC:           false, // VPCs are root network boundaries
	graphrag.NodeTypeCloudSecurityGroup: false, // Security groups are root firewall rules
	graphrag.NodeTypeCloudInstance:      false, // Cloud instances can be discovered independently
	graphrag.NodeTypeCloudFunction:      false, // Cloud functions can be discovered independently
	graphrag.NodeTypeCloudStorage:       false, // Cloud storage is root storage
	graphrag.NodeTypeCloudDatabase:      false, // Cloud databases are root data stores
	graphrag.NodeTypeCloudQueue:         false, // Cloud queues are root messaging
	graphrag.NodeTypeCloudAPIGateway:    false, // API gateways are root services
	graphrag.NodeTypeCloudCDN:           false, // CDNs are root delivery networks
	graphrag.NodeTypeCloudDNSZone:       false, // DNS zones are root DNS management
	graphrag.NodeTypeCloudCertificate:   false, // Cloud certificates can be discovered independently
	graphrag.NodeTypeCloudKMSKey:        false, // KMS keys can be discovered independently
	graphrag.NodeTypeCloudIAMRole:       false, // IAM roles are root access control
	graphrag.NodeTypeCloudIAMPolicy:     false, // IAM policies are root policies
	graphrag.NodeTypeCloudTrail:         false, // Cloud trails are root audit logs
	graphrag.NodeTypeCloudMetric:        false, // Cloud metrics can be discovered independently
	graphrag.NodeTypeCloudAlarm:         false, // Cloud alarms can be discovered independently
	graphrag.NodeTypeCloudRegion:        false, // Cloud regions are root geographic entities

	// Security Finding Types - Root nodes
	graphrag.NodeTypeFinding:    false, // Findings are root security discoveries
	graphrag.NodeTypeEvidence:   false, // Evidence can be discovered independently
	graphrag.NodeTypeMitigation: false, // Mitigations can be discovered independently
	graphrag.NodeTypeTactic:     false, // Tactics are root attack categories
	graphrag.NodeTypeTechnique:  false, // Techniques are root attack methods

	// ========== Child Nodes (Parent Required) ==========
	// These nodes represent hierarchical entities that semantically belong to a parent
	// and must declare the parent relationship via BelongsTo() before storage

	// Asset Hierarchy - Child nodes
	graphrag.NodeTypeSubdomain: true, // Subdomains belong to domains
	graphrag.NodeTypePort:      true, // Ports belong to hosts
	graphrag.NodeTypeService:   true, // Services belong to ports

	// Web/API Types - Child nodes
	graphrag.NodeTypeEndpoint:        true, // Endpoints belong to services
	graphrag.NodeTypeAPIEndpoint:     true, // API endpoints belong to APIs or services
	graphrag.NodeTypeParameter:       true, // Parameters belong to endpoints
	graphrag.NodeTypeHeader:          true, // Headers belong to endpoints
	graphrag.NodeTypeCookie:          true, // Cookies belong to endpoints
	graphrag.NodeTypeForm:            true, // Forms belong to endpoints
	graphrag.NodeTypeFormField:       true, // Form fields belong to forms
	graphrag.NodeTypeWebSocket:       true, // WebSockets belong to endpoints
	graphrag.NodeTypeGraphQLSchema:   true, // GraphQL schemas belong to APIs
	graphrag.NodeTypeGraphQLQuery:    true, // GraphQL queries belong to schemas
	graphrag.NodeTypeGraphQLMutation: true, // GraphQL mutations belong to schemas
	graphrag.NodeTypeRESTResource:    true, // REST resources belong to APIs
	graphrag.NodeTypeHTTPMethod:      true, // HTTP methods belong to endpoints
	graphrag.NodeTypeContentType:     true, // Content types belong to endpoints
	graphrag.NodeTypeCORSPolicy:      true, // CORS policies belong to endpoints
	graphrag.NodeTypeRateLimit:       true, // Rate limits belong to endpoints
	graphrag.NodeTypeRequestBody:     true, // Request bodies belong to endpoints

	// Identity & Access Types - Child nodes
	graphrag.NodeTypePermission: true, // Permissions belong to roles
	graphrag.NodeTypeOAuthScope: true, // OAuth scopes belong to OAuth clients

	// AI/LLM Types - Child nodes
	graphrag.NodeTypeLLMResponse:  true, // LLM responses belong to LLM calls
	graphrag.NodeTypeTokenUsage:   true, // Token usage belongs to LLM calls
	graphrag.NodeTypeFineTune:     true, // Fine-tunes belong to base models
	graphrag.NodeTypeModelVersion: true, // Model versions belong to model registries
	graphrag.NodeTypeBatchJob:     true, // Batch jobs belong to inference endpoints
	graphrag.NodeTypeTrainingRun:  true, // Training runs belong to models

	// AI Agent Types - Child nodes
	graphrag.NodeTypeToolCall:      true, // Tool calls belong to agent runs
	graphrag.NodeTypeReasoningStep: true, // Reasoning steps belong to agent runs
	graphrag.NodeTypeMemoryEntry:   true, // Memory entries belong to agent memory
	graphrag.NodeTypePlanningStep:  true, // Planning steps belong to agent runs

	// MCP Types - Child nodes
	graphrag.NodeTypeMCPTool:     true, // MCP tools belong to MCP servers
	graphrag.NodeTypeMCPResource: true, // MCP resources belong to MCP servers
	graphrag.NodeTypeMCPPrompt:   true, // MCP prompts belong to MCP servers
	graphrag.NodeTypeMCPSampling: true, // MCP sampling belongs to MCP clients
	graphrag.NodeTypeMCPRoots:    true, // MCP roots belong to MCP servers

	// RAG Types - Child nodes
	graphrag.NodeTypeDocumentChunk:   true, // Document chunks belong to documents
	graphrag.NodeTypeEmbedding:       true, // Embeddings belong to chunks or documents
	graphrag.NodeTypeRetrievalResult: true, // Retrieval results belong to queries

	// Data Types - Child nodes
	graphrag.NodeTypeTable:           true, // Tables belong to databases
	graphrag.NodeTypeColumn:          true, // Columns belong to tables
	graphrag.NodeTypeIndex:           true, // Indices belong to tables
	graphrag.NodeTypeView:            true, // Views belong to databases
	graphrag.NodeTypeStoredProcedure: true, // Stored procedures belong to databases
	graphrag.NodeTypeTrigger:         true, // Triggers belong to tables
	graphrag.NodeTypeFile:            true, // Files belong to storage buckets or filesystems
	graphrag.NodeTypeObject:          true, // Objects belong to storage buckets

	// Kubernetes Types - Child nodes
	graphrag.NodeTypeK8sPod:                true, // Pods belong to namespaces
	graphrag.NodeTypeK8sDeployment:         true, // Deployments belong to namespaces
	graphrag.NodeTypeK8sService:            true, // Services belong to namespaces
	graphrag.NodeTypeK8sIngress:            true, // Ingresses belong to namespaces
	graphrag.NodeTypeK8sPVC:                true, // PVCs belong to namespaces
	graphrag.NodeTypeK8sStatefulSet:        true, // StatefulSets belong to namespaces
	graphrag.NodeTypeK8sDaemonSet:          true, // DaemonSets belong to namespaces
	graphrag.NodeTypeK8sJob:                true, // Jobs belong to namespaces
	graphrag.NodeTypeK8sCronJob:            true, // CronJobs belong to namespaces
	graphrag.NodeTypeK8sRoleBinding:        true, // RoleBindings belong to namespaces
	graphrag.NodeTypeK8sClusterRoleBinding: true, // ClusterRoleBindings belong to clusters

	// Cloud Provider Types - Child nodes
	graphrag.NodeTypeCloudSubnet:  true, // Subnets belong to VPCs
	graphrag.NodeTypeRoute:        true, // Routes belong to routers or VPCs
	graphrag.NodeTypeFirewallRule: true, // Firewall rules belong to firewalls
}

// ValidateGraphNode validates a node before storage, ensuring that nodes requiring
// parent relationships have declared them via BelongsTo().
//
// This validation is part of the mission-scoped storage implementation (Phase 1, Task 2.2).
// It enforces the graph structure rules defined in RequiresParent.
//
// Returns an error if:
//   - The node type requires a parent but ParentRef() returns nil
//   - The node type is unknown (not in RequiresParent map)
//
// Example error messages:
//   - "port requires a parent. Use node.BelongsTo(parent) to set the parent relationship"
//   - "unknown node type 'invalid_type' - not found in taxonomy"
//
// Usage:
//
//	port := domain.NewPort(443, "tcp")
//	err := domain.ValidateGraphNode(port)
//	// Returns error: "port requires a parent..."
//
//	port.BelongsTo(host)
//	err = domain.ValidateGraphNode(port)
//	// Returns nil - validation passes
func ValidateGraphNode(node GraphNode) error {
	nodeType := node.NodeType()

	// Check if node type exists in taxonomy
	requiresParent, exists := RequiresParent[nodeType]
	if !exists {
		return fmt.Errorf("unknown node type '%s' - not found in taxonomy", nodeType)
	}

	// If node requires a parent, ensure ParentRef() is set
	if requiresParent {
		if node.ParentRef() == nil {
			return fmt.Errorf(
				"%s requires a parent. Use node.BelongsTo(parent) to set the parent relationship",
				nodeType,
			)
		}
	}

	return nil
}
