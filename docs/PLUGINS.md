# Gibson Plugin Development Guide

Complete reference for building stateful service plugins that integrate with the Gibson framework.

## Table of Contents

1. [Overview](#overview)
2. [Plugin Interface](#plugin-interface)
3. [Lifecycle Management](#lifecycle-management)
4. [Method Definitions](#method-definitions)
5. [JSON Schema Validation](#json-schema-validation)
6. [Building a Plugin](#building-a-plugin)
7. [Production Features](#production-features)
8. [Health Checks](#health-checks)
9. [Testing Plugins](#testing-plugins)
10. [Serving Plugins](#serving-plugins)
11. [Complete Examples](#complete-examples)
12. [Best Practices](#best-practices)

---

## Overview

Gibson plugins are **stateful services** that extend the framework with external integrations, APIs, and custom functionality. Key characteristics:

- **Stateful** - Maintain connections, caches, and state
- **JSON I/O** - Flexible map-based input/output
- **Lifecycle Managed** - Initialize/Shutdown hooks
- **Method-Based** - Named methods with JSON schemas
- **gRPC Distribution** - Serve plugins over the network

```
┌─────────────────────────────────────────────────────────────────┐
│                          AGENT                                   │
│                             │                                    │
│                             ▼                                    │
│       harness.QueryPlugin(ctx, "shodan", "search", params)      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      PLUGIN REGISTRY                             │
│  ┌─────────┐ ┌─────────────┐ ┌──────────┐ ┌─────────────────┐   │
│  │ shodan  │ │scope-ingest │ │ vector-db│ │ threat-intel    │   │
│  └─────────┘ └─────────────┘ └──────────┘ └─────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       YOUR PLUGIN                                │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   Initialize(ctx, config)                │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │    │
│  │  │ API      │ │  Cache   │ │   Pool   │ │  State   │   │    │
│  │  │ Client   │ │          │ │          │ │          │   │    │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │    │
│  │                                                          │    │
│  │            Query(ctx, method, params) -> any            │    │
│  │                                                          │    │
│  │                    Shutdown(ctx)                         │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### Plugins vs Tools

| Aspect | Plugins | Tools |
|--------|---------|-------|
| **State** | Maintains state (connections, caches) | Stateless |
| **I/O** | JSON maps (flexible) | Protocol Buffers (type-safe) |
| **Lifecycle** | Initialize/Shutdown | None |
| **Use Cases** | APIs, databases, services | Atomic operations |
| **Examples** | Shodan API, scope parser, vector DB | nmap, httpx, nuclei |

---

## Plugin Interface

Every plugin must implement the `Plugin` interface:

```go
type Plugin interface {
    // Identity
    Name() string                              // Unique identifier (e.g., "shodan")
    Version() string                           // Semantic version (e.g., "1.0.0")
    Description() string                       // Human-readable description

    // Method Definitions
    Methods() []MethodDescriptor               // Available methods with schemas

    // Lifecycle
    Initialize(ctx context.Context, config map[string]any) error
    Shutdown(ctx context.Context) error

    // Execution
    Query(ctx context.Context, method string, params map[string]any) (any, error)

    // Health
    Health(ctx context.Context) types.HealthStatus
}
```

### Method Descriptor

```go
type MethodDescriptor struct {
    Name         string      // Method name (e.g., "search")
    Description  string      // What the method does
    InputSchema  schema.JSON // JSON Schema for input validation
    OutputSchema schema.JSON // JSON Schema for output validation
}
```

---

## Lifecycle Management

Plugins have explicit lifecycle hooks for resource management.

### Initialization

Called once when the plugin is registered:

```go
func (p *MyPlugin) Initialize(ctx context.Context, config map[string]any) error {
    // Parse configuration
    apiKey, ok := config["api_key"].(string)
    if !ok || apiKey == "" {
        return fmt.Errorf("api_key required")
    }

    baseURL, _ := config["base_url"].(string)
    if baseURL == "" {
        baseURL = "https://api.default.com"
    }

    // Initialize HTTP client
    p.client = &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }

    // Initialize connection pool
    p.pool = newConnectionPool(baseURL, 10)

    // Initialize cache
    cacheTTL := 1 * time.Hour
    if ttl, ok := config["cache_ttl"].(int); ok {
        cacheTTL = time.Duration(ttl) * time.Second
    }
    p.cache = newCache(cacheTTL)

    // Initialize rate limiter
    rps := 10
    if r, ok := config["rate_limit"].(int); ok {
        rps = r
    }
    p.rateLimiter = newRateLimiter(rps)

    // Store config
    p.config = Config{
        APIKey:     apiKey,
        BaseURL:    baseURL,
        CacheTTL:   cacheTTL,
        RateLimit:  rps,
    }

    // Test connectivity
    if err := p.ping(ctx); err != nil {
        return fmt.Errorf("connectivity test failed: %w", err)
    }

    return nil
}
```

### Shutdown

Called when the plugin is unregistered or the framework shuts down:

```go
func (p *MyPlugin) Shutdown(ctx context.Context) error {
    var errs []error

    // Close connection pool
    if p.pool != nil {
        if err := p.pool.Close(); err != nil {
            errs = append(errs, fmt.Errorf("pool close: %w", err))
        }
    }

    // Flush cache
    if p.cache != nil {
        p.cache.Flush()
    }

    // Close any open connections
    if p.client != nil {
        p.client.CloseIdleConnections()
    }

    // Cancel background goroutines
    if p.cancel != nil {
        p.cancel()
    }

    // Wait for background tasks with timeout
    done := make(chan struct{})
    go func() {
        p.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        // Clean shutdown
    case <-ctx.Done():
        errs = append(errs, fmt.Errorf("shutdown timeout: %w", ctx.Err()))
    }

    if len(errs) > 0 {
        return fmt.Errorf("shutdown errors: %v", errs)
    }

    return nil
}
```

### Lifecycle States

```go
type PluginStatus string

const (
    PluginStatusUninitialized PluginStatus = "uninitialized"
    PluginStatusInitializing  PluginStatus = "initializing"
    PluginStatusRunning       PluginStatus = "running"
    PluginStatusStopping      PluginStatus = "stopping"
    PluginStatusStopped       PluginStatus = "stopped"
    PluginStatusError         PluginStatus = "error"
)
```

---

## Method Definitions

Plugins expose named methods that agents can call.

### Defining Methods

```go
func (p *MyPlugin) Methods() []plugin.MethodDescriptor {
    return []plugin.MethodDescriptor{
        {
            Name:        "search",
            Description: "Search for targets by query",
            InputSchema: schema.Object(map[string]schema.JSON{
                "query": schema.StringWithDesc("Search query string"),
                "limit": schema.Int().WithDefault(100).WithMin(1).WithMax(1000),
                "page":  schema.Int().WithDefault(1).WithMin(1),
                "filters": schema.Object(map[string]schema.JSON{
                    "country": schema.String(),
                    "port":    schema.Int(),
                    "product": schema.String(),
                }),
            }, "query"),  // "query" is required
            OutputSchema: schema.Object(map[string]schema.JSON{
                "results": schema.Array(schema.Object(map[string]schema.JSON{
                    "ip":       schema.String(),
                    "port":     schema.Int(),
                    "product":  schema.String(),
                    "version":  schema.String(),
                    "metadata": schema.Any(),
                })),
                "total":  schema.Int(),
                "page":   schema.Int(),
                "cached": schema.Bool(),
            }),
        },
        {
            Name:        "lookup",
            Description: "Lookup detailed information for a specific target",
            InputSchema: schema.Object(map[string]schema.JSON{
                "target": schema.StringWithDesc("IP address or hostname"),
            }, "target"),
            OutputSchema: schema.Object(map[string]schema.JSON{
                "ip":        schema.String(),
                "hostnames": schema.Array(schema.String()),
                "ports":     schema.Array(schema.Int()),
                "services":  schema.Array(schema.Any()),
                "vulns":     schema.Array(schema.String()),
                "last_seen": schema.String(),
            }),
        },
        {
            Name:        "count",
            Description: "Count results matching a query",
            InputSchema: schema.Object(map[string]schema.JSON{
                "query": schema.String(),
            }, "query"),
            OutputSchema: schema.Object(map[string]schema.JSON{
                "count": schema.Int(),
            }),
        },
    }
}
```

### Query Handler

```go
func (p *MyPlugin) Query(ctx context.Context, method string, params map[string]any) (any, error) {
    // Route to appropriate handler
    switch method {
    case "search":
        return p.search(ctx, params)
    case "lookup":
        return p.lookup(ctx, params)
    case "count":
        return p.count(ctx, params)
    default:
        return nil, fmt.Errorf("unknown method: %s", method)
    }
}

func (p *MyPlugin) search(ctx context.Context, params map[string]any) (any, error) {
    // Extract parameters
    query := params["query"].(string)

    limit := 100
    if l, ok := params["limit"].(int); ok {
        limit = l
    }

    page := 1
    if pg, ok := params["page"].(int); ok {
        page = pg
    }

    // Check cache
    cacheKey := fmt.Sprintf("search:%s:%d:%d", query, limit, page)
    if cached, ok := p.cache.Get(cacheKey); ok {
        result := cached.(map[string]any)
        result["cached"] = true
        return result, nil
    }

    // Apply rate limiting
    if err := p.rateLimiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit: %w", err)
    }

    // Make API request
    results, total, err := p.client.Search(ctx, query, limit, page)
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }

    response := map[string]any{
        "results": results,
        "total":   total,
        "page":    page,
        "cached":  false,
    }

    // Cache results
    p.cache.Set(cacheKey, response)

    return response, nil
}

func (p *MyPlugin) lookup(ctx context.Context, params map[string]any) (any, error) {
    target := params["target"].(string)

    // Apply rate limiting
    if err := p.rateLimiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit: %w", err)
    }

    // Make API request
    info, err := p.client.Lookup(ctx, target)
    if err != nil {
        return nil, fmt.Errorf("lookup failed: %w", err)
    }

    return map[string]any{
        "ip":        info.IP,
        "hostnames": info.Hostnames,
        "ports":     info.Ports,
        "services":  info.Services,
        "vulns":     info.Vulns,
        "last_seen": info.LastSeen.Format(time.RFC3339),
    }, nil
}

func (p *MyPlugin) count(ctx context.Context, params map[string]any) (any, error) {
    query := params["query"].(string)

    count, err := p.client.Count(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("count failed: %w", err)
    }

    return map[string]any{
        "count": count,
    }, nil
}
```

---

## JSON Schema Validation

The SDK provides schema helpers for input/output validation.

### Schema Types

```go
import "github.com/zero-day-ai/sdk/schema"

// Primitive types
schema.String()                          // Any string
schema.StringWithDesc("description")     // With description
schema.Int()                             // Integer
schema.Number()                          // Float64
schema.Bool()                            // Boolean
schema.Any()                             // Any type

// Modifiers
schema.String().
    WithDefault("default").              // Default value
    WithMinLength(1).                    // Min length
    WithMaxLength(100).                  // Max length
    WithPattern("^[a-z]+$").             // Regex pattern
    WithFormat("email")                  // Standard format

schema.Int().
    WithDefault(10).
    WithMin(1).
    WithMax(1000)

schema.Number().
    WithMin(0.0).
    WithMax(1.0)

// Arrays
schema.Array(schema.String())            // Array of strings
schema.Array(schema.Object(...))         // Array of objects

// Objects
schema.Object(
    map[string]schema.JSON{
        "name":  schema.String(),
        "age":   schema.Int(),
        "email": schema.String().WithFormat("email"),
    },
    "name", "age",                       // Required fields
)

// Enums
schema.Enum("low", "medium", "high")
```

### Validation Example

```go
// Input schema
inputSchema := schema.Object(map[string]schema.JSON{
    "query": schema.StringWithDesc("Search query").
        WithMinLength(1).
        WithMaxLength(1000),
    "limit": schema.Int().
        WithDefault(100).
        WithMin(1).
        WithMax(1000),
    "severity": schema.Enum("low", "medium", "high", "critical"),
    "tags": schema.Array(schema.String()).
        WithMinItems(0).
        WithMaxItems(10),
}, "query")  // "query" is required

// Validate input
if err := inputSchema.Validate(params); err != nil {
    return nil, fmt.Errorf("invalid input: %w", err)
}
```

---

## Building a Plugin

### Step 1: Project Structure

```
myplugin/
├── plugin.go              # Plugin implementation
├── client.go              # API client
├── cache.go               # Caching layer
├── ratelimit.go           # Rate limiter
├── main.go                # Entry point
├── component.yaml         # Component metadata
├── go.mod
├── go.sum
└── Makefile
```

### Step 2: Define Configuration

```go
// config.go
package myplugin

import "time"

type Config struct {
    // API Settings
    APIKey     string        `json:"api_key"`
    BaseURL    string        `json:"base_url"`

    // Caching
    CacheTTL   time.Duration `json:"cache_ttl"`
    CacheSize  int           `json:"cache_size"`

    // Rate Limiting
    RateLimit  int           `json:"rate_limit"`  // Requests per second

    // Timeouts
    Timeout    time.Duration `json:"timeout"`

    // Circuit Breaker
    CBThreshold int          `json:"cb_threshold"`  // Failures before open
    CBTimeout   time.Duration `json:"cb_timeout"`    // Time before retry
}

func DefaultConfig() Config {
    return Config{
        BaseURL:     "https://api.example.com",
        CacheTTL:    1 * time.Hour,
        CacheSize:   10000,
        RateLimit:   10,
        Timeout:     30 * time.Second,
        CBThreshold: 5,
        CBTimeout:   60 * time.Second,
    }
}

func ParseConfig(raw map[string]any) (Config, error) {
    cfg := DefaultConfig()

    if key, ok := raw["api_key"].(string); ok {
        cfg.APIKey = key
    }

    if url, ok := raw["base_url"].(string); ok && url != "" {
        cfg.BaseURL = url
    }

    if ttl, ok := raw["cache_ttl"].(int); ok {
        cfg.CacheTTL = time.Duration(ttl) * time.Second
    }

    if rps, ok := raw["rate_limit"].(int); ok {
        cfg.RateLimit = rps
    }

    if timeout, ok := raw["timeout"].(int); ok {
        cfg.Timeout = time.Duration(timeout) * time.Second
    }

    return cfg, nil
}
```

### Step 3: Implement the Plugin

```go
// plugin.go
package myplugin

import (
    "context"
    "fmt"
    "sync"

    "github.com/zero-day-ai/sdk/plugin"
    "github.com/zero-day-ai/sdk/schema"
    "github.com/zero-day-ai/sdk/types"
)

type Plugin struct {
    mu          sync.RWMutex
    config      Config
    client      *APIClient
    cache       *Cache
    rateLimiter *RateLimiter
    cb          *CircuitBreaker

    // Background tasks
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

func New() *Plugin {
    return &Plugin{}
}

// ═══════════════════════════════════════════════════════════════
// IDENTITY
// ═══════════════════════════════════════════════════════════════

func (p *Plugin) Name() string {
    return "myplugin"
}

func (p *Plugin) Version() string {
    return "1.0.0"
}

func (p *Plugin) Description() string {
    return "Integration with external security intelligence API"
}

// ═══════════════════════════════════════════════════════════════
// METHODS
// ═══════════════════════════════════════════════════════════════

func (p *Plugin) Methods() []plugin.MethodDescriptor {
    return []plugin.MethodDescriptor{
        {
            Name:        "search",
            Description: "Search for targets matching a query",
            InputSchema: schema.Object(map[string]schema.JSON{
                "query": schema.StringWithDesc("Search query"),
                "limit": schema.Int().WithDefault(100).WithMin(1).WithMax(1000),
                "page":  schema.Int().WithDefault(1).WithMin(1),
            }, "query"),
            OutputSchema: schema.Object(map[string]schema.JSON{
                "results": schema.Array(schema.Any()),
                "total":   schema.Int(),
                "cached":  schema.Bool(),
            }),
        },
        {
            Name:        "lookup",
            Description: "Get detailed information for a target",
            InputSchema: schema.Object(map[string]schema.JSON{
                "target": schema.StringWithDesc("IP or hostname"),
            }, "target"),
            OutputSchema: schema.Any(),
        },
        {
            Name:        "enrich",
            Description: "Enrich target data with additional context",
            InputSchema: schema.Object(map[string]schema.JSON{
                "targets": schema.Array(schema.String()),
                "fields":  schema.Array(schema.Enum("geo", "asn", "vulns", "ports")),
            }, "targets"),
            OutputSchema: schema.Object(map[string]schema.JSON{
                "enriched": schema.Array(schema.Any()),
                "errors":   schema.Array(schema.String()),
            }),
        },
    }
}

// ═══════════════════════════════════════════════════════════════
// LIFECYCLE
// ═══════════════════════════════════════════════════════════════

func (p *Plugin) Initialize(ctx context.Context, config map[string]any) error {
    // Parse configuration
    cfg, err := ParseConfig(config)
    if err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }

    if cfg.APIKey == "" {
        return fmt.Errorf("api_key is required")
    }

    // Store config
    p.config = cfg

    // Create background context
    p.ctx, p.cancel = context.WithCancel(context.Background())

    // Initialize API client
    p.client = NewAPIClient(cfg.BaseURL, cfg.APIKey, cfg.Timeout)

    // Initialize cache
    p.cache = NewCache(cfg.CacheTTL, cfg.CacheSize)

    // Initialize rate limiter
    p.rateLimiter = NewRateLimiter(cfg.RateLimit)

    // Initialize circuit breaker
    p.cb = NewCircuitBreaker(cfg.CBThreshold, cfg.CBTimeout)

    // Start background cache cleanup
    p.wg.Add(1)
    go func() {
        defer p.wg.Done()
        p.cache.CleanupLoop(p.ctx)
    }()

    // Test connectivity
    if err := p.client.Ping(ctx); err != nil {
        return fmt.Errorf("connectivity test failed: %w", err)
    }

    return nil
}

func (p *Plugin) Shutdown(ctx context.Context) error {
    // Cancel background tasks
    if p.cancel != nil {
        p.cancel()
    }

    // Wait for background tasks
    done := make(chan struct{})
    go func() {
        p.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
    case <-ctx.Done():
        return ctx.Err()
    }

    // Flush cache
    if p.cache != nil {
        p.cache.Flush()
    }

    // Close client
    if p.client != nil {
        p.client.Close()
    }

    return nil
}

// ═══════════════════════════════════════════════════════════════
// QUERY HANDLER
// ═══════════════════════════════════════════════════════════════

func (p *Plugin) Query(ctx context.Context, method string, params map[string]any) (any, error) {
    // Check circuit breaker
    if !p.cb.Allow() {
        return nil, fmt.Errorf("service unavailable (circuit breaker open)")
    }

    var result any
    var err error

    switch method {
    case "search":
        result, err = p.search(ctx, params)
    case "lookup":
        result, err = p.lookup(ctx, params)
    case "enrich":
        result, err = p.enrich(ctx, params)
    default:
        return nil, fmt.Errorf("unknown method: %s", method)
    }

    // Update circuit breaker
    if err != nil {
        p.cb.RecordFailure()
    } else {
        p.cb.RecordSuccess()
    }

    return result, err
}

func (p *Plugin) search(ctx context.Context, params map[string]any) (any, error) {
    query := params["query"].(string)
    limit := getIntParam(params, "limit", 100)
    page := getIntParam(params, "page", 1)

    // Check cache
    cacheKey := fmt.Sprintf("search:%s:%d:%d", query, limit, page)
    if cached, ok := p.cache.Get(cacheKey); ok {
        result := cached.(map[string]any)
        result["cached"] = true
        return result, nil
    }

    // Rate limit
    if err := p.rateLimiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit exceeded: %w", err)
    }

    // Make request
    results, total, err := p.client.Search(ctx, query, limit, page)
    if err != nil {
        return nil, err
    }

    response := map[string]any{
        "results": results,
        "total":   total,
        "cached":  false,
    }

    // Cache result
    p.cache.Set(cacheKey, response)

    return response, nil
}

func (p *Plugin) lookup(ctx context.Context, params map[string]any) (any, error) {
    target := params["target"].(string)

    // Check cache
    cacheKey := fmt.Sprintf("lookup:%s", target)
    if cached, ok := p.cache.Get(cacheKey); ok {
        return cached, nil
    }

    // Rate limit
    if err := p.rateLimiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit exceeded: %w", err)
    }

    // Make request
    info, err := p.client.Lookup(ctx, target)
    if err != nil {
        return nil, err
    }

    // Cache result
    p.cache.Set(cacheKey, info)

    return info, nil
}

func (p *Plugin) enrich(ctx context.Context, params map[string]any) (any, error) {
    targets := toStringSlice(params["targets"])
    fields := toStringSlice(params["fields"])

    var enriched []any
    var errors []string

    for _, target := range targets {
        // Rate limit per target
        if err := p.rateLimiter.Wait(ctx); err != nil {
            errors = append(errors, fmt.Sprintf("%s: rate limit", target))
            continue
        }

        info, err := p.client.Enrich(ctx, target, fields)
        if err != nil {
            errors = append(errors, fmt.Sprintf("%s: %v", target, err))
            continue
        }

        enriched = append(enriched, info)
    }

    return map[string]any{
        "enriched": enriched,
        "errors":   errors,
    }, nil
}

// ═══════════════════════════════════════════════════════════════
// HEALTH
// ═══════════════════════════════════════════════════════════════

func (p *Plugin) Health(ctx context.Context) types.HealthStatus {
    // Check circuit breaker state
    if !p.cb.Allow() {
        return types.NewUnhealthyStatus("circuit breaker open", map[string]any{
            "failures": p.cb.Failures(),
        })
    }

    // Test API connectivity
    if err := p.client.Ping(ctx); err != nil {
        return types.NewUnhealthyStatus("API unreachable", map[string]any{
            "error": err.Error(),
        })
    }

    // Check cache health
    cacheStats := p.cache.Stats()

    return types.NewHealthyStatus("operational", map[string]any{
        "cache_size":     cacheStats.Size,
        "cache_hit_rate": cacheStats.HitRate,
        "circuit_state":  p.cb.State(),
    })
}

// ═══════════════════════════════════════════════════════════════
// HELPERS
// ═══════════════════════════════════════════════════════════════

func getIntParam(params map[string]any, key string, defaultVal int) int {
    if v, ok := params[key].(int); ok {
        return v
    }
    if v, ok := params[key].(float64); ok {
        return int(v)
    }
    return defaultVal
}

func toStringSlice(v any) []string {
    if v == nil {
        return nil
    }
    slice, ok := v.([]any)
    if !ok {
        return nil
    }
    result := make([]string, 0, len(slice))
    for _, item := range slice {
        if s, ok := item.(string); ok {
            result = append(result, s)
        }
    }
    return result
}
```

### Step 4: Entry Point

```go
// main.go
package main

import (
    "os"
    "os/signal"
    "syscall"

    "github.com/zero-day-ai/plugins/myplugin"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    plugin := myplugin.New()

    // Handle graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        // Server will handle graceful shutdown
    }()

    // Serve via gRPC
    serve.Plugin(plugin,
        serve.WithPort(50053),
        serve.WithHealthEndpoint("/health"),
        serve.WithGracefulTimeout(30*time.Second),
    )
}
```

### Step 5: Component Metadata

```yaml
# component.yaml
name: myplugin
version: 1.0.0
type: plugin
description: Integration with external security intelligence API

methods:
  - name: search
    description: Search for targets matching a query
    input_schema:
      type: object
      properties:
        query:
          type: string
          description: Search query
        limit:
          type: integer
          default: 100
          minimum: 1
          maximum: 1000
        page:
          type: integer
          default: 1
          minimum: 1
      required:
        - query
    output_schema:
      type: object
      properties:
        results:
          type: array
        total:
          type: integer
        cached:
          type: boolean

  - name: lookup
    description: Get detailed information for a target
    input_schema:
      type: object
      properties:
        target:
          type: string
          description: IP or hostname
      required:
        - target

  - name: enrich
    description: Enrich target data with additional context

config_options:
  - name: api_key
    type: string
    description: API authentication key
    required: true
    env: MY_API_KEY
  - name: base_url
    type: string
    description: API base URL
    default: "https://api.example.com"
  - name: cache_ttl
    type: integer
    description: Cache TTL in seconds
    default: 3600
  - name: rate_limit
    type: integer
    description: Requests per second
    default: 10
```

---

## Production Features

### Caching

```go
// cache.go
package myplugin

import (
    "context"
    "sync"
    "time"
)

type Cache struct {
    mu      sync.RWMutex
    items   map[string]*CacheItem
    ttl     time.Duration
    maxSize int

    // Stats
    hits   int64
    misses int64
}

type CacheItem struct {
    Value     any
    ExpiresAt time.Time
}

func NewCache(ttl time.Duration, maxSize int) *Cache {
    return &Cache{
        items:   make(map[string]*CacheItem),
        ttl:     ttl,
        maxSize: maxSize,
    }
}

func (c *Cache) Get(key string) (any, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    item, ok := c.items[key]
    if !ok {
        c.misses++
        return nil, false
    }

    if time.Now().After(item.ExpiresAt) {
        c.misses++
        return nil, false
    }

    c.hits++
    return item.Value, true
}

func (c *Cache) Set(key string, value any) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Evict if at capacity
    if len(c.items) >= c.maxSize {
        c.evictOldest()
    }

    c.items[key] = &CacheItem{
        Value:     value,
        ExpiresAt: time.Now().Add(c.ttl),
    }
}

func (c *Cache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.items, key)
}

func (c *Cache) Flush() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items = make(map[string]*CacheItem)
}

func (c *Cache) evictOldest() {
    var oldestKey string
    var oldestTime time.Time

    for key, item := range c.items {
        if oldestKey == "" || item.ExpiresAt.Before(oldestTime) {
            oldestKey = key
            oldestTime = item.ExpiresAt
        }
    }

    if oldestKey != "" {
        delete(c.items, oldestKey)
    }
}

func (c *Cache) CleanupLoop(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            c.cleanup()
        }
    }
}

func (c *Cache) cleanup() {
    c.mu.Lock()
    defer c.mu.Unlock()

    now := time.Now()
    for key, item := range c.items {
        if now.After(item.ExpiresAt) {
            delete(c.items, key)
        }
    }
}

type CacheStats struct {
    Size    int
    HitRate float64
}

func (c *Cache) Stats() CacheStats {
    c.mu.RLock()
    defer c.mu.RUnlock()

    total := c.hits + c.misses
    hitRate := 0.0
    if total > 0 {
        hitRate = float64(c.hits) / float64(total)
    }

    return CacheStats{
        Size:    len(c.items),
        HitRate: hitRate,
    }
}
```

### Rate Limiting

```go
// ratelimit.go
package myplugin

import (
    "context"
    "sync"
    "time"
)

type RateLimiter struct {
    mu       sync.Mutex
    tokens   int
    maxTokens int
    interval time.Duration
    lastRefill time.Time
}

func NewRateLimiter(rps int) *RateLimiter {
    return &RateLimiter{
        tokens:     rps,
        maxTokens:  rps,
        interval:   time.Second / time.Duration(rps),
        lastRefill: time.Now(),
    }
}

func (r *RateLimiter) Wait(ctx context.Context) error {
    r.mu.Lock()

    // Refill tokens based on elapsed time
    now := time.Now()
    elapsed := now.Sub(r.lastRefill)
    newTokens := int(elapsed / r.interval)
    if newTokens > 0 {
        r.tokens = min(r.tokens+newTokens, r.maxTokens)
        r.lastRefill = now
    }

    // Check if we have tokens
    if r.tokens > 0 {
        r.tokens--
        r.mu.Unlock()
        return nil
    }

    // Calculate wait time
    waitTime := r.interval - (now.Sub(r.lastRefill) % r.interval)
    r.mu.Unlock()

    // Wait with context
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(waitTime):
        return r.Wait(ctx)  // Retry after wait
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

### Circuit Breaker

```go
// circuitbreaker.go
package myplugin

import (
    "sync"
    "time"
)

type CircuitState string

const (
    CircuitClosed   CircuitState = "closed"
    CircuitOpen     CircuitState = "open"
    CircuitHalfOpen CircuitState = "half-open"
)

type CircuitBreaker struct {
    mu        sync.RWMutex
    state     CircuitState
    failures  int
    threshold int
    timeout   time.Duration
    openUntil time.Time
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        state:     CircuitClosed,
        threshold: threshold,
        timeout:   timeout,
    }
}

func (cb *CircuitBreaker) Allow() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.state {
    case CircuitClosed:
        return true
    case CircuitOpen:
        if time.Now().After(cb.openUntil) {
            cb.state = CircuitHalfOpen
            return true
        }
        return false
    case CircuitHalfOpen:
        return true
    }
    return false
}

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failures = 0
    cb.state = CircuitClosed
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failures++

    if cb.state == CircuitHalfOpen {
        cb.state = CircuitOpen
        cb.openUntil = time.Now().Add(cb.timeout)
        return
    }

    if cb.failures >= cb.threshold {
        cb.state = CircuitOpen
        cb.openUntil = time.Now().Add(cb.timeout)
    }
}

func (cb *CircuitBreaker) State() CircuitState {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.state
}

func (cb *CircuitBreaker) Failures() int {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.failures
}
```

---

## Health Checks

### Comprehensive Health Reporting

```go
func (p *Plugin) Health(ctx context.Context) types.HealthStatus {
    details := make(map[string]any)
    issues := []string{}

    // Check API connectivity
    apiHealthy := true
    if err := p.client.Ping(ctx); err != nil {
        apiHealthy = false
        issues = append(issues, fmt.Sprintf("API: %v", err))
    }
    details["api_healthy"] = apiHealthy

    // Check circuit breaker
    cbState := p.cb.State()
    details["circuit_breaker"] = string(cbState)
    if cbState == CircuitOpen {
        issues = append(issues, "circuit breaker open")
    }

    // Check rate limiter
    details["rate_limit_remaining"] = p.rateLimiter.Available()

    // Check cache
    cacheStats := p.cache.Stats()
    details["cache_size"] = cacheStats.Size
    details["cache_hit_rate"] = fmt.Sprintf("%.2f%%", cacheStats.HitRate*100)

    // Determine overall status
    if !apiHealthy || cbState == CircuitOpen {
        return types.NewUnhealthyStatus(
            strings.Join(issues, "; "),
            details,
        )
    }

    if cbState == CircuitHalfOpen {
        return types.NewDegradedStatus(
            "circuit breaker recovering",
            details,
        )
    }

    return types.NewHealthyStatus("operational", details)
}
```

---

## Testing Plugins

### Unit Tests

```go
// plugin_test.go
package myplugin

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPlugin_Initialize(t *testing.T) {
    plugin := New()

    // Missing API key
    err := plugin.Initialize(context.Background(), map[string]any{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "api_key")

    // Valid config
    err = plugin.Initialize(context.Background(), map[string]any{
        "api_key":    "test-key",
        "base_url":   "https://api.test.com",
        "cache_ttl":  3600,
        "rate_limit": 10,
    })
    // May fail if no network, but config should be valid
}

func TestPlugin_Methods(t *testing.T) {
    plugin := New()
    methods := plugin.Methods()

    assert.Len(t, methods, 3)

    // Check search method
    var searchMethod *plugin.MethodDescriptor
    for _, m := range methods {
        if m.Name == "search" {
            searchMethod = &m
            break
        }
    }
    require.NotNil(t, searchMethod)
    assert.Equal(t, "search", searchMethod.Name)
    assert.NotEmpty(t, searchMethod.Description)
}

func TestPlugin_Query_UnknownMethod(t *testing.T) {
    plugin := New()
    plugin.Initialize(context.Background(), map[string]any{
        "api_key": "test",
    })

    _, err := plugin.Query(context.Background(), "unknown", nil)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unknown method")
}

func TestCache(t *testing.T) {
    cache := NewCache(1*time.Second, 100)

    // Set and get
    cache.Set("key1", "value1")
    val, ok := cache.Get("key1")
    assert.True(t, ok)
    assert.Equal(t, "value1", val)

    // Miss
    _, ok = cache.Get("nonexistent")
    assert.False(t, ok)

    // Expiration
    time.Sleep(1100 * time.Millisecond)
    _, ok = cache.Get("key1")
    assert.False(t, ok)
}

func TestRateLimiter(t *testing.T) {
    limiter := NewRateLimiter(10)  // 10 RPS

    ctx := context.Background()

    // Should allow initial requests
    for i := 0; i < 10; i++ {
        err := limiter.Wait(ctx)
        assert.NoError(t, err)
    }

    // Next request should wait
    ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
    defer cancel()

    start := time.Now()
    err := limiter.Wait(ctx)
    elapsed := time.Since(start)

    // Should either succeed after wait or timeout
    if err == nil {
        assert.True(t, elapsed > 50*time.Millisecond)
    } else {
        assert.Equal(t, context.DeadlineExceeded, err)
    }
}

func TestCircuitBreaker(t *testing.T) {
    cb := NewCircuitBreaker(3, 100*time.Millisecond)

    // Initially closed
    assert.True(t, cb.Allow())
    assert.Equal(t, CircuitClosed, cb.State())

    // Record failures
    cb.RecordFailure()
    cb.RecordFailure()
    assert.True(t, cb.Allow())  // Still closed

    cb.RecordFailure()  // Threshold reached
    assert.Equal(t, CircuitOpen, cb.State())
    assert.False(t, cb.Allow())

    // Wait for timeout
    time.Sleep(150 * time.Millisecond)
    assert.True(t, cb.Allow())  // Half-open
    assert.Equal(t, CircuitHalfOpen, cb.State())

    // Success recovers
    cb.RecordSuccess()
    assert.Equal(t, CircuitClosed, cb.State())
}
```

### Integration Tests

```go
// +build integration

package myplugin

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
)

func TestPlugin_Integration(t *testing.T) {
    apiKey := os.Getenv("MY_API_KEY")
    if apiKey == "" {
        t.Skip("MY_API_KEY not set")
    }

    plugin := New()

    // Initialize
    err := plugin.Initialize(context.Background(), map[string]any{
        "api_key": apiKey,
    })
    require.NoError(t, err)
    defer plugin.Shutdown(context.Background())

    // Health check
    health := plugin.Health(context.Background())
    require.True(t, health.IsHealthy(), "plugin not healthy: %s", health.Message)

    // Search
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    result, err := plugin.Query(ctx, "search", map[string]any{
        "query": "nginx",
        "limit": 10,
    })
    require.NoError(t, err)

    resultMap := result.(map[string]any)
    require.Contains(t, resultMap, "results")
    require.Contains(t, resultMap, "total")
}
```

---

## Serving Plugins

### gRPC Server

```go
package main

import (
    "github.com/zero-day-ai/plugins/myplugin"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    plugin := myplugin.New()

    serve.Plugin(plugin,
        serve.WithPort(50053),
        serve.WithTLS("cert.pem", "key.pem"),
        serve.WithGracefulTimeout(30*time.Second),
    )
}
```

---

## Complete Examples

### Shodan-like Intelligence Plugin

```go
package shodan

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "sync"
    "time"

    "github.com/zero-day-ai/sdk/plugin"
    "github.com/zero-day-ai/sdk/schema"
    "github.com/zero-day-ai/sdk/types"
)

type Plugin struct {
    mu          sync.RWMutex
    apiKey      string
    baseURL     string
    client      *http.Client
    cache       *Cache
    rateLimiter *RateLimiter
}

func New() *Plugin {
    return &Plugin{}
}

func (p *Plugin) Name() string        { return "shodan" }
func (p *Plugin) Version() string     { return "1.0.0" }
func (p *Plugin) Description() string { return "Shodan internet intelligence API integration" }

func (p *Plugin) Methods() []plugin.MethodDescriptor {
    return []plugin.MethodDescriptor{
        {
            Name:        "search",
            Description: "Search Shodan for hosts matching a query",
            InputSchema: schema.Object(map[string]schema.JSON{
                "query": schema.StringWithDesc("Shodan search query (e.g., 'nginx country:US')"),
                "limit": schema.Int().WithDefault(100).WithMin(1).WithMax(1000),
                "page":  schema.Int().WithDefault(1).WithMin(1),
            }, "query"),
            OutputSchema: schema.Object(map[string]schema.JSON{
                "matches": schema.Array(schema.Any()),
                "total":   schema.Int(),
            }),
        },
        {
            Name:        "host",
            Description: "Get information about a specific IP",
            InputSchema: schema.Object(map[string]schema.JSON{
                "ip": schema.StringWithDesc("IP address to lookup"),
            }, "ip"),
            OutputSchema: schema.Any(),
        },
        {
            Name:        "dns_resolve",
            Description: "Resolve hostnames to IPs",
            InputSchema: schema.Object(map[string]schema.JSON{
                "hostnames": schema.Array(schema.String()),
            }, "hostnames"),
            OutputSchema: schema.Object(map[string]schema.JSON{
                "resolved": schema.Any(),
            }),
        },
        {
            Name:        "dns_reverse",
            Description: "Reverse DNS lookup for IPs",
            InputSchema: schema.Object(map[string]schema.JSON{
                "ips": schema.Array(schema.String()),
            }, "ips"),
            OutputSchema: schema.Object(map[string]schema.JSON{
                "hostnames": schema.Any(),
            }),
        },
        {
            Name:        "exploits",
            Description: "Search for exploits related to a query",
            InputSchema: schema.Object(map[string]schema.JSON{
                "query": schema.String(),
            }, "query"),
            OutputSchema: schema.Object(map[string]schema.JSON{
                "exploits": schema.Array(schema.Any()),
                "total":    schema.Int(),
            }),
        },
    }
}

func (p *Plugin) Initialize(ctx context.Context, config map[string]any) error {
    apiKey, ok := config["api_key"].(string)
    if !ok || apiKey == "" {
        return fmt.Errorf("api_key required")
    }

    p.apiKey = apiKey
    p.baseURL = "https://api.shodan.io"

    if baseURL, ok := config["base_url"].(string); ok && baseURL != "" {
        p.baseURL = baseURL
    }

    p.client = &http.Client{
        Timeout: 30 * time.Second,
    }

    cacheTTL := 1 * time.Hour
    if ttl, ok := config["cache_ttl"].(int); ok {
        cacheTTL = time.Duration(ttl) * time.Second
    }
    p.cache = NewCache(cacheTTL, 10000)

    rps := 1  // Shodan free tier is 1 RPS
    if r, ok := config["rate_limit"].(int); ok {
        rps = r
    }
    p.rateLimiter = NewRateLimiter(rps)

    // Verify API key
    if err := p.verifyKey(ctx); err != nil {
        return fmt.Errorf("API key verification failed: %w", err)
    }

    return nil
}

func (p *Plugin) Shutdown(ctx context.Context) error {
    if p.cache != nil {
        p.cache.Flush()
    }
    return nil
}

func (p *Plugin) Query(ctx context.Context, method string, params map[string]any) (any, error) {
    switch method {
    case "search":
        return p.search(ctx, params)
    case "host":
        return p.host(ctx, params)
    case "dns_resolve":
        return p.dnsResolve(ctx, params)
    case "dns_reverse":
        return p.dnsReverse(ctx, params)
    case "exploits":
        return p.exploits(ctx, params)
    default:
        return nil, fmt.Errorf("unknown method: %s", method)
    }
}

func (p *Plugin) search(ctx context.Context, params map[string]any) (any, error) {
    query := params["query"].(string)
    page := 1
    if pg, ok := params["page"].(int); ok {
        page = pg
    }

    // Check cache
    cacheKey := fmt.Sprintf("search:%s:%d", query, page)
    if cached, ok := p.cache.Get(cacheKey); ok {
        return cached, nil
    }

    // Rate limit
    if err := p.rateLimiter.Wait(ctx); err != nil {
        return nil, err
    }

    // Build URL
    endpoint := fmt.Sprintf("%s/shodan/host/search", p.baseURL)
    u, _ := url.Parse(endpoint)
    q := u.Query()
    q.Set("key", p.apiKey)
    q.Set("query", query)
    q.Set("page", fmt.Sprintf("%d", page))
    u.RawQuery = q.Encode()

    // Make request
    resp, err := p.client.Get(u.String())
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: %s", resp.Status)
    }

    var result map[string]any
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    // Cache result
    p.cache.Set(cacheKey, result)

    return result, nil
}

func (p *Plugin) host(ctx context.Context, params map[string]any) (any, error) {
    ip := params["ip"].(string)

    // Check cache
    cacheKey := fmt.Sprintf("host:%s", ip)
    if cached, ok := p.cache.Get(cacheKey); ok {
        return cached, nil
    }

    // Rate limit
    if err := p.rateLimiter.Wait(ctx); err != nil {
        return nil, err
    }

    // Build URL
    endpoint := fmt.Sprintf("%s/shodan/host/%s?key=%s", p.baseURL, ip, p.apiKey)

    resp, err := p.client.Get(endpoint)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: %s", resp.Status)
    }

    var result map[string]any
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    p.cache.Set(cacheKey, result)
    return result, nil
}

func (p *Plugin) dnsResolve(ctx context.Context, params map[string]any) (any, error) {
    hostnames := toStringSlice(params["hostnames"])
    if len(hostnames) == 0 {
        return nil, fmt.Errorf("hostnames required")
    }

    if err := p.rateLimiter.Wait(ctx); err != nil {
        return nil, err
    }

    endpoint := fmt.Sprintf("%s/dns/resolve?key=%s&hostnames=%s",
        p.baseURL, p.apiKey, url.QueryEscape(strings.Join(hostnames, ",")))

    resp, err := p.client.Get(endpoint)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]any
    json.NewDecoder(resp.Body).Decode(&result)

    return map[string]any{"resolved": result}, nil
}

func (p *Plugin) dnsReverse(ctx context.Context, params map[string]any) (any, error) {
    ips := toStringSlice(params["ips"])
    if len(ips) == 0 {
        return nil, fmt.Errorf("ips required")
    }

    if err := p.rateLimiter.Wait(ctx); err != nil {
        return nil, err
    }

    endpoint := fmt.Sprintf("%s/dns/reverse?key=%s&ips=%s",
        p.baseURL, p.apiKey, url.QueryEscape(strings.Join(ips, ",")))

    resp, err := p.client.Get(endpoint)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]any
    json.NewDecoder(resp.Body).Decode(&result)

    return map[string]any{"hostnames": result}, nil
}

func (p *Plugin) exploits(ctx context.Context, params map[string]any) (any, error) {
    query := params["query"].(string)

    if err := p.rateLimiter.Wait(ctx); err != nil {
        return nil, err
    }

    endpoint := fmt.Sprintf("%s/api-ms-test/search?key=%s&query=%s",
        "https://exploits.shodan.io", p.apiKey, url.QueryEscape(query))

    resp, err := p.client.Get(endpoint)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]any
    json.NewDecoder(resp.Body).Decode(&result)

    return result, nil
}

func (p *Plugin) verifyKey(ctx context.Context) error {
    endpoint := fmt.Sprintf("%s/api-info?key=%s", p.baseURL, p.apiKey)

    req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
    resp, err := p.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("invalid API key")
    }

    return nil
}

func (p *Plugin) Health(ctx context.Context) types.HealthStatus {
    if err := p.verifyKey(ctx); err != nil {
        return types.NewUnhealthyStatus("API unreachable", map[string]any{
            "error": err.Error(),
        })
    }

    return types.NewHealthyStatus("connected to Shodan API")
}

func toStringSlice(v any) []string {
    if v == nil {
        return nil
    }
    slice, ok := v.([]any)
    if !ok {
        return nil
    }
    result := make([]string, 0, len(slice))
    for _, item := range slice {
        if s, ok := item.(string); ok {
            result = append(result, s)
        }
    }
    return result
}
```

---

## Best Practices

### 1. Validate Configuration Early

```go
func (p *Plugin) Initialize(ctx context.Context, config map[string]any) error {
    // Validate required fields immediately
    apiKey, ok := config["api_key"].(string)
    if !ok || apiKey == "" {
        return fmt.Errorf("api_key is required")
    }

    // Validate types and ranges
    if rps, ok := config["rate_limit"].(int); ok {
        if rps < 1 || rps > 100 {
            return fmt.Errorf("rate_limit must be 1-100")
        }
    }

    // Test connectivity before completing initialization
    if err := p.testConnection(ctx); err != nil {
        return fmt.Errorf("connectivity test failed: %w", err)
    }

    return nil
}
```

### 2. Implement Graceful Shutdown

```go
func (p *Plugin) Shutdown(ctx context.Context) error {
    // Signal background goroutines to stop
    if p.cancel != nil {
        p.cancel()
    }

    // Wait with timeout
    done := make(chan struct{})
    go func() {
        p.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        // Clean shutdown
    case <-ctx.Done():
        return fmt.Errorf("shutdown timeout")
    }

    // Clean up resources
    p.cache.Flush()
    p.client.CloseIdleConnections()

    return nil
}
```

### 3. Cache Aggressively

```go
func (p *Plugin) lookup(ctx context.Context, params map[string]any) (any, error) {
    target := params["target"].(string)

    // Always check cache first
    cacheKey := fmt.Sprintf("lookup:%s", target)
    if cached, ok := p.cache.Get(cacheKey); ok {
        return cached, nil
    }

    // ... make API call ...

    // Cache successful results
    p.cache.Set(cacheKey, result)

    return result, nil
}
```

### 4. Use Rate Limiting

```go
func (p *Plugin) Query(ctx context.Context, method string, params map[string]any) (any, error) {
    // Always rate limit API calls
    if err := p.rateLimiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit: %w", err)
    }

    // ... rest of implementation
}
```

### 5. Implement Circuit Breakers

```go
func (p *Plugin) Query(ctx context.Context, method string, params map[string]any) (any, error) {
    // Check circuit breaker
    if !p.cb.Allow() {
        return nil, fmt.Errorf("service unavailable (circuit breaker open)")
    }

    result, err := p.doQuery(ctx, method, params)

    // Update circuit breaker
    if err != nil {
        p.cb.RecordFailure()
    } else {
        p.cb.RecordSuccess()
    }

    return result, err
}
```

### 6. Return Structured Errors

```go
// Good: Structured error information
return nil, &PluginError{
    Code:    "API_ERROR",
    Message: "Shodan API returned error",
    Details: map[string]any{
        "status_code": resp.StatusCode,
        "body":        body,
    },
    Retryable: resp.StatusCode >= 500,
}

// Bad: Generic error
return nil, fmt.Errorf("API error")
```

### 7. Document Methods Thoroughly

```go
{
    Name: "search",
    Description: `Search for hosts matching a Shodan query.

Query syntax supports filters like:
- country:US - Filter by country
- port:22 - Filter by port
- product:nginx - Filter by product
- vuln:CVE-2021-44228 - Filter by vulnerability

Example: "nginx port:80 country:US"`,
    InputSchema: schema.Object(...),
}
```

### 8. Handle Context Cancellation

```go
func (p *Plugin) search(ctx context.Context, params map[string]any) (any, error) {
    // Check context before expensive operations
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Use context for HTTP requests
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := p.client.Do(req)

    // ...
}
```

### 9. Thread-Safe State Access

```go
func (p *Plugin) updateStats(success bool) {
    p.mu.Lock()
    defer p.mu.Unlock()

    if success {
        p.successCount++
    } else {
        p.failureCount++
    }
}

func (p *Plugin) getStats() Stats {
    p.mu.RLock()
    defer p.mu.RUnlock()

    return Stats{
        Successes: p.successCount,
        Failures:  p.failureCount,
    }
}
```

### 10. Comprehensive Health Checks

```go
func (p *Plugin) Health(ctx context.Context) types.HealthStatus {
    details := map[string]any{}
    issues := []string{}

    // Check API
    if err := p.pingAPI(ctx); err != nil {
        issues = append(issues, fmt.Sprintf("API: %v", err))
    }
    details["api_healthy"] = len(issues) == 0

    // Check cache
    cacheStats := p.cache.Stats()
    details["cache_size"] = cacheStats.Size
    details["cache_hit_rate"] = cacheStats.HitRate

    // Check circuit breaker
    details["circuit_state"] = p.cb.State()
    if p.cb.State() == CircuitOpen {
        issues = append(issues, "circuit breaker open")
    }

    // Determine status
    if len(issues) > 0 {
        return types.NewUnhealthyStatus(strings.Join(issues, "; "), details)
    }

    return types.NewHealthyStatus("operational", details)
}
```
