package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/graphrag"
)

// TestRequiresParent_Coverage verifies that all node types from the taxonomy
// are accounted for in the RequiresParent map.
func TestRequiresParent_Coverage(t *testing.T) {
	// List of all node type constants from taxonomy_generated.go
	// This test ensures we don't miss any node types when categorizing
	allNodeTypes := []string{
		// Core Execution Types
		graphrag.NodeTypeAgentRun,
		graphrag.NodeTypeMission,
		graphrag.NodeTypeToolExecution,
		graphrag.NodeTypeLlmCall,
		graphrag.NodeTypeIntelligence,

		// Asset Discovery Types
		graphrag.NodeTypeDomain,
		graphrag.NodeTypeSubdomain,
		graphrag.NodeTypeHost,
		graphrag.NodeTypePort,
		graphrag.NodeTypeService,
		graphrag.NodeTypeCertificate,
		graphrag.NodeTypeTechnology,

		// Network Infrastructure Types
		graphrag.NodeTypeDNSRecord,
		graphrag.NodeTypeFirewall,
		graphrag.NodeTypeFirewallRule,
		graphrag.NodeTypeRouter,
		graphrag.NodeTypeRoute,
		graphrag.NodeTypeLoadBalancer,
		graphrag.NodeTypeProxy,
		graphrag.NodeTypeVPN,
		graphrag.NodeTypeNetwork,
		graphrag.NodeTypeVLAN,
		graphrag.NodeTypeNetworkInterface,
		graphrag.NodeTypeNetworkZone,
		graphrag.NodeTypeNetworkACL,
		graphrag.NodeTypeNATGateway,
		graphrag.NodeTypeBGPPeer,

		// Web/API Types
		graphrag.NodeTypeApi,
		graphrag.NodeTypeEndpoint,
		graphrag.NodeTypeAPIEndpoint,
		graphrag.NodeTypeParameter,
		graphrag.NodeTypeHeader,
		graphrag.NodeTypeCookie,
		graphrag.NodeTypeForm,
		graphrag.NodeTypeFormField,
		graphrag.NodeTypeWebSocket,
		graphrag.NodeTypeGraphQLSchema,
		graphrag.NodeTypeGraphQLQuery,
		graphrag.NodeTypeGraphQLMutation,
		graphrag.NodeTypeRESTResource,
		graphrag.NodeTypeHTTPMethod,
		graphrag.NodeTypeContentType,
		graphrag.NodeTypeCORSPolicy,
		graphrag.NodeTypeRateLimit,
		graphrag.NodeTypeRequestBody,

		// Identity & Access Types
		graphrag.NodeTypeUser,
		graphrag.NodeTypeGroup,
		graphrag.NodeTypeRole,
		graphrag.NodeTypePermission,
		graphrag.NodeTypePolicy,
		graphrag.NodeTypeCredential,
		graphrag.NodeTypeAPIKey,
		graphrag.NodeTypeToken,
		graphrag.NodeTypeOAuthClient,
		graphrag.NodeTypeOAuthScope,
		graphrag.NodeTypeSAMLProvider,
		graphrag.NodeTypeIdentityProvider,
		graphrag.NodeTypeServiceAccount,
		graphrag.NodeTypeSession,
		graphrag.NodeTypeAccessKey,
		graphrag.NodeTypeMFADevice,

		// AI/LLM Types
		graphrag.NodeTypeLLM,
		graphrag.NodeTypeLLMDeployment,
		graphrag.NodeTypePrompt,
		graphrag.NodeTypeSystemPrompt,
		graphrag.NodeTypeGuardrail,
		graphrag.NodeTypeContentFilter,
		graphrag.NodeTypeLLMResponse,
		graphrag.NodeTypeTokenUsage,
		graphrag.NodeTypeEmbeddingModel,
		graphrag.NodeTypeFineTune,
		graphrag.NodeTypeModelRegistry,
		graphrag.NodeTypeModelVersion,
		graphrag.NodeTypeInferenceEndpoint,
		graphrag.NodeTypeBatchJob,
		graphrag.NodeTypeTrainingRun,
		graphrag.NodeTypeDataset,

		// AI Agent Types
		graphrag.NodeTypeAIAgent,
		graphrag.NodeTypeAgentConfig,
		graphrag.NodeTypeAgentMemory,
		graphrag.NodeTypeAgentTool,
		graphrag.NodeTypeChain,
		graphrag.NodeTypeWorkflow,
		graphrag.NodeTypeCrew,
		graphrag.NodeTypeAgentTask,
		graphrag.NodeTypeAgentRole,
		graphrag.NodeTypeToolCall,
		graphrag.NodeTypeReasoningStep,
		graphrag.NodeTypeMemoryEntry,
		graphrag.NodeTypeAgentLoop,
		graphrag.NodeTypePlanningStep,
		graphrag.NodeTypeAgentArtifact,

		// MCP Types
		graphrag.NodeTypeMCPServer,
		graphrag.NodeTypeMCPTool,
		graphrag.NodeTypeMCPResource,
		graphrag.NodeTypeMCPPrompt,
		graphrag.NodeTypeMCPClient,
		graphrag.NodeTypeMCPTransport,
		graphrag.NodeTypeMCPCapability,
		graphrag.NodeTypeMCPSampling,
		graphrag.NodeTypeMCPRoots,

		// RAG Types
		graphrag.NodeTypeVectorStore,
		graphrag.NodeTypeVectorIndex,
		graphrag.NodeTypeDocument,
		graphrag.NodeTypeDocumentChunk,
		graphrag.NodeTypeKnowledgeBase,
		graphrag.NodeTypeRetriever,
		graphrag.NodeTypeRAGPipeline,
		graphrag.NodeTypeEmbedding,
		graphrag.NodeTypeReranker,
		graphrag.NodeTypeChunkingStrategy,
		graphrag.NodeTypeRetrievalResult,

		// Data Types
		graphrag.NodeTypeDatabase,
		graphrag.NodeTypeTable,
		graphrag.NodeTypeColumn,
		graphrag.NodeTypeIndex,
		graphrag.NodeTypeView,
		graphrag.NodeTypeStoredProcedure,
		graphrag.NodeTypeTrigger,
		graphrag.NodeTypeFile,
		graphrag.NodeTypeStorageBucket,
		graphrag.NodeTypeObject,
		graphrag.NodeTypeQueue,
		graphrag.NodeTypeTopic,
		graphrag.NodeTypeStream,
		graphrag.NodeTypeCache,
		graphrag.NodeTypeSchema,
		graphrag.NodeTypeDataPipeline,

		// Container Types
		graphrag.NodeTypeContainer,
		graphrag.NodeTypeContainerImage,
		graphrag.NodeTypeContainerRegistry,
		graphrag.NodeTypeDockerfile,

		// Kubernetes Types
		graphrag.NodeTypeK8sCluster,
		graphrag.NodeTypeK8sNamespace,
		graphrag.NodeTypeK8sPod,
		graphrag.NodeTypeK8sDeployment,
		graphrag.NodeTypeK8sService,
		graphrag.NodeTypeK8sIngress,
		graphrag.NodeTypeK8sConfigMap,
		graphrag.NodeTypeK8sSecret,
		graphrag.NodeTypeK8sPVC,
		graphrag.NodeTypeK8sPV,
		graphrag.NodeTypeK8sStatefulSet,
		graphrag.NodeTypeK8sDaemonSet,
		graphrag.NodeTypeK8sJob,
		graphrag.NodeTypeK8sCronJob,
		graphrag.NodeTypeK8sServiceAccount,
		graphrag.NodeTypeK8sRole,
		graphrag.NodeTypeK8sRoleBinding,
		graphrag.NodeTypeK8sClusterRole,
		graphrag.NodeTypeK8sClusterRoleBinding,
		graphrag.NodeTypeK8sNetworkPolicy,
		graphrag.NodeTypeK8sLimitRange,
		graphrag.NodeTypeK8sResourceQuota,

		// Cloud Provider Types
		graphrag.NodeTypeCloudAsset,
		graphrag.NodeTypeCloudAccount,
		graphrag.NodeTypeCloudVPC,
		graphrag.NodeTypeCloudSubnet,
		graphrag.NodeTypeCloudSecurityGroup,
		graphrag.NodeTypeCloudInstance,
		graphrag.NodeTypeCloudFunction,
		graphrag.NodeTypeCloudStorage,
		graphrag.NodeTypeCloudDatabase,
		graphrag.NodeTypeCloudQueue,
		graphrag.NodeTypeCloudAPIGateway,
		graphrag.NodeTypeCloudCDN,
		graphrag.NodeTypeCloudDNSZone,
		graphrag.NodeTypeCloudCertificate,
		graphrag.NodeTypeCloudKMSKey,
		graphrag.NodeTypeCloudIAMRole,
		graphrag.NodeTypeCloudIAMPolicy,
		graphrag.NodeTypeCloudTrail,
		graphrag.NodeTypeCloudMetric,
		graphrag.NodeTypeCloudAlarm,
		graphrag.NodeTypeCloudRegion,

		// Security Finding Types
		graphrag.NodeTypeFinding,
		graphrag.NodeTypeEvidence,
		graphrag.NodeTypeMitigation,
		graphrag.NodeTypeTactic,
		graphrag.NodeTypeTechnique,
	}

	// Verify every node type has an entry in RequiresParent
	for _, nodeType := range allNodeTypes {
		_, exists := RequiresParent[nodeType]
		assert.True(t, exists, "Node type '%s' is missing from RequiresParent map", nodeType)
	}

	// Verify no extra entries in RequiresParent
	assert.Equal(t, len(allNodeTypes), len(RequiresParent),
		"RequiresParent has different number of entries than taxonomy")
}

// TestValidateGraphNode_RootNodes tests validation of root nodes (no parent required)
func TestValidateGraphNode_RootNodes(t *testing.T) {
	tests := []struct {
		name    string
		node    GraphNode
		wantErr bool
	}{
		{
			name: "host without parent is valid",
			node: &Host{
				IP: "192.168.1.1",
			},
			wantErr: false,
		},
		{
			name: "domain without parent is valid",
			node: &Domain{
				Name: "example.com",
			},
			wantErr: false,
		},
		{
			name: "finding without parent is valid",
			node: &Finding{
				ID: "finding-1",
			},
			wantErr: false,
		},
		{
			name: "technique without parent is valid",
			node: &Technique{
				ID:   "GIB-T1001",
				Name: "Prompt Injection",
			},
			wantErr: false,
		},
		{
			name: "cloud account without parent is valid",
			node: &CloudAccount{
				AccountID: "123456789012",
				Provider:  "aws",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGraphNode(tt.node)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateGraphNode_ChildNodes tests validation of child nodes (parent required)
func TestValidateGraphNode_ChildNodes(t *testing.T) {
	tests := []struct {
		name        string
		node        GraphNode
		setupParent bool
		wantErr     bool
		errContains string
	}{
		{
			name: "port without ANY parent info fails validation",
			node: &Port{
				// No HostID, no parent set - this should fail
				Number:   443,
				Protocol: "tcp",
			},
			wantErr:     true,
			errContains: "port requires a parent",
		},
		{
			name: "port with BelongsTo() passes validation",
			node: &Port{
				Number:   443,
				Protocol: "tcp",
			},
			setupParent: true,
			wantErr:     false,
		},
		{
			name: "service without parent fails validation",
			node: &Service{
				// PortID empty - should fail
				Name: "https",
			},
			wantErr:     true,
			errContains: "service requires a parent",
		},
		{
			name: "endpoint without parent fails validation",
			node: &Endpoint{
				// ServiceID empty - should fail
				URL:    "/api/users",
				Method: "GET",
			},
			wantErr:     true,
			errContains: "endpoint requires a parent",
		},
		{
			name: "subdomain without parent fails validation",
			node: &Subdomain{
				// ParentDomain empty - should fail
				Name: "api.example.com",
			},
			wantErr:     true,
			errContains: "subdomain requires a parent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For "with parent" test cases, use BelongsTo() as intended
			if tt.setupParent {
				port := tt.node.(*Port)
				host := &Host{IP: "192.168.1.1"}
				port.BelongsTo(host)
			}

			err := ValidateGraphNode(tt.node)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateGraphNode_UnknownNodeType tests handling of unknown node types
func TestValidateGraphNode_UnknownNodeType(t *testing.T) {
	// Create a mock node with an invalid type
	mockNode := &mockGraphNode{
		nodeType: "invalid_node_type",
	}

	err := ValidateGraphNode(mockNode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown node type")
	assert.Contains(t, err.Error(), "invalid_node_type")
}

// TestRequiresParent_RootNodeCategories verifies root node categorization logic
func TestRequiresParent_RootNodeCategories(t *testing.T) {
	// These node types should all be root nodes (false in RequiresParent)
	rootNodeTypes := []string{
		// Top-level discovery entities
		graphrag.NodeTypeHost,
		graphrag.NodeTypeDomain,
		graphrag.NodeTypeCloudAccount,
		graphrag.NodeTypeFinding,
		graphrag.NodeTypeTechnique,
		graphrag.NodeTypeTactic,

		// Infrastructure roots
		graphrag.NodeTypeFirewall,
		graphrag.NodeTypeRouter,
		graphrag.NodeTypeLoadBalancer,
		graphrag.NodeTypeK8sCluster,

		// Services and APIs
		graphrag.NodeTypeApi,
		graphrag.NodeTypeCloudAPIGateway,
		graphrag.NodeTypeMCPServer,

		// Identity roots
		graphrag.NodeTypeUser,
		graphrag.NodeTypeRole,
		graphrag.NodeTypePolicy,
	}

	for _, nodeType := range rootNodeTypes {
		requiresParent, exists := RequiresParent[nodeType]
		assert.True(t, exists, "Root node type '%s' not found in map", nodeType)
		assert.False(t, requiresParent, "Node type '%s' should be a root node (false)", nodeType)
	}
}

// TestRequiresParent_ChildNodeCategories verifies child node categorization logic
func TestRequiresParent_ChildNodeCategories(t *testing.T) {
	// These node types should all be child nodes (true in RequiresParent)
	childNodeTypes := []string{
		// Asset hierarchy children
		graphrag.NodeTypePort,
		graphrag.NodeTypeService,
		graphrag.NodeTypeEndpoint,
		graphrag.NodeTypeSubdomain,

		// Web/API children
		graphrag.NodeTypeParameter,
		graphrag.NodeTypeHeader,
		graphrag.NodeTypeCookie,
		graphrag.NodeTypeFormField,

		// Data hierarchy children
		graphrag.NodeTypeTable,
		graphrag.NodeTypeColumn,
		graphrag.NodeTypeIndex,

		// K8s hierarchy children
		graphrag.NodeTypeK8sPod,
		graphrag.NodeTypeK8sDeployment,
		graphrag.NodeTypeK8sService,

		// Cloud hierarchy children
		graphrag.NodeTypeCloudSubnet,
		graphrag.NodeTypeFirewallRule,

		// AI/LLM hierarchy children
		graphrag.NodeTypeLLMResponse,
		graphrag.NodeTypeTokenUsage,
		graphrag.NodeTypeToolCall,

		// RAG hierarchy children
		graphrag.NodeTypeDocumentChunk,
		graphrag.NodeTypeEmbedding,
	}

	for _, nodeType := range childNodeTypes {
		requiresParent, exists := RequiresParent[nodeType]
		assert.True(t, exists, "Child node type '%s' not found in map", nodeType)
		assert.True(t, requiresParent, "Node type '%s' should be a child node (true)", nodeType)
	}
}

// TestValidateGraphNode_RealWorldScenarios tests realistic agent usage patterns
func TestValidateGraphNode_RealWorldScenarios(t *testing.T) {
	t.Run("network recon stores host (valid)", func(t *testing.T) {
		host := &Host{
			IP:       "192.168.1.100",
			Hostname: "web-server",
			State:    "up",
		}
		err := ValidateGraphNode(host)
		assert.NoError(t, err)
	})

	t.Run("network recon stores port without parent ref (invalid)", func(t *testing.T) {
		port := &Port{
			// No HostID, no parent - should fail
			Number:   443,
			Protocol: "tcp",
			State:    "open",
		}
		err := ValidateGraphNode(port)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "port requires a parent")
	})

	t.Run("tech fingerprinting stores service without parent (invalid)", func(t *testing.T) {
		service := &Service{
			// PortID empty - should fail
			Name:    "https",
			Version: "nginx/1.18.0",
		}
		err := ValidateGraphNode(service)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service requires a parent")
	})

	t.Run("finding submission (valid)", func(t *testing.T) {
		finding := &Finding{
			ID: "finding-123",
		}
		err := ValidateGraphNode(finding)
		assert.NoError(t, err)
	})
}

// TestValidateGraphNode_ErrorMessages verifies error message quality
func TestValidateGraphNode_ErrorMessages(t *testing.T) {
	tests := []struct {
		name              string
		node              GraphNode
		expectedErrSubstr []string
	}{
		{
			name: "missing parent error is actionable",
			node: &Port{
				// No HostID - should fail with actionable error
				Number:   80,
				Protocol: "tcp",
			},
			expectedErrSubstr: []string{
				"port",
				"requires a parent",
				"BelongsTo",
			},
		},
		{
			name: "unknown type error is clear",
			node: &mockGraphNode{
				nodeType: "bogus_type",
			},
			expectedErrSubstr: []string{
				"unknown node type",
				"bogus_type",
				"taxonomy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGraphNode(tt.node)
			require.Error(t, err)

			errMsg := err.Error()
			for _, substr := range tt.expectedErrSubstr {
				assert.Contains(t, errMsg, substr,
					"Error message should contain '%s'", substr)
			}
		})
	}
}

// mockGraphNode is a test helper for creating nodes with custom types
type mockGraphNode struct {
	nodeType   string
	parent     *NodeRef
	properties map[string]any
}

func (m *mockGraphNode) NodeType() string {
	return m.nodeType
}

func (m *mockGraphNode) IdentifyingProperties() map[string]any {
	if m.properties != nil {
		return m.properties
	}
	return map[string]any{"id": "test-id"}
}

func (m *mockGraphNode) Properties() map[string]any {
	return m.IdentifyingProperties()
}

func (m *mockGraphNode) ParentRef() *NodeRef {
	return m.parent
}

func (m *mockGraphNode) RelationshipType() string {
	if m.parent != nil {
		return "TEST_RELATIONSHIP"
	}
	return ""
}
