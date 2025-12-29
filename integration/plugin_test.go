package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/schema"
)

// TestPluginCreation tests creating a plugin using SDK entry points.
func TestPluginCreation(t *testing.T) {
	t.Run("with basic configuration", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("test-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("A test plugin")

		p, err := plugin.New(cfg)

		require.NoError(t, err)
		require.NotNil(t, p)

		assert.Equal(t, "test-plugin", p.Name())
		assert.Equal(t, "1.0.0", p.Version())
		assert.Equal(t, "A test plugin", p.Description())
	})

	t.Run("with methods", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("calculator-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Calculator operations")

		// Add methods
		cfg.AddMethod("add", func(ctx context.Context, params map[string]any) (any, error) {
			x := params["x"].(float64)
			y := params["y"].(float64)
			return x + y, nil
		}, schema.Object(map[string]schema.JSON{
			"x": schema.Number(),
			"y": schema.Number(),
		}), schema.Number())

		cfg.AddMethodWithDesc("multiply", "Multiplies two numbers", func(ctx context.Context, params map[string]any) (any, error) {
			x := params["x"].(float64)
			y := params["y"].(float64)
			return x * y, nil
		}, schema.Object(map[string]schema.JSON{
			"x": schema.Number(),
			"y": schema.Number(),
		}), schema.Number())

		p, err := plugin.New(cfg)

		require.NoError(t, err)
		methods := p.Methods()
		assert.Len(t, methods, 2)

		// Check method names
		methodNames := make(map[string]bool)
		for _, m := range methods {
			methodNames[m.Name] = true
		}
		assert.True(t, methodNames["add"])
		assert.True(t, methodNames["multiply"])
	})

	t.Run("missing required name", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("No name")

		p, err := plugin.New(cfg)

		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("missing required version", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("no-version")
		cfg.SetDescription("Missing version")

		// Clear the default version
		cfg.SetVersion("")

		p, err := plugin.New(cfg)

		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Contains(t, err.Error(), "version")
	})
}

// TestPluginInitialization tests plugin initialization.
func TestPluginInitialization(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		var initConfig map[string]any

		cfg := plugin.NewConfig()
		cfg.SetName("init-test-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Tests initialization")

		cfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
			initConfig = config
			return nil
		})

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		config := map[string]any{
			"api_key": "test-key",
			"timeout": 30,
		}

		err = p.Initialize(ctx, config)
		require.NoError(t, err)

		assert.Equal(t, "test-key", initConfig["api_key"])
		assert.Equal(t, 30, initConfig["timeout"])
	})

	t.Run("initialization error", func(t *testing.T) {
		expectedErr := errors.New("initialization failed")

		cfg := plugin.NewConfig()
		cfg.SetName("fail-init-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Fails during init")

		cfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
			return expectedErr
		})

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = p.Initialize(ctx, map[string]any{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initialization failed")
	})

	t.Run("double initialization", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("double-init-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Tests double init")

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()

		// First initialization should succeed
		err = p.Initialize(ctx, map[string]any{})
		require.NoError(t, err)

		// Second initialization should fail
		err = p.Initialize(ctx, map[string]any{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already initialized")
	})
}

// TestPluginMethodInvocation tests invoking plugin methods.
func TestPluginMethodInvocation(t *testing.T) {
	t.Run("invoke valid method", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("math-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Mathematical operations")

		cfg.AddMethod("square", func(ctx context.Context, params map[string]any) (any, error) {
			x := params["x"].(float64)
			return x * x, nil
		}, schema.Object(map[string]schema.JSON{
			"x": schema.Number(),
		}), schema.Number())

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		// Initialize plugin first
		ctx := context.Background()
		err = p.Initialize(ctx, map[string]any{})
		require.NoError(t, err)

		// Invoke method
		result, err := p.Query(ctx, "square", map[string]any{"x": 5.0})
		require.NoError(t, err)
		assert.Equal(t, 25.0, result)
	})

	t.Run("invoke non-existent method", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("empty-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("No methods")

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = p.Initialize(ctx, map[string]any{})
		require.NoError(t, err)

		result, err := p.Query(ctx, "nonexistent", map[string]any{})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "method not found")
	})

	t.Run("invoke with invalid input", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("validator-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Validates input")

		cfg.AddMethod("validate", func(ctx context.Context, params map[string]any) (any, error) {
			return true, nil
		}, schema.Object(map[string]schema.JSON{
			"required_field": schema.String(),
		}, "required_field"), schema.Bool())

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = p.Initialize(ctx, map[string]any{})
		require.NoError(t, err)

		// Missing required field
		result, err := p.Query(ctx, "validate", map[string]any{})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid input")
	})

	t.Run("method returns error", func(t *testing.T) {
		expectedErr := errors.New("method execution failed")

		cfg := plugin.NewConfig()
		cfg.SetName("failing-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Method that fails")

		cfg.AddMethod("fail", func(ctx context.Context, params map[string]any) (any, error) {
			return nil, expectedErr
		}, schema.Object(map[string]schema.JSON{}), schema.Object(map[string]schema.JSON{}))

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = p.Initialize(ctx, map[string]any{})
		require.NoError(t, err)

		result, err := p.Query(ctx, "fail", map[string]any{})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
}

// TestPluginWithMultipleMethods tests plugins with multiple methods.
func TestPluginWithMultipleMethods(t *testing.T) {
	cfg := plugin.NewConfig()
	cfg.SetName("multi-method-plugin")
	cfg.SetVersion("1.0.0")
	cfg.SetDescription("Plugin with multiple methods")

	// Add math operations
	cfg.AddMethod("add", func(ctx context.Context, params map[string]any) (any, error) {
		x := params["x"].(float64)
		y := params["y"].(float64)
		return x + y, nil
	}, schema.Object(map[string]schema.JSON{
		"x": schema.Number(),
		"y": schema.Number(),
	}), schema.Number())

	cfg.AddMethod("subtract", func(ctx context.Context, params map[string]any) (any, error) {
		x := params["x"].(float64)
		y := params["y"].(float64)
		return x - y, nil
	}, schema.Object(map[string]schema.JSON{
		"x": schema.Number(),
		"y": schema.Number(),
	}), schema.Number())

	cfg.AddMethod("multiply", func(ctx context.Context, params map[string]any) (any, error) {
		x := params["x"].(float64)
		y := params["y"].(float64)
		return x * y, nil
	}, schema.Object(map[string]schema.JSON{
		"x": schema.Number(),
		"y": schema.Number(),
	}), schema.Number())

	p, err := plugin.New(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = p.Initialize(ctx, map[string]any{})
	require.NoError(t, err)

	t.Run("invoke all methods", func(t *testing.T) {
		params := map[string]any{"x": 10.0, "y": 5.0}

		// Test add
		result, err := p.Query(ctx, "add", params)
		require.NoError(t, err)
		assert.Equal(t, 15.0, result)

		// Test subtract
		result, err = p.Query(ctx, "subtract", params)
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)

		// Test multiply
		result, err = p.Query(ctx, "multiply", params)
		require.NoError(t, err)
		assert.Equal(t, 50.0, result)
	})

	t.Run("check method descriptors", func(t *testing.T) {
		methods := p.Methods()
		assert.Len(t, methods, 3)

		methodNames := make(map[string]bool)
		for _, m := range methods {
			methodNames[m.Name] = true
			assert.NotNil(t, m.InputSchema)
			assert.NotNil(t, m.OutputSchema)
		}

		assert.True(t, methodNames["add"])
		assert.True(t, methodNames["subtract"])
		assert.True(t, methodNames["multiply"])
	})
}

// TestPluginShutdown tests plugin shutdown.
func TestPluginShutdown(t *testing.T) {
	t.Run("successful shutdown", func(t *testing.T) {
		var shutdownCalled bool

		cfg := plugin.NewConfig()
		cfg.SetName("shutdown-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Tests shutdown")

		cfg.SetShutdownFunc(func(ctx context.Context) error {
			shutdownCalled = true
			return nil
		})

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = p.Initialize(ctx, map[string]any{})
		require.NoError(t, err)

		err = p.Shutdown(ctx)
		require.NoError(t, err)
		assert.True(t, shutdownCalled)
	})

	t.Run("shutdown error", func(t *testing.T) {
		expectedErr := errors.New("shutdown failed")

		cfg := plugin.NewConfig()
		cfg.SetName("fail-shutdown-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Fails during shutdown")

		cfg.SetShutdownFunc(func(ctx context.Context) error {
			return expectedErr
		})

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = p.Initialize(ctx, map[string]any{})
		require.NoError(t, err)

		err = p.Shutdown(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shutdown failed")
	})

	t.Run("shutdown without initialization", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("no-init-shutdown-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Shutdown without init")

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = p.Shutdown(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}

// TestPluginHealthStatus tests plugin health status reporting.
func TestPluginHealthStatus(t *testing.T) {
	t.Run("healthy when initialized", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("healthy-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Healthy plugin")

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()

		// Before initialization - should be unhealthy
		health := p.Health(ctx)
		assert.False(t, health.IsHealthy())
		assert.Contains(t, health.Message, "not initialized")

		// After initialization - should be healthy
		err = p.Initialize(ctx, map[string]any{})
		require.NoError(t, err)

		health = p.Health(ctx)
		assert.True(t, health.IsHealthy())
		assert.Contains(t, health.Message, "operational")
	})

	t.Run("unhealthy after shutdown", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("shutdown-health-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Tests health after shutdown")

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = p.Initialize(ctx, map[string]any{})
		require.NoError(t, err)

		// Should be healthy
		health := p.Health(ctx)
		assert.True(t, health.IsHealthy())

		// Shutdown
		err = p.Shutdown(ctx)
		require.NoError(t, err)

		// Should be unhealthy
		health = p.Health(ctx)
		assert.False(t, health.IsHealthy())
	})
}

// TestPluginRealWorldScenarios tests realistic plugin scenarios.
func TestPluginRealWorldScenarios(t *testing.T) {
	t.Run("database plugin", func(t *testing.T) {
		var db map[string]any

		cfg := plugin.NewConfig()
		cfg.SetName("database-plugin")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Simple database operations")

		cfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
			db = make(map[string]any)
			return nil
		})

		cfg.AddMethod("set", func(ctx context.Context, params map[string]any) (any, error) {
			key := params["key"].(string)
			value := params["value"]
			db[key] = value
			return true, nil
		}, schema.Object(map[string]schema.JSON{
			"key":   schema.String(),
			"value": schema.Object(map[string]schema.JSON{}),
		}, "key", "value"), schema.Bool())

		cfg.AddMethod("get", func(ctx context.Context, params map[string]any) (any, error) {
			key := params["key"].(string)
			value, ok := db[key]
			if !ok {
				return nil, errors.New("key not found")
			}
			return value, nil
		}, schema.Object(map[string]schema.JSON{
			"key": schema.String(),
		}, "key"), schema.Object(map[string]schema.JSON{}))

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = p.Initialize(ctx, map[string]any{})
		require.NoError(t, err)

		// Set a value
		result, err := p.Query(ctx, "set", map[string]any{
			"key":   "user:123",
			"value": map[string]any{"name": "Alice", "age": 30.0},
		})
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Get the value
		result, err = p.Query(ctx, "get", map[string]any{
			"key": "user:123",
		})
		require.NoError(t, err)
		user := result.(map[string]any)
		assert.Equal(t, "Alice", user["name"])
		assert.Equal(t, 30.0, user["age"])
	})

	t.Run("LLM provider plugin", func(t *testing.T) {
		cfg := plugin.NewConfig()
		cfg.SetName("llm-provider")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Mock LLM provider")

		var apiKey string

		cfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
			apiKey = config["api_key"].(string)
			if apiKey == "" {
				return errors.New("api_key is required")
			}
			return nil
		})

		cfg.AddMethodWithDesc("complete", "Generates a completion", func(ctx context.Context, params map[string]any) (any, error) {
			prompt := params["prompt"].(string)
			return map[string]any{
				"text":  "Mock response to: " + prompt,
				"model": "mock-model",
			}, nil
		}, schema.Object(map[string]schema.JSON{
			"prompt": schema.String(),
		}, "prompt"), schema.Object(map[string]schema.JSON{
			"text":  schema.String(),
			"model": schema.String(),
		}))

		cfg.AddMethodWithDesc("embed", "Generates embeddings", func(ctx context.Context, params map[string]any) (any, error) {
			// Mock embedding vector
			return map[string]any{
				"embedding": []any{0.1, 0.2, 0.3},
				"model":     "mock-embed-model",
			}, nil
		}, schema.Object(map[string]schema.JSON{
			"text": schema.String(),
		}, "text"), schema.Object(map[string]schema.JSON{
			"embedding": schema.Array(schema.Number()),
			"model":     schema.String(),
		}))

		p, err := plugin.New(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = p.Initialize(ctx, map[string]any{
			"api_key": "test-key-123",
		})
		require.NoError(t, err)

		// Test completion
		result, err := p.Query(ctx, "complete", map[string]any{
			"prompt": "Hello, world!",
		})
		require.NoError(t, err)
		completion := result.(map[string]any)
		assert.Contains(t, completion["text"], "Mock response")

		// Test embedding
		result, err = p.Query(ctx, "embed", map[string]any{
			"text": "Sample text",
		})
		require.NoError(t, err)
		embedding := result.(map[string]any)
		assert.NotNil(t, embedding["embedding"])
	})
}
