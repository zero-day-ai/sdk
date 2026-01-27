package queue

import (
	"testing"
	"time"
)

func TestWorkItem_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		item    WorkItem
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid work item",
			item: WorkItem{
				JobID:       "job-123",
				Index:       0,
				Total:       1,
				Tool:        "nmap",
				InputJSON:   `{"target": "127.0.0.1"}`,
				InputType:   "zero_day.tools.nmap.v1.ScanRequest",
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				TraceID:     "trace-456",
				SpanID:      "span-789",
				SubmittedAt: time.Now().UnixMilli(),
			},
			wantErr: false,
		},
		{
			name: "missing job_id",
			item: WorkItem{
				Index:       0,
				Total:       1,
				Tool:        "nmap",
				InputJSON:   `{"target": "127.0.0.1"}`,
				InputType:   "zero_day.tools.nmap.v1.ScanRequest",
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				SubmittedAt: time.Now().UnixMilli(),
			},
			wantErr: true,
			errMsg:  "job_id is required",
		},
		{
			name: "negative index",
			item: WorkItem{
				JobID:       "job-123",
				Index:       -1,
				Total:       1,
				Tool:        "nmap",
				InputJSON:   `{"target": "127.0.0.1"}`,
				InputType:   "zero_day.tools.nmap.v1.ScanRequest",
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				SubmittedAt: time.Now().UnixMilli(),
			},
			wantErr: true,
			errMsg:  "index must be non-negative, got -1",
		},
		{
			name: "zero total",
			item: WorkItem{
				JobID:       "job-123",
				Index:       0,
				Total:       0,
				Tool:        "nmap",
				InputJSON:   `{"target": "127.0.0.1"}`,
				InputType:   "zero_day.tools.nmap.v1.ScanRequest",
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				SubmittedAt: time.Now().UnixMilli(),
			},
			wantErr: true,
			errMsg:  "total must be positive, got 0",
		},
		{
			name: "index out of bounds",
			item: WorkItem{
				JobID:       "job-123",
				Index:       5,
				Total:       3,
				Tool:        "nmap",
				InputJSON:   `{"target": "127.0.0.1"}`,
				InputType:   "zero_day.tools.nmap.v1.ScanRequest",
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				SubmittedAt: time.Now().UnixMilli(),
			},
			wantErr: true,
			errMsg:  "index 5 is out of bounds for total 3",
		},
		{
			name: "missing tool name",
			item: WorkItem{
				JobID:       "job-123",
				Index:       0,
				Total:       1,
				InputJSON:   `{"target": "127.0.0.1"}`,
				InputType:   "zero_day.tools.nmap.v1.ScanRequest",
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				SubmittedAt: time.Now().UnixMilli(),
			},
			wantErr: true,
			errMsg:  "tool name is required",
		},
		{
			name: "missing input_json",
			item: WorkItem{
				JobID:       "job-123",
				Index:       0,
				Total:       1,
				Tool:        "nmap",
				InputType:   "zero_day.tools.nmap.v1.ScanRequest",
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				SubmittedAt: time.Now().UnixMilli(),
			},
			wantErr: true,
			errMsg:  "input_json is required",
		},
		{
			name: "missing input_type",
			item: WorkItem{
				JobID:       "job-123",
				Index:       0,
				Total:       1,
				Tool:        "nmap",
				InputJSON:   `{"target": "127.0.0.1"}`,
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				SubmittedAt: time.Now().UnixMilli(),
			},
			wantErr: true,
			errMsg:  "input_type is required",
		},
		{
			name: "missing output_type",
			item: WorkItem{
				JobID:       "job-123",
				Index:       0,
				Total:       1,
				Tool:        "nmap",
				InputJSON:   `{"target": "127.0.0.1"}`,
				InputType:   "zero_day.tools.nmap.v1.ScanRequest",
				SubmittedAt: time.Now().UnixMilli(),
			},
			wantErr: true,
			errMsg:  "output_type is required",
		},
		{
			name: "invalid submitted_at",
			item: WorkItem{
				JobID:       "job-123",
				Index:       0,
				Total:       1,
				Tool:        "nmap",
				InputJSON:   `{"target": "127.0.0.1"}`,
				InputType:   "zero_day.tools.nmap.v1.ScanRequest",
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				SubmittedAt: -1,
			},
			wantErr: true,
			errMsg:  "submitted_at must be positive, got -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkItem.IsValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("WorkItem.IsValid() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestWorkItem_Age(t *testing.T) {
	now := time.Now().UnixMilli()

	tests := []struct {
		name        string
		submittedAt int64
		wantMin     time.Duration
		wantMax     time.Duration
	}{
		{
			name:        "recent submission",
			submittedAt: now,
			wantMin:     0,
			wantMax:     100 * time.Millisecond,
		},
		{
			name:        "one second old",
			submittedAt: now - 1000,
			wantMin:     900 * time.Millisecond,
			wantMax:     1100 * time.Millisecond,
		},
		{
			name:        "zero timestamp",
			submittedAt: 0,
			wantMin:     0,
			wantMax:     0,
		},
		{
			name:        "negative timestamp",
			submittedAt: -1,
			wantMin:     0,
			wantMax:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := WorkItem{SubmittedAt: tt.submittedAt}
			age := item.Age()
			if age < tt.wantMin || age > tt.wantMax {
				t.Errorf("WorkItem.Age() = %v, want between %v and %v", age, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestResult_IsValid(t *testing.T) {
	now := time.Now().UnixMilli()

	tests := []struct {
		name    string
		result  Result
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid success result",
			result: Result{
				JobID:       "job-123",
				Index:       0,
				OutputJSON:  `{"result": "success"}`,
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				WorkerID:    "worker-1",
				StartedAt:   now - 1000,
				CompletedAt: now,
			},
			wantErr: false,
		},
		{
			name: "valid error result",
			result: Result{
				JobID:       "job-123",
				Index:       0,
				Error:       "tool execution failed",
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				WorkerID:    "worker-1",
				StartedAt:   now - 1000,
				CompletedAt: now,
			},
			wantErr: false,
		},
		{
			name: "missing job_id",
			result: Result{
				Index:       0,
				OutputJSON:  `{"result": "success"}`,
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				WorkerID:    "worker-1",
				StartedAt:   now - 1000,
				CompletedAt: now,
			},
			wantErr: true,
			errMsg:  "job_id is required",
		},
		{
			name: "negative index",
			result: Result{
				JobID:       "job-123",
				Index:       -1,
				OutputJSON:  `{"result": "success"}`,
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				WorkerID:    "worker-1",
				StartedAt:   now - 1000,
				CompletedAt: now,
			},
			wantErr: true,
			errMsg:  "index must be non-negative, got -1",
		},
		{
			name: "missing output_type",
			result: Result{
				JobID:       "job-123",
				Index:       0,
				OutputJSON:  `{"result": "success"}`,
				WorkerID:    "worker-1",
				StartedAt:   now - 1000,
				CompletedAt: now,
			},
			wantErr: true,
			errMsg:  "output_type is required",
		},
		{
			name: "missing worker_id",
			result: Result{
				JobID:       "job-123",
				Index:       0,
				OutputJSON:  `{"result": "success"}`,
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				StartedAt:   now - 1000,
				CompletedAt: now,
			},
			wantErr: true,
			errMsg:  "worker_id is required",
		},
		{
			name: "invalid started_at",
			result: Result{
				JobID:       "job-123",
				Index:       0,
				OutputJSON:  `{"result": "success"}`,
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				WorkerID:    "worker-1",
				StartedAt:   0,
				CompletedAt: now,
			},
			wantErr: true,
			errMsg:  "started_at must be positive, got 0",
		},
		{
			name: "invalid completed_at",
			result: Result{
				JobID:       "job-123",
				Index:       0,
				OutputJSON:  `{"result": "success"}`,
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				WorkerID:    "worker-1",
				StartedAt:   now,
				CompletedAt: 0,
			},
			wantErr: true,
			errMsg:  "completed_at must be positive, got 0",
		},
		{
			name: "completed_at before started_at",
			result: Result{
				JobID:       "job-123",
				Index:       0,
				OutputJSON:  `{"result": "success"}`,
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				WorkerID:    "worker-1",
				StartedAt:   now,
				CompletedAt: now - 1000,
			},
			wantErr: true,
			errMsg:  "completed_at",
		},
		{
			name: "missing output_json without error",
			result: Result{
				JobID:       "job-123",
				Index:       0,
				OutputType:  "zero_day.tools.nmap.v1.ScanResponse",
				WorkerID:    "worker-1",
				StartedAt:   now - 1000,
				CompletedAt: now,
			},
			wantErr: true,
			errMsg:  "output_json is required when error is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Result.IsValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				// For some error messages we just check contains, not exact match
				if tt.errMsg == "completed_at" {
					if err.Error()[:12] != "completed_at" {
						t.Errorf("Result.IsValid() error = %v, want to start with %v", err.Error(), tt.errMsg)
					}
				} else if err.Error() != tt.errMsg {
					t.Errorf("Result.IsValid() error = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestResult_HasError(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name:   "no error",
			result: Result{Error: ""},
			want:   false,
		},
		{
			name:   "has error",
			result: Result{Error: "something went wrong"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.HasError(); got != tt.want {
				t.Errorf("Result.HasError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_Duration(t *testing.T) {
	now := time.Now().UnixMilli()

	tests := []struct {
		name      string
		startedAt int64
		completed int64
		want      time.Duration
	}{
		{
			name:      "one second duration",
			startedAt: now - 1000,
			completed: now,
			want:      1000 * time.Millisecond,
		},
		{
			name:      "100ms duration",
			startedAt: now - 100,
			completed: now,
			want:      100 * time.Millisecond,
		},
		{
			name:      "zero started_at",
			startedAt: 0,
			completed: now,
			want:      0,
		},
		{
			name:      "zero completed_at",
			startedAt: now,
			completed: 0,
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{
				StartedAt:   tt.startedAt,
				CompletedAt: tt.completed,
			}
			got := r.Duration()
			if got != tt.want {
				t.Errorf("Result.Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolMeta_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		meta    ToolMeta
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid tool meta",
			meta: ToolMeta{
				Name:              "nmap",
				Version:           "1.0.0",
				Description:       "Network scanner",
				InputMessageType:  "zero_day.tools.nmap.v1.ScanRequest",
				OutputMessageType: "zero_day.tools.nmap.v1.ScanResponse",
				Schema:            `{"type": "object"}`,
				Tags:              []string{"network", "discovery"},
				WorkerCount:       3,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			meta: ToolMeta{
				Version:           "1.0.0",
				InputMessageType:  "zero_day.tools.nmap.v1.ScanRequest",
				OutputMessageType: "zero_day.tools.nmap.v1.ScanResponse",
			},
			wantErr: true,
			errMsg:  "tool name is required",
		},
		{
			name: "missing version",
			meta: ToolMeta{
				Name:              "nmap",
				InputMessageType:  "zero_day.tools.nmap.v1.ScanRequest",
				OutputMessageType: "zero_day.tools.nmap.v1.ScanResponse",
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "missing input_type",
			meta: ToolMeta{
				Name:              "nmap",
				Version:           "1.0.0",
				OutputMessageType: "zero_day.tools.nmap.v1.ScanResponse",
			},
			wantErr: true,
			errMsg:  "input_type is required",
		},
		{
			name: "missing output_type",
			meta: ToolMeta{
				Name:             "nmap",
				Version:          "1.0.0",
				InputMessageType: "zero_day.tools.nmap.v1.ScanRequest",
			},
			wantErr: true,
			errMsg:  "output_type is required",
		},
		{
			name: "negative worker_count",
			meta: ToolMeta{
				Name:              "nmap",
				Version:           "1.0.0",
				InputMessageType:  "zero_day.tools.nmap.v1.ScanRequest",
				OutputMessageType: "zero_day.tools.nmap.v1.ScanResponse",
				WorkerCount:       -1,
			},
			wantErr: true,
			errMsg:  "worker_count must be non-negative, got -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToolMeta.IsValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ToolMeta.IsValid() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestToolMeta_SupportsInput(t *testing.T) {
	meta := ToolMeta{
		InputMessageType: "zero_day.tools.nmap.v1.ScanRequest",
	}

	tests := []struct {
		name      string
		inputType string
		want      bool
	}{
		{
			name:      "matching input type",
			inputType: "zero_day.tools.nmap.v1.ScanRequest",
			want:      true,
		},
		{
			name:      "non-matching input type",
			inputType: "zero_day.tools.httpx.v1.Request",
			want:      false,
		},
		{
			name:      "empty input type",
			inputType: "",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := meta.SupportsInput(tt.inputType); got != tt.want {
				t.Errorf("ToolMeta.SupportsInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolMeta_HasTag(t *testing.T) {
	meta := ToolMeta{
		Tags: []string{"network", "discovery", "scanning"},
	}

	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{
			name: "has tag",
			tag:  "network",
			want: true,
		},
		{
			name: "has another tag",
			tag:  "discovery",
			want: true,
		},
		{
			name: "does not have tag",
			tag:  "web",
			want: false,
		},
		{
			name: "empty tag",
			tag:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := meta.HasTag(tt.tag); got != tt.want {
				t.Errorf("ToolMeta.HasTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolMeta_HasTag_EmptyTags(t *testing.T) {
	meta := ToolMeta{
		Tags: []string{},
	}

	if meta.HasTag("network") {
		t.Error("ToolMeta.HasTag() should return false for empty tags")
	}
}

func TestToolMeta_HasTag_NilTags(t *testing.T) {
	meta := ToolMeta{
		Tags: nil,
	}

	if meta.HasTag("network") {
		t.Error("ToolMeta.HasTag() should return false for nil tags")
	}
}
