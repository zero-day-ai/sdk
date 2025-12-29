package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/zero-day-ai/sdk/schema"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.methods == nil {
		t.Error("expected methods slice to be initialized")
	}
	if cfg.initFunc == nil {
		t.Error("expected default initFunc")
	}
	if cfg.shutdownFunc == nil {
		t.Error("expected default shutdownFunc")
	}

	// Test default init and shutdown functions
	ctx := context.Background()
	if err := cfg.initFunc(ctx, nil); err != nil {
		t.Errorf("default initFunc should not error: %v", err)
	}
	if err := cfg.shutdownFunc(ctx); err != nil {
		t.Errorf("default shutdownFunc should not error: %v", err)
	}
}

func TestConfigSetters(t *testing.T) {
	cfg := NewConfig()

	cfg.SetName("testPlugin")
	if cfg.name != "testPlugin" {
		t.Errorf("expected name 'testPlugin', got %s", cfg.name)
	}

	cfg.SetVersion("1.0.0")
	if cfg.version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", cfg.version)
	}

	cfg.SetDescription("Test plugin description")
	if cfg.description != "Test plugin description" {
		t.Errorf("expected description 'Test plugin description', got %s", cfg.description)
	}
}

func TestConfigAddMethod(t *testing.T) {
	cfg := NewConfig()

	handler := func(ctx context.Context, params map[string]any) (any, error) {
		return "result", nil
	}

	cfg.AddMethod("testMethod", handler, schema.String(), schema.String())

	if len(cfg.methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(cfg.methods))
	}

	method := cfg.methods[0]
	if method.descriptor.Name != "testMethod" {
		t.Errorf("expected method name 'testMethod', got %s", method.descriptor.Name)
	}
	if method.descriptor.Description != "" {
		t.Errorf("expected empty description, got %s", method.descriptor.Description)
	}
	if method.handler == nil {
		t.Error("expected non-nil handler")
	}
}

func TestConfigAddMethodWithDesc(t *testing.T) {
	cfg := NewConfig()

	handler := func(ctx context.Context, params map[string]any) (any, error) {
		return 42, nil
	}

	cfg.AddMethodWithDesc("calculate", "Performs a calculation", handler, schema.Int(), schema.Int())

	if len(cfg.methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(cfg.methods))
	}

	method := cfg.methods[0]
	if method.descriptor.Name != "calculate" {
		t.Errorf("expected method name 'calculate', got %s", method.descriptor.Name)
	}
	if method.descriptor.Description != "Performs a calculation" {
		t.Errorf("expected description 'Performs a calculation', got %s", method.descriptor.Description)
	}
}

func TestConfigSetInitFunc(t *testing.T) {
	cfg := NewConfig()
	called := false

	cfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
		called = true
		return nil
	})

	if err := cfg.initFunc(context.Background(), nil); err != nil {
		t.Errorf("initFunc returned error: %v", err)
	}
	if !called {
		t.Error("expected initFunc to be called")
	}
}

func TestConfigSetShutdownFunc(t *testing.T) {
	cfg := NewConfig()
	called := false

	cfg.SetShutdownFunc(func(ctx context.Context) error {
		called = true
		return nil
	})

	if err := cfg.shutdownFunc(context.Background()); err != nil {
		t.Errorf("shutdownFunc returned error: %v", err)
	}
	if !called {
		t.Error("expected shutdownFunc to be called")
	}
}

func TestNew_NilConfig(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
	if err.Error() != "config cannot be nil" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNew_MissingName(t *testing.T) {
	cfg := NewConfig()
	cfg.SetVersion("1.0.0")

	_, err := New(cfg)
	if err == nil {
		t.Error("expected error for missing name")
	}
	if err.Error() != "plugin name is required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNew_MissingVersion(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("testPlugin")

	_, err := New(cfg)
	if err == nil {
		t.Error("expected error for missing version")
	}
	if err.Error() != "plugin version is required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNew_EmptyMethodName(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("testPlugin")
	cfg.SetVersion("1.0.0")
	cfg.AddMethod("", func(ctx context.Context, params map[string]any) (any, error) {
		return nil, nil
	}, schema.String(), schema.String())

	_, err := New(cfg)
	if err == nil {
		t.Error("expected error for empty method name")
	}
	if err.Error() != "method name cannot be empty" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNew_DuplicateMethodName(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("testPlugin")
	cfg.SetVersion("1.0.0")

	handler := func(ctx context.Context, params map[string]any) (any, error) {
		return nil, nil
	}

	cfg.AddMethod("duplicate", handler, schema.String(), schema.String())
	cfg.AddMethod("duplicate", handler, schema.Int(), schema.Int())

	_, err := New(cfg)
	if err == nil {
		t.Error("expected error for duplicate method name")
	}
	if err.Error() != "duplicate method name: duplicate" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNew_Success(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("testPlugin")
	cfg.SetVersion("1.0.0")
	cfg.SetDescription("A test plugin")
	cfg.AddMethod("method1", func(ctx context.Context, params map[string]any) (any, error) {
		return "result", nil
	}, schema.String(), schema.String())

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil plugin")
	}

	// Verify plugin implements Plugin interface
	var _ Plugin = p
}

func TestPluginName(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	if p.Name() != "myPlugin" {
		t.Errorf("expected name 'myPlugin', got %s", p.Name())
	}
}

func TestPluginVersion(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("2.1.3")

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	if p.Version() != "2.1.3" {
		t.Errorf("expected version '2.1.3', got %s", p.Version())
	}
}

func TestPluginDescription(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")
	cfg.SetDescription("My awesome plugin")

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	if p.Description() != "My awesome plugin" {
		t.Errorf("expected description 'My awesome plugin', got %s", p.Description())
	}
}

func TestPluginMethods(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")
	cfg.AddMethodWithDesc("method1", "First method", func(ctx context.Context, params map[string]any) (any, error) {
		return nil, nil
	}, schema.String(), schema.Int())
	cfg.AddMethodWithDesc("method2", "Second method", func(ctx context.Context, params map[string]any) (any, error) {
		return nil, nil
	}, schema.Bool(), schema.String())

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	methods := p.Methods()
	if len(methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(methods))
	}

	// Verify methods are in the list
	methodNames := make(map[string]bool)
	for _, m := range methods {
		methodNames[m.Name] = true
	}
	if !methodNames["method1"] {
		t.Error("expected method1 in methods list")
	}
	if !methodNames["method2"] {
		t.Error("expected method2 in methods list")
	}
}

func TestPluginQuery_MethodNotFound(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	_, err = p.Query(ctx, "nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent method")
	}
	if err.Error() != "method not found: nonexistent" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPluginQuery_InvalidInput(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")
	cfg.AddMethod("testMethod", func(ctx context.Context, params map[string]any) (any, error) {
		return "result", nil
	}, schema.Object(map[string]schema.JSON{
		"required_field": schema.String(),
	}, "required_field"), schema.String())

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	// Pass empty params when required_field is required
	_, err = p.Query(ctx, "testMethod", map[string]any{})
	if err == nil {
		t.Error("expected error for invalid input")
	}
}

func TestPluginQuery_HandlerError(t *testing.T) {
	expectedErr := errors.New("handler error")
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")
	cfg.AddMethod("errorMethod", func(ctx context.Context, params map[string]any) (any, error) {
		return nil, expectedErr
	}, schema.Object(map[string]schema.JSON{
		"input": schema.String(),
	}, "input"), schema.String())

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	_, err = p.Query(ctx, "errorMethod", map[string]any{"input": "test"})
	if err == nil {
		t.Error("expected error from handler")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected handler error, got: %v", err)
	}
}

func TestPluginQuery_InvalidOutput(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")
	cfg.AddMethod("badOutput", func(ctx context.Context, params map[string]any) (any, error) {
		// Return string when schema expects int
		return "not an int", nil
	}, schema.Object(map[string]schema.JSON{
		"input": schema.String(),
	}, "input"), schema.Int())

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	_, err = p.Query(ctx, "badOutput", map[string]any{"input": "test"})
	if err == nil {
		t.Error("expected error for invalid output")
	}
}

func TestPluginQuery_Success(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")
	cfg.AddMethodWithDesc("greet", "Greets a person", func(ctx context.Context, params map[string]any) (any, error) {
		name := params["name"].(string)
		return map[string]any{
			"greeting": "Hello, " + name,
		}, nil
	}, schema.Object(map[string]schema.JSON{
		"name": schema.String(),
	}, "name"), schema.Object(map[string]schema.JSON{
		"greeting": schema.String(),
	}, "greeting"))

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	result, err := p.Query(ctx, "greet", map[string]any{"name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any result, got %T", result)
	}

	greeting, ok := resultMap["greeting"].(string)
	if !ok {
		t.Fatalf("expected string greeting, got %T", resultMap["greeting"])
	}

	if greeting != "Hello, World" {
		t.Errorf("expected greeting 'Hello, World', got %s", greeting)
	}
}

func TestPluginInitialize_Success(t *testing.T) {
	initCalled := false
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")
	cfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
		initCalled = true
		if timeout, ok := config["timeout"].(int); ok {
			if timeout != 30 {
				return errors.New("unexpected timeout value")
			}
		}
		return nil
	})

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	err = p.Initialize(ctx, map[string]any{"timeout": 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !initCalled {
		t.Error("expected init function to be called")
	}
}

func TestPluginInitialize_AlreadyInitialized(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	err = p.Initialize(ctx, nil)
	if err != nil {
		t.Fatalf("first initialization failed: %v", err)
	}

	// Try to initialize again
	err = p.Initialize(ctx, nil)
	if err == nil {
		t.Error("expected error for double initialization")
	}
	if err.Error() != "plugin already initialized" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPluginInitialize_Error(t *testing.T) {
	expectedErr := errors.New("init error")
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")
	cfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
		return expectedErr
	})

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	err = p.Initialize(ctx, nil)
	if err == nil {
		t.Error("expected initialization error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected init error, got: %v", err)
	}
}

func TestPluginShutdown_NotInitialized(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	err = p.Shutdown(ctx)
	if err == nil {
		t.Error("expected error for shutdown without initialization")
	}
	if err.Error() != "plugin not initialized" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPluginShutdown_Success(t *testing.T) {
	shutdownCalled := false
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")
	cfg.SetShutdownFunc(func(ctx context.Context) error {
		shutdownCalled = true
		return nil
	})

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	err = p.Initialize(ctx, nil)
	if err != nil {
		t.Fatalf("initialization failed: %v", err)
	}

	err = p.Shutdown(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !shutdownCalled {
		t.Error("expected shutdown function to be called")
	}
}

func TestPluginShutdown_Error(t *testing.T) {
	expectedErr := errors.New("shutdown error")
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")
	cfg.SetShutdownFunc(func(ctx context.Context) error {
		return expectedErr
	})

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	err = p.Initialize(ctx, nil)
	if err != nil {
		t.Fatalf("initialization failed: %v", err)
	}

	err = p.Shutdown(ctx)
	if err == nil {
		t.Error("expected shutdown error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected shutdown error, got: %v", err)
	}
}

func TestPluginHealth_NotInitialized(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	status := p.Health(ctx)

	if !status.IsUnhealthy() {
		t.Error("expected unhealthy status for uninitialized plugin")
	}
	if status.Message != "plugin not initialized" {
		t.Errorf("unexpected message: %s", status.Message)
	}
}

func TestPluginHealth_Initialized(t *testing.T) {
	cfg := NewConfig()
	cfg.SetName("myPlugin")
	cfg.SetVersion("1.0.0")

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	err = p.Initialize(ctx, nil)
	if err != nil {
		t.Fatalf("initialization failed: %v", err)
	}

	status := p.Health(ctx)

	if !status.IsHealthy() {
		t.Error("expected healthy status for initialized plugin")
	}
	if status.Message != "plugin operational" {
		t.Errorf("unexpected message: %s", status.Message)
	}
}

func TestPluginLifecycle(t *testing.T) {
	// Test complete plugin lifecycle
	initCalled := false
	shutdownCalled := false

	cfg := NewConfig()
	cfg.SetName("lifecyclePlugin")
	cfg.SetVersion("1.0.0")
	cfg.SetDescription("Tests plugin lifecycle")
	cfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
		initCalled = true
		return nil
	})
	cfg.SetShutdownFunc(func(ctx context.Context) error {
		shutdownCalled = true
		return nil
	})
	cfg.AddMethodWithDesc("echo", "Echoes input", func(ctx context.Context, params map[string]any) (any, error) {
		return params["message"], nil
	}, schema.Object(map[string]schema.JSON{
		"message": schema.String(),
	}, "message"), schema.String())

	// Create plugin
	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()

	// Check initial health (should be unhealthy)
	status := p.Health(ctx)
	if !status.IsUnhealthy() {
		t.Error("expected unhealthy status before initialization")
	}

	// Initialize
	err = p.Initialize(ctx, nil)
	if err != nil {
		t.Fatalf("initialization failed: %v", err)
	}
	if !initCalled {
		t.Error("expected init to be called")
	}

	// Check health after initialization (should be healthy)
	status = p.Health(ctx)
	if !status.IsHealthy() {
		t.Error("expected healthy status after initialization")
	}

	// Use the plugin
	result, err := p.Query(ctx, "echo", map[string]any{"message": "test"})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result != "test" {
		t.Errorf("expected 'test', got %v", result)
	}

	// Shutdown
	err = p.Shutdown(ctx)
	if err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
	if !shutdownCalled {
		t.Error("expected shutdown to be called")
	}

	// Check health after shutdown (should be unhealthy)
	status = p.Health(ctx)
	if !status.IsUnhealthy() {
		t.Error("expected unhealthy status after shutdown")
	}
}

func TestPluginThreadSafety(t *testing.T) {
	// Test concurrent access to plugin
	cfg := NewConfig()
	cfg.SetName("threadSafePlugin")
	cfg.SetVersion("1.0.0")
	cfg.AddMethod("compute", func(ctx context.Context, params map[string]any) (any, error) {
		value := params["value"].(int)
		return value * 2, nil
	}, schema.Object(map[string]schema.JSON{
		"value": schema.Int(),
	}, "value"), schema.Int())

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()
	err = p.Initialize(ctx, nil)
	if err != nil {
		t.Fatalf("initialization failed: %v", err)
	}

	// Launch multiple goroutines to query concurrently
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(value int) {
			result, err := p.Query(ctx, "compute", map[string]any{"value": value})
			if err != nil {
				t.Errorf("query failed: %v", err)
			}
			expected := value * 2
			if result != expected {
				t.Errorf("expected %d, got %v", expected, result)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
