package sdk

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/finding"
)

func TestFramework_Lifecycle(t *testing.T) {
	t.Run("start and shutdown", func(t *testing.T) {
		fw, err := NewFramework()
		if err != nil {
			t.Fatalf("failed to create framework: %v", err)
		}

		ctx := context.Background()

		// Start framework
		err = fw.Start(ctx)
		if err != nil {
			t.Fatalf("failed to start framework: %v", err)
		}

		// Starting again should fail
		err = fw.Start(ctx)
		if err == nil {
			t.Error("expected error when starting already started framework")
		}

		// Shutdown framework
		err = fw.Shutdown(ctx)
		if err != nil {
			t.Fatalf("failed to shutdown framework: %v", err)
		}

		// Shutting down again should not error
		err = fw.Shutdown(ctx)
		if err != nil {
			t.Errorf("unexpected error on second shutdown: %v", err)
		}
	})
}

func TestFramework_MissionManagement(t *testing.T) {
	fw, err := NewFramework()
	if err != nil {
		t.Fatalf("failed to create framework: %v", err)
	}

	ctx := context.Background()

	t.Run("create mission", func(t *testing.T) {
		mission, err := fw.CreateMission(ctx,
			WithMissionName("Test Mission"),
			WithMissionDescription("Testing the system"),
			WithMissionTarget("target-1"),
			WithMissionAgents("agent-1", "agent-2"),
		)

		if err != nil {
			t.Fatalf("failed to create mission: %v", err)
		}

		if mission.ID == "" {
			t.Error("expected mission ID to be set")
		}
		if mission.Name != "Test Mission" {
			t.Errorf("expected name 'Test Mission', got %s", mission.Name)
		}
		if mission.Description != "Testing the system" {
			t.Errorf("expected description 'Testing the system', got %s", mission.Description)
		}
		if mission.Status != "pending" {
			t.Errorf("expected status 'pending', got %s", mission.Status)
		}
		if mission.TargetID != "target-1" {
			t.Errorf("expected target ID 'target-1', got %s", mission.TargetID)
		}
		if len(mission.AgentNames) != 2 {
			t.Errorf("expected 2 agents, got %d", len(mission.AgentNames))
		}
	})

	t.Run("get mission", func(t *testing.T) {
		mission, err := fw.CreateMission(ctx, WithMissionName("Get Test"))
		if err != nil {
			t.Fatalf("failed to create mission: %v", err)
		}

		retrieved, err := fw.GetMission(ctx, mission.ID)
		if err != nil {
			t.Fatalf("failed to get mission: %v", err)
		}

		if retrieved.ID != mission.ID {
			t.Errorf("expected ID %s, got %s", mission.ID, retrieved.ID)
		}
		if retrieved.Name != mission.Name {
			t.Errorf("expected name %s, got %s", mission.Name, retrieved.Name)
		}
	})

	t.Run("get non-existent mission", func(t *testing.T) {
		_, err := fw.GetMission(ctx, "non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent mission")
		}
	})

	t.Run("start mission", func(t *testing.T) {
		mission, err := fw.CreateMission(ctx, WithMissionName("Start Test"))
		if err != nil {
			t.Fatalf("failed to create mission: %v", err)
		}

		err = fw.StartMission(ctx, mission.ID)
		if err != nil {
			t.Fatalf("failed to start mission: %v", err)
		}

		retrieved, err := fw.GetMission(ctx, mission.ID)
		if err != nil {
			t.Fatalf("failed to get mission: %v", err)
		}

		if retrieved.Status != "running" {
			t.Errorf("expected status 'running', got %s", retrieved.Status)
		}
		if retrieved.StartedAt == nil {
			t.Error("expected StartedAt to be set")
		}
	})

	t.Run("start non-pending mission", func(t *testing.T) {
		mission, err := fw.CreateMission(ctx, WithMissionName("Already Running"))
		if err != nil {
			t.Fatalf("failed to create mission: %v", err)
		}

		err = fw.StartMission(ctx, mission.ID)
		if err != nil {
			t.Fatalf("failed to start mission: %v", err)
		}

		// Try to start again
		err = fw.StartMission(ctx, mission.ID)
		if err == nil {
			t.Error("expected error when starting non-pending mission")
		}
	})

	t.Run("stop mission", func(t *testing.T) {
		mission, err := fw.CreateMission(ctx, WithMissionName("Stop Test"))
		if err != nil {
			t.Fatalf("failed to create mission: %v", err)
		}

		err = fw.StartMission(ctx, mission.ID)
		if err != nil {
			t.Fatalf("failed to start mission: %v", err)
		}

		err = fw.StopMission(ctx, mission.ID)
		if err != nil {
			t.Fatalf("failed to stop mission: %v", err)
		}

		retrieved, err := fw.GetMission(ctx, mission.ID)
		if err != nil {
			t.Fatalf("failed to get mission: %v", err)
		}

		if retrieved.Status != "stopped" {
			t.Errorf("expected status 'stopped', got %s", retrieved.Status)
		}
		if retrieved.CompletedAt == nil {
			t.Error("expected CompletedAt to be set")
		}
	})

	t.Run("list missions", func(t *testing.T) {
		// Create a new framework to avoid interference from other tests
		fw2, err := NewFramework()
		if err != nil {
			t.Fatalf("failed to create framework: %v", err)
		}

		// Create several missions
		for i := 0; i < 5; i++ {
			_, err := fw2.CreateMission(ctx, WithMissionName("List Test Mission"))
			if err != nil {
				t.Fatalf("failed to create mission: %v", err)
			}
		}

		missions, err := fw2.ListMissions(ctx)
		if err != nil {
			t.Fatalf("failed to list missions: %v", err)
		}

		if len(missions) != 5 {
			t.Errorf("expected 5 missions, got %d", len(missions))
		}
	})

	t.Run("list missions with limit", func(t *testing.T) {
		fw3, err := NewFramework()
		if err != nil {
			t.Fatalf("failed to create framework: %v", err)
		}

		for i := 0; i < 10; i++ {
			_, err := fw3.CreateMission(ctx, WithMissionName("Limit Test"))
			if err != nil {
				t.Fatalf("failed to create mission: %v", err)
			}
		}

		missions, err := fw3.ListMissions(ctx, WithLimit(3))
		if err != nil {
			t.Fatalf("failed to list missions: %v", err)
		}

		if len(missions) != 3 {
			t.Errorf("expected 3 missions, got %d", len(missions))
		}
	})

	t.Run("list missions with offset", func(t *testing.T) {
		fw4, err := NewFramework()
		if err != nil {
			t.Fatalf("failed to create framework: %v", err)
		}

		for i := 0; i < 10; i++ {
			_, err := fw4.CreateMission(ctx, WithMissionName("Offset Test"))
			if err != nil {
				t.Fatalf("failed to create mission: %v", err)
			}
		}

		missions, err := fw4.ListMissions(ctx, WithOffset(5))
		if err != nil {
			t.Fatalf("failed to list missions: %v", err)
		}

		if len(missions) != 5 {
			t.Errorf("expected 5 missions, got %d", len(missions))
		}
	})
}

func TestFramework_AgentRegistry(t *testing.T) {
	fw, err := NewFramework()
	if err != nil {
		t.Fatalf("failed to create framework: %v", err)
	}

	registry := fw.Agents()

	t.Run("register agent", func(t *testing.T) {
		a, err := NewAgent(
			WithName("test-agent"),
			WithVersion("1.0.0"),
			WithDescription("Test"),
			WithExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
				return agent.NewSuccessResult(nil), nil
			}),
		)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		err = registry.Register(a)
		if err != nil {
			t.Fatalf("failed to register agent: %v", err)
		}
	})

	t.Run("register duplicate agent", func(t *testing.T) {
		a, err := NewAgent(
			WithName("duplicate"),
			WithVersion("1.0.0"),
			WithDescription("Test"),
			WithExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
				return agent.NewSuccessResult(nil), nil
			}),
		)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		err = registry.Register(a)
		if err != nil {
			t.Fatalf("failed to register agent: %v", err)
		}

		// Try to register again
		err = registry.Register(a)
		if err == nil {
			t.Error("expected error when registering duplicate agent")
		}
	})

	t.Run("get agent", func(t *testing.T) {
		a, err := NewAgent(
			WithName("get-test"),
			WithVersion("1.0.0"),
			WithDescription("Test"),
			WithExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
				return agent.NewSuccessResult(nil), nil
			}),
		)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		err = registry.Register(a)
		if err != nil {
			t.Fatalf("failed to register agent: %v", err)
		}

		retrieved, err := registry.Get("get-test")
		if err != nil {
			t.Fatalf("failed to get agent: %v", err)
		}

		if retrieved.Name() != "get-test" {
			t.Errorf("expected name 'get-test', got %s", retrieved.Name())
		}
	})

	t.Run("get non-existent agent", func(t *testing.T) {
		_, err := registry.Get("non-existent")
		if err == nil {
			t.Error("expected error for non-existent agent")
		}
	})

	t.Run("list agents", func(t *testing.T) {
		descriptors := registry.List()
		if len(descriptors) < 3 {
			t.Errorf("expected at least 3 agents, got %d", len(descriptors))
		}
	})

	t.Run("unregister agent", func(t *testing.T) {
		a, err := NewAgent(
			WithName("unregister-test"),
			WithVersion("1.0.0"),
			WithDescription("Test"),
			WithExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
				return agent.NewSuccessResult(nil), nil
			}),
		)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		err = registry.Register(a)
		if err != nil {
			t.Fatalf("failed to register agent: %v", err)
		}

		err = registry.Unregister("unregister-test")
		if err != nil {
			t.Fatalf("failed to unregister agent: %v", err)
		}

		_, err = registry.Get("unregister-test")
		if err == nil {
			t.Error("expected error after unregistering agent")
		}
	})
}

func TestFramework_ToolRegistry(t *testing.T) {
	fw, err := NewFramework()
	if err != nil {
		t.Fatalf("failed to create framework: %v", err)
	}

	registry := fw.Tools()

	t.Run("register tool", func(t *testing.T) {
		tl, err := NewTool(
			WithToolName("test-tool"),
			WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return nil, nil
			}),
		)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		err = registry.Register(tl)
		if err != nil {
			t.Fatalf("failed to register tool: %v", err)
		}
	})

	t.Run("register duplicate tool", func(t *testing.T) {
		tl, err := NewTool(
			WithToolName("duplicate-tool"),
			WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return nil, nil
			}),
		)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		err = registry.Register(tl)
		if err != nil {
			t.Fatalf("failed to register tool: %v", err)
		}

		err = registry.Register(tl)
		if err == nil {
			t.Error("expected error when registering duplicate tool")
		}
	})

	t.Run("get tool", func(t *testing.T) {
		tl, err := NewTool(
			WithToolName("get-tool"),
			WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return nil, nil
			}),
		)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		err = registry.Register(tl)
		if err != nil {
			t.Fatalf("failed to register tool: %v", err)
		}

		retrieved, err := registry.Get("get-tool")
		if err != nil {
			t.Fatalf("failed to get tool: %v", err)
		}

		if retrieved.Name() != "get-tool" {
			t.Errorf("expected name 'get-tool', got %s", retrieved.Name())
		}
	})

	t.Run("list tools", func(t *testing.T) {
		descriptors := registry.List()
		if len(descriptors) < 3 {
			t.Errorf("expected at least 3 tools, got %d", len(descriptors))
		}
	})

	t.Run("unregister tool", func(t *testing.T) {
		tl, err := NewTool(
			WithToolName("unregister-tool"),
			WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return nil, nil
			}),
		)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		err = registry.Register(tl)
		if err != nil {
			t.Fatalf("failed to register tool: %v", err)
		}

		err = registry.Unregister("unregister-tool")
		if err != nil {
			t.Fatalf("failed to unregister tool: %v", err)
		}

		_, err = registry.Get("unregister-tool")
		if err == nil {
			t.Error("expected error after unregistering tool")
		}
	})
}

func TestFramework_PluginRegistry(t *testing.T) {
	fw, err := NewFramework()
	if err != nil {
		t.Fatalf("failed to create framework: %v", err)
	}

	registry := fw.Plugins()

	t.Run("register plugin", func(t *testing.T) {
		p, err := NewPlugin(WithPluginName("test-plugin"))
		if err != nil {
			t.Fatalf("failed to create plugin: %v", err)
		}

		err = registry.Register(p)
		if err != nil {
			t.Fatalf("failed to register plugin: %v", err)
		}
	})

	t.Run("get plugin", func(t *testing.T) {
		p, err := NewPlugin(WithPluginName("get-plugin"))
		if err != nil {
			t.Fatalf("failed to create plugin: %v", err)
		}

		err = registry.Register(p)
		if err != nil {
			t.Fatalf("failed to register plugin: %v", err)
		}

		retrieved, err := registry.Get("get-plugin")
		if err != nil {
			t.Fatalf("failed to get plugin: %v", err)
		}

		if retrieved.Name() != "get-plugin" {
			t.Errorf("expected name 'get-plugin', got %s", retrieved.Name())
		}
	})

	t.Run("list plugins", func(t *testing.T) {
		descriptors := registry.List()
		if len(descriptors) < 2 {
			t.Errorf("expected at least 2 plugins, got %d", len(descriptors))
		}
	})
}

func TestFramework_Findings(t *testing.T) {
	fw, err := NewFramework()
	if err != nil {
		t.Fatalf("failed to create framework: %v", err)
	}

	ctx := context.Background()

	t.Run("get findings - empty", func(t *testing.T) {
		findings, err := fw.GetFindings(ctx, finding.Filter{})
		if err != nil {
			t.Fatalf("failed to get findings: %v", err)
		}

		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("export findings JSON", func(t *testing.T) {
		var buf bytes.Buffer
		err := fw.ExportFindings(ctx, finding.FormatJSON, &buf)
		if err != nil {
			t.Fatalf("failed to export findings: %v", err)
		}

		output := buf.String()
		// Output should contain JSON array brackets (even if empty)
		if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
			t.Errorf("expected JSON array in output, got: %s", output)
		}
	})

	t.Run("export findings CSV", func(t *testing.T) {
		var buf bytes.Buffer
		err := fw.ExportFindings(ctx, finding.FormatCSV, &buf)
		if err != nil {
			t.Fatalf("failed to export findings: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "ID,Title,Severity") {
			t.Error("expected CSV header in output")
		}
	})

	t.Run("export findings HTML", func(t *testing.T) {
		var buf bytes.Buffer
		err := fw.ExportFindings(ctx, finding.FormatHTML, &buf)
		if err != nil {
			t.Fatalf("failed to export findings: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "<html>") {
			t.Error("expected HTML in output")
		}
	})

	t.Run("export findings SARIF", func(t *testing.T) {
		var buf bytes.Buffer
		err := fw.ExportFindings(ctx, finding.FormatSARIF, &buf)
		if err != nil {
			t.Fatalf("failed to export findings: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "sarif-schema") {
			t.Error("expected SARIF schema in output")
		}
	})

	t.Run("export invalid format", func(t *testing.T) {
		var buf bytes.Buffer
		err := fw.ExportFindings(ctx, "invalid", &buf)
		if err == nil {
			t.Error("expected error for invalid export format")
		}
	})
}

func TestFramework_ShutdownStopsMissions(t *testing.T) {
	fw, err := NewFramework()
	if err != nil {
		t.Fatalf("failed to create framework: %v", err)
	}

	ctx := context.Background()

	// Start framework
	err = fw.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start framework: %v", err)
	}

	// Create and start a mission
	mission, err := fw.CreateMission(ctx, WithMissionName("Shutdown Test"))
	if err != nil {
		t.Fatalf("failed to create mission: %v", err)
	}

	err = fw.StartMission(ctx, mission.ID)
	if err != nil {
		t.Fatalf("failed to start mission: %v", err)
	}

	// Shutdown framework
	err = fw.Shutdown(ctx)
	if err != nil {
		t.Fatalf("failed to shutdown framework: %v", err)
	}

	// Verify mission was stopped
	retrieved, err := fw.GetMission(ctx, mission.ID)
	if err != nil {
		t.Fatalf("failed to get mission: %v", err)
	}

	if retrieved.Status != "stopped" {
		t.Errorf("expected status 'stopped' after shutdown, got %s", retrieved.Status)
	}
}

func TestMission(t *testing.T) {
	t.Run("mission fields", func(t *testing.T) {
		now := time.Now()
		mission := &Mission{
			ID:          "test-id",
			Name:        "Test Mission",
			Description: "Testing",
			Status:      "pending",
			TargetID:    "target-1",
			AgentNames:  []string{"agent-1"},
			CreatedAt:   now,
			Metadata:    map[string]interface{}{"key": "value"},
		}

		if mission.ID != "test-id" {
			t.Errorf("expected ID 'test-id', got %s", mission.ID)
		}
		if mission.Name != "Test Mission" {
			t.Errorf("expected name 'Test Mission', got %s", mission.Name)
		}
		if mission.Metadata["key"] != "value" {
			t.Errorf("expected metadata['key'] = 'value', got %v", mission.Metadata["key"])
		}
	})
}
