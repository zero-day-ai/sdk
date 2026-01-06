package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/finding"
	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/tool"
	"go.opentelemetry.io/otel/trace"
)

// Framework provides the main SDK interface for interacting with the Gibson system.
// It manages missions, registries, findings, and lifecycle operations.
//
// The Framework acts as the central orchestrator, coordinating between:
//   - Agents: Autonomous security testing components
//   - Tools: Executable utilities used by agents
//   - Plugins: Extension points for custom functionality
//   - Missions: Orchestrated testing campaigns
//   - Findings: Security vulnerabilities discovered during testing
type Framework interface {
	// Mission management

	// CreateMission creates a new testing mission with the provided configuration.
	// Returns the created mission or an error if creation fails.
	CreateMission(ctx context.Context, opts ...MissionOption) (*Mission, error)

	// StartMission initiates execution of a mission.
	// The mission will begin executing tasks with configured agents.
	StartMission(ctx context.Context, missionID string) error

	// StopMission halts execution of a running mission.
	// In-flight tasks will be cancelled gracefully.
	StopMission(ctx context.Context, missionID string) error

	// GetMission retrieves mission details by ID.
	// Returns an error if the mission is not found.
	GetMission(ctx context.Context, missionID string) (*Mission, error)

	// ListMissions returns a list of missions matching the provided options.
	// Supports filtering, pagination, and sorting.
	ListMissions(ctx context.Context, opts ...ListOption) ([]*Mission, error)

	// Registry access

	// Agents returns the agent registry for registering and discovering agents.
	Agents() AgentRegistry

	// Tools returns the tool registry for registering and discovering tools.
	Tools() ToolRegistry

	// Plugins returns the plugin registry for registering and discovering plugins.
	Plugins() PluginRegistry

	// Findings

	// GetFindings retrieves findings matching the provided filter criteria.
	// Returns all findings if filter is nil.
	GetFindings(ctx context.Context, filter finding.Filter) ([]finding.Finding, error)

	// ExportFindings exports findings in the specified format to the writer.
	// Supported formats: JSON, SARIF, CSV, HTML.
	ExportFindings(ctx context.Context, format finding.ExportFormat, w io.Writer) error

	// Lifecycle

	// Start initializes the framework and prepares it for operation.
	// This should be called before using any framework functionality.
	Start(ctx context.Context) error

	// Shutdown gracefully stops the framework and releases resources.
	// This should be called when the framework is no longer needed.
	Shutdown(ctx context.Context) error
}

// Mission represents a testing campaign executed by the Gibson framework.
// A mission coordinates one or more agents to test a target system.
type Mission struct {
	// ID is the unique identifier for this mission.
	ID string `json:"id"`

	// Name is a human-readable name for the mission.
	Name string `json:"name"`

	// Description explains what this mission is testing.
	Description string `json:"description"`

	// Status indicates the current state of the mission.
	// Values: "pending", "running", "stopped", "completed", "failed"
	Status string `json:"status"`

	// TargetID identifies the system being tested.
	TargetID string `json:"target_id,omitempty"`

	// AgentNames lists the agents assigned to this mission.
	AgentNames []string `json:"agent_names,omitempty"`

	// CreatedAt is the timestamp when the mission was created.
	CreatedAt time.Time `json:"created_at"`

	// StartedAt is the timestamp when the mission execution began.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is the timestamp when the mission finished.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Metadata stores additional mission-specific information.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgentRegistry manages agent registration and discovery.
type AgentRegistry interface {
	// Register adds an agent to the registry.
	// Returns an error if an agent with the same name already exists.
	Register(a agent.Agent) error

	// Get retrieves an agent by name.
	// Returns an error if the agent is not found.
	Get(name string) (agent.Agent, error)

	// List returns descriptors for all registered agents.
	List() []agent.Descriptor

	// Unregister removes an agent from the registry.
	Unregister(name string) error
}

// ToolRegistry manages tool registration and discovery.
type ToolRegistry interface {
	// Register adds a tool to the registry.
	// Returns an error if a tool with the same name already exists.
	Register(t tool.Tool) error

	// Get retrieves a tool by name.
	// Returns an error if the tool is not found.
	Get(name string) (tool.Tool, error)

	// List returns descriptors for all registered tools.
	List() []tool.Descriptor

	// Unregister removes a tool from the registry.
	Unregister(name string) error
}

// PluginRegistry manages plugin registration and discovery.
type PluginRegistry interface {
	// Register adds a plugin to the registry.
	// Returns an error if a plugin with the same name already exists.
	Register(p plugin.Plugin) error

	// Get retrieves a plugin by name.
	// Returns an error if the plugin is not found.
	Get(name string) (plugin.Plugin, error)

	// List returns descriptors for all registered plugins.
	List() []plugin.Descriptor

	// Unregister removes a plugin from the registry.
	Unregister(name string) error
}

// defaultFramework is the concrete implementation of Framework.
type defaultFramework struct {
	logger        *slog.Logger
	tracer        trace.Tracer
	configPath    string
	agents        *agentRegistry
	tools         *toolRegistry
	plugins       *pluginRegistry
	missions      map[string]*Mission
	missionsMu    sync.RWMutex
	findingsStore map[string][]findingRecord
	findingsMu    sync.RWMutex
	started       bool
}

// findingRecord wraps a finding with its mission context.
type findingRecord struct {
	MissionID string
	Finding   finding.Finding
}

// CreateMission creates a new mission with the provided options.
func (f *defaultFramework) CreateMission(ctx context.Context, opts ...MissionOption) (*Mission, error) {
	cfg := &missionConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	mission := &Mission{
		ID:          uuid.New().String(),
		Name:        cfg.name,
		Description: cfg.description,
		Status:      "pending",
		TargetID:    cfg.targetID,
		AgentNames:  cfg.agentNames,
		CreatedAt:   time.Now(),
		Metadata:    cfg.metadata,
	}

	f.missionsMu.Lock()
	f.missions[mission.ID] = mission
	f.missionsMu.Unlock()

	f.logger.Info("mission created",
		slog.String("mission_id", mission.ID),
		slog.String("name", mission.Name),
	)

	return mission, nil
}

// StartMission starts execution of a mission.
func (f *defaultFramework) StartMission(ctx context.Context, missionID string) error {
	f.missionsMu.Lock()
	defer f.missionsMu.Unlock()

	mission, ok := f.missions[missionID]
	if !ok {
		return fmt.Errorf("mission not found: %s", missionID)
	}

	if mission.Status != "pending" {
		return fmt.Errorf("mission cannot be started from status: %s", mission.Status)
	}

	now := time.Now()
	mission.Status = "running"
	mission.StartedAt = &now

	f.logger.Info("mission started",
		slog.String("mission_id", missionID),
	)

	return nil
}

// StopMission stops a running mission.
func (f *defaultFramework) StopMission(ctx context.Context, missionID string) error {
	f.missionsMu.Lock()
	defer f.missionsMu.Unlock()

	mission, ok := f.missions[missionID]
	if !ok {
		return fmt.Errorf("mission not found: %s", missionID)
	}

	if mission.Status != "running" {
		return fmt.Errorf("mission is not running: %s", mission.Status)
	}

	now := time.Now()
	mission.Status = "stopped"
	mission.CompletedAt = &now

	f.logger.Info("mission stopped",
		slog.String("mission_id", missionID),
	)

	return nil
}

// GetMission retrieves a mission by ID.
func (f *defaultFramework) GetMission(ctx context.Context, missionID string) (*Mission, error) {
	f.missionsMu.RLock()
	defer f.missionsMu.RUnlock()

	mission, ok := f.missions[missionID]
	if !ok {
		return nil, fmt.Errorf("mission not found: %s", missionID)
	}

	return mission, nil
}

// ListMissions returns a list of missions.
func (f *defaultFramework) ListMissions(ctx context.Context, opts ...ListOption) ([]*Mission, error) {
	cfg := &listConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	f.missionsMu.RLock()
	defer f.missionsMu.RUnlock()

	// Collect all missions
	missions := make([]*Mission, 0, len(f.missions))
	for _, mission := range f.missions {
		missions = append(missions, mission)
	}

	// Apply offset
	if cfg.offset > 0 {
		if cfg.offset >= len(missions) {
			return []*Mission{}, nil
		}
		missions = missions[cfg.offset:]
	}

	// Apply limit
	if cfg.limit > 0 && cfg.limit < len(missions) {
		missions = missions[:cfg.limit]
	}

	return missions, nil
}

// Agents returns the agent registry.
func (f *defaultFramework) Agents() AgentRegistry {
	return f.agents
}

// Tools returns the tool registry.
func (f *defaultFramework) Tools() ToolRegistry {
	return f.tools
}

// Plugins returns the plugin registry.
func (f *defaultFramework) Plugins() PluginRegistry {
	return f.plugins
}

// GetFindings retrieves findings matching the filter.
func (f *defaultFramework) GetFindings(ctx context.Context, filter finding.Filter) ([]finding.Finding, error) {
	f.findingsMu.RLock()
	defer f.findingsMu.RUnlock()

	var results []finding.Finding

	// Collect all findings
	for _, records := range f.findingsStore {
		for _, record := range records {
			if filter.Matches(record.Finding) {
				results = append(results, record.Finding)
			}
		}
	}

	// Apply pagination
	if filter.Offset > 0 {
		if filter.Offset >= len(results) {
			return []finding.Finding{}, nil
		}
		results = results[filter.Offset:]
	}

	if filter.Limit > 0 && filter.Limit < len(results) {
		results = results[:filter.Limit]
	}

	return results, nil
}

// ExportFindings exports findings in the specified format.
func (f *defaultFramework) ExportFindings(ctx context.Context, format finding.ExportFormat, w io.Writer) error {
	if !format.IsValid() {
		return fmt.Errorf("invalid export format: %s", format)
	}

	// Get all findings
	allFindings, err := f.GetFindings(ctx, finding.Filter{})
	if err != nil {
		return fmt.Errorf("failed to get findings: %w", err)
	}

	// Ensure we have a non-nil slice for JSON encoding
	if allFindings == nil {
		allFindings = []finding.Finding{}
	}

	// Export based on format
	switch format {
	case finding.FormatJSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(allFindings)

	case finding.FormatCSV:
		// Simple CSV export (headers + one line per finding)
		_, err := fmt.Fprintf(w, "ID,Title,Severity,Category,Status,CreatedAt\n")
		if err != nil {
			return err
		}
		for _, f := range allFindings {
			_, err := fmt.Fprintf(w, "%s,%s,%s,%s,%s,%s\n",
				f.ID, f.Title, f.Severity, f.Category, f.Status, f.CreatedAt.Format(time.RFC3339))
			if err != nil {
				return err
			}
		}
		return nil

	case finding.FormatHTML:
		// Simple HTML export
		_, err := fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Gibson Findings Report</title></head>
<body>
<h1>Security Findings Report</h1>
<p>Generated: %s</p>
<table border="1">
<tr><th>ID</th><th>Title</th><th>Severity</th><th>Category</th><th>Status</th></tr>
`, time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}
		for _, f := range allFindings {
			_, err := fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
				f.ID, f.Title, f.Severity, f.Category, f.Status)
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(w, "</table>\n</body>\n</html>")
		return err

	case finding.FormatSARIF:
		// Placeholder SARIF export
		sarifReport := map[string]interface{}{
			"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
			"version": "2.1.0",
			"runs": []map[string]interface{}{
				{
					"tool": map[string]interface{}{
						"driver": map[string]interface{}{
							"name": "Gibson",
						},
					},
					"results": []map[string]interface{}{},
				},
			},
		}
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(sarifReport)

	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// Start initializes the framework.
func (f *defaultFramework) Start(ctx context.Context) error {
	if f.started {
		return fmt.Errorf("framework already started")
	}

	f.logger.Info("starting Gibson framework")
	f.started = true
	return nil
}

// Shutdown gracefully stops the framework.
func (f *defaultFramework) Shutdown(ctx context.Context) error {
	if !f.started {
		return nil
	}

	f.logger.Info("shutting down Gibson framework")

	// Stop all running missions
	f.missionsMu.Lock()
	for id, mission := range f.missions {
		if mission.Status == "running" {
			now := time.Now()
			mission.Status = "stopped"
			mission.CompletedAt = &now
			f.logger.Info("stopped mission during shutdown", slog.String("mission_id", id))
		}
	}
	f.missionsMu.Unlock()

	f.started = false
	return nil
}

// agentRegistry is the concrete implementation of AgentRegistry.
type agentRegistry struct {
	logger *slog.Logger
	agents map[string]agent.Agent
	mu     sync.RWMutex
}

func newAgentRegistry(logger *slog.Logger) *agentRegistry {
	return &agentRegistry{
		logger: logger,
		agents: make(map[string]agent.Agent),
	}
}

func (r *agentRegistry) Register(a agent.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[a.Name()]; exists {
		return fmt.Errorf("agent already registered: %s", a.Name())
	}

	r.agents[a.Name()] = a
	r.logger.Info("agent registered",
		slog.String("name", a.Name()),
		slog.String("version", a.Version()),
	)
	return nil
}

func (r *agentRegistry) Get(name string) (agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.agents[name]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", name)
	}
	return a, nil
}

func (r *agentRegistry) List() []agent.Descriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	descriptors := make([]agent.Descriptor, 0, len(r.agents))
	for _, a := range r.agents {
		descriptors = append(descriptors, agent.Descriptor{
			Name:        a.Name(),
			Version:     a.Version(),
			Description: a.Description(),
		})
	}
	return descriptors
}

func (r *agentRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[name]; !exists {
		return fmt.Errorf("agent not found: %s", name)
	}

	delete(r.agents, name)
	r.logger.Info("agent unregistered", slog.String("name", name))
	return nil
}

// toolRegistry is the concrete implementation of ToolRegistry.
type toolRegistry struct {
	logger *slog.Logger
	tools  map[string]tool.Tool
	mu     sync.RWMutex
}

func newToolRegistry(logger *slog.Logger) *toolRegistry {
	return &toolRegistry{
		logger: logger,
		tools:  make(map[string]tool.Tool),
	}
}

func (r *toolRegistry) Register(t tool.Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[t.Name()]; exists {
		return fmt.Errorf("tool already registered: %s", t.Name())
	}

	r.tools[t.Name()] = t
	r.logger.Info("tool registered",
		slog.String("name", t.Name()),
		slog.String("version", t.Version()),
	)
	return nil
}

func (r *toolRegistry) Get(name string) (tool.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return t, nil
}

func (r *toolRegistry) List() []tool.Descriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	descriptors := make([]tool.Descriptor, 0, len(r.tools))
	for _, t := range r.tools {
		descriptors = append(descriptors, tool.Descriptor{
			Name:        t.Name(),
			Version:     t.Version(),
			Description: t.Description(),
		})
	}
	return descriptors
}

func (r *toolRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("tool not found: %s", name)
	}

	delete(r.tools, name)
	r.logger.Info("tool unregistered", slog.String("name", name))
	return nil
}

// pluginRegistry is the concrete implementation of PluginRegistry.
type pluginRegistry struct {
	logger  *slog.Logger
	plugins map[string]plugin.Plugin
	mu      sync.RWMutex
}

func newPluginRegistry(logger *slog.Logger) *pluginRegistry {
	return &pluginRegistry{
		logger:  logger,
		plugins: make(map[string]plugin.Plugin),
	}
}

func (r *pluginRegistry) Register(p plugin.Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[p.Name()]; exists {
		return fmt.Errorf("plugin already registered: %s", p.Name())
	}

	r.plugins[p.Name()] = p
	r.logger.Info("plugin registered",
		slog.String("name", p.Name()),
		slog.String("version", p.Version()),
	)
	return nil
}

func (r *pluginRegistry) Get(name string) (plugin.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return p, nil
}

func (r *pluginRegistry) List() []plugin.Descriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	descriptors := make([]plugin.Descriptor, 0, len(r.plugins))
	for _, p := range r.plugins {
		descriptors = append(descriptors, plugin.ToDescriptor(p))
	}
	return descriptors
}

func (r *pluginRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	delete(r.plugins, name)
	r.logger.Info("plugin unregistered", slog.String("name", name))
	return nil
}
