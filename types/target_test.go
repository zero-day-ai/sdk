package types

import (
	"encoding/json"
	"testing"
)

func TestTargetType_String(t *testing.T) {
	tests := []struct {
		name       string
		targetType TargetType
		want       string
	}{
		{"llm chat", TargetTypeLLMChat, "llm_chat"},
		{"llm api", TargetTypeLLMAPI, "llm_api"},
		{"rag", TargetTypeRAG, "rag"},
		{"agent", TargetTypeAgent, "agent"},
		{"copilot", TargetTypeCopilot, "copilot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.targetType.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTargetType_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		targetType TargetType
		want       bool
	}{
		{"valid llm chat", TargetTypeLLMChat, true},
		{"valid llm api", TargetTypeLLMAPI, true},
		{"valid rag", TargetTypeRAG, true},
		{"valid agent", TargetTypeAgent, true},
		{"valid copilot", TargetTypeCopilot, true},
		{"invalid empty", TargetType(""), false},
		{"invalid unknown", TargetType("unknown"), false},
		{"invalid custom", TargetType("custom_type"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.targetType.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTargetInfo_Validate(t *testing.T) {
	tests := []struct {
		name    string
		target  TargetInfo
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid target",
			target: TargetInfo{
				ID:   "target-1",
				Name: "Test Target",
				URL:  "https://api.example.com",
				Type: TargetTypeLLMAPI,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			target: TargetInfo{
				Name: "Test Target",
				URL:  "https://api.example.com",
				Type: TargetTypeLLMAPI,
			},
			wantErr: true,
			errMsg:  "ID",
		},
		{
			name: "missing name",
			target: TargetInfo{
				ID:   "target-1",
				URL:  "https://api.example.com",
				Type: TargetTypeLLMAPI,
			},
			wantErr: true,
			errMsg:  "Name",
		},
		{
			name: "missing URL",
			target: TargetInfo{
				ID:   "target-1",
				Name: "Test Target",
				Type: TargetTypeLLMAPI,
			},
			wantErr: true,
			errMsg:  "URL",
		},
		{
			name: "invalid type",
			target: TargetInfo{
				ID:   "target-1",
				Name: "Test Target",
				URL:  "https://api.example.com",
				Type: TargetType("invalid"),
			},
			wantErr: true,
			errMsg:  "Type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if verr, ok := err.(*ValidationError); ok {
					if verr.Field != tt.errMsg {
						t.Errorf("Validate() error field = %v, want %v", verr.Field, tt.errMsg)
					}
				}
			}
		})
	}
}

func TestTargetInfo_GetHeader(t *testing.T) {
	target := &TargetInfo{
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"Content-Type":  "application/json",
		},
	}

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing header", "Authorization", "Bearer token123"},
		{"another existing header", "Content-Type", "application/json"},
		{"non-existent header", "X-Custom-Header", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := target.GetHeader(tt.key); got != tt.want {
				t.Errorf("GetHeader(%v) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}

	// Test with nil headers
	emptyTarget := &TargetInfo{}
	if got := emptyTarget.GetHeader("any-key"); got != "" {
		t.Errorf("GetHeader on nil headers = %v, want empty string", got)
	}
}

func TestTargetInfo_SetHeader(t *testing.T) {
	target := &TargetInfo{}

	target.SetHeader("Authorization", "Bearer token123")
	if got := target.GetHeader("Authorization"); got != "Bearer token123" {
		t.Errorf("After SetHeader, GetHeader() = %v, want %v", got, "Bearer token123")
	}

	// Update existing header
	target.SetHeader("Authorization", "Bearer new-token")
	if got := target.GetHeader("Authorization"); got != "Bearer new-token" {
		t.Errorf("After updating header, GetHeader() = %v, want %v", got, "Bearer new-token")
	}

	// Add another header
	target.SetHeader("Content-Type", "application/json")
	if got := target.GetHeader("Content-Type"); got != "application/json" {
		t.Errorf("After adding second header, GetHeader() = %v, want %v", got, "application/json")
	}
}

func TestTargetInfo_GetMetadata(t *testing.T) {
	target := &TargetInfo{
		Metadata: map[string]any{
			"model":       "gpt-4",
			"max_tokens":  4096,
			"temperature": 0.7,
		},
	}

	tests := []struct {
		name    string
		key     string
		wantVal any
		wantOk  bool
	}{
		{"existing string", "model", "gpt-4", true},
		{"existing int", "max_tokens", 4096, true},
		{"existing float", "temperature", 0.7, true},
		{"non-existent", "unknown", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := target.GetMetadata(tt.key)
			if gotOk != tt.wantOk {
				t.Errorf("GetMetadata(%v) ok = %v, want %v", tt.key, gotOk, tt.wantOk)
			}
			if tt.wantOk && gotVal != tt.wantVal {
				t.Errorf("GetMetadata(%v) val = %v, want %v", tt.key, gotVal, tt.wantVal)
			}
		})
	}

	// Test with nil metadata
	emptyTarget := &TargetInfo{}
	if _, ok := emptyTarget.GetMetadata("any-key"); ok {
		t.Errorf("GetMetadata on nil metadata should return false")
	}
}

func TestTargetInfo_SetMetadata(t *testing.T) {
	target := &TargetInfo{}

	target.SetMetadata("model", "gpt-4")
	val, ok := target.GetMetadata("model")
	if !ok {
		t.Fatal("SetMetadata failed to set value")
	}
	if val != "gpt-4" {
		t.Errorf("After SetMetadata, GetMetadata() = %v, want %v", val, "gpt-4")
	}

	// Update existing metadata
	target.SetMetadata("model", "gpt-4-turbo")
	val, ok = target.GetMetadata("model")
	if !ok {
		t.Fatal("GetMetadata failed to retrieve updated value")
	}
	if val != "gpt-4-turbo" {
		t.Errorf("After updating metadata, GetMetadata() = %v, want %v", val, "gpt-4-turbo")
	}

	// Add different types
	target.SetMetadata("max_tokens", 4096)
	target.SetMetadata("temperature", 0.7)

	intVal, _ := target.GetMetadata("max_tokens")
	if intVal != 4096 {
		t.Errorf("Integer metadata = %v, want %v", intVal, 4096)
	}

	floatVal, _ := target.GetMetadata("temperature")
	if floatVal != 0.7 {
		t.Errorf("Float metadata = %v, want %v", floatVal, 0.7)
	}
}

func TestTargetInfo_JSONMarshaling(t *testing.T) {
	original := TargetInfo{
		ID:       "target-1",
		Name:     "Test Target",
		URL:      "https://api.example.com",
		Type:     TargetTypeLLMAPI,
		Provider: "openai",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
		},
		Metadata: map[string]any{
			"model":      "gpt-4",
			"max_tokens": 4096,
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var unmarshaled TargetInfo
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify basic fields
	if unmarshaled.ID != original.ID {
		t.Errorf("ID = %v, want %v", unmarshaled.ID, original.ID)
	}

	if unmarshaled.Name != original.Name {
		t.Errorf("Name = %v, want %v", unmarshaled.Name, original.Name)
	}

	if unmarshaled.URL != original.URL {
		t.Errorf("URL = %v, want %v", unmarshaled.URL, original.URL)
	}

	if unmarshaled.Type != original.Type {
		t.Errorf("Type = %v, want %v", unmarshaled.Type, original.Type)
	}

	if unmarshaled.Provider != original.Provider {
		t.Errorf("Provider = %v, want %v", unmarshaled.Provider, original.Provider)
	}

	// Verify headers
	if unmarshaled.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Headers[Authorization] = %v, want %v", unmarshaled.Headers["Authorization"], "Bearer token123")
	}

	// Verify metadata
	if unmarshaled.Metadata["model"] != "gpt-4" {
		t.Errorf("Metadata[model] = %v, want %v", unmarshaled.Metadata["model"], "gpt-4")
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "TestField",
		Message: "test error message",
	}

	expected := "TestField: test error message"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %v, want %v", got, expected)
	}
}
