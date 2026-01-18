# Plugins Quickstart Guide

Extend the Gibson framework with custom data providers and integrations.

## What is a Plugin?

A plugin is an extension that provides data access and custom functionality:
- Exposes named methods with validated parameters
- Provides access to external data sources (databases, APIs, knowledge graphs)
- Can maintain state and connections
- Supports initialization and graceful shutdown

**Plugins vs Tools:**
- **Tools** perform actions (make requests, scan ports, run commands)
- **Plugins** provide data access (query databases, fetch configurations, access knowledge)

## Minimal Plugin (5 Minutes)

```go
package main

import (
    "context"
    "log"

    "github.com/zero-day-ai/sdk/plugin"
    "github.com/zero-day-ai/sdk/schema"
)

func main() {
    cfg := plugin.NewConfig().
        SetName("greeter").
        SetVersion("1.0.0").
        SetDescription("A simple greeting plugin")

    // Add a method
    cfg.AddMethodWithDesc(
        "greet",
        "Returns a personalized greeting",
        func(ctx context.Context, params map[string]any) (any, error) {
            name := params["name"].(string)
            return map[string]any{
                "message": "Hello, " + name + "!",
            }, nil
        },
        schema.Object(map[string]schema.JSON{
            "name": schema.StringWithDesc("Name to greet"),
        }, "name"),
        schema.Object(map[string]schema.JSON{
            "message": schema.String(),
        }, "message"),
    )

    p, err := plugin.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Initialize
    ctx := context.Background()
    if err := p.Initialize(ctx, nil); err != nil {
        log.Fatal(err)
    }

    // Query the method
    result, err := p.Query(ctx, "greet", map[string]any{"name": "World"})
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Result: %v", result) // {message: Hello, World!}
}
```

## Plugin Configuration

### Required Fields

| Field | Description |
|-------|-------------|
| `Name` | Unique identifier (kebab-case) |
| `Version` | Semantic version |
| `Description` | What the plugin provides |
| `Methods` | At least one method |

### Optional Fields

| Field | Description |
|-------|-------------|
| `InitFunc` | Initialization logic (connections, setup) |
| `ShutdownFunc` | Cleanup logic |
| `HealthFunc` | Health check for monitoring |

## Adding Methods

Methods are the core of plugins. Each method has:
- A unique name
- Input parameters with schema validation
- Output schema
- Handler function

### Basic Method

```go
cfg.AddMethodWithDesc(
    "method_name",
    "Description of what this method does",
    handlerFunc,
    inputSchema,
    outputSchema,
)
```

### Multiple Methods

```go
cfg := plugin.NewConfig().
    SetName("user-service").
    SetVersion("1.0.0").
    SetDescription("User data access plugin")

// Get user by ID
cfg.AddMethodWithDesc(
    "get_user",
    "Retrieve user by ID",
    func(ctx context.Context, params map[string]any) (any, error) {
        userID := params["user_id"].(string)
        // Fetch from database
        return map[string]any{
            "id":    userID,
            "name":  "John Doe",
            "email": "john@example.com",
        }, nil
    },
    schema.Object(map[string]schema.JSON{
        "user_id": schema.String(),
    }, "user_id"),
    schema.Object(map[string]schema.JSON{
        "id":    schema.String(),
        "name":  schema.String(),
        "email": schema.String(),
    }, "id", "name"),
)

// Search users
cfg.AddMethodWithDesc(
    "search_users",
    "Search users by query",
    func(ctx context.Context, params map[string]any) (any, error) {
        query := params["query"].(string)
        limit := 10
        if l, ok := params["limit"].(float64); ok {
            limit = int(l)
        }
        // Search logic
        return map[string]any{
            "users": []map[string]any{},
            "total": 0,
        }, nil
    },
    schema.Object(map[string]schema.JSON{
        "query": schema.String(),
        "limit": schema.Int(),
    }, "query"),
    schema.Object(map[string]schema.JSON{
        "users": schema.Array(schema.Object(map[string]schema.JSON{})),
        "total": schema.Int(),
    }, "users"),
)

// Delete user
cfg.AddMethodWithDesc(
    "delete_user",
    "Delete user by ID",
    func(ctx context.Context, params map[string]any) (any, error) {
        userID := params["user_id"].(string)
        // Delete logic
        return map[string]any{"deleted": true}, nil
    },
    schema.Object(map[string]schema.JSON{
        "user_id": schema.String(),
    }, "user_id"),
    schema.Object(map[string]schema.JSON{
        "deleted": schema.Bool(),
    }, "deleted"),
)
```

## Stateful Plugins

Plugins can maintain state (connections, caches, etc.):

```go
package main

import (
    "context"
    "database/sql"
    "sync"

    "github.com/zero-day-ai/sdk/plugin"
    "github.com/zero-day-ai/sdk/schema"
    "github.com/zero-day-ai/sdk/types"
    _ "github.com/lib/pq"
)

type DBPlugin struct {
    db   *sql.DB
    mu   sync.RWMutex
    cfg  map[string]any
}

func NewDBPlugin() *DBPlugin {
    return &DBPlugin{}
}

func (p *DBPlugin) Build() (*plugin.Plugin, error) {
    cfg := plugin.NewConfig().
        SetName("postgres").
        SetVersion("1.0.0").
        SetDescription("PostgreSQL database plugin").
        SetInitFunc(p.init).
        SetShutdownFunc(p.shutdown).
        SetHealthFunc(p.health)

    // Query method
    cfg.AddMethodWithDesc(
        "query",
        "Execute a SQL query",
        p.query,
        schema.Object(map[string]schema.JSON{
            "sql":    schema.StringWithDesc("SQL query to execute"),
            "params": schema.Array(schema.String()),
        }, "sql"),
        schema.Object(map[string]schema.JSON{
            "rows":          schema.Array(schema.Object(map[string]schema.JSON{})),
            "rows_affected": schema.Int(),
        }),
    )

    // Execute method (for INSERT/UPDATE/DELETE)
    cfg.AddMethodWithDesc(
        "execute",
        "Execute a SQL statement",
        p.execute,
        schema.Object(map[string]schema.JSON{
            "sql":    schema.String(),
            "params": schema.Array(schema.String()),
        }, "sql"),
        schema.Object(map[string]schema.JSON{
            "rows_affected": schema.Int(),
            "last_insert_id": schema.Int(),
        }),
    )

    return plugin.New(cfg)
}

func (p *DBPlugin) init(ctx context.Context, config map[string]any) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    connStr := config["connection_string"].(string)
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return err
    }

    if err := db.PingContext(ctx); err != nil {
        return err
    }

    p.db = db
    p.cfg = config
    return nil
}

func (p *DBPlugin) shutdown(ctx context.Context) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    if p.db != nil {
        return p.db.Close()
    }
    return nil
}

func (p *DBPlugin) health(ctx context.Context) types.HealthStatus {
    p.mu.RLock()
    defer p.mu.RUnlock()

    if p.db == nil {
        return types.NewUnhealthyStatus("not initialized", nil)
    }

    if err := p.db.PingContext(ctx); err != nil {
        return types.NewUnhealthyStatus("database unreachable", map[string]any{
            "error": err.Error(),
        })
    }

    return types.NewHealthyStatus("database connected")
}

func (p *DBPlugin) query(ctx context.Context, params map[string]any) (any, error) {
    p.mu.RLock()
    defer p.mu.RUnlock()

    sqlQuery := params["sql"].(string)

    // Convert params
    var args []any
    if sqlParams, ok := params["params"].([]any); ok {
        args = sqlParams
    }

    rows, err := p.db.QueryContext(ctx, sqlQuery, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Scan rows into maps
    columns, _ := rows.Columns()
    var results []map[string]any

    for rows.Next() {
        values := make([]any, len(columns))
        valuePtrs := make([]any, len(columns))
        for i := range values {
            valuePtrs[i] = &values[i]
        }

        rows.Scan(valuePtrs...)

        row := make(map[string]any)
        for i, col := range columns {
            row[col] = values[i]
        }
        results = append(results, row)
    }

    return map[string]any{
        "rows":          results,
        "rows_affected": len(results),
    }, nil
}

func (p *DBPlugin) execute(ctx context.Context, params map[string]any) (any, error) {
    p.mu.RLock()
    defer p.mu.RUnlock()

    sqlQuery := params["sql"].(string)

    var args []any
    if sqlParams, ok := params["params"].([]any); ok {
        args = sqlParams
    }

    result, err := p.db.ExecContext(ctx, sqlQuery, args...)
    if err != nil {
        return nil, err
    }

    rowsAffected, _ := result.RowsAffected()
    lastID, _ := result.LastInsertId()

    return map[string]any{
        "rows_affected":  rowsAffected,
        "last_insert_id": lastID,
    }, nil
}
```

## Using Plugins from Agents

Agents access plugins via the harness:

```go
func execute(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    // Query a plugin method
    result, err := h.QueryPlugin(ctx, "postgres", "query", map[string]any{
        "sql": "SELECT * FROM vulnerabilities WHERE severity = $1",
        "params": []string{"critical"},
    })
    if err != nil {
        return agent.NewFailedResult(err), err
    }

    rows := result.(map[string]any)["rows"].([]any)
    // Process results...

    // List available plugins
    plugins, err := h.ListPlugins(ctx)
    for _, p := range plugins {
        logger.Info("available plugin", "name", p.Name, "version", p.Version)
    }

    return agent.NewSuccessResult("done"), nil
}
```

## GraphRAG Plugin Example

A plugin for knowledge graph access:

```go
package main

import (
    "context"

    "github.com/zero-day-ai/sdk/plugin"
    "github.com/zero-day-ai/sdk/schema"
)

func main() {
    cfg := plugin.NewConfig().
        SetName("knowledge-graph").
        SetVersion("1.0.0").
        SetDescription("Knowledge graph query plugin")

    // Semantic search
    cfg.AddMethodWithDesc(
        "search",
        "Semantic search over the knowledge graph",
        func(ctx context.Context, params map[string]any) (any, error) {
            query := params["query"].(string)
            topK := 10
            if k, ok := params["top_k"].(float64); ok {
                topK = int(k)
            }

            // Vector search implementation
            results := performSemanticSearch(query, topK)

            return map[string]any{
                "results": results,
                "count":   len(results),
            }, nil
        },
        schema.Object(map[string]schema.JSON{
            "query":      schema.StringWithDesc("Search query"),
            "top_k":      schema.Int(),
            "node_types": schema.Array(schema.String()),
            "min_score":  schema.Number(),
        }, "query"),
        schema.Object(map[string]schema.JSON{
            "results": schema.Array(schema.Object(map[string]schema.JSON{})),
            "count":   schema.Int(),
        }),
    )

    // Find related nodes
    cfg.AddMethodWithDesc(
        "find_related",
        "Find nodes related to a given node",
        func(ctx context.Context, params map[string]any) (any, error) {
            nodeID := params["node_id"].(string)
            relType := params["relationship_type"].(string)

            // Graph traversal
            related := findRelatedNodes(nodeID, relType)

            return map[string]any{
                "nodes": related,
            }, nil
        },
        schema.Object(map[string]schema.JSON{
            "node_id":           schema.String(),
            "relationship_type": schema.String(),
            "max_hops":          schema.Int(),
        }, "node_id", "relationship_type"),
        schema.Object(map[string]schema.JSON{
            "nodes": schema.Array(schema.Object(map[string]schema.JSON{})),
        }),
    )

    // Store node
    cfg.AddMethodWithDesc(
        "store_node",
        "Store a new node in the knowledge graph",
        func(ctx context.Context, params map[string]any) (any, error) {
            nodeType := params["type"].(string)
            content := params["content"].(string)
            properties := params["properties"].(map[string]any)

            id := storeNode(nodeType, content, properties)

            return map[string]any{
                "id":      id,
                "success": true,
            }, nil
        },
        schema.Object(map[string]schema.JSON{
            "type":       schema.String(),
            "content":    schema.String(),
            "properties": schema.Object(map[string]schema.JSON{}),
        }, "type", "content"),
        schema.Object(map[string]schema.JSON{
            "id":      schema.String(),
            "success": schema.Bool(),
        }),
    )

    p, _ := plugin.New(cfg)
    // Serve or register
}
```

## Serving Plugins

### gRPC Mode

```go
import "github.com/zero-day-ai/sdk/serve"

// Local mode
err := serve.Plugin(myPlugin,
    serve.WithPort(50053),
    serve.WithLocalMode("~/.gibson/run/plugins/my-plugin.sock"),
)

// Remote mode with TLS
err := serve.Plugin(myPlugin,
    serve.WithPort(50053),
    serve.WithTLS("cert.pem", "key.pem"),
)
```

### Configuration

Remote plugins must be configured in `~/.gibson/config.yaml`:

```yaml
remote_plugins:
  knowledge-graph:
    address: "graphdb.internal:50053"
    protocol: grpc
    health_check:
      type: grpc
      interval: 30s
      timeout: 5s
```

## Lifecycle Management

```go
cfg := plugin.NewConfig().
    SetName("my-plugin").
    SetVersion("1.0.0").
    SetDescription("Plugin with lifecycle").
    SetInitFunc(func(ctx context.Context, config map[string]any) error {
        // Called once when plugin starts
        // - Open connections
        // - Load configurations
        // - Initialize caches
        return nil
    }).
    SetShutdownFunc(func(ctx context.Context) error {
        // Called when plugin is stopping
        // - Close connections
        // - Flush caches
        // - Clean up resources
        return nil
    }).
    SetHealthFunc(func(ctx context.Context) types.HealthStatus {
        // Called periodically for monitoring
        // Return healthy or unhealthy status
        return types.NewHealthyStatus("operational")
    })
```

## Error Handling

Return errors from method handlers:

```go
cfg.AddMethodWithDesc(
    "get_secret",
    "Retrieve a secret value",
    func(ctx context.Context, params map[string]any) (any, error) {
        key := params["key"].(string)

        // Check authorization
        if !isAuthorized(ctx, key) {
            return nil, fmt.Errorf("unauthorized access to secret: %s", key)
        }

        value, err := vault.Get(key)
        if err != nil {
            return nil, fmt.Errorf("failed to retrieve secret: %w", err)
        }

        return map[string]any{"value": value}, nil
    },
    // schemas...
)
```

## Best Practices

1. **Keep methods focused** - One method, one purpose
2. **Use descriptive method names** - `get_user`, `search_findings`, `store_node`
3. **Validate parameters** - Even with schema validation, validate business logic
4. **Handle timeouts** - Respect context cancellation
5. **Implement health checks** - Return meaningful health status
6. **Use connection pooling** - For database plugins, pool connections
7. **Document thoroughly** - Use descriptions on methods and parameters
8. **Make thread-safe** - Plugins are called concurrently; use mutexes where needed
9. **Graceful shutdown** - Clean up resources in shutdown handler

## Testing Plugins

```go
func TestMyPlugin(t *testing.T) {
    p, err := NewMyPlugin().Build()
    require.NoError(t, err)

    ctx := context.Background()

    // Initialize with test config
    err = p.Initialize(ctx, map[string]any{
        "connection_string": "postgres://test:test@localhost/test",
    })
    require.NoError(t, err)
    defer p.Shutdown(ctx)

    // Test a method
    result, err := p.Query(ctx, "search", map[string]any{
        "query": "test query",
        "top_k": 5,
    })
    require.NoError(t, err)
    assert.NotNil(t, result)

    // Test health
    health := p.Health(ctx)
    assert.True(t, health.Healthy)
}
```

## Next Steps

- See `examples/` for more plugin examples
- Read the [Agents Guide](AGENTS.md) to understand how agents use plugins
- Read the [Tools Guide](TOOLS.md) for comparison with tools
- Check the [main README](../README.md) for deployment options
