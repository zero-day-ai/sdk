package serve

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/memory"
)

// CallbackMemoryStore implements memory.Store by forwarding all operations
// to the orchestrator via gRPC callbacks.
type CallbackMemoryStore struct {
	client  *CallbackClient
	working *callbackWorkingMemory
}

// NewCallbackMemoryStore creates a new memory store that uses gRPC callbacks.
func NewCallbackMemoryStore(client *CallbackClient) *CallbackMemoryStore {
	store := &CallbackMemoryStore{
		client: client,
	}
	store.working = &callbackWorkingMemory{client: client}
	return store
}

// Working returns the working memory tier (ephemeral, in-memory).
func (m *CallbackMemoryStore) Working() memory.WorkingMemory {
	return m.working
}

// Mission returns the mission memory tier (persistent per-mission).
// Currently forwards to the same callback mechanism as working memory.
func (m *CallbackMemoryStore) Mission() memory.MissionMemory {
	return &callbackMissionMemory{client: m.client}
}

// LongTerm returns the long-term memory tier (vector-based).
// Currently returns a stub implementation that returns ErrNotImplemented.
func (m *CallbackMemoryStore) LongTerm() memory.LongTermMemory {
	return &callbackLongTermMemory{client: m.client}
}

// ============================================================================
// Working Memory Implementation
// ============================================================================

type callbackWorkingMemory struct {
	client *CallbackClient
}

// Get retrieves a value by key from the orchestrator's memory store.
func (m *callbackWorkingMemory) Get(ctx context.Context, key string) (any, error) {
	if !m.client.IsConnected() {
		return nil, fmt.Errorf("callback client not connected")
	}

	req := &proto.MemoryGetRequest{
		Key: key,
	}

	resp, err := m.client.MemoryGet(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("memory get failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("memory get error: %s", resp.Error.Message)
	}

	if !resp.Found {
		return nil, memory.ErrNotFound
	}

	// Deserialize the value from JSON
	var value any
	if err := json.Unmarshal([]byte(resp.ValueJson), &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return value, nil
}

// Set stores a value with the given key in the orchestrator's memory store.
func (m *callbackWorkingMemory) Set(ctx context.Context, key string, value any) error {
	if !m.client.IsConnected() {
		return fmt.Errorf("callback client not connected")
	}

	// Serialize the value to JSON
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	req := &proto.MemorySetRequest{
		Key:       key,
		ValueJson: string(valueJSON),
	}

	resp, err := m.client.MemorySet(ctx, req)
	if err != nil {
		return fmt.Errorf("memory set failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("memory set error: %s", resp.Error.Message)
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
// Mission Memory Implementation (Stub)
// ============================================================================

type callbackMissionMemory struct {
	client *CallbackClient
}

func (m *callbackMissionMemory) Get(ctx context.Context, key string) (*memory.Item, error) {
	return nil, memory.ErrNotImplemented
}

func (m *callbackMissionMemory) Set(ctx context.Context, key string, value any, metadata map[string]any) error {
	return memory.ErrNotImplemented
}

func (m *callbackMissionMemory) Delete(ctx context.Context, key string) error {
	return memory.ErrNotImplemented
}

func (m *callbackMissionMemory) Search(ctx context.Context, query string, limit int) ([]memory.Result, error) {
	return nil, memory.ErrNotImplemented
}

func (m *callbackMissionMemory) History(ctx context.Context, limit int) ([]memory.Item, error) {
	return nil, memory.ErrNotImplemented
}

// ============================================================================
// Long-Term Memory Implementation (Stub)
// ============================================================================

type callbackLongTermMemory struct {
	client *CallbackClient
}

func (m *callbackLongTermMemory) Store(ctx context.Context, content string, metadata map[string]any) (string, error) {
	return "", memory.ErrNotImplemented
}

func (m *callbackLongTermMemory) Search(ctx context.Context, query string, topK int, filters map[string]any) ([]memory.Result, error) {
	return nil, memory.ErrNotImplemented
}

func (m *callbackLongTermMemory) Delete(ctx context.Context, id string) error {
	return memory.ErrNotImplemented
}
