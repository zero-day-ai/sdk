package sdk

import (
	"context"
	"testing"

	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/schema"
)

// TestPluginIntegration tests the integration between NewPlugin and the plugin registry
func TestPluginIntegration(t *testing.T) {
	t.Run("NewPlugin creates valid plugin", func(t *testing.T) {
		p, err := NewPlugin(
			WithPluginName("test-integration-plugin"),
			WithPluginVersion("1.0.0"),
			WithPluginDescription("A test plugin for integration testing"),
		)

		if err != nil {
			t.Fatalf("failed to create plugin: %v", err)
		}

		if p.Name() != "test-integration-plugin" {
			t.Errorf("expected name 'test-integration-plugin', got %s", p.Name())
		}
		if p.Version() != "1.0.0" {
			t.Errorf("expected version '1.0.0', got %s", p.Version())
		}
		if p.Description() != "A test plugin for integration testing" {
			t.Errorf("expected description 'A test plugin for integration testing', got %s", p.Description())
		}

		// Verify it implements plugin.Plugin interface
		var _ plugin.Plugin = p
	})

	t.Run("NewPlugin with method", func(t *testing.T) {
		handler := func(ctx context.Context, params map[string]any) (any, error) {
			name := params["name"].(string)
			return map[string]any{"greeting": "Hello, " + name}, nil
		}

		p, err := NewPlugin(
			WithPluginName("greeter-plugin"),
			WithPluginVersion("1.0.0"),
			WithPluginDescription("A greeting plugin"),
			WithMethod("greet", plugin.MethodHandler(handler),
				schema.Object(map[string]schema.JSON{
					"name": schema.String(),
				}, "name"),
				schema.Object(map[string]schema.JSON{
					"greeting": schema.String(),
				}, "greeting")),
		)

		if err != nil {
			t.Fatalf("failed to create plugin: %v", err)
		}

		// Initialize the plugin
		ctx := context.Background()
		if err := p.Initialize(ctx, nil); err != nil {
			t.Fatalf("failed to initialize plugin: %v", err)
		}

		// Query the method
		result, err := p.Query(ctx, "greet", map[string]any{"name": "World"})
		if err != nil {
			t.Fatalf("failed to query plugin: %v", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any result, got %T", result)
		}

		if resultMap["greeting"] != "Hello, World" {
			t.Errorf("expected 'Hello, World', got %v", resultMap["greeting"])
		}

		// Verify methods are listed
		methods := p.Methods()
		if len(methods) != 1 {
			t.Fatalf("expected 1 method, got %d", len(methods))
		}
		if methods[0].Name != "greet" {
			t.Errorf("expected method 'greet', got %s", methods[0].Name)
		}
	})

	t.Run("Plugin registration in framework", func(t *testing.T) {
		fw, err := NewFramework()
		if err != nil {
			t.Fatalf("failed to create framework: %v", err)
		}

		p, err := NewPlugin(
			WithPluginName("registry-test-plugin"),
			WithPluginVersion("2.0.0"),
		)
		if err != nil {
			t.Fatalf("failed to create plugin: %v", err)
		}

		// Register the plugin
		if err := fw.Plugins().Register(p); err != nil {
			t.Fatalf("failed to register plugin: %v", err)
		}

		// Retrieve the plugin
		retrieved, err := fw.Plugins().Get("registry-test-plugin")
		if err != nil {
			t.Fatalf("failed to get plugin: %v", err)
		}

		if retrieved.Name() != "registry-test-plugin" {
			t.Errorf("expected name 'registry-test-plugin', got %s", retrieved.Name())
		}
		if retrieved.Version() != "2.0.0" {
			t.Errorf("expected version '2.0.0', got %s", retrieved.Version())
		}

		// List plugins
		descriptors := fw.Plugins().List()
		if len(descriptors) != 1 {
			t.Fatalf("expected 1 plugin in registry, got %d", len(descriptors))
		}
		if descriptors[0].Name != "registry-test-plugin" {
			t.Errorf("expected plugin name 'registry-test-plugin', got %s", descriptors[0].Name)
		}

		// Unregister the plugin
		if err := fw.Plugins().Unregister("registry-test-plugin"); err != nil {
			t.Fatalf("failed to unregister plugin: %v", err)
		}

		// Verify it's gone
		_, err = fw.Plugins().Get("registry-test-plugin")
		if err == nil {
			t.Error("expected error when getting unregistered plugin")
		}
	})

	t.Run("Plugin with init and shutdown functions", func(t *testing.T) {
		initCalled := false
		shutdownCalled := false

		p, err := NewPlugin(
			WithPluginName("lifecycle-plugin"),
			WithPluginVersion("1.0.0"),
			WithPluginInitFunc(func() error {
				initCalled = true
				return nil
			}),
			WithPluginShutdownFunc(func() error {
				shutdownCalled = true
				return nil
			}),
		)

		if err != nil {
			t.Fatalf("failed to create plugin: %v", err)
		}

		ctx := context.Background()

		// Initialize
		if err := p.Initialize(ctx, nil); err != nil {
			t.Fatalf("failed to initialize plugin: %v", err)
		}
		if !initCalled {
			t.Error("expected init function to be called")
		}

		// Shutdown
		if err := p.Shutdown(ctx); err != nil {
			t.Fatalf("failed to shutdown plugin: %v", err)
		}
		if !shutdownCalled {
			t.Error("expected shutdown function to be called")
		}
	})
}
