package memory

import (
	"encoding/json"
	"testing"
	"time"
)

func TestItem_GetMetadata(t *testing.T) {
	tests := []struct {
		name     string
		item     *Item
		key      string
		wantVal  any
		wantBool bool
	}{
		{
			name: "existing key",
			item: &Item{
				Metadata: map[string]any{
					"category": "test",
					"priority": 1,
				},
			},
			key:      "category",
			wantVal:  "test",
			wantBool: true,
		},
		{
			name: "missing key",
			item: &Item{
				Metadata: map[string]any{
					"category": "test",
				},
			},
			key:      "priority",
			wantVal:  nil,
			wantBool: false,
		},
		{
			name:     "nil metadata",
			item:     &Item{},
			key:      "any",
			wantVal:  nil,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotBool := tt.item.GetMetadata(tt.key)
			if gotBool != tt.wantBool {
				t.Errorf("GetMetadata() bool = %v, want %v", gotBool, tt.wantBool)
			}
			if gotVal != tt.wantVal {
				t.Errorf("GetMetadata() val = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestItem_SetMetadata(t *testing.T) {
	tests := []struct {
		name     string
		item     *Item
		key      string
		value    any
		wantMeta map[string]any
	}{
		{
			name:  "set on existing metadata",
			item:  &Item{Metadata: map[string]any{"existing": "value"}},
			key:   "new",
			value: "data",
			wantMeta: map[string]any{
				"existing": "value",
				"new":      "data",
			},
		},
		{
			name:  "set on nil metadata",
			item:  &Item{},
			key:   "first",
			value: 123,
			wantMeta: map[string]any{
				"first": 123,
			},
		},
		{
			name:  "overwrite existing key",
			item:  &Item{Metadata: map[string]any{"key": "old"}},
			key:   "key",
			value: "new",
			wantMeta: map[string]any{
				"key": "new",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.item.SetMetadata(tt.key, tt.value)
			if len(tt.item.Metadata) != len(tt.wantMeta) {
				t.Errorf("metadata length = %v, want %v", len(tt.item.Metadata), len(tt.wantMeta))
			}
			for k, wantV := range tt.wantMeta {
				gotV, ok := tt.item.Metadata[k]
				if !ok {
					t.Errorf("missing metadata key %q", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("metadata[%q] = %v, want %v", k, gotV, wantV)
				}
			}
		})
	}
}

func TestItem_HasMetadata(t *testing.T) {
	tests := []struct {
		name string
		item *Item
		key  string
		want bool
	}{
		{
			name: "has key",
			item: &Item{Metadata: map[string]any{"key": "value"}},
			key:  "key",
			want: true,
		},
		{
			name: "missing key",
			item: &Item{Metadata: map[string]any{"key": "value"}},
			key:  "other",
			want: false,
		},
		{
			name: "nil metadata",
			item: &Item{},
			key:  "any",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.HasMetadata(tt.key); got != tt.want {
				t.Errorf("HasMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_Clone(t *testing.T) {
	now := time.Now()
	original := &Item{
		Key:   "test",
		Value: map[string]any{"nested": "data"},
		Metadata: map[string]any{
			"category": "test",
			"tags":     []string{"a", "b"},
		},
		CreatedAt: now,
		UpdatedAt: now.Add(time.Hour),
	}

	clone := original.Clone()

	// Verify clone has same values
	if clone.Key != original.Key {
		t.Errorf("clone.Key = %v, want %v", clone.Key, original.Key)
	}
	if !clone.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("clone.CreatedAt = %v, want %v", clone.CreatedAt, original.CreatedAt)
	}
	if !clone.UpdatedAt.Equal(original.UpdatedAt) {
		t.Errorf("clone.UpdatedAt = %v, want %v", clone.UpdatedAt, original.UpdatedAt)
	}

	// Verify deep copy - modify clone shouldn't affect original
	cloneMap := clone.Value.(map[string]any)
	cloneMap["nested"] = "modified"

	originalMap := original.Value.(map[string]any)
	if originalMap["nested"] == "modified" {
		t.Error("modifying clone affected original value")
	}

	// Verify metadata is deep copied
	cloneMeta := clone.Metadata["tags"].([]any)
	cloneMeta[0] = "modified"

	originalTags := original.Metadata["tags"].([]string)
	if originalTags[0] == "modified" {
		t.Error("modifying clone affected original metadata")
	}
}

func TestItem_Age(t *testing.T) {
	now := time.Now()
	item := &Item{
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now,
	}

	age := item.Age()
	if age < 55*time.Minute || age > 65*time.Minute {
		t.Errorf("Age() = %v, want ~1 hour", age)
	}
}

func TestItem_TimeSinceUpdate(t *testing.T) {
	now := time.Now()
	item := &Item{
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-10 * time.Minute),
	}

	timeSince := item.TimeSinceUpdate()
	if timeSince < 9*time.Minute || timeSince > 11*time.Minute {
		t.Errorf("TimeSinceUpdate() = %v, want ~10 minutes", timeSince)
	}
}

func TestItem_IsModified(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		createdAt time.Time
		updatedAt time.Time
		want      bool
	}{
		{
			name:      "not modified",
			createdAt: now,
			updatedAt: now,
			want:      false,
		},
		{
			name:      "modified",
			createdAt: now,
			updatedAt: now.Add(time.Hour),
			want:      true,
		},
		{
			name:      "modified before created (edge case)",
			createdAt: now,
			updatedAt: now.Add(-time.Hour),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &Item{
				CreatedAt: tt.createdAt,
				UpdatedAt: tt.updatedAt,
			}
			if got := item.IsModified(); got != tt.want {
				t.Errorf("IsModified() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_String(t *testing.T) {
	item := &Item{
		Key:   "test",
		Value: "data",
		Metadata: map[string]any{
			"category": "test",
		},
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	str := item.String()

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(str), &parsed); err != nil {
		t.Fatalf("String() returned invalid JSON: %v", err)
	}

	// Verify key fields are present
	if parsed["key"] != "test" {
		t.Errorf("JSON missing or incorrect key field")
	}
	if parsed["value"] != "data" {
		t.Errorf("JSON missing or incorrect value field")
	}
}

func TestResult_String(t *testing.T) {
	result := &Result{
		Item: Item{
			Key:       "test",
			Value:     "data",
			CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		Score: 0.95,
	}

	str := result.String()

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(str), &parsed); err != nil {
		t.Fatalf("String() returned invalid JSON: %v", err)
	}

	// Verify score field is present
	score, ok := parsed["score"].(float64)
	if !ok {
		t.Fatal("JSON missing or incorrect score field")
	}
	if score != 0.95 {
		t.Errorf("score = %v, want 0.95", score)
	}
}

func TestCloneValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{
			name:  "nil",
			value: nil,
		},
		{
			name:  "string",
			value: "test",
		},
		{
			name:  "number",
			value: 42,
		},
		{
			name:  "slice",
			value: []string{"a", "b", "c"},
		},
		{
			name:  "map",
			value: map[string]any{"key": "value", "num": 123},
		},
		{
			name: "nested",
			value: map[string]any{
				"level1": map[string]any{
					"level2": []any{1, 2, 3},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := cloneValue(tt.value)

			// Verify values are equal
			origJSON, _ := json.Marshal(tt.value)
			cloneJSON, _ := json.Marshal(clone)

			if string(origJSON) != string(cloneJSON) {
				t.Errorf("cloneValue() produced different value:\norig:  %s\nclone: %s",
					origJSON, cloneJSON)
			}
		})
	}
}
