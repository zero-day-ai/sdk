package serve

import (
	"context"
	"fmt"
	"time"

	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/memory"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// CallbackMemoryStore implements memory.Store by forwarding all operations
// to the orchestrator via gRPC callbacks.
type CallbackMemoryStore struct {
	client  *CallbackClient
	tracer  trace.Tracer
	working *callbackWorkingMemory
}

// NewCallbackMemoryStore creates a new memory store that uses gRPC callbacks.
func NewCallbackMemoryStore(client *CallbackClient, tracer trace.Tracer) *CallbackMemoryStore {
	store := &CallbackMemoryStore{
		client: client,
		tracer: tracer,
	}
	store.working = &callbackWorkingMemory{client: client, tracer: tracer}
	return store
}

// Working returns the working memory tier (ephemeral, in-memory).
func (m *CallbackMemoryStore) Working() memory.WorkingMemory {
	return m.working
}

// Mission returns the mission memory tier (persistent per-mission).
func (m *CallbackMemoryStore) Mission() memory.MissionMemory {
	return &callbackMissionMemory{client: m.client, tracer: m.tracer}
}

// LongTerm returns the long-term memory tier (vector-based).
func (m *CallbackMemoryStore) LongTerm() memory.LongTermMemory {
	return &callbackLongTermMemory{client: m.client, tracer: m.tracer}
}

// ============================================================================
// Working Memory Implementation
// ============================================================================

type callbackWorkingMemory struct {
	client *CallbackClient
	tracer trace.Tracer
}

// Get retrieves a value by key from the orchestrator's memory store.
func (m *callbackWorkingMemory) Get(ctx context.Context, key string) (any, error) {
	// Start span for memory get
	ctx, span := m.tracer.Start(ctx, "gibson.memory.get",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.key", key),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	req := &proto.MemoryGetRequest{
		Key: key,
	}

	resp, err := m.client.MemoryGet(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("memory get failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("memory get error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	if !resp.Found {
		span.SetAttributes(attribute.Bool("gibson.memory.found", false))
		return nil, memory.ErrNotFound
	}

	span.SetAttributes(attribute.Bool("gibson.memory.found", true))

	// Convert the value from TypedValue
	value := FromTypedValue(resp.Value)

	return value, nil
}

// Set stores a value with the given key in the orchestrator's memory store.
func (m *callbackWorkingMemory) Set(ctx context.Context, key string, value any) error {
	// Start span for memory set
	ctx, span := m.tracer.Start(ctx, "gibson.memory.set",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.key", key),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// Convert value to TypedValue
	req := &proto.MemorySetRequest{
		Context: m.client.contextInfo(),
		Key:     key,
		Value:   ToTypedValue(value),
	}

	resp, err := m.client.MemorySet(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("memory set failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("memory set error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return err
	}

	return nil
}

// Delete removes a value by key from the orchestrator's memory store.
func (m *callbackWorkingMemory) Delete(ctx context.Context, key string) error {
	if !m.client.IsConnected() {
		return fmt.Errorf("callback client not connected")
	}

	req := &proto.MemoryDeleteRequest{
		Key: key,
	}

	resp, err := m.client.MemoryDelete(ctx, req)
	if err != nil {
		return fmt.Errorf("memory delete failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("memory delete error: %s", resp.Error.Message)
	}

	return nil
}

// Clear removes all values from working memory.
func (m *callbackWorkingMemory) Clear(ctx context.Context) error {
	// Use List with empty prefix to get all keys, then delete them
	req := &proto.MemoryListRequest{
		Prefix: "",
	}

	resp, err := m.client.MemoryList(ctx, req)
	if err != nil {
		return fmt.Errorf("memory list failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("memory list error: %s", resp.Error.Message)
	}

	// Delete each key
	for _, key := range resp.Keys {
		if err := m.Delete(ctx, key); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", key, err)
		}
	}

	return nil
}

// Keys returns all keys currently in working memory.
func (m *callbackWorkingMemory) Keys(ctx context.Context) ([]string, error) {
	if !m.client.IsConnected() {
		return nil, fmt.Errorf("callback client not connected")
	}

	req := &proto.MemoryListRequest{
		Prefix: "",
	}

	resp, err := m.client.MemoryList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("memory list failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("memory list error: %s", resp.Error.Message)
	}

	return resp.Keys, nil
}

// ============================================================================
// Mission Memory Implementation
// ============================================================================

type callbackMissionMemory struct {
	client *CallbackClient
	tracer trace.Tracer
}

func (m *callbackMissionMemory) Get(ctx context.Context, key string) (*memory.Item, error) {
	ctx, span := m.tracer.Start(ctx, "gibson.memory.mission.get",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.key", key),
			attribute.String("gibson.memory.tier", "mission"),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	req := &proto.MemoryGetRequest{
		Key:  key,
		Tier: proto.MemoryTier_MEMORY_TIER_MISSION,
	}

	resp, err := m.client.MemoryGet(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("mission memory get failed: %w", err)
	}

	if resp.Error != nil {
		if resp.Error.Code == proto.ErrorCode_ERROR_CODE_NOT_FOUND {
			span.SetAttributes(attribute.Bool("gibson.memory.found", false))
			return nil, memory.ErrNotFound
		}
		err := fmt.Errorf("mission memory get error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	if !resp.Found {
		span.SetAttributes(attribute.Bool("gibson.memory.found", false))
		return nil, memory.ErrNotFound
	}

	span.SetAttributes(attribute.Bool("gibson.memory.found", true))

	// Convert the value from TypedValue
	value := FromTypedValue(resp.Value)

	// Convert metadata from TypedMap
	metadata := FromTypedMap(resp.Metadata)

	// Parse timestamps
	var createdAt, updatedAt time.Time
	if resp.CreatedAt != "" {
		createdAt, _ = time.Parse(time.RFC3339, resp.CreatedAt)
	}
	if resp.UpdatedAt != "" {
		updatedAt, _ = time.Parse(time.RFC3339, resp.UpdatedAt)
	}

	return &memory.Item{
		Key:       key,
		Value:     value,
		Metadata:  metadata,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (m *callbackMissionMemory) Set(ctx context.Context, key string, value any, metadata map[string]any) error {
	ctx, span := m.tracer.Start(ctx, "gibson.memory.mission.set",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.key", key),
			attribute.String("gibson.memory.tier", "mission"),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	req := &proto.MemorySetRequest{
		Key:      key,
		Value:    ToTypedValue(value),
		Metadata: ToTypedMap(metadata),
		Tier:     proto.MemoryTier_MEMORY_TIER_MISSION,
	}

	resp, err := m.client.MemorySet(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("mission memory set failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("mission memory set error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return err
	}

	return nil
}

func (m *callbackMissionMemory) Delete(ctx context.Context, key string) error {
	ctx, span := m.tracer.Start(ctx, "gibson.memory.mission.delete",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.key", key),
			attribute.String("gibson.memory.tier", "mission"),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	req := &proto.MemoryDeleteRequest{
		Key:  key,
		Tier: proto.MemoryTier_MEMORY_TIER_MISSION,
	}

	resp, err := m.client.MemoryDelete(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("mission memory delete failed: %w", err)
	}

	if resp.Error != nil {
		if resp.Error.Code == proto.ErrorCode_ERROR_CODE_NOT_FOUND {
			return memory.ErrNotFound
		}
		err := fmt.Errorf("mission memory delete error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return err
	}

	return nil
}

func (m *callbackMissionMemory) Search(ctx context.Context, query string, limit int) ([]memory.Result, error) {
	ctx, span := m.tracer.Start(ctx, "gibson.memory.mission.search",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.query", query),
			attribute.Int("gibson.memory.limit", limit),
			attribute.String("gibson.memory.tier", "mission"),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	req := &proto.MissionMemorySearchRequest{
		Query: query,
		Limit: int32(limit),
	}

	resp, err := m.client.MissionMemorySearch(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("mission memory search failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("mission memory search error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	results := make([]memory.Result, 0, len(resp.Results))
	for _, r := range resp.Results {
		// Convert value from TypedValue
		value := FromTypedValue(r.Value)

		// Convert metadata from TypedMap
		metadata := FromTypedMap(r.Metadata)

		// Parse timestamps
		var createdAt, updatedAt time.Time
		if r.CreatedAt != "" {
			createdAt, _ = time.Parse(time.RFC3339, r.CreatedAt)
		}
		if r.UpdatedAt != "" {
			updatedAt, _ = time.Parse(time.RFC3339, r.UpdatedAt)
		}

		results = append(results, memory.Result{
			Item: memory.Item{
				Key:       r.Key,
				Value:     value,
				Metadata:  metadata,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			Score: r.Score,
		})
	}

	span.SetAttributes(attribute.Int("gibson.memory.results", len(results)))
	return results, nil
}

func (m *callbackMissionMemory) History(ctx context.Context, limit int) ([]memory.Item, error) {
	ctx, span := m.tracer.Start(ctx, "gibson.memory.mission.history",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int("gibson.memory.limit", limit),
			attribute.String("gibson.memory.tier", "mission"),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	req := &proto.MissionMemoryHistoryRequest{
		Limit: int32(limit),
	}

	resp, err := m.client.MissionMemoryHistory(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("mission memory history failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("mission memory history error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	items := make([]memory.Item, 0, len(resp.Items))
	for _, item := range resp.Items {
		// Convert value from TypedValue
		value := FromTypedValue(item.Value)

		// Convert metadata from TypedMap
		metadata := FromTypedMap(item.Metadata)

		// Parse timestamps
		var createdAt, updatedAt time.Time
		if item.CreatedAt != "" {
			createdAt, _ = time.Parse(time.RFC3339, item.CreatedAt)
		}
		if item.UpdatedAt != "" {
			updatedAt, _ = time.Parse(time.RFC3339, item.UpdatedAt)
		}

		items = append(items, memory.Item{
			Key:       item.Key,
			Value:     value,
			Metadata:  metadata,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	span.SetAttributes(attribute.Int("gibson.memory.items", len(items)))
	return items, nil
}

func (m *callbackMissionMemory) GetPreviousRunValue(ctx context.Context, key string) (any, error) {
	ctx, span := m.tracer.Start(ctx, "gibson.memory.mission.get_previous_run_value",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.key", key),
			attribute.String("gibson.memory.tier", "mission"),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	req := &proto.MissionMemoryGetPreviousRunValueRequest{
		Key: key,
	}

	resp, err := m.client.MissionMemoryGetPreviousRunValue(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("mission memory get previous run value failed: %w", err)
	}

	if resp.Error != nil {
		// Use custom string codes for mission memory continuity errors
		// These aren't standardized error codes in proto.ErrorCode
		if resp.Error.Message == "NO_PREVIOUS_RUN" {
			return nil, memory.ErrNoPreviousRun
		}
		if resp.Error.Message == "CONTINUITY_NOT_SUPPORTED" {
			return nil, memory.ErrContinuityNotSupported
		}
		if resp.Error.Code == proto.ErrorCode_ERROR_CODE_NOT_FOUND {
			span.SetAttributes(attribute.Bool("gibson.memory.found", false))
			return nil, memory.ErrNotFound
		}
		err := fmt.Errorf("mission memory get previous run value error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	if !resp.Found {
		span.SetAttributes(attribute.Bool("gibson.memory.found", false))
		return nil, memory.ErrNotFound
	}

	span.SetAttributes(attribute.Bool("gibson.memory.found", true))

	// Convert the value from TypedValue
	value := FromTypedValue(resp.Value)

	return value, nil
}

func (m *callbackMissionMemory) GetValueHistory(ctx context.Context, key string) ([]memory.HistoricalValue, error) {
	ctx, span := m.tracer.Start(ctx, "gibson.memory.mission.get_value_history",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.key", key),
			attribute.String("gibson.memory.tier", "mission"),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	req := &proto.MissionMemoryGetValueHistoryRequest{
		Key: key,
	}

	resp, err := m.client.MissionMemoryGetValueHistory(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("mission memory get value history failed: %w", err)
	}

	if resp.Error != nil {
		// Use custom string codes for mission memory continuity errors
		if resp.Error.Message == "CONTINUITY_NOT_SUPPORTED" {
			return nil, memory.ErrContinuityNotSupported
		}
		err := fmt.Errorf("mission memory get value history error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	values := make([]memory.HistoricalValue, 0, len(resp.Values))
	for _, v := range resp.Values {
		// Convert value from TypedValue
		value := FromTypedValue(v.Value)

		// Parse timestamp
		var storedAt time.Time
		if v.StoredAt != "" {
			storedAt, _ = time.Parse(time.RFC3339, v.StoredAt)
		}

		values = append(values, memory.HistoricalValue{
			Value:     value,
			RunNumber: int(v.RunNumber),
			MissionID: v.MissionId,
			StoredAt:  storedAt,
		})
	}

	span.SetAttributes(attribute.Int("gibson.memory.values", len(values)))
	return values, nil
}

func (m *callbackMissionMemory) ContinuityMode() memory.MemoryContinuityMode {
	// Note: ContinuityMode doesn't take a context, but we need one for the RPC call.
	// We use a background context with a reasonable timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if !m.client.IsConnected() {
		return memory.MemoryIsolated // Default if not connected
	}

	req := &proto.MissionMemoryContinuityModeRequest{}

	resp, err := m.client.MissionMemoryContinuityMode(ctx, req)
	if err != nil {
		return memory.MemoryIsolated // Default on error
	}

	if resp.Error != nil {
		return memory.MemoryIsolated // Default on error
	}

	// Convert string to MemoryContinuityMode
	switch resp.Mode {
	case "isolated":
		return memory.MemoryIsolated
	case "inherit":
		return memory.MemoryInherit
	case "shared":
		return memory.MemoryShared
	default:
		return memory.MemoryIsolated
	}
}

// ============================================================================
// Long-Term Memory Implementation
// ============================================================================

type callbackLongTermMemory struct {
	client *CallbackClient
	tracer trace.Tracer
}

func (m *callbackLongTermMemory) Store(ctx context.Context, content string, metadata map[string]any) (string, error) {
	ctx, span := m.tracer.Start(ctx, "gibson.memory.longterm.store",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.tier", "longterm"),
			attribute.Int("gibson.memory.content_length", len(content)),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	req := &proto.LongTermMemoryStoreRequest{
		Content:  content,
		Metadata: ToTypedMap(metadata),
	}

	resp, err := m.client.LongTermMemoryStore(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("longterm memory store failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("longterm memory store error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return "", err
	}

	span.SetAttributes(attribute.String("gibson.memory.id", resp.Id))
	return resp.Id, nil
}

func (m *callbackLongTermMemory) Search(ctx context.Context, query string, topK int, filters map[string]any) ([]memory.Result, error) {
	ctx, span := m.tracer.Start(ctx, "gibson.memory.longterm.search",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.query", query),
			attribute.Int("gibson.memory.top_k", topK),
			attribute.String("gibson.memory.tier", "longterm"),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	req := &proto.LongTermMemorySearchRequest{
		Query:   query,
		TopK:    int32(topK),
		Filters: ToTypedMap(filters),
	}

	resp, err := m.client.LongTermMemorySearch(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("longterm memory search failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("longterm memory search error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	results := make([]memory.Result, 0, len(resp.Results))
	for _, r := range resp.Results {
		// Convert metadata from TypedMap
		metadata := FromTypedMap(r.Metadata)

		// Parse timestamps
		var createdAt time.Time
		if r.CreatedAt != "" {
			createdAt, _ = time.Parse(time.RFC3339, r.CreatedAt)
		}

		results = append(results, memory.Result{
			Item: memory.Item{
				Key:       r.Id,
				Value:     r.Content,
				Metadata:  metadata,
				CreatedAt: createdAt,
				UpdatedAt: createdAt, // LongTerm items are immutable
			},
			Score: r.Score,
		})
	}

	span.SetAttributes(attribute.Int("gibson.memory.results", len(results)))
	return results, nil
}

func (m *callbackLongTermMemory) Delete(ctx context.Context, id string) error {
	ctx, span := m.tracer.Start(ctx, "gibson.memory.longterm.delete",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.memory.id", id),
			attribute.String("gibson.memory.tier", "longterm"),
		),
	)
	defer span.End()

	if !m.client.IsConnected() {
		err := fmt.Errorf("callback client not connected")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	req := &proto.LongTermMemoryDeleteRequest{
		Id: id,
	}

	resp, err := m.client.LongTermMemoryDelete(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("longterm memory delete failed: %w", err)
	}

	if resp.Error != nil {
		if resp.Error.Code == proto.ErrorCode_ERROR_CODE_NOT_FOUND {
			return memory.ErrNotFound
		}
		err := fmt.Errorf("longterm memory delete error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return err
	}

	return nil
}
