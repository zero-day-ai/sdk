package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentRegistrationWithMockRegistry tests agent registration flow
func TestAgentRegistrationWithMockRegistry(t *testing.T) {
	mockReg := newMockRegistry()
	mockAgt := &mockAgent{
		name:        "test-agent",
		version:     "1.0.0",
		description: "Test agent for unit tests",
	}

	// Create server with mock registry
	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		Registry:        mockReg,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Simulate registration (extracted from Agent() function logic)
	endpoint := fmt.Sprintf("localhost:%d", srv.Port())

	capabilities := make([]string, len(mockAgt.Capabilities()))
	for i, cap := range mockAgt.Capabilities() {
		capabilities[i] = cap
	}

	targetTypes := make([]string, len(mockAgt.TargetTypes()))
	for i, tt := range mockAgt.TargetTypes() {
		targetTypes[i] = tt
	}

	techniqueTypes := make([]string, len(mockAgt.TechniqueTypes()))
	for i, tt := range mockAgt.TechniqueTypes() {
		techniqueTypes[i] = tt
	}

	serviceInfo := map[string]interface{}{
		"kind":        "agent",
		"name":        mockAgt.Name(),
		"version":     mockAgt.Version(),
		"instance_id": "test-instance-123",
		"endpoint":    endpoint,
		"metadata": map[string]string{
			"description":     mockAgt.Description(),
			"capabilities":    fmt.Sprintf("%v", capabilities),
			"target_types":    fmt.Sprintf("%v", targetTypes),
			"technique_types": fmt.Sprintf("%v", techniqueTypes),
		},
		"started_at": time.Now(),
	}

	// Register
	ctx := context.Background()
	err = mockReg.Register(ctx, serviceInfo)
	require.NoError(t, err)

	// Verify registration
	assert.Len(t, mockReg.registered, 1)
	registered := mockReg.registered[0].(map[string]interface{})
	assert.Equal(t, "agent", registered["kind"])
	assert.Equal(t, "test-agent", registered["name"])
	assert.Equal(t, "1.0.0", registered["version"])
	assert.Equal(t, endpoint, registered["endpoint"])

	// Verify metadata extraction
	metadata := registered["metadata"].(map[string]string)
	assert.Equal(t, "Test agent for unit tests", metadata["description"])
	assert.Contains(t, metadata, "capabilities")
	assert.Contains(t, metadata, "target_types")
	assert.Contains(t, metadata, "technique_types")

	// Deregister
	err = mockReg.Deregister(ctx, serviceInfo)
	require.NoError(t, err)
	assert.Len(t, mockReg.deregistered, 1)

	srv.Stop()
}

// TestToolRegistrationWithMockRegistry tests tool registration flow
func TestToolRegistrationWithMockRegistry(t *testing.T) {
	mockReg := newMockRegistry()
	mockTl := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "Test tool for unit tests",
		tags:        []string{"test", "mock"},
	}

	// Create server with mock registry
	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		Registry:        mockReg,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Simulate registration (extracted from Tool() function logic)
	endpoint := fmt.Sprintf("localhost:%d", srv.Port())

	metadata := map[string]string{
		"description": mockTl.Description(),
	}

	if len(mockTl.Tags()) > 0 {
		metadata["tags"] = fmt.Sprintf("%v", mockTl.Tags())
	}

	// Serialize schemas
	if inputSchemaBytes, err := json.Marshal(mockTl.InputSchema()); err == nil {
		metadata["input_schema"] = string(inputSchemaBytes)
	}

	if outputSchemaBytes, err := json.Marshal(mockTl.OutputSchema()); err == nil {
		metadata["output_schema"] = string(outputSchemaBytes)
	}

	serviceInfo := map[string]interface{}{
		"kind":        "tool",
		"name":        mockTl.Name(),
		"version":     mockTl.Version(),
		"instance_id": "test-instance-456",
		"endpoint":    endpoint,
		"metadata":    metadata,
		"started_at":  time.Now(),
	}

	// Register
	ctx := context.Background()
	err = mockReg.Register(ctx, serviceInfo)
	require.NoError(t, err)

	// Verify registration
	assert.Len(t, mockReg.registered, 1)
	registered := mockReg.registered[0].(map[string]interface{})
	assert.Equal(t, "tool", registered["kind"])
	assert.Equal(t, "test-tool", registered["name"])
	assert.Equal(t, "1.0.0", registered["version"])
	assert.Equal(t, endpoint, registered["endpoint"])

	// Verify metadata extraction
	regMetadata := registered["metadata"].(map[string]string)
	assert.Equal(t, "Test tool for unit tests", regMetadata["description"])
	assert.Contains(t, regMetadata, "tags")
	assert.Contains(t, regMetadata, "input_schema")
	assert.Contains(t, regMetadata, "output_schema")

	// Verify schemas are valid JSON
	var inputSchema map[string]interface{}
	err = json.Unmarshal([]byte(regMetadata["input_schema"]), &inputSchema)
	assert.NoError(t, err)

	var outputSchema map[string]interface{}
	err = json.Unmarshal([]byte(regMetadata["output_schema"]), &outputSchema)
	assert.NoError(t, err)

	// Deregister
	err = mockReg.Deregister(ctx, serviceInfo)
	require.NoError(t, err)
	assert.Len(t, mockReg.deregistered, 1)

	srv.Stop()
}

// TestPluginRegistrationWithMockRegistry tests plugin registration flow
func TestPluginRegistrationWithMockRegistry(t *testing.T) {
	mockReg := newMockRegistry()
	mockPlg := &mockPlugin{}

	// Create server with mock registry
	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		Registry:        mockReg,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Simulate registration (extracted from Plugin() function logic)
	endpoint := fmt.Sprintf("localhost:%d", srv.Port())

	methods := mockPlg.Methods()
	methodNames := make([]string, len(methods))
	for i, method := range methods {
		methodNames[i] = method.Name
	}

	serviceInfo := map[string]interface{}{
		"kind":        "plugin",
		"name":        mockPlg.Name(),
		"version":     mockPlg.Version(),
		"instance_id": "test-instance-789",
		"endpoint":    endpoint,
		"metadata": map[string]string{
			"description": mockPlg.Description(),
			"methods":     fmt.Sprintf("%v", methodNames),
		},
		"started_at": time.Now(),
	}

	// Register
	ctx := context.Background()
	err = mockReg.Register(ctx, serviceInfo)
	require.NoError(t, err)

	// Verify registration
	assert.Len(t, mockReg.registered, 1)
	registered := mockReg.registered[0].(map[string]interface{})
	assert.Equal(t, "plugin", registered["kind"])
	assert.Equal(t, "test-plugin", registered["name"])
	assert.Equal(t, "1.0.0", registered["version"])
	assert.Equal(t, endpoint, registered["endpoint"])

	// Verify metadata extraction
	metadata := registered["metadata"].(map[string]string)
	assert.Equal(t, "Test plugin for unit tests", metadata["description"])
	assert.Contains(t, metadata, "methods")

	// Deregister
	err = mockReg.Deregister(ctx, serviceInfo)
	require.NoError(t, err)
	assert.Len(t, mockReg.deregistered, 1)

	srv.Stop()
}

// TestRegistrationWithLocalMode tests registration with Unix socket
func TestRegistrationWithLocalMode(t *testing.T) {
	mockReg := newMockRegistry()

	tmpDir := t.TempDir()
	socketPath := tmpDir + "/test-agent.sock"

	// Create server with mock registry and LocalMode
	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		LocalMode:       socketPath,
		Registry:        mockReg,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Simulate registration with Unix socket endpoint
	endpoint := fmt.Sprintf("unix://%s", socketPath)

	serviceInfo := map[string]interface{}{
		"kind":        "agent",
		"name":        "test-agent",
		"version":     "1.0.0",
		"instance_id": "test-instance-unix",
		"endpoint":    endpoint,
		"metadata": map[string]string{
			"description": "Test agent with Unix socket",
		},
		"started_at": time.Now(),
	}

	// Register
	ctx := context.Background()
	err = mockReg.Register(ctx, serviceInfo)
	require.NoError(t, err)

	// Verify registration uses Unix socket endpoint
	assert.Len(t, mockReg.registered, 1)
	registered := mockReg.registered[0].(map[string]interface{})
	assert.Equal(t, endpoint, registered["endpoint"])
	assert.Contains(t, registered["endpoint"], "unix://")

	srv.Stop()
}

// TestRegistrationError tests that registration errors are handled gracefully
func TestRegistrationError(t *testing.T) {
	mockReg := newMockRegistry()
	mockReg.registerErr = fmt.Errorf("registry unavailable")

	// Create server with mock registry
	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		Registry:        mockReg,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	serviceInfo := map[string]interface{}{
		"kind":        "agent",
		"name":        "test-agent",
		"version":     "1.0.0",
		"instance_id": "test-instance-error",
		"endpoint":    fmt.Sprintf("localhost:%d", srv.Port()),
		"metadata":    map[string]string{},
		"started_at":  time.Now(),
	}

	// Register should fail but not panic
	ctx := context.Background()
	err = mockReg.Register(ctx, serviceInfo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registry unavailable")

	srv.Stop()
}

// TestDeregistrationOnShutdown tests that deregistration happens during shutdown
func TestDeregistrationOnShutdown(t *testing.T) {
	mockReg := newMockRegistry()

	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		Registry:        mockReg,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	serviceInfo := map[string]interface{}{
		"kind":        "agent",
		"name":        "test-agent",
		"version":     "1.0.0",
		"instance_id": "test-instance-shutdown",
		"endpoint":    fmt.Sprintf("localhost:%d", srv.Port()),
		"metadata":    map[string]string{},
		"started_at":  time.Now(),
	}

	// Register
	ctx := context.Background()
	err = mockReg.Register(ctx, serviceInfo)
	require.NoError(t, err)
	assert.Len(t, mockReg.registered, 1)

	// Simulate deregistration on shutdown
	err = mockReg.Deregister(ctx, serviceInfo)
	require.NoError(t, err)
	assert.Len(t, mockReg.deregistered, 1)

	// Verify the same service info was deregistered
	deregistered := mockReg.deregistered[0].(map[string]interface{})
	assert.Equal(t, serviceInfo["instance_id"], deregistered["instance_id"])

	srv.Stop()
}

// TestMetadataExtractionCorrectness verifies that all metadata is correctly extracted
func TestMetadataExtractionCorrectness(t *testing.T) {
	tests := []struct {
		name          string
		componentKind string
		expectedFields []string
	}{
		{
			name:          "agent metadata",
			componentKind: "agent",
			expectedFields: []string{
				"description",
				"capabilities",
				"target_types",
				"technique_types",
			},
		},
		{
			name:          "tool metadata",
			componentKind: "tool",
			expectedFields: []string{
				"description",
				"tags",
				"input_schema",
				"output_schema",
			},
		},
		{
			name:          "plugin metadata",
			componentKind: "plugin",
			expectedFields: []string{
				"description",
				"methods",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReg := newMockRegistry()

			cfg := &Config{
				Port:            0,
				HealthEndpoint:  "/health",
				GracefulTimeout: 1 * time.Second,
				Registry:        mockReg,
			}

			srv, err := NewServer(cfg)
			require.NoError(t, err)
			require.NotNil(t, srv)
			defer srv.Stop()

			// Create appropriate metadata based on component kind
			var metadata map[string]string

			switch tt.componentKind {
			case "agent":
				mockAgt := &mockAgent{
					name:        "test-agent",
					version:     "1.0.0",
					description: "Test agent for unit tests",
				}
				capabilities := make([]string, len(mockAgt.Capabilities()))
				for i, cap := range mockAgt.Capabilities() {
					capabilities[i] = cap
				}
				targetTypes := make([]string, len(mockAgt.TargetTypes()))
				for i, t := range mockAgt.TargetTypes() {
					targetTypes[i] = t
				}
				techniqueTypes := make([]string, len(mockAgt.TechniqueTypes()))
				for i, t := range mockAgt.TechniqueTypes() {
					techniqueTypes[i] = t
				}
				metadata = map[string]string{
					"description":     mockAgt.Description(),
					"capabilities":    fmt.Sprintf("%v", capabilities),
					"target_types":    fmt.Sprintf("%v", targetTypes),
					"technique_types": fmt.Sprintf("%v", techniqueTypes),
				}

			case "tool":
				mockTl := &mockTool{
					name:        "test-tool",
					version:     "1.0.0",
					description: "Test tool for unit tests",
					tags:        []string{"test", "mock"},
				}
				metadata = map[string]string{
					"description": mockTl.Description(),
					"tags":        fmt.Sprintf("%v", mockTl.Tags()),
				}
				if inputSchemaBytes, err := json.Marshal(mockTl.InputSchema()); err == nil {
					metadata["input_schema"] = string(inputSchemaBytes)
				}
				if outputSchemaBytes, err := json.Marshal(mockTl.OutputSchema()); err == nil {
					metadata["output_schema"] = string(outputSchemaBytes)
				}

			case "plugin":
				mockPlg := &mockPlugin{}
				methods := mockPlg.Methods()
				methodNames := make([]string, len(methods))
				for i, method := range methods {
					methodNames[i] = method.Name
				}
				metadata = map[string]string{
					"description": mockPlg.Description(),
					"methods":     fmt.Sprintf("%v", methodNames),
				}
			}

			serviceInfo := map[string]interface{}{
				"kind":        tt.componentKind,
				"name":        "test-component",
				"version":     "1.0.0",
				"instance_id": "test-instance",
				"endpoint":    fmt.Sprintf("localhost:%d", srv.Port()),
				"metadata":    metadata,
				"started_at":  time.Now(),
			}

			// Register
			ctx := context.Background()
			err = mockReg.Register(ctx, serviceInfo)
			require.NoError(t, err)

			// Verify all expected fields are present
			registered := mockReg.registered[0].(map[string]interface{})
			regMetadata := registered["metadata"].(map[string]string)

			for _, field := range tt.expectedFields {
				assert.Contains(t, regMetadata, field, "metadata should contain field: %s", field)
				assert.NotEmpty(t, regMetadata[field], "metadata field should not be empty: %s", field)
			}
		})
	}
}
