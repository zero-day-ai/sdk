# Evaluation Results Logger

## Overview

The evaluation results logger provides persistent storage for evaluation metrics in JSONL (JSON Lines) format. Each evaluation result is written as a single JSON line, making it easy to stream, analyze, and integrate with data processing pipelines.

## Components

### 1. Logger Interface

Defined in `eval.go` (lines 274-281):

```go
type Logger interface {
    Log(sample Sample, result Result) error
    Close() error
}
```

The interface allows for multiple logger implementations. Currently, JSONLLogger is the primary implementation.

### 2. LogEntry Structure

Defined in `logger.go`:

```go
type LogEntry struct {
    Timestamp    time.Time          `json:"timestamp"`
    SampleID     string             `json:"sample_id"`
    TaskID       string             `json:"task_id,omitempty"`
    Scores       map[string]float64 `json:"scores"`
    OverallScore float64            `json:"overall_score"`
    Duration     int64              `json:"duration_ms"`
    Details      map[string]any     `json:"details,omitempty"`
}
```

**Key Features:**
- Flattens `map[string]ScoreResult` to `map[string]float64` for simpler analysis
- Stores duration in milliseconds for easy aggregation
- Includes optional details map for scorer diagnostics, metadata, and errors
- Omits empty fields in JSON output

### 3. JSONLLogger Implementation

**Thread-Safe:** Uses `sync.Mutex` to protect concurrent writes.

**File Handling:**
- Opens file in append mode (`os.O_APPEND`)
- Creates file if it doesn't exist with 0644 permissions
- Flushes after each write (`file.Sync()`) to ensure data persistence
- Properly closes and releases file handle on `Close()`

**Methods:**

#### NewJSONLLogger(path string) (Logger, error)
Creates a new JSONL logger writing to the specified file path.

```go
logger, err := eval.NewJSONLLogger("evals.jsonl")
if err != nil {
    return err
}
defer logger.Close()
```

#### Log(sample Sample, result Result) error
Writes a sample and result as a single JSON line. Automatically:
- Extracts task ID from `sample.Task.ID`
- Flattens scores to simple map
- Includes scorer details with `{scorer_name}_details` keys
- Adds sample metadata and tags to details
- Records errors in details if present
- Flushes immediately after write

#### Close() error
Flushes buffered data and closes the file. Should be called when done logging (typically via `defer`).

## Usage Examples

### Basic Usage

```go
import "github.com/zero-day-ai/sdk/eval"

func TestMyAgent(t *testing.T) {
    eval.Run(t, "my_test", func(e *eval.E) {
        // Create logger
        logger, err := eval.NewJSONLLogger("evals.jsonl")
        if err != nil {
            t.Fatal(err)
        }
        defer logger.Close()

        // Configure E with logger
        e.WithLogger(logger)

        // Run evaluation - results automatically logged
        sample := eval.Sample{
            ID: "test-001",
            Task: agent.Task{ID: "task-001", Goal: "Test goal"},
        }
        result := e.Score(sample, scorer1, scorer2)

        // Result is automatically logged via e.Log()
    })
}
```

### Manual Logging

```go
logger, err := eval.NewJSONLLogger("manual.jsonl")
if err != nil {
    return err
}
defer logger.Close()

sample := eval.Sample{
    ID: "sample-001",
    Task: agent.Task{ID: "task-001"},
}

result := eval.Result{
    SampleID: "sample-001",
    Scores: map[string]eval.ScoreResult{
        "accuracy": {Score: 0.85, Details: map[string]any{"precision": 0.9}},
    },
    OverallScore: 0.85,
    Duration: 200 * time.Millisecond,
    Timestamp: time.Now(),
}

if err := logger.Log(sample, result); err != nil {
    return err
}
```

### Concurrent Usage

The logger is thread-safe and can be used from multiple goroutines:

```go
logger, _ := eval.NewJSONLLogger("concurrent.jsonl")
defer logger.Close()

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // Each goroutine can safely log
        logger.Log(sample, result)
    }(i)
}
wg.Wait()
```

## JSONL Format

Each line in the log file is a complete JSON object:

```json
{"timestamp":"2024-01-15T10:30:00Z","sample_id":"test-001","task_id":"task-001","scores":{"accuracy":0.92,"completeness":0.88},"overall_score":0.90,"duration_ms":250,"details":{"accuracy_details":{"precision":0.9,"recall":0.85},"sample_tags":["security","prompt-injection"]}}
{"timestamp":"2024-01-15T10:30:15Z","sample_id":"test-002","task_id":"task-002","scores":{"accuracy":0.78},"overall_score":0.78,"duration_ms":180,"details":{}}
```

## Analysis

JSONL format makes it easy to analyze results:

### Using jq
```bash
# Get average overall score
cat evals.jsonl | jq -s 'map(.overall_score) | add / length'

# Filter by sample ID
cat evals.jsonl | jq 'select(.sample_id == "test-001")'

# Extract all scorer scores
cat evals.jsonl | jq '.scores'
```

### Using Python
```python
import json

with open('evals.jsonl') as f:
    for line in f:
        entry = json.loads(line)
        print(f"{entry['sample_id']}: {entry['overall_score']}")
```

### Using Go
```go
file, _ := os.Open("evals.jsonl")
defer file.Close()

scanner := bufio.NewScanner(file)
for scanner.Scan() {
    var entry eval.LogEntry
    json.Unmarshal(scanner.Bytes(), &entry)
    fmt.Printf("%s: %.2f\n", entry.SampleID, entry.OverallScore)
}
```

## Testing

Comprehensive tests are provided in `logger_test.go`:

- File creation and append mode
- JSON serialization/deserialization
- Concurrent write safety
- Error handling
- Multiple entries
- Optional fields (task ID, details)

Run tests:
```bash
cd opensource/sdk
go test ./eval -run TestJSONLLogger -v
go test ./eval -run TestNewJSONLLogger -v
```

## Implementation Details

### Thread Safety

The logger uses a `sync.Mutex` to protect the file handle during concurrent writes. This ensures:
- No corrupted JSON lines
- Proper line boundaries
- Sequential writes from concurrent goroutines

### Data Persistence

Each write is followed by `file.Sync()` to ensure data is flushed to disk immediately. This prevents data loss if the process crashes but has a slight performance impact.

For high-throughput scenarios, consider:
- Batching writes
- Using a buffered writer with periodic flushes
- Implementing an async logger with a background goroutine

### Error Handling

All errors are wrapped with context:
```go
fmt.Errorf("failed to write log entry: %w", err)
```

This provides clear error messages with file paths when operations fail.

## Future Enhancements

Potential improvements:

1. **Buffered Logger**: Batch writes with periodic flushes for better performance
2. **Async Logger**: Background goroutine with channel-based queuing
3. **Compression**: Gzip compression for long-running evaluations
4. **Rotation**: Automatic file rotation based on size or time
5. **Remote Logging**: Export to external systems (S3, PostgreSQL, etc.)
6. **Structured Filtering**: Filter what gets logged based on criteria

## Related Files

- `eval.go` - Logger interface definition (lines 274-281)
- `types.go` - Sample and Result type definitions
- `scorer.go` - ScoreResult type definition
- `logger_test.go` - Comprehensive test suite
- `example_logger/main.go` - Standalone example program
