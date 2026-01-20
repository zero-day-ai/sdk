package domain

import (
	"github.com/zero-day-ai/sdk/graphrag"
)

// LLM represents a Large Language Model in the knowledge graph.
// An LLM is identified by its provider and model ID (e.g., anthropic/claude-3-opus).
// LLM is a root-level node with no parent relationships.
//
// Example:
//
//	llm := &LLM{
//	    Provider:      "anthropic",
//	    ModelID:       "claude-3-opus-20240229",
//	    Name:          "Claude 3 Opus",
//	    Version:       "20240229",
//	    ContextWindow: 200000,
//	    MaxTokens:     4096,
//	    Capabilities:  []string{"chat", "vision", "function_calling"},
//	}
//
// Identifying Properties:
//   - provider (required): Provider name (anthropic, openai, google, etc.)
//   - model_id (required): Model identifier (claude-3-opus, gpt-4, etc.)
//
// Relationships:
//   - None (root node)
//   - Children: LLMDeployment nodes, LLMCall nodes, FineTune nodes
type LLM struct {
	// Provider is the LLM provider name.
	// This is an identifying property and is required.
	// Examples: "anthropic", "openai", "google", "meta"
	Provider string

	// ModelID is the unique model identifier from the provider.
	// This is an identifying property and is required.
	// Examples: "claude-3-opus-20240229", "gpt-4-turbo", "gemini-pro"
	ModelID string

	// Name is the human-readable model name.
	// Optional. Example: "Claude 3 Opus", "GPT-4 Turbo"
	Name string

	// Version is the model version string.
	// Optional. Example: "20240229", "0125"
	Version string

	// ContextWindow is the maximum context window size in tokens.
	// Optional. Example: 200000, 128000
	ContextWindow int

	// MaxTokens is the maximum output tokens per response.
	// Optional. Example: 4096, 8192
	MaxTokens int

	// Capabilities is a list of model capabilities.
	// Optional. Examples: ["chat", "vision", "function_calling", "streaming"]
	Capabilities []string
}

// NodeType returns the canonical node type for LLM nodes.
// Implements GraphNode interface.
func (l *LLM) NodeType() string {
	return graphrag.NodeTypeLLM
}

// IdentifyingProperties returns the properties that uniquely identify this LLM.
// For LLM nodes, provider and model_id are both identifying properties.
// Implements GraphNode interface.
func (l *LLM) IdentifyingProperties() map[string]any {
	return map[string]any{
		"provider": l.Provider,
		"model_id": l.ModelID,
	}
}

// Properties returns all properties to set on the LLM node.
// Includes both identifying properties and optional descriptive properties.
// Implements GraphNode interface.
func (l *LLM) Properties() map[string]any {
	props := map[string]any{
		"provider": l.Provider,
		"model_id": l.ModelID,
	}

	if l.Name != "" {
		props[graphrag.PropName] = l.Name
	}
	if l.Version != "" {
		props["version"] = l.Version
	}
	if l.ContextWindow > 0 {
		props["context_window"] = l.ContextWindow
	}
	if l.MaxTokens > 0 {
		props["max_tokens"] = l.MaxTokens
	}
	if len(l.Capabilities) > 0 {
		props["capabilities"] = l.Capabilities
	}

	return props
}

// ParentRef returns nil because LLM is a root node with no parent.
// Implements GraphNode interface.
func (l *LLM) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because LLM is a root node.
// Implements GraphNode interface.
func (l *LLM) RelationshipType() string {
	return ""
}

// LLMDeployment represents a deployment instance of an LLM with a specific endpoint.
// An LLMDeployment is identified by its endpoint URL and model ID.
// LLMDeployment nodes are children of LLM nodes.
//
// Example:
//
//	deployment := &LLMDeployment{
//	    Endpoint:    "https://api.anthropic.com/v1",
//	    ModelID:     "claude-3-opus-20240229",
//	    Provider:    "anthropic",
//	    Region:      "us-west-2",
//	    Environment: "production",
//	}
//
// Identifying Properties:
//   - endpoint (required): Deployment endpoint URL
//   - model_id (required): Model identifier being deployed
//
// Relationships:
//   - Parent: LLM node (via DEPLOYS relationship)
type LLMDeployment struct {
	// Endpoint is the URL of the deployment endpoint.
	// This is an identifying property and is required.
	// Example: "https://api.anthropic.com/v1"
	Endpoint string

	// ModelID is the model identifier for this deployment.
	// This is an identifying property and is required.
	// Example: "claude-3-opus-20240229"
	ModelID string

	// Provider is the LLM provider name.
	// Optional. Example: "anthropic", "openai"
	Provider string

	// Region is the deployment region.
	// Optional. Example: "us-west-2", "eu-central-1"
	Region string

	// Environment is the deployment environment.
	// Optional. Example: "production", "staging", "development"
	Environment string
}

// NodeType returns the canonical node type for LLMDeployment nodes.
// Implements GraphNode interface.
func (d *LLMDeployment) NodeType() string {
	return graphrag.NodeTypeLLMDeployment
}

// IdentifyingProperties returns the properties that uniquely identify this deployment.
// For LLMDeployment nodes, endpoint and model_id are both identifying properties.
// Implements GraphNode interface.
func (d *LLMDeployment) IdentifyingProperties() map[string]any {
	return map[string]any{
		"endpoint": d.Endpoint,
		"model_id": d.ModelID,
	}
}

// Properties returns all properties to set on the LLMDeployment node.
// Implements GraphNode interface.
func (d *LLMDeployment) Properties() map[string]any {
	props := map[string]any{
		"endpoint": d.Endpoint,
		"model_id": d.ModelID,
	}

	if d.Provider != "" {
		props["provider"] = d.Provider
	}
	if d.Region != "" {
		props["region"] = d.Region
	}
	if d.Environment != "" {
		props["environment"] = d.Environment
	}

	return props
}

// ParentRef returns a reference to the parent LLM node.
// Implements GraphNode interface.
func (d *LLMDeployment) ParentRef() *NodeRef {
	if d.Provider == "" || d.ModelID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeLLM,
		Properties: map[string]any{
			"provider": d.Provider,
			"model_id": d.ModelID,
		},
	}
}

// RelationshipType returns the relationship type to the parent LLM node.
// Implements GraphNode interface.
func (d *LLMDeployment) RelationshipType() string {
	return graphrag.RelTypeDeployedAs
}

// Prompt represents a prompt template or saved prompt in the knowledge graph.
// A Prompt is identified by its unique ID.
// Prompt is a root-level node with no parent relationships.
//
// Example:
//
//	prompt := &Prompt{
//	    ID:          "prompt-123",
//	    Name:        "Security Analysis",
//	    Template:    "Analyze the following for security vulnerabilities: {{input}}",
//	    Version:     "1.0",
//	    Description: "Analyzes code or configurations for security issues",
//	}
//
// Identifying Properties:
//   - id (required): Unique prompt identifier
//
// Relationships:
//   - None (root node)
type Prompt struct {
	// ID is the unique prompt identifier.
	// This is an identifying property and is required.
	ID string

	// Name is the human-readable prompt name.
	// Optional. Example: "Security Analysis", "Code Review"
	Name string

	// Template is the prompt template content.
	// Optional. May contain template variables like {{input}}
	Template string

	// Version is the prompt version string.
	// Optional. Example: "1.0", "2.3"
	Version string

	// Description describes the prompt's purpose.
	// Optional.
	Description string
}

// NodeType returns the canonical node type for Prompt nodes.
// Implements GraphNode interface.
func (p *Prompt) NodeType() string {
	return graphrag.NodeTypePrompt
}

// IdentifyingProperties returns the properties that uniquely identify this prompt.
// For Prompt nodes, only id is identifying.
// Implements GraphNode interface.
func (p *Prompt) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": p.ID,
	}
}

// Properties returns all properties to set on the Prompt node.
// Implements GraphNode interface.
func (p *Prompt) Properties() map[string]any {
	props := map[string]any{
		"id": p.ID,
	}

	if p.Name != "" {
		props[graphrag.PropName] = p.Name
	}
	if p.Template != "" {
		props["template"] = p.Template
	}
	if p.Version != "" {
		props["version"] = p.Version
	}
	if p.Description != "" {
		props[graphrag.PropDescription] = p.Description
	}

	return props
}

// ParentRef returns nil because Prompt is a root node with no parent.
// Implements GraphNode interface.
func (p *Prompt) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Prompt is a root node.
// Implements GraphNode interface.
func (p *Prompt) RelationshipType() string {
	return ""
}

// SystemPrompt represents a system-level prompt for an application or agent.
// A SystemPrompt is identified by its application ID.
// SystemPrompt is a root-level node with no parent relationships.
//
// Example:
//
//	systemPrompt := &SystemPrompt{
//	    ApplicationID: "my-security-agent",
//	    Content:       "You are a security testing agent...",
//	    Version:       "1.2",
//	}
//
// Identifying Properties:
//   - application_id (required): Application or agent identifier
//
// Relationships:
//   - None (root node)
type SystemPrompt struct {
	// ApplicationID is the application or agent identifier.
	// This is an identifying property and is required.
	ApplicationID string

	// Content is the system prompt content.
	// Optional.
	Content string

	// Version is the prompt version.
	// Optional. Example: "1.2"
	Version string
}

// NodeType returns the canonical node type for SystemPrompt nodes.
// Implements GraphNode interface.
func (s *SystemPrompt) NodeType() string {
	return graphrag.NodeTypeSystemPrompt
}

// IdentifyingProperties returns the properties that uniquely identify this system prompt.
// For SystemPrompt nodes, only application_id is identifying.
// Implements GraphNode interface.
func (s *SystemPrompt) IdentifyingProperties() map[string]any {
	return map[string]any{
		"application_id": s.ApplicationID,
	}
}

// Properties returns all properties to set on the SystemPrompt node.
// Implements GraphNode interface.
func (s *SystemPrompt) Properties() map[string]any {
	props := map[string]any{
		"application_id": s.ApplicationID,
	}

	if s.Content != "" {
		props["content"] = s.Content
	}
	if s.Version != "" {
		props["version"] = s.Version
	}

	return props
}

// ParentRef returns nil because SystemPrompt is a root node with no parent.
// Implements GraphNode interface.
func (s *SystemPrompt) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because SystemPrompt is a root node.
// Implements GraphNode interface.
func (s *SystemPrompt) RelationshipType() string {
	return ""
}

// Guardrail represents a safety guardrail or content filter for LLM interactions.
// A Guardrail is identified by its name.
// Guardrail is a root-level node with no parent relationships.
//
// Example:
//
//	guardrail := &Guardrail{
//	    Name:        "pii-filter",
//	    Type:        "content_filter",
//	    Description: "Filters personally identifiable information",
//	    Enabled:     true,
//	}
//
// Identifying Properties:
//   - name (required): Guardrail name
//
// Relationships:
//   - None (root node)
type Guardrail struct {
	// Name is the guardrail name.
	// This is an identifying property and is required.
	Name string

	// Type is the guardrail type.
	// Optional. Examples: "content_filter", "rate_limit", "toxicity_detection"
	Type string

	// Description describes the guardrail's purpose.
	// Optional.
	Description string

	// Enabled indicates if the guardrail is active.
	// Optional.
	Enabled bool
}

// NodeType returns the canonical node type for Guardrail nodes.
// Implements GraphNode interface.
func (g *Guardrail) NodeType() string {
	return graphrag.NodeTypeGuardrail
}

// IdentifyingProperties returns the properties that uniquely identify this guardrail.
// For Guardrail nodes, only name is identifying.
// Implements GraphNode interface.
func (g *Guardrail) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: g.Name,
	}
}

// Properties returns all properties to set on the Guardrail node.
// Implements GraphNode interface.
func (g *Guardrail) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: g.Name,
	}

	if g.Type != "" {
		props["type"] = g.Type
	}
	if g.Description != "" {
		props[graphrag.PropDescription] = g.Description
	}
	// Always include enabled field as it's a boolean with meaningful false value
	props["enabled"] = g.Enabled

	return props
}

// ParentRef returns nil because Guardrail is a root node with no parent.
// Implements GraphNode interface.
func (g *Guardrail) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Guardrail is a root node.
// Implements GraphNode interface.
func (g *Guardrail) RelationshipType() string {
	return ""
}

// ContentFilter represents a content filtering policy for LLM inputs/outputs.
// A ContentFilter is identified by its name.
// ContentFilter is a root-level node with no parent relationships.
//
// Example:
//
//	filter := &ContentFilter{
//	    Name:     "profanity-filter",
//	    Type:     "blocklist",
//	    Severity: "high",
//	    Action:   "block",
//	}
//
// Identifying Properties:
//   - name (required): Filter name
//
// Relationships:
//   - None (root node)
type ContentFilter struct {
	// Name is the content filter name.
	// This is an identifying property and is required.
	Name string

	// Type is the filter type.
	// Optional. Examples: "blocklist", "allowlist", "regex", "ml_classifier"
	Type string

	// Severity is the severity level for violations.
	// Optional. Examples: "low", "medium", "high", "critical"
	Severity string

	// Action is the action taken on filter match.
	// Optional. Examples: "block", "warn", "log", "redact"
	Action string
}

// NodeType returns the canonical node type for ContentFilter nodes.
// Implements GraphNode interface.
func (c *ContentFilter) NodeType() string {
	return graphrag.NodeTypeContentFilter
}

// IdentifyingProperties returns the properties that uniquely identify this content filter.
// For ContentFilter nodes, only name is identifying.
// Implements GraphNode interface.
func (c *ContentFilter) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
	}
}

// Properties returns all properties to set on the ContentFilter node.
// Implements GraphNode interface.
func (c *ContentFilter) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: c.Name,
	}

	if c.Type != "" {
		props["type"] = c.Type
	}
	if c.Severity != "" {
		props[graphrag.PropSeverity] = c.Severity
	}
	if c.Action != "" {
		props["action"] = c.Action
	}

	return props
}

// ParentRef returns nil because ContentFilter is a root node with no parent.
// Implements GraphNode interface.
func (c *ContentFilter) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because ContentFilter is a root node.
// Implements GraphNode interface.
func (c *ContentFilter) RelationshipType() string {
	return ""
}

// LLMResponse represents a response from an LLM call.
// An LLMResponse is identified by its response ID.
// LLMResponse nodes are children of LLMCall nodes.
//
// Example:
//
//	response := &LLMResponse{
//	    ResponseID:    "resp-123",
//	    LLMCallID:     "call-456",
//	    Content:       "The code appears to have a SQL injection vulnerability...",
//	    FinishReason:  "stop",
//	    PromptTokens:  1500,
//	    OutputTokens:  250,
//	}
//
// Identifying Properties:
//   - response_id (required): Unique response identifier
//
// Relationships:
//   - Parent: LLMCall node (via HAS_RESPONSE relationship)
type LLMResponse struct {
	// ResponseID is the unique response identifier.
	// This is an identifying property and is required.
	ResponseID string

	// LLMCallID is the LLM call that produced this response.
	// Required for parent relationship.
	LLMCallID string

	// Content is the response content/text.
	// Optional.
	Content string

	// FinishReason is the reason the generation stopped.
	// Optional. Examples: "stop", "length", "content_filter"
	FinishReason string

	// PromptTokens is the number of tokens in the prompt.
	// Optional.
	PromptTokens int

	// OutputTokens is the number of tokens in the output.
	// Optional.
	OutputTokens int
}

// NodeType returns the canonical node type for LLMResponse nodes.
// Implements GraphNode interface.
func (r *LLMResponse) NodeType() string {
	return graphrag.NodeTypeLLMResponse
}

// IdentifyingProperties returns the properties that uniquely identify this response.
// For LLMResponse nodes, only response_id is identifying.
// Implements GraphNode interface.
func (r *LLMResponse) IdentifyingProperties() map[string]any {
	return map[string]any{
		"response_id": r.ResponseID,
	}
}

// Properties returns all properties to set on the LLMResponse node.
// Implements GraphNode interface.
func (r *LLMResponse) Properties() map[string]any {
	props := map[string]any{
		"response_id": r.ResponseID,
	}

	if r.Content != "" {
		props["content"] = r.Content
	}
	if r.FinishReason != "" {
		props["finish_reason"] = r.FinishReason
	}
	if r.PromptTokens > 0 {
		props["prompt_tokens"] = r.PromptTokens
	}
	if r.OutputTokens > 0 {
		props["output_tokens"] = r.OutputTokens
	}

	return props
}

// ParentRef returns a reference to the parent LLMCall node.
// Implements GraphNode interface.
func (r *LLMResponse) ParentRef() *NodeRef {
	if r.LLMCallID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeLlmCall,
		Properties: map[string]any{
			"id": r.LLMCallID,
		},
	}
}

// RelationshipType returns the relationship type to the parent LLMCall node.
// Implements GraphNode interface.
func (r *LLMResponse) RelationshipType() string {
	return graphrag.RelTypeGeneratesResponse
}

// TokenUsage represents token usage statistics for a request.
// TokenUsage is identified by its request ID.
// TokenUsage is a root-level node with no parent relationships.
//
// Example:
//
//	usage := &TokenUsage{
//	    RequestID:    "req-789",
//	    PromptTokens: 1500,
//	    OutputTokens: 250,
//	    TotalTokens:  1750,
//	    Cost:         0.0525,
//	}
//
// Identifying Properties:
//   - request_id (required): Request identifier
//
// Relationships:
//   - None (root node)
type TokenUsage struct {
	// RequestID is the request identifier.
	// This is an identifying property and is required.
	RequestID string

	// PromptTokens is the number of tokens in the prompt.
	// Optional.
	PromptTokens int

	// OutputTokens is the number of tokens in the output.
	// Optional.
	OutputTokens int

	// TotalTokens is the total token count.
	// Optional.
	TotalTokens int

	// Cost is the estimated cost in USD.
	// Optional.
	Cost float64
}

// NodeType returns the canonical node type for TokenUsage nodes.
// Implements GraphNode interface.
func (t *TokenUsage) NodeType() string {
	return graphrag.NodeTypeTokenUsage
}

// IdentifyingProperties returns the properties that uniquely identify this token usage record.
// For TokenUsage nodes, only request_id is identifying.
// Implements GraphNode interface.
func (t *TokenUsage) IdentifyingProperties() map[string]any {
	return map[string]any{
		"request_id": t.RequestID,
	}
}

// Properties returns all properties to set on the TokenUsage node.
// Implements GraphNode interface.
func (t *TokenUsage) Properties() map[string]any {
	props := map[string]any{
		"request_id": t.RequestID,
	}

	if t.PromptTokens > 0 {
		props["prompt_tokens"] = t.PromptTokens
	}
	if t.OutputTokens > 0 {
		props["output_tokens"] = t.OutputTokens
	}
	if t.TotalTokens > 0 {
		props["total_tokens"] = t.TotalTokens
	}
	if t.Cost > 0 {
		props["cost"] = t.Cost
	}

	return props
}

// ParentRef returns nil because TokenUsage is a root node with no parent.
// Implements GraphNode interface.
func (t *TokenUsage) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because TokenUsage is a root node.
// Implements GraphNode interface.
func (t *TokenUsage) RelationshipType() string {
	return ""
}

// EmbeddingModel represents a text embedding model in the knowledge graph.
// An EmbeddingModel is identified by its provider and model ID.
// EmbeddingModel is a root-level node with no parent relationships.
//
// Example:
//
//	model := &EmbeddingModel{
//	    Provider:   "openai",
//	    ModelID:    "text-embedding-3-large",
//	    Dimensions: 3072,
//	    MaxTokens:  8191,
//	}
//
// Identifying Properties:
//   - provider (required): Provider name
//   - model_id (required): Model identifier
//
// Relationships:
//   - None (root node)
type EmbeddingModel struct {
	// Provider is the embedding model provider.
	// This is an identifying property and is required.
	Provider string

	// ModelID is the model identifier.
	// This is an identifying property and is required.
	ModelID string

	// Dimensions is the embedding dimension size.
	// Optional. Example: 1536, 3072
	Dimensions int

	// MaxTokens is the maximum input tokens.
	// Optional. Example: 8191
	MaxTokens int
}

// NodeType returns the canonical node type for EmbeddingModel nodes.
// Implements GraphNode interface.
func (e *EmbeddingModel) NodeType() string {
	return graphrag.NodeTypeEmbeddingModel
}

// IdentifyingProperties returns the properties that uniquely identify this embedding model.
// For EmbeddingModel nodes, provider and model_id are both identifying.
// Implements GraphNode interface.
func (e *EmbeddingModel) IdentifyingProperties() map[string]any {
	return map[string]any{
		"provider": e.Provider,
		"model_id": e.ModelID,
	}
}

// Properties returns all properties to set on the EmbeddingModel node.
// Implements GraphNode interface.
func (e *EmbeddingModel) Properties() map[string]any {
	props := map[string]any{
		"provider": e.Provider,
		"model_id": e.ModelID,
	}

	if e.Dimensions > 0 {
		props["dimensions"] = e.Dimensions
	}
	if e.MaxTokens > 0 {
		props["max_tokens"] = e.MaxTokens
	}

	return props
}

// ParentRef returns nil because EmbeddingModel is a root node with no parent.
// Implements GraphNode interface.
func (e *EmbeddingModel) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because EmbeddingModel is a root node.
// Implements GraphNode interface.
func (e *EmbeddingModel) RelationshipType() string {
	return ""
}

// FineTune represents a fine-tuned model training job.
// A FineTune is identified by its unique ID.
// FineTune nodes are children of LLM nodes (base model).
//
// Example:
//
//	fineTune := &FineTune{
//	    ID:          "ft-abc123",
//	    BaseModel:   "gpt-4",
//	    Provider:    "openai",
//	    Status:      "completed",
//	    TrainingSet: "dataset-xyz",
//	}
//
// Identifying Properties:
//   - id (required): Fine-tune job identifier
//
// Relationships:
//   - Parent: LLM node (base model via FINE_TUNES relationship)
type FineTune struct {
	// ID is the fine-tune job identifier.
	// This is an identifying property and is required.
	ID string

	// BaseModel is the base model identifier.
	// Required for parent relationship.
	BaseModel string

	// Provider is the model provider.
	// Required for parent relationship.
	Provider string

	// Status is the fine-tuning job status.
	// Optional. Examples: "pending", "running", "completed", "failed"
	Status string

	// TrainingSet is the training dataset identifier.
	// Optional.
	TrainingSet string
}

// NodeType returns the canonical node type for FineTune nodes.
// Implements GraphNode interface.
func (f *FineTune) NodeType() string {
	return graphrag.NodeTypeFineTune
}

// IdentifyingProperties returns the properties that uniquely identify this fine-tune job.
// For FineTune nodes, only id is identifying.
// Implements GraphNode interface.
func (f *FineTune) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": f.ID,
	}
}

// Properties returns all properties to set on the FineTune node.
// Implements GraphNode interface.
func (f *FineTune) Properties() map[string]any {
	props := map[string]any{
		"id": f.ID,
	}

	if f.Status != "" {
		props["status"] = f.Status
	}
	if f.TrainingSet != "" {
		props["training_set"] = f.TrainingSet
	}

	return props
}

// ParentRef returns a reference to the parent LLM node (base model).
// Implements GraphNode interface.
func (f *FineTune) ParentRef() *NodeRef {
	if f.Provider == "" || f.BaseModel == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeLLM,
		Properties: map[string]any{
			"provider": f.Provider,
			"model_id": f.BaseModel,
		},
	}
}

// RelationshipType returns the relationship type to the parent LLM node.
// Implements GraphNode interface.
func (f *FineTune) RelationshipType() string {
	return graphrag.RelTypeFineTunedFrom
}

// ModelRegistry represents a model registry service.
// A ModelRegistry is identified by its name.
// ModelRegistry is a root-level node with no parent relationships.
//
// Example:
//
//	registry := &ModelRegistry{
//	    Name:     "huggingface",
//	    URL:      "https://huggingface.co",
//	    Type:     "public",
//	}
//
// Identifying Properties:
//   - name (required): Registry name
//
// Relationships:
//   - None (root node)
//   - Children: ModelVersion nodes
type ModelRegistry struct {
	// Name is the registry name.
	// This is an identifying property and is required.
	Name string

	// URL is the registry URL.
	// Optional.
	URL string

	// Type is the registry type.
	// Optional. Examples: "public", "private", "enterprise"
	Type string
}

// NodeType returns the canonical node type for ModelRegistry nodes.
// Implements GraphNode interface.
func (m *ModelRegistry) NodeType() string {
	return graphrag.NodeTypeModelRegistry
}

// IdentifyingProperties returns the properties that uniquely identify this registry.
// For ModelRegistry nodes, only name is identifying.
// Implements GraphNode interface.
func (m *ModelRegistry) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: m.Name,
	}
}

// Properties returns all properties to set on the ModelRegistry node.
// Implements GraphNode interface.
func (m *ModelRegistry) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: m.Name,
	}

	if m.URL != "" {
		props[graphrag.PropURL] = m.URL
	}
	if m.Type != "" {
		props["type"] = m.Type
	}

	return props
}

// ParentRef returns nil because ModelRegistry is a root node with no parent.
// Implements GraphNode interface.
func (m *ModelRegistry) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because ModelRegistry is a root node.
// Implements GraphNode interface.
func (m *ModelRegistry) RelationshipType() string {
	return ""
}

// ModelVersion represents a specific version of a model in a registry.
// A ModelVersion is identified by its registry ID and version string.
// ModelVersion nodes are children of ModelRegistry nodes.
//
// Example:
//
//	version := &ModelVersion{
//	    RegistryID: "huggingface",
//	    Version:    "v2.1.0",
//	    ModelName:  "bert-base-uncased",
//	    Size:       440,
//	}
//
// Identifying Properties:
//   - registry_id (required): Parent registry name
//   - version (required): Version string
//
// Relationships:
//   - Parent: ModelRegistry node (via HAS_VERSION relationship)
type ModelVersion struct {
	// RegistryID is the parent registry name.
	// This is an identifying property and is required.
	RegistryID string

	// Version is the version string.
	// This is an identifying property and is required.
	Version string

	// ModelName is the model name in the registry.
	// Optional.
	ModelName string

	// Size is the model size in megabytes.
	// Optional.
	Size int
}

// NodeType returns the canonical node type for ModelVersion nodes.
// Implements GraphNode interface.
func (m *ModelVersion) NodeType() string {
	return graphrag.NodeTypeModelVersion
}

// IdentifyingProperties returns the properties that uniquely identify this model version.
// For ModelVersion nodes, registry_id and version are both identifying.
// Implements GraphNode interface.
func (m *ModelVersion) IdentifyingProperties() map[string]any {
	return map[string]any{
		"registry_id": m.RegistryID,
		"version":     m.Version,
	}
}

// Properties returns all properties to set on the ModelVersion node.
// Implements GraphNode interface.
func (m *ModelVersion) Properties() map[string]any {
	props := map[string]any{
		"registry_id": m.RegistryID,
		"version":     m.Version,
	}

	if m.ModelName != "" {
		props["model_name"] = m.ModelName
	}
	if m.Size > 0 {
		props["size"] = m.Size
	}

	return props
}

// ParentRef returns a reference to the parent ModelRegistry node.
// Implements GraphNode interface.
func (m *ModelVersion) ParentRef() *NodeRef {
	if m.RegistryID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeModelRegistry,
		Properties: map[string]any{
			graphrag.PropName: m.RegistryID,
		},
	}
}

// RelationshipType returns the relationship type to the parent ModelRegistry node.
// Implements GraphNode interface.
func (m *ModelVersion) RelationshipType() string {
	// Using empty string for now as RelTypeHasVersion doesn't exist
	// This will be added in Phase 1 taxonomy generation
	return ""
}

// InferenceEndpoint represents an inference API endpoint.
// An InferenceEndpoint is identified by its URL.
// InferenceEndpoint is a root-level node with no parent relationships.
//
// Example:
//
//	endpoint := &InferenceEndpoint{
//	    URL:      "https://api.example.com/v1/infer",
//	    Provider: "custom",
//	    Status:   "active",
//	}
//
// Identifying Properties:
//   - url (required): Endpoint URL
//
// Relationships:
//   - None (root node)
type InferenceEndpoint struct {
	// URL is the endpoint URL.
	// This is an identifying property and is required.
	URL string

	// Provider is the provider name.
	// Optional.
	Provider string

	// Status is the endpoint status.
	// Optional. Examples: "active", "inactive", "deprecated"
	Status string
}

// NodeType returns the canonical node type for InferenceEndpoint nodes.
// Implements GraphNode interface.
func (i *InferenceEndpoint) NodeType() string {
	return graphrag.NodeTypeInferenceEndpoint
}

// IdentifyingProperties returns the properties that uniquely identify this endpoint.
// For InferenceEndpoint nodes, only url is identifying.
// Implements GraphNode interface.
func (i *InferenceEndpoint) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropURL: i.URL,
	}
}

// Properties returns all properties to set on the InferenceEndpoint node.
// Implements GraphNode interface.
func (i *InferenceEndpoint) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropURL: i.URL,
	}

	if i.Provider != "" {
		props["provider"] = i.Provider
	}
	if i.Status != "" {
		props["status"] = i.Status
	}

	return props
}

// ParentRef returns nil because InferenceEndpoint is a root node with no parent.
// Implements GraphNode interface.
func (i *InferenceEndpoint) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because InferenceEndpoint is a root node.
// Implements GraphNode interface.
func (i *InferenceEndpoint) RelationshipType() string {
	return ""
}

// BatchJob represents a batch inference job.
// A BatchJob is identified by its unique ID.
// BatchJob is a root-level node with no parent relationships.
//
// Example:
//
//	job := &BatchJob{
//	    ID:       "batch-789",
//	    Status:   "completed",
//	    Model:    "gpt-4",
//	    Records:  1000,
//	}
//
// Identifying Properties:
//   - id (required): Batch job identifier
//
// Relationships:
//   - None (root node)
type BatchJob struct {
	// ID is the batch job identifier.
	// This is an identifying property and is required.
	ID string

	// Status is the job status.
	// Optional. Examples: "pending", "running", "completed", "failed"
	Status string

	// Model is the model used for the batch job.
	// Optional.
	Model string

	// Records is the number of records processed.
	// Optional.
	Records int
}

// NodeType returns the canonical node type for BatchJob nodes.
// Implements GraphNode interface.
func (b *BatchJob) NodeType() string {
	return graphrag.NodeTypeBatchJob
}

// IdentifyingProperties returns the properties that uniquely identify this batch job.
// For BatchJob nodes, only id is identifying.
// Implements GraphNode interface.
func (b *BatchJob) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": b.ID,
	}
}

// Properties returns all properties to set on the BatchJob node.
// Implements GraphNode interface.
func (b *BatchJob) Properties() map[string]any {
	props := map[string]any{
		"id": b.ID,
	}

	if b.Status != "" {
		props["status"] = b.Status
	}
	if b.Model != "" {
		props[graphrag.PropModel] = b.Model
	}
	if b.Records > 0 {
		props["records"] = b.Records
	}

	return props
}

// ParentRef returns nil because BatchJob is a root node with no parent.
// Implements GraphNode interface.
func (b *BatchJob) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because BatchJob is a root node.
// Implements GraphNode interface.
func (b *BatchJob) RelationshipType() string {
	return ""
}

// TrainingRun represents a model training run or experiment.
// A TrainingRun is identified by its unique ID.
// TrainingRun is a root-level node with no parent relationships.
//
// Example:
//
//	run := &TrainingRun{
//	    ID:      "run-456",
//	    Model:   "bert-base",
//	    Status:  "completed",
//	    Epochs:  10,
//	    Loss:    0.234,
//	}
//
// Identifying Properties:
//   - id (required): Training run identifier
//
// Relationships:
//   - None (root node)
type TrainingRun struct {
	// ID is the training run identifier.
	// This is an identifying property and is required.
	ID string

	// Model is the model being trained.
	// Optional.
	Model string

	// Status is the training status.
	// Optional. Examples: "pending", "running", "completed", "failed"
	Status string

	// Epochs is the number of training epochs.
	// Optional.
	Epochs int

	// Loss is the final training loss.
	// Optional.
	Loss float64
}

// NodeType returns the canonical node type for TrainingRun nodes.
// Implements GraphNode interface.
func (t *TrainingRun) NodeType() string {
	return graphrag.NodeTypeTrainingRun
}

// IdentifyingProperties returns the properties that uniquely identify this training run.
// For TrainingRun nodes, only id is identifying.
// Implements GraphNode interface.
func (t *TrainingRun) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": t.ID,
	}
}

// Properties returns all properties to set on the TrainingRun node.
// Implements GraphNode interface.
func (t *TrainingRun) Properties() map[string]any {
	props := map[string]any{
		"id": t.ID,
	}

	if t.Model != "" {
		props[graphrag.PropModel] = t.Model
	}
	if t.Status != "" {
		props["status"] = t.Status
	}
	if t.Epochs > 0 {
		props["epochs"] = t.Epochs
	}
	if t.Loss > 0 {
		props["loss"] = t.Loss
	}

	return props
}

// ParentRef returns nil because TrainingRun is a root node with no parent.
// Implements GraphNode interface.
func (t *TrainingRun) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because TrainingRun is a root node.
// Implements GraphNode interface.
func (t *TrainingRun) RelationshipType() string {
	return ""
}

// Dataset represents a training or evaluation dataset.
// A Dataset is identified by its unique ID.
// Dataset is a root-level node with no parent relationships.
//
// Example:
//
//	dataset := &Dataset{
//	    ID:      "ds-123",
//	    Name:    "security-prompts",
//	    Type:    "training",
//	    Size:    50000,
//	    Format:  "jsonl",
//	}
//
// Identifying Properties:
//   - id (required): Dataset identifier
//
// Relationships:
//   - None (root node)
type Dataset struct {
	// ID is the dataset identifier.
	// This is an identifying property and is required.
	ID string

	// Name is the dataset name.
	// Optional.
	Name string

	// Type is the dataset type.
	// Optional. Examples: "training", "validation", "test"
	Type string

	// Size is the number of records in the dataset.
	// Optional.
	Size int

	// Format is the dataset format.
	// Optional. Examples: "jsonl", "csv", "parquet"
	Format string
}

// NodeType returns the canonical node type for Dataset nodes.
// Implements GraphNode interface.
func (d *Dataset) NodeType() string {
	return graphrag.NodeTypeDataset
}

// IdentifyingProperties returns the properties that uniquely identify this dataset.
// For Dataset nodes, only id is identifying.
// Implements GraphNode interface.
func (d *Dataset) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": d.ID,
	}
}

// Properties returns all properties to set on the Dataset node.
// Implements GraphNode interface.
func (d *Dataset) Properties() map[string]any {
	props := map[string]any{
		"id": d.ID,
	}

	if d.Name != "" {
		props[graphrag.PropName] = d.Name
	}
	if d.Type != "" {
		props["type"] = d.Type
	}
	if d.Size > 0 {
		props["size"] = d.Size
	}
	if d.Format != "" {
		props["format"] = d.Format
	}

	return props
}

// ParentRef returns nil because Dataset is a root node with no parent.
// Implements GraphNode interface.
func (d *Dataset) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Dataset is a root node.
// Implements GraphNode interface.
func (d *Dataset) RelationshipType() string {
	return ""
}
