package types

import (
	"encoding/json"
	"testing"
)

func TestTechniqueType_String(t *testing.T) {
	tests := []struct {
		name          string
		techniqueType TechniqueType
		want          string
	}{
		{"prompt injection", TechniquePromptInjection, "prompt_injection"},
		{"jailbreak", TechniqueJailbreak, "jailbreak"},
		{"data extraction", TechniqueDataExtraction, "data_extraction"},
		{"model manipulation", TechniqueModelManipulation, "model_manipulation"},
		{"dos", TechniqueDOS, "dos"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.techniqueType.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTechniqueType_IsValid(t *testing.T) {
	tests := []struct {
		name          string
		techniqueType TechniqueType
		want          bool
	}{
		{"valid prompt injection", TechniquePromptInjection, true},
		{"valid jailbreak", TechniqueJailbreak, true},
		{"valid data extraction", TechniqueDataExtraction, true},
		{"valid model manipulation", TechniqueModelManipulation, true},
		{"valid dos", TechniqueDOS, true},
		{"invalid empty", TechniqueType(""), false},
		{"invalid unknown", TechniqueType("unknown"), false},
		{"invalid custom", TechniqueType("custom_technique"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.techniqueType.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTechniqueType_Description(t *testing.T) {
	tests := []struct {
		name          string
		techniqueType TechniqueType
		wantContains  string
	}{
		{"prompt injection", TechniquePromptInjection, "Prompt injection"},
		{"jailbreak", TechniqueJailbreak, "Jailbreak"},
		{"data extraction", TechniqueDataExtraction, "Data extraction"},
		{"model manipulation", TechniqueModelManipulation, "Model manipulation"},
		{"dos", TechniqueDOS, "Denial-of-service"},
		{"unknown", TechniqueType("unknown"), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.techniqueType.Description()
			if got == "" {
				t.Error("Description() returned empty string")
			}
			// Just verify we get a non-empty description
			if len(got) == 0 {
				t.Errorf("Description() = empty, want non-empty")
			}
		})
	}
}

func TestTechniqueType_Severity(t *testing.T) {
	tests := []struct {
		name          string
		techniqueType TechniqueType
		want          string
	}{
		{"prompt injection", TechniquePromptInjection, "high"},
		{"jailbreak", TechniqueJailbreak, "high"},
		{"data extraction", TechniqueDataExtraction, "critical"},
		{"model manipulation", TechniqueModelManipulation, "medium"},
		{"dos", TechniqueDOS, "high"},
		{"unknown", TechniqueType("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.techniqueType.Severity(); got != tt.want {
				t.Errorf("Severity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTechniqueInfo_Validate(t *testing.T) {
	tests := []struct {
		name      string
		technique TechniqueInfo
		wantErr   bool
		errField  string
	}{
		{
			name: "valid technique",
			technique: TechniqueInfo{
				Type: TechniquePromptInjection,
				Name: "System Prompt Override",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			technique: TechniqueInfo{
				Type: TechniqueType("invalid"),
				Name: "Test Technique",
			},
			wantErr:  true,
			errField: "Type",
		},
		{
			name: "missing name",
			technique: TechniqueInfo{
				Type: TechniqueJailbreak,
			},
			wantErr:  true,
			errField: "Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.technique.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if verr, ok := err.(*ValidationError); ok {
					if verr.Field != tt.errField {
						t.Errorf("Validate() error field = %v, want %v", verr.Field, tt.errField)
					}
				}
			}
		})
	}
}

func TestTechniqueInfo_HasTag(t *testing.T) {
	technique := &TechniqueInfo{
		Tags: []string{"injection", "high-risk", "automated"},
	}

	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{"existing tag", "injection", true},
		{"another existing tag", "high-risk", true},
		{"non-existent tag", "manual", false},
		{"empty tag", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := technique.HasTag(tt.tag); got != tt.want {
				t.Errorf("HasTag(%v) = %v, want %v", tt.tag, got, tt.want)
			}
		})
	}

	// Test with nil tags
	emptyTechnique := &TechniqueInfo{}
	if emptyTechnique.HasTag("any-tag") {
		t.Error("HasTag on nil tags should return false")
	}
}

func TestTechniqueInfo_AddTag(t *testing.T) {
	technique := &TechniqueInfo{}

	// Add first tag
	technique.AddTag("injection")
	if !technique.HasTag("injection") {
		t.Error("AddTag failed to add tag")
	}

	// Add duplicate tag (should not add twice)
	technique.AddTag("injection")
	count := 0
	for _, tag := range technique.Tags {
		if tag == "injection" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("AddTag added duplicate tag, count = %d, want 1", count)
	}

	// Add different tag
	technique.AddTag("high-risk")
	if !technique.HasTag("high-risk") {
		t.Error("AddTag failed to add second tag")
	}

	if len(technique.Tags) != 2 {
		t.Errorf("Tags length = %d, want 2", len(technique.Tags))
	}
}

func TestTechniqueInfo_GetMetadata(t *testing.T) {
	technique := &TechniqueInfo{
		Metadata: map[string]any{
			"success_rate": 0.85,
			"attempts":     150,
			"reference":    "https://example.com/technique",
		},
	}

	tests := []struct {
		name    string
		key     string
		wantVal any
		wantOk  bool
	}{
		{"existing float", "success_rate", 0.85, true},
		{"existing int", "attempts", 150, true},
		{"existing string", "reference", "https://example.com/technique", true},
		{"non-existent", "unknown", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := technique.GetMetadata(tt.key)
			if gotOk != tt.wantOk {
				t.Errorf("GetMetadata(%v) ok = %v, want %v", tt.key, gotOk, tt.wantOk)
			}
			if tt.wantOk && gotVal != tt.wantVal {
				t.Errorf("GetMetadata(%v) val = %v, want %v", tt.key, gotVal, tt.wantVal)
			}
		})
	}

	// Test with nil metadata
	emptyTechnique := &TechniqueInfo{}
	if _, ok := emptyTechnique.GetMetadata("any-key"); ok {
		t.Error("GetMetadata on nil metadata should return false")
	}
}

func TestTechniqueInfo_SetMetadata(t *testing.T) {
	technique := &TechniqueInfo{}

	// Set first value
	technique.SetMetadata("success_rate", 0.85)
	val, ok := technique.GetMetadata("success_rate")
	if !ok {
		t.Fatal("SetMetadata failed to set value")
	}
	if val != 0.85 {
		t.Errorf("After SetMetadata, GetMetadata() = %v, want %v", val, 0.85)
	}

	// Update existing value
	technique.SetMetadata("success_rate", 0.90)
	val, ok = technique.GetMetadata("success_rate")
	if !ok {
		t.Fatal("GetMetadata failed to retrieve updated value")
	}
	if val != 0.90 {
		t.Errorf("After updating metadata, GetMetadata() = %v, want %v", val, 0.90)
	}

	// Add different types
	technique.SetMetadata("attempts", 150)
	technique.SetMetadata("reference", "https://example.com")

	intVal, _ := technique.GetMetadata("attempts")
	if intVal != 150 {
		t.Errorf("Integer metadata = %v, want %v", intVal, 150)
	}

	strVal, _ := technique.GetMetadata("reference")
	if strVal != "https://example.com" {
		t.Errorf("String metadata = %v, want %v", strVal, "https://example.com")
	}
}

func TestTechniqueInfo_JSONMarshaling(t *testing.T) {
	original := TechniqueInfo{
		Type:        TechniquePromptInjection,
		Name:        "System Prompt Override",
		Description: "Attempts to override system instructions",
		Tags:        []string{"injection", "high-risk"},
		Metadata: map[string]any{
			"success_rate": 0.85,
			"attempts":     150,
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var unmarshaled TechniqueInfo
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify fields
	if unmarshaled.Type != original.Type {
		t.Errorf("Type = %v, want %v", unmarshaled.Type, original.Type)
	}

	if unmarshaled.Name != original.Name {
		t.Errorf("Name = %v, want %v", unmarshaled.Name, original.Name)
	}

	if unmarshaled.Description != original.Description {
		t.Errorf("Description = %v, want %v", unmarshaled.Description, original.Description)
	}

	// Verify tags
	if len(unmarshaled.Tags) != len(original.Tags) {
		t.Errorf("Tags length = %v, want %v", len(unmarshaled.Tags), len(original.Tags))
	}

	// Verify metadata (note: JSON unmarshaling converts numbers to float64)
	if unmarshaled.Metadata["success_rate"] != 0.85 {
		t.Errorf("Metadata[success_rate] = %v, want %v", unmarshaled.Metadata["success_rate"], 0.85)
	}

	if unmarshaled.Metadata["attempts"] != float64(150) {
		t.Errorf("Metadata[attempts] = %v, want %v", unmarshaled.Metadata["attempts"], 150)
	}
}

func TestTechniqueTypeConstants(t *testing.T) {
	// Verify constants have expected values
	if TechniquePromptInjection != "prompt_injection" {
		t.Errorf("TechniquePromptInjection = %v, want %v", TechniquePromptInjection, "prompt_injection")
	}

	if TechniqueJailbreak != "jailbreak" {
		t.Errorf("TechniqueJailbreak = %v, want %v", TechniqueJailbreak, "jailbreak")
	}

	if TechniqueDataExtraction != "data_extraction" {
		t.Errorf("TechniqueDataExtraction = %v, want %v", TechniqueDataExtraction, "data_extraction")
	}

	if TechniqueModelManipulation != "model_manipulation" {
		t.Errorf("TechniqueModelManipulation = %v, want %v", TechniqueModelManipulation, "model_manipulation")
	}

	if TechniqueDOS != "dos" {
		t.Errorf("TechniqueDOS = %v, want %v", TechniqueDOS, "dos")
	}
}
