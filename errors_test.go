package sdk

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestSentinelErrors verifies that all sentinel errors are defined correctly.
func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrAgentNotFound",
			err:  ErrAgentNotFound,
			want: "agent not found",
		},
		{
			name: "ErrToolNotFound",
			err:  ErrToolNotFound,
			want: "tool not found",
		},
		{
			name: "ErrPluginNotFound",
			err:  ErrPluginNotFound,
			want: "plugin not found",
		},
		{
			name: "ErrInvalidConfig",
			err:  ErrInvalidConfig,
			want: "invalid configuration",
		},
		{
			name: "ErrSlotNotSatisfied",
			err:  ErrSlotNotSatisfied,
			want: "slot requirements not satisfied",
		},
		{
			name: "ErrExecutionFailed",
			err:  ErrExecutionFailed,
			want: "execution failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("sentinel error %s is nil", tt.name)
			}
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("error message = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSDKErrorError verifies the Error() method formatting.
func TestSDKErrorError(t *testing.T) {
	tests := []struct {
		name    string
		err     *SDKError
		want    string
		wantErr bool
	}{
		{
			name: "basic error",
			err: &SDKError{
				Op:   "Client.CreateAgent",
				Kind: KindExecution,
				Err:  ErrExecutionFailed,
			},
			want: "sdk: Client.CreateAgent (execution): execution failed",
		},
		{
			name: "error with context",
			err: &SDKError{
				Op:   "Agent.Execute",
				Kind: KindExecution,
				Err:  ErrExecutionFailed,
				Context: map[string]any{
					"agent_id": "test-agent",
					"timeout":  30,
				},
			},
			want: "sdk: Agent.Execute (execution): execution failed [context:",
		},
		{
			name: "error without underlying error",
			err: &SDKError{
				Op:   "Tool.Validate",
				Kind: KindValidation,
			},
			want: "sdk: Tool.Validate: validation",
		},
		{
			name: "error with wrapped error",
			err: &SDKError{
				Op:   "Plugin.Initialize",
				Kind: KindConfiguration,
				Err:  fmt.Errorf("failed to load config: %w", ErrInvalidConfig),
			},
			want: "sdk: Plugin.Initialize (configuration): failed to load config: invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !strings.Contains(got, tt.want) {
				t.Errorf("Error() = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

// TestSDKErrorUnwrap verifies the Unwrap() method.
func TestSDKErrorUnwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := &SDKError{
		Op:   "Test.Operation",
		Kind: KindExecution,
		Err:  underlyingErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlyingErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, underlyingErr)
	}

	// Test with nil underlying error
	errNil := &SDKError{
		Op:   "Test.Operation",
		Kind: KindExecution,
	}
	if unwrapped := errNil.Unwrap(); unwrapped != nil {
		t.Errorf("Unwrap() with nil Err = %v, want nil", unwrapped)
	}
}

// TestSDKErrorIs verifies the Is() method and errors.Is() compatibility.
func TestSDKErrorIs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name: "matches underlying sentinel error",
			err: &SDKError{
				Op:   "Agent.Execute",
				Kind: KindExecution,
				Err:  ErrExecutionFailed,
			},
			target: ErrExecutionFailed,
			want:   true,
		},
		{
			name: "matches wrapped error",
			err: &SDKError{
				Op:   "Tool.Execute",
				Kind: KindExecution,
				Err:  fmt.Errorf("wrapped: %w", ErrToolNotFound),
			},
			target: ErrToolNotFound,
			want:   true,
		},
		{
			name: "matches SDKError by kind",
			err: &SDKError{
				Op:   "Agent.Execute",
				Kind: KindExecution,
				Err:  ErrExecutionFailed,
			},
			target: &SDKError{Kind: KindExecution},
			want:   true,
		},
		{
			name: "matches SDKError by kind and op",
			err: &SDKError{
				Op:   "Agent.Execute",
				Kind: KindExecution,
				Err:  ErrExecutionFailed,
			},
			target: &SDKError{
				Op:   "Agent.Execute",
				Kind: KindExecution,
			},
			want: true,
		},
		{
			name: "does not match different kind",
			err: &SDKError{
				Op:   "Agent.Execute",
				Kind: KindExecution,
				Err:  ErrExecutionFailed,
			},
			target: &SDKError{Kind: KindValidation},
			want:   false,
		},
		{
			name: "does not match different underlying error",
			err: &SDKError{
				Op:   "Agent.Execute",
				Kind: KindExecution,
				Err:  ErrExecutionFailed,
			},
			target: ErrAgentNotFound,
			want:   false,
		},
		{
			name: "does not match nil",
			err: &SDKError{
				Op:   "Agent.Execute",
				Kind: KindExecution,
				Err:  ErrExecutionFailed,
			},
			target: nil,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.Is(tt.err, tt.target)
			if got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSDKErrorAs verifies errors.As() compatibility.
func TestSDKErrorAs(t *testing.T) {
	originalErr := &SDKError{
		Op:   "Agent.Execute",
		Kind: KindExecution,
		Err:  ErrExecutionFailed,
		Context: map[string]any{
			"agent_id": "test-agent",
		},
	}

	wrappedErr := fmt.Errorf("outer error: %w", originalErr)

	var sdkErr *SDKError
	if !errors.As(wrappedErr, &sdkErr) {
		t.Fatal("errors.As() failed to extract SDKError")
	}

	if sdkErr.Op != originalErr.Op {
		t.Errorf("Op = %q, want %q", sdkErr.Op, originalErr.Op)
	}
	if sdkErr.Kind != originalErr.Kind {
		t.Errorf("Kind = %q, want %q", sdkErr.Kind, originalErr.Kind)
	}
	if sdkErr.Context["agent_id"] != "test-agent" {
		t.Errorf("Context[agent_id] = %v, want test-agent", sdkErr.Context["agent_id"])
	}
}

// TestSDKErrorWithContext verifies the WithContext() method.
func TestSDKErrorWithContext(t *testing.T) {
	original := &SDKError{
		Op:   "Agent.Execute",
		Kind: KindExecution,
		Err:  ErrExecutionFailed,
	}

	// Add context
	withCtx := original.WithContext(map[string]any{
		"agent_id": "test-agent",
		"timeout":  30,
	})

	// Verify new error has context
	if withCtx.Context["agent_id"] != "test-agent" {
		t.Errorf("Context[agent_id] = %v, want test-agent", withCtx.Context["agent_id"])
	}
	if withCtx.Context["timeout"] != 30 {
		t.Errorf("Context[timeout] = %v, want 30", withCtx.Context["timeout"])
	}

	// Verify original error is unchanged
	if original.Context != nil {
		t.Error("original error Context was modified")
	}

	// Add more context
	withMoreCtx := withCtx.WithContext(map[string]any{
		"retry_count": 3,
	})

	// Verify all context is present
	if withMoreCtx.Context["agent_id"] != "test-agent" {
		t.Error("agent_id context was lost")
	}
	if withMoreCtx.Context["retry_count"] != 3 {
		t.Error("retry_count context was not added")
	}
}

// TestNewErrorFunctions verifies all the New*Error() constructor functions.
func TestNewErrorFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string, error) *SDKError
		wantKind string
	}{
		{
			name:     "NewNotFoundError",
			fn:       NewNotFoundError,
			wantKind: KindNotFound,
		},
		{
			name:     "NewValidationError",
			fn:       NewValidationError,
			wantKind: KindValidation,
		},
		{
			name:     "NewExecutionError",
			fn:       NewExecutionError,
			wantKind: KindExecution,
		},
		{
			name:     "NewConfigurationError",
			fn:       NewConfigurationError,
			wantKind: KindConfiguration,
		},
		{
			name:     "NewNetworkError",
			fn:       NewNetworkError,
			wantKind: KindNetwork,
		},
		{
			name:     "NewPermissionError",
			fn:       NewPermissionError,
			wantKind: KindPermission,
		},
		{
			name:     "NewTimeoutError",
			fn:       NewTimeoutError,
			wantKind: KindTimeout,
		},
		{
			name:     "NewInternalError",
			fn:       NewInternalError,
			wantKind: KindInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := "Test.Operation"
			underlyingErr := errors.New("test error")

			err := tt.fn(op, underlyingErr)

			if err.Op != op {
				t.Errorf("Op = %q, want %q", err.Op, op)
			}
			if err.Kind != tt.wantKind {
				t.Errorf("Kind = %q, want %q", err.Kind, tt.wantKind)
			}
			if !errors.Is(err, underlyingErr) {
				t.Error("underlying error not preserved")
			}
		})
	}
}

// TestErrorKinds verifies all error kind constants are defined.
func TestErrorKinds(t *testing.T) {
	kinds := []struct {
		name  string
		value string
	}{
		{"KindNotFound", KindNotFound},
		{"KindValidation", KindValidation},
		{"KindExecution", KindExecution},
		{"KindConfiguration", KindConfiguration},
		{"KindNetwork", KindNetwork},
		{"KindPermission", KindPermission},
		{"KindTimeout", KindTimeout},
		{"KindInternal", KindInternal},
	}

	for _, k := range kinds {
		t.Run(k.name, func(t *testing.T) {
			if k.value == "" {
				t.Errorf("constant %s is empty", k.name)
			}
		})
	}
}

// TestErrorChaining verifies that error chains work correctly.
func TestErrorChaining(t *testing.T) {
	// Create a chain: baseErr -> wrappedErr -> sdkErr -> outerErr
	baseErr := errors.New("base error")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
	sdkErr := &SDKError{
		Op:   "Agent.Execute",
		Kind: KindExecution,
		Err:  wrappedErr,
	}
	outerErr := fmt.Errorf("outer: %w", sdkErr)

	// Verify we can find the base error
	if !errors.Is(outerErr, baseErr) {
		t.Error("failed to find base error in chain")
	}

	// Verify we can find the SDK error
	var extractedSDK *SDKError
	if !errors.As(outerErr, &extractedSDK) {
		t.Error("failed to extract SDK error from chain")
	}

	if extractedSDK.Op != "Agent.Execute" {
		t.Errorf("extracted SDK error has wrong Op: %q", extractedSDK.Op)
	}
}

// BenchmarkSDKErrorCreation benchmarks error creation.
func BenchmarkSDKErrorCreation(b *testing.B) {
	b.Run("basic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = &SDKError{
				Op:   "Agent.Execute",
				Kind: KindExecution,
				Err:  ErrExecutionFailed,
			}
		}
	})

	b.Run("with_context", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := &SDKError{
				Op:   "Agent.Execute",
				Kind: KindExecution,
				Err:  ErrExecutionFailed,
			}
			_ = err.WithContext(map[string]any{
				"agent_id": "test-agent",
			})
		}
	})
}

// BenchmarkSDKErrorError benchmarks the Error() method.
func BenchmarkSDKErrorError(b *testing.B) {
	err := &SDKError{
		Op:   "Agent.Execute",
		Kind: KindExecution,
		Err:  ErrExecutionFailed,
		Context: map[string]any{
			"agent_id": "test-agent",
			"timeout":  30,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

// BenchmarkErrorsIs benchmarks errors.Is() with SDKError.
func BenchmarkErrorsIs(b *testing.B) {
	err := &SDKError{
		Op:   "Agent.Execute",
		Kind: KindExecution,
		Err:  ErrExecutionFailed,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = errors.Is(err, ErrExecutionFailed)
	}
}
