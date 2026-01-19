package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/types"
)

func TestHTTPAPISchema(t *testing.T) {
	// Validate schema definition
	err := HTTPAPISchema.Validate()
	require.NoError(t, err, "HTTPAPISchema should be valid")

	tests := []struct {
		name       string
		connection map[string]any
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid minimal connection",
			connection: map[string]any{
				"url": "https://api.example.com/v1/chat",
			},
			wantErr: false,
		},
		{
			name: "valid full connection",
			connection: map[string]any{
				"url":    "https://api.example.com/v1/chat",
				"method": "POST",
				"headers": map[string]any{
					"Authorization": "Bearer token123",
					"Content-Type":  "application/json",
				},
				"timeout": 60,
			},
			wantErr: false,
		},
		{
			name: "missing required url",
			connection: map[string]any{
				"method": "GET",
			},
			wantErr: true,
			errMsg:  "url",
		},
		{
			name: "invalid method enum",
			connection: map[string]any{
				"url":    "https://api.example.com",
				"method": "INVALID",
			},
			wantErr: true,
			errMsg:  "not one of the allowed values",
		},
		{
			name: "invalid timeout type",
			connection: map[string]any{
				"url":     "https://api.example.com",
				"timeout": "not-a-number",
			},
			wantErr: true,
			errMsg:  "expected integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HTTPAPISchema.ValidateConnection(tt.connection)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLLMChatSchema(t *testing.T) {
	err := LLMChatSchema.Validate()
	require.NoError(t, err, "LLMChatSchema should be valid")

	tests := []struct {
		name       string
		connection map[string]any
		wantErr    bool
	}{
		{
			name: "valid minimal connection",
			connection: map[string]any{
				"url": "https://api.openai.com/v1/chat/completions",
			},
			wantErr: false,
		},
		{
			name: "valid full connection",
			connection: map[string]any{
				"url":           "https://api.anthropic.com/v1/messages",
				"model":         "claude-3-opus-20240229",
				"provider":      "anthropic",
				"system_prompt": "You are a helpful assistant",
				"headers": map[string]any{
					"x-api-key": "sk-ant-...",
				},
			},
			wantErr: false,
		},
		{
			name: "missing required url",
			connection: map[string]any{
				"model": "gpt-4",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LLMChatSchema.ValidateConnection(tt.connection)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLLMAPISchema(t *testing.T) {
	err := LLMAPISchema.Validate()
	require.NoError(t, err, "LLMAPISchema should be valid")

	tests := []struct {
		name       string
		connection map[string]any
		wantErr    bool
	}{
		{
			name: "valid connection",
			connection: map[string]any{
				"url":      "https://api.openai.com/v1/completions",
				"model":    "gpt-4",
				"provider": "openai",
			},
			wantErr: false,
		},
		{
			name: "missing url",
			connection: map[string]any{
				"model": "gpt-4",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LLMAPISchema.ValidateConnection(tt.connection)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKubernetesSchema(t *testing.T) {
	err := KubernetesSchema.Validate()
	require.NoError(t, err, "KubernetesSchema should be valid")

	tests := []struct {
		name       string
		connection map[string]any
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid minimal connection",
			connection: map[string]any{
				"cluster": "prod-cluster",
			},
			wantErr: false,
		},
		{
			name: "valid full connection",
			connection: map[string]any{
				"cluster":    "prod-cluster",
				"namespace":  "ml-pipeline",
				"kubeconfig": "/home/user/.kube/config",
				"api_server": "https://api.prod-cluster.example.com:6443",
			},
			wantErr: false,
		},
		{
			name: "missing required cluster",
			connection: map[string]any{
				"namespace": "default",
			},
			wantErr: true,
			errMsg:  "cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := KubernetesSchema.ValidateConnection(tt.connection)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSmartContractSchema(t *testing.T) {
	err := SmartContractSchema.Validate()
	require.NoError(t, err, "SmartContractSchema should be valid")

	tests := []struct {
		name       string
		connection map[string]any
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid ethereum contract",
			connection: map[string]any{
				"chain":   "ethereum",
				"address": "0x1234567890123456789012345678901234567890",
			},
			wantErr: false,
		},
		{
			name: "valid polygon contract with rpc",
			connection: map[string]any{
				"chain":   "polygon",
				"address": "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				"rpc_url": "https://polygon-rpc.com",
				"abi":     `[{"type":"function","name":"predict","inputs":[]}]`,
			},
			wantErr: false,
		},
		{
			name: "missing required chain",
			connection: map[string]any{
				"address": "0x1234567890123456789012345678901234567890",
			},
			wantErr: true,
			errMsg:  "chain",
		},
		{
			name: "missing required address",
			connection: map[string]any{
				"chain": "ethereum",
			},
			wantErr: true,
			errMsg:  "address",
		},
		{
			name: "invalid chain enum",
			connection: map[string]any{
				"chain":   "invalid-chain",
				"address": "0x1234567890123456789012345678901234567890",
			},
			wantErr: true,
			errMsg:  "not one of the allowed values",
		},
		{
			name: "invalid address format",
			connection: map[string]any{
				"chain":   "ethereum",
				"address": "not-a-valid-address",
			},
			wantErr: true,
			errMsg:  "does not match pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SmartContractSchema.ValidateConnection(tt.connection)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetBuiltinSchema(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		wantNil  bool
	}{
		{
			name:     "http_api schema exists",
			typeName: "http_api",
			wantNil:  false,
		},
		{
			name:     "llm_chat schema exists",
			typeName: "llm_chat",
			wantNil:  false,
		},
		{
			name:     "llm_api schema exists",
			typeName: "llm_api",
			wantNil:  false,
		},
		{
			name:     "kubernetes schema exists",
			typeName: "kubernetes",
			wantNil:  false,
		},
		{
			name:     "smart_contract schema exists",
			typeName: "smart_contract",
			wantNil:  false,
		},
		{
			name:     "unknown schema returns nil",
			typeName: "unknown_type",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GetBuiltinSchema(tt.typeName)
			if tt.wantNil {
				assert.Nil(t, schema)
			} else {
				require.NotNil(t, schema)
				assert.Equal(t, tt.typeName, schema.Type)

				// Verify schema is valid
				err := schema.Validate()
				assert.NoError(t, err, "builtin schema should be valid")
			}
		})
	}
}

func TestListBuiltinSchemas(t *testing.T) {
	schemas := ListBuiltinSchemas()

	assert.Len(t, schemas, 5, "should have 5 built-in schemas")

	expected := []string{"http_api", "llm_chat", "llm_api", "kubernetes", "smart_contract"}
	assert.Equal(t, expected, schemas)

	// Verify each listed schema can be retrieved
	for _, typeName := range schemas {
		schema := GetBuiltinSchema(typeName)
		assert.NotNil(t, schema, "listed schema %s should be retrievable", typeName)
	}
}

func TestBuiltinSchemasAreValid(t *testing.T) {
	// Ensure all built-in schemas pass their own validation
	schemas := []struct {
		name   string
		schema *types.TargetSchema
	}{
		{"http_api", &HTTPAPISchema},
		{"llm_chat", &LLMChatSchema},
		{"llm_api", &LLMAPISchema},
		{"kubernetes", &KubernetesSchema},
		{"smart_contract", &SmartContractSchema},
	}

	for _, tt := range schemas {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Validate()
			assert.NoError(t, err, "%s schema should be valid", tt.name)
			assert.Equal(t, tt.name, tt.schema.Type)
			assert.NotEmpty(t, tt.schema.Version)
			assert.NotEmpty(t, tt.schema.Description)
		})
	}
}
