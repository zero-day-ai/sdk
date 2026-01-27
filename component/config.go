// Package component provides loading and parsing of component.yaml configuration files.
// Component configurations define tool and plugin metadata, dependencies, and runtime settings.
package component

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents a component.yaml configuration file.
// This is the primary configuration for tools and plugins in the Gibson ecosystem.
type Config struct {
	// Identity
	Kind        string `yaml:"kind,omitempty"` // "tool" or "plugin" (alternative to Type)
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Type        string `yaml:"type,omitempty"` // "tool" or "plugin"
	Description string `yaml:"description"`

	// Categorization
	Tags []string `yaml:"tags,omitempty"`

	// MITRE ATT&CK mappings
	MITREAttack *MITREAttackConfig `yaml:"mitre_attack,omitempty"`

	// Dependencies
	Dependencies *DependenciesConfig `yaml:"dependencies,omitempty"`

	// Proto message types
	Proto *ProtoConfig `yaml:"proto,omitempty"`

	// Worker configuration (for queue-based execution)
	Worker *WorkerConfig `yaml:"worker,omitempty"`

	// Build configuration
	Build *BuildConfig `yaml:"build,omitempty"`

	// Additional metadata
	Author     string `yaml:"author,omitempty"`
	License    string `yaml:"license,omitempty"`
	Repository string `yaml:"repository,omitempty"`
}

// MITREAttackConfig contains MITRE ATT&CK framework mappings.
type MITREAttackConfig struct {
	Tactics    []string `yaml:"tactics,omitempty"`
	Techniques []string `yaml:"techniques,omitempty"`
}

// DependenciesConfig defines external dependencies required by the component.
type DependenciesConfig struct {
	Binaries []BinaryDependency `yaml:"binaries,omitempty"`
}

// BinaryDependency describes a required external binary.
type BinaryDependency struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version,omitempty"` // Version constraint (e.g., ">=2.0.0")
	Install string `yaml:"install,omitempty"` // Installation command
}

// ProtoConfig defines Protocol Buffer message types for tool I/O.
type ProtoConfig struct {
	Input  string `yaml:"input"`  // Input message type (e.g., "gibson.tools.nmap.ScanRequest")
	Output string `yaml:"output"` // Output message type (e.g., "gibson.tools.nmap.ScanResponse")
}

// WorkerConfig defines configuration for queue-based worker execution.
type WorkerConfig struct {
	// Concurrency is the default number of concurrent worker goroutines.
	// Tools can specify their optimal concurrency based on resource usage:
	// - I/O-bound tools (e.g., nmap, httpx): higher concurrency (4-8)
	// - CPU-bound tools: lower concurrency (1-2)
	// Default: 4
	Concurrency int `yaml:"concurrency,omitempty"`

	// ShutdownTimeout is the time to wait for graceful shutdown.
	// Format: Go duration string (e.g., "30s", "1m")
	// Default: 30s
	ShutdownTimeout string `yaml:"shutdown_timeout,omitempty"`

	// QueuePrefix is the Redis key prefix for this tool's queue.
	// Default: "tool" (resulting in "tool:<name>:queue")
	QueuePrefix string `yaml:"queue_prefix,omitempty"`

	// HeartbeatInterval is the interval between health heartbeats.
	// Format: Go duration string (e.g., "10s")
	// Default: 10s
	HeartbeatInterval string `yaml:"heartbeat_interval,omitempty"`

	// MaxRetries is the maximum number of times to retry a failed work item.
	// Default: 0 (no retries)
	MaxRetries int `yaml:"max_retries,omitempty"`
}

// GetShutdownTimeout parses the shutdown timeout string and returns a duration.
// Returns the default value if not set or invalid.
func (w *WorkerConfig) GetShutdownTimeout() time.Duration {
	if w == nil || w.ShutdownTimeout == "" {
		return 30 * time.Second
	}
	d, err := time.ParseDuration(w.ShutdownTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

// GetHeartbeatInterval parses the heartbeat interval string and returns a duration.
// Returns the default value if not set or invalid.
func (w *WorkerConfig) GetHeartbeatInterval() time.Duration {
	if w == nil || w.HeartbeatInterval == "" {
		return 10 * time.Second
	}
	d, err := time.ParseDuration(w.HeartbeatInterval)
	if err != nil {
		return 10 * time.Second
	}
	return d
}

// GetConcurrency returns the configured concurrency or the default value.
func (w *WorkerConfig) GetConcurrency() int {
	if w == nil || w.Concurrency <= 0 {
		return 4
	}
	return w.Concurrency
}

// GetQueuePrefix returns the queue prefix or the default value.
func (w *WorkerConfig) GetQueuePrefix() string {
	if w == nil || w.QueuePrefix == "" {
		return "tool"
	}
	return w.QueuePrefix
}

// BuildConfig defines build configuration for the component.
type BuildConfig struct {
	Command string `yaml:"command,omitempty"` // Build command (e.g., "make build")
}

// Load reads and parses a component.yaml file from the given path.
// If the path is a directory, it looks for component.yaml or component.yml in that directory.
func Load(path string) (*Config, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	var configPath string
	if info.IsDir() {
		// Try component.yaml first, then component.yml
		yamlPath := filepath.Join(path, "component.yaml")
		if _, err := os.Stat(yamlPath); err == nil {
			configPath = yamlPath
		} else {
			ymlPath := filepath.Join(path, "component.yml")
			if _, err := os.Stat(ymlPath); err == nil {
				configPath = ymlPath
			} else {
				return nil, fmt.Errorf("no component.yaml or component.yml found in %s", path)
			}
		}
	} else {
		configPath = path
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadFromDir searches for component.yaml starting from the given directory
// and walking up to parent directories until found or root is reached.
func LoadFromDir(dir string) (*Config, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	for {
		config, err := Load(absDir)
		if err == nil {
			return config, nil
		}

		// Move to parent directory
		parent := filepath.Dir(absDir)
		if parent == absDir {
			// Reached root
			return nil, fmt.Errorf("no component.yaml found in %s or parent directories", dir)
		}
		absDir = parent
	}
}

// LoadFromCurrentDir loads component.yaml from the current working directory.
func LoadFromCurrentDir() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	return LoadFromDir(cwd)
}
