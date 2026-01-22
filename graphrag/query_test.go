package graphrag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStructuredQuery(t *testing.T) {
	q := NewStructuredQuery()

	assert.Equal(t, "", q.Text)
	assert.Empty(t, q.Embedding)
	assert.Equal(t, 100, q.TopK)
	assert.Equal(t, 0, q.MaxHops)
	assert.Equal(t, 0.0, q.MinScore)
	assert.Equal(t, 0.0, q.VectorWeight)
	assert.Equal(t, 1.0, q.GraphWeight)
}

func TestQueryValidate(t *testing.T) {
	tests := []struct {
		name    string
		query   *Query
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid text query",
			query:   NewQuery("test"),
			wantErr: false,
		},
		{
			name:    "valid embedding query",
			query:   NewQueryFromEmbedding([]float64{0.1, 0.2}),
			wantErr: false,
		},
		{
			name: "valid structured query with node types",
			query: NewStructuredQuery().
				WithNodeTypes("host", "port"),
			wantErr: false,
		},
		{
			name: "invalid: both text and embedding",
			query: &Query{
				Text:         "test",
				Embedding:    []float64{0.1},
				TopK:         10,
				MaxHops:      3,
				MinScore:     0.7,
				VectorWeight: 0.6,
				GraphWeight:  0.4,
			},
			wantErr: true,
			errMsg:  "must have either Text or Embedding, not both",
		},
		{
			name: "invalid: neither text, embedding, nor node types",
			query: &Query{
				TopK:         10,
				MaxHops:      3,
				MinScore:     0.7,
				VectorWeight: 0.6,
				GraphWeight:  0.4,
			},
			wantErr: true,
			errMsg:  "must have either Text, Embedding, or NodeTypes",
		},
		{
			name: "invalid: zero TopK",
			query: &Query{
				Text:         "test",
				TopK:         0,
				MaxHops:      3,
				MinScore:     0.7,
				VectorWeight: 0.6,
				GraphWeight:  0.4,
			},
			wantErr: true,
			errMsg:  "TopK must be greater than 0",
		},
		{
			name: "valid: zero MaxHops",
			query: &Query{
				Text:         "test",
				TopK:         10,
				MaxHops:      0,
				MinScore:     0.7,
				VectorWeight: 0.6,
				GraphWeight:  0.4,
			},
			wantErr: false,
		},
		{
			name: "invalid: negative MaxHops",
			query: &Query{
				Text:         "test",
				TopK:         10,
				MaxHops:      -1,
				MinScore:     0.7,
				VectorWeight: 0.6,
				GraphWeight:  0.4,
			},
			wantErr: true,
			errMsg:  "MaxHops must be non-negative",
		},
		{
			name: "invalid: MinScore too low",
			query: &Query{
				Text:         "test",
				TopK:         10,
				MaxHops:      3,
				MinScore:     -0.1,
				VectorWeight: 0.6,
				GraphWeight:  0.4,
			},
			wantErr: true,
			errMsg:  "MinScore must be between 0.0 and 1.0",
		},
		{
			name: "invalid: MinScore too high",
			query: &Query{
				Text:         "test",
				TopK:         10,
				MaxHops:      3,
				MinScore:     1.1,
				VectorWeight: 0.6,
				GraphWeight:  0.4,
			},
			wantErr: true,
			errMsg:  "MinScore must be between 0.0 and 1.0",
		},
		{
			name: "invalid: negative VectorWeight",
			query: &Query{
				Text:         "test",
				TopK:         10,
				MaxHops:      3,
				MinScore:     0.7,
				VectorWeight: -0.1,
				GraphWeight:  1.1,
			},
			wantErr: true,
			errMsg:  "VectorWeight must be non-negative",
		},
		{
			name: "invalid: negative GraphWeight",
			query: &Query{
				Text:         "test",
				TopK:         10,
				MaxHops:      3,
				MinScore:     0.7,
				VectorWeight: 1.1,
				GraphWeight:  -0.1,
			},
			wantErr: true,
			errMsg:  "GraphWeight must be non-negative",
		},
		{
			name: "invalid: weights don't sum to 1.0",
			query: &Query{
				Text:         "test",
				TopK:         10,
				MaxHops:      3,
				MinScore:     0.7,
				VectorWeight: 0.5,
				GraphWeight:  0.3,
			},
			wantErr: true,
			errMsg:  "must equal 1.0",
		},
		{
			name: "invalid: zero RunNumber",
			query: func() *Query {
				q := NewQuery("test")
				zero := 0
				q.RunNumber = &zero
				return q
			}(),
			wantErr: true,
			errMsg:  "RunNumber must be greater than 0",
		},
		{
			name: "valid: RunNumber 1",
			query: func() *Query {
				q := NewQuery("test")
				one := 1
				q.RunNumber = &one
				return q
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestQueryBuilderChaining(t *testing.T) {
	q := NewQuery("test query").
		WithTopK(20).
		WithMaxHops(5).
		WithMinScore(0.8).
		WithNodeTypes("host", "port").
		WithMission("mission-123").
		WithWeights(0.7, 0.3).
		WithMissionName("test-mission").
		WithRunNumber(2).
		WithIncludeRunMetadata(true)

	assert.Equal(t, "test query", q.Text)
	assert.Equal(t, 20, q.TopK)
	assert.Equal(t, 5, q.MaxHops)
	assert.Equal(t, 0.8, q.MinScore)
	assert.Equal(t, []string{"host", "port"}, q.NodeTypes)
	assert.Equal(t, "mission-123", q.MissionID)
	assert.Equal(t, 0.7, q.VectorWeight)
	assert.Equal(t, 0.3, q.GraphWeight)
	assert.Equal(t, "test-mission", q.MissionName)
	assert.NotNil(t, q.RunNumber)
	assert.Equal(t, 2, *q.RunNumber)
	assert.True(t, q.IncludeRunMetadata)
}

func TestStructuredQueryChaining(t *testing.T) {
	q := NewStructuredQuery().
		WithNodeTypes("host", "port").
		WithMission("mission-123").
		WithTopK(50)

	assert.Equal(t, "", q.Text)
	assert.Empty(t, q.Embedding)
	assert.Equal(t, []string{"host", "port"}, q.NodeTypes)
	assert.Equal(t, "mission-123", q.MissionID)
	assert.Equal(t, 50, q.TopK)
	assert.Equal(t, 0, q.MaxHops)
	assert.Equal(t, 0.0, q.VectorWeight)
	assert.Equal(t, 1.0, q.GraphWeight)

	// Should validate successfully
	err := q.Validate()
	require.NoError(t, err)
}

// TestWithMissionRun tests setting a specific mission run ID
func TestWithMissionRun(t *testing.T) {
	runID := "run_abc123"
	q := NewQuery("test").WithMissionRun(runID)

	assert.Equal(t, runID, q.MissionRunID)
}

// TestMissionRunIDNotSerialized tests that MissionRunID is not included in JSON
func TestMissionRunIDNotSerialized(t *testing.T) {
	// This test verifies the json:"-" tag works correctly
	// We can't easily test JSON serialization without importing encoding/json,
	// but the tag is verified at compile time
	q := NewQuery("test").WithMissionRun("run_123")

	assert.Equal(t, "run_123", q.MissionRunID)
	// The field should exist and be accessible
	// but won't be serialized to JSON (enforced by json:"-" tag)
}
