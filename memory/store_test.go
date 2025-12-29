package memory

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockStore is a mock implementation of Store for testing.
type mockStore struct {
	working  *mockWorkingMemory
	mission  *mockMissionMemory
	longTerm *mockLongTermMemory
}

func newMockStore() *mockStore {
	return &mockStore{
		working:  newMockWorkingMemory(),
		mission:  newMockMissionMemory(),
		longTerm: newMockLongTermMemory(),
	}
}

func (m *mockStore) Working() WorkingMemory {
	return m.working
}

func (m *mockStore) Mission() MissionMemory {
	return m.mission
}

func (m *mockStore) LongTerm() LongTermMemory {
	return m.longTerm
}

// mockWorkingMemory is a mock implementation of WorkingMemory.
type mockWorkingMemory struct {
	mu   sync.RWMutex
	data map[string]any
}

func newMockWorkingMemory() *mockWorkingMemory {
	return &mockWorkingMemory{
		data: make(map[string]any),
	}
}

func (m *mockWorkingMemory) Get(ctx context.Context, key string) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return val, nil
}

func (m *mockWorkingMemory) Set(ctx context.Context, key string, value any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if key == "" {
		return ErrInvalidKey
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	return nil
}

func (m *mockWorkingMemory) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.data[key]; !ok {
		return ErrNotFound
	}
	delete(m.data, key)
	return nil
}

func (m *mockWorkingMemory) Clear(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]any)
	return nil
}

func (m *mockWorkingMemory) Keys(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// mockMissionMemory is a mock implementation of MissionMemory.
type mockMissionMemory struct {
	mu    sync.RWMutex
	items map[string]*Item
}

func newMockMissionMemory() *mockMissionMemory {
	return &mockMissionMemory{
		items: make(map[string]*Item),
	}
}

func (m *mockMissionMemory) Get(ctx context.Context, key string) (*Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	item, ok := m.items[key]
	if !ok {
		return nil, ErrNotFound
	}
	return item.Clone(), nil
}

func (m *mockMissionMemory) Set(ctx context.Context, key string, value any, metadata map[string]any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if key == "" {
		return ErrInvalidKey
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if existing, ok := m.items[key]; ok {
		// Update existing
		existing.Value = value
		existing.Metadata = metadata
		existing.UpdatedAt = now
	} else {
		// Create new
		m.items[key] = &Item{
			Key:       key,
			Value:     value,
			Metadata:  metadata,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	return nil
}

func (m *mockMissionMemory) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.items[key]; !ok {
		return ErrNotFound
	}
	delete(m.items, key)
	return nil
}

func (m *mockMissionMemory) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []Result
	queryLower := strings.ToLower(query)

	for _, item := range m.items {
		score := 0.0

		// Simple scoring: check if query appears in key or value
		keyLower := strings.ToLower(item.Key)
		if strings.Contains(keyLower, queryLower) {
			score += 0.5
		}

		if valStr, ok := item.Value.(string); ok {
			valLower := strings.ToLower(valStr)
			if strings.Contains(valLower, queryLower) {
				score += 0.5
			}
		}

		if score > 0 {
			results = append(results, Result{
				Item:  *item.Clone(),
				Score: score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (m *mockMissionMemory) History(ctx context.Context, limit int) ([]Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]Item, 0, len(m.items))
	for _, item := range m.items {
		items = append(items, *item.Clone())
	}

	// Sort by UpdatedAt descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})

	// Apply limit
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}

// mockLongTermMemory is a mock implementation of LongTermMemory.
type mockLongTermMemory struct {
	mu      sync.RWMutex
	items   map[string]*Item
	counter int
}

func newMockLongTermMemory() *mockLongTermMemory {
	return &mockLongTermMemory{
		items: make(map[string]*Item),
	}
}

func (m *mockLongTermMemory) Store(ctx context.Context, content string, metadata map[string]any) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	id := fmt.Sprintf("ltm-%d", m.counter)

	now := time.Now()
	m.items[id] = &Item{
		Key:       id,
		Value:     content,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return id, nil
}

func (m *mockLongTermMemory) Search(ctx context.Context, query string, topK int, filters map[string]any) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []Result
	queryLower := strings.ToLower(query)

	for _, item := range m.items {
		// Apply filters
		if filters != nil {
			match := true
			for k, v := range filters {
				metaVal, ok := item.GetMetadata(k)
				if !ok || metaVal != v {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		// Simple semantic scoring based on word overlap
		score := 0.0
		if content, ok := item.Value.(string); ok {
			contentLower := strings.ToLower(content)
			queryWords := strings.Fields(queryLower)
			for _, word := range queryWords {
				if strings.Contains(contentLower, word) {
					score += 1.0 / float64(len(queryWords))
				}
			}
		}

		if score > 0 {
			results = append(results, Result{
				Item:  *item.Clone(),
				Score: score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply topK limit
	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

func (m *mockLongTermMemory) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.items[id]; !ok {
		return ErrNotFound
	}
	delete(m.items, id)
	return nil
}

// Tests start here

func TestMockStore(t *testing.T) {
	store := newMockStore()

	if store.Working() == nil {
		t.Error("Working() returned nil")
	}
	if store.Mission() == nil {
		t.Error("Mission() returned nil")
	}
	if store.LongTerm() == nil {
		t.Error("LongTerm() returned nil")
	}
}

func TestWorkingMemory(t *testing.T) {
	ctx := context.Background()
	working := newMockWorkingMemory()

	t.Run("Set and Get", func(t *testing.T) {
		err := working.Set(ctx, "key1", "value1")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		val, err := working.Get(ctx, "key1")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "value1" {
			t.Errorf("Get() = %v, want value1", val)
		}
	})

	t.Run("Get nonexistent", func(t *testing.T) {
		_, err := working.Get(ctx, "nonexistent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Get() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("Set with empty key", func(t *testing.T) {
		err := working.Set(ctx, "", "value")
		if !errors.Is(err, ErrInvalidKey) {
			t.Errorf("Set() error = %v, want ErrInvalidKey", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		working.Set(ctx, "key2", "value2")

		err := working.Delete(ctx, "key2")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err = working.Get(ctx, "key2")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Get() after Delete error = %v, want ErrNotFound", err)
		}
	})

	t.Run("Delete nonexistent", func(t *testing.T) {
		err := working.Delete(ctx, "nonexistent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Delete() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("Keys", func(t *testing.T) {
		working.Clear(ctx)
		working.Set(ctx, "key1", "val1")
		working.Set(ctx, "key2", "val2")
		working.Set(ctx, "key3", "val3")

		keys, err := working.Keys(ctx)
		if err != nil {
			t.Fatalf("Keys() error = %v", err)
		}

		want := []string{"key1", "key2", "key3"}
		if len(keys) != len(want) {
			t.Fatalf("Keys() length = %v, want %v", len(keys), len(want))
		}
		for i, k := range keys {
			if k != want[i] {
				t.Errorf("Keys()[%d] = %v, want %v", i, k, want[i])
			}
		}
	})

	t.Run("Clear", func(t *testing.T) {
		working.Set(ctx, "key1", "val1")
		working.Set(ctx, "key2", "val2")

		err := working.Clear(ctx)
		if err != nil {
			t.Fatalf("Clear() error = %v", err)
		}

		keys, err := working.Keys(ctx)
		if err != nil {
			t.Fatalf("Keys() error = %v", err)
		}
		if len(keys) != 0 {
			t.Errorf("Keys() after Clear length = %v, want 0", len(keys))
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := working.Set(ctx, "key", "value")
		if err == nil {
			t.Error("Set() with cancelled context should fail")
		}
	})
}

func TestMissionMemory(t *testing.T) {
	ctx := context.Background()
	mission := newMockMissionMemory()

	t.Run("Set and Get", func(t *testing.T) {
		metadata := map[string]any{"category": "test"}
		err := mission.Set(ctx, "key1", "value1", metadata)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		item, err := mission.Get(ctx, "key1")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if item.Key != "key1" {
			t.Errorf("item.Key = %v, want key1", item.Key)
		}
		if item.Value != "value1" {
			t.Errorf("item.Value = %v, want value1", item.Value)
		}
		if cat, ok := item.GetMetadata("category"); !ok || cat != "test" {
			t.Errorf("item.Metadata[category] = %v, want test", cat)
		}
	})

	t.Run("Update existing", func(t *testing.T) {
		mission.Set(ctx, "key2", "original", nil)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamp

		mission.Set(ctx, "key2", "updated", map[string]any{"new": "meta"})

		item, err := mission.Get(ctx, "key2")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if item.Value != "updated" {
			t.Errorf("item.Value = %v, want updated", item.Value)
		}
		if !item.IsModified() {
			t.Error("item should be marked as modified")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		mission.Set(ctx, "key3", "value3", nil)

		err := mission.Delete(ctx, "key3")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err = mission.Get(ctx, "key3")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Get() after Delete error = %v, want ErrNotFound", err)
		}
	})

	t.Run("Search", func(t *testing.T) {
		mission.Set(ctx, "pref_theme", "dark mode", map[string]any{"type": "ui"})
		mission.Set(ctx, "pref_lang", "english", map[string]any{"type": "i18n"})
		mission.Set(ctx, "data", "some other data", nil)

		results, err := mission.Search(ctx, "dark", 10)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if len(results) == 0 {
			t.Fatal("Search() returned no results")
		}

		found := false
		for _, r := range results {
			if r.Key == "pref_theme" {
				found = true
				if r.Score <= 0 {
					t.Errorf("result.Score = %v, want > 0", r.Score)
				}
			}
		}
		if !found {
			t.Error("Search() did not find expected result")
		}
	})

	t.Run("Search with limit", func(t *testing.T) {
		mission.Set(ctx, "match1", "test data", nil)
		mission.Set(ctx, "match2", "test data", nil)
		mission.Set(ctx, "match3", "test data", nil)

		results, err := mission.Search(ctx, "test", 2)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if len(results) > 2 {
			t.Errorf("Search() returned %d results, want <= 2", len(results))
		}
	})

	t.Run("History", func(t *testing.T) {
		mission.items = make(map[string]*Item) // Clear

		// Add items with different timestamps
		now := time.Now()
		mission.items["old"] = &Item{
			Key:       "old",
			Value:     "data",
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		}
		mission.items["middle"] = &Item{
			Key:       "middle",
			Value:     "data",
			CreatedAt: now.Add(-1 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour),
		}
		mission.items["recent"] = &Item{
			Key:       "recent",
			Value:     "data",
			CreatedAt: now,
			UpdatedAt: now,
		}

		history, err := mission.History(ctx, 10)
		if err != nil {
			t.Fatalf("History() error = %v", err)
		}

		if len(history) != 3 {
			t.Fatalf("History() length = %v, want 3", len(history))
		}

		// Should be ordered by most recent first
		if history[0].Key != "recent" {
			t.Errorf("History()[0].Key = %v, want recent", history[0].Key)
		}
		if history[2].Key != "old" {
			t.Errorf("History()[2].Key = %v, want old", history[2].Key)
		}
	})

	t.Run("History with limit", func(t *testing.T) {
		history, err := mission.History(ctx, 2)
		if err != nil {
			t.Fatalf("History() error = %v", err)
		}

		if len(history) > 2 {
			t.Errorf("History() length = %v, want <= 2", len(history))
		}
	})
}

func TestLongTermMemory(t *testing.T) {
	ctx := context.Background()
	longTerm := newMockLongTermMemory()

	t.Run("Store and Search", func(t *testing.T) {
		metadata := map[string]any{
			"category": "programming",
			"language": "go",
		}

		id, err := longTerm.Store(ctx, "Go is a statically typed language", metadata)
		if err != nil {
			t.Fatalf("Store() error = %v", err)
		}
		if id == "" {
			t.Error("Store() returned empty id")
		}

		results, err := longTerm.Search(ctx, "statically typed", 10, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if len(results) == 0 {
			t.Fatal("Search() returned no results")
		}

		found := false
		for _, r := range results {
			if r.Key == id {
				found = true
				if r.Score <= 0 {
					t.Errorf("result.Score = %v, want > 0", r.Score)
				}
			}
		}
		if !found {
			t.Error("Search() did not find stored item")
		}
	})

	t.Run("Search with filters", func(t *testing.T) {
		longTerm.Store(ctx, "Python is dynamically typed", map[string]any{
			"category": "programming",
			"language": "python",
		})
		longTerm.Store(ctx, "JavaScript is also dynamically typed", map[string]any{
			"category": "programming",
			"language": "javascript",
		})

		filters := map[string]any{"language": "python"}
		results, err := longTerm.Search(ctx, "dynamically typed", 10, filters)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		for _, r := range results {
			lang, ok := r.GetMetadata("language")
			if !ok || lang != "python" {
				t.Errorf("Search() with filters returned item with language = %v, want python", lang)
			}
		}
	})

	t.Run("Search with topK limit", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			longTerm.Store(ctx, "Common content for testing", nil)
		}

		results, err := longTerm.Search(ctx, "common content", 3, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if len(results) > 3 {
			t.Errorf("Search() returned %d results, want <= 3", len(results))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		id, _ := longTerm.Store(ctx, "Content to delete", nil)

		err := longTerm.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		results, err := longTerm.Search(ctx, "Content to delete", 10, nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		for _, r := range results {
			if r.Key == id {
				t.Error("Search() found deleted item")
			}
		}
	})

	t.Run("Delete nonexistent", func(t *testing.T) {
		err := longTerm.Delete(ctx, "nonexistent-id")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Delete() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("Concurrent operations", func(t *testing.T) {
		var wg sync.WaitGroup

		// Store concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				content := fmt.Sprintf("Concurrent content %d", n)
				longTerm.Store(ctx, content, map[string]any{"index": n})
			}(i)
		}

		wg.Wait()

		// Search concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				longTerm.Search(ctx, "concurrent", 10, nil)
			}()
		}

		wg.Wait()
	})
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrNotFound",
			err:  ErrNotFound,
			want: "memory: item not found",
		},
		{
			name: "ErrInvalidKey",
			err:  ErrInvalidKey,
			want: "memory: invalid key",
		},
		{
			name: "ErrInvalidValue",
			err:  ErrInvalidValue,
			want: "memory: invalid value",
		},
		{
			name: "ErrStorageFailed",
			err:  ErrStorageFailed,
			want: "memory: storage operation failed",
		},
		{
			name: "ErrNotImplemented",
			err:  ErrNotImplemented,
			want: "memory: operation not implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("error.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
