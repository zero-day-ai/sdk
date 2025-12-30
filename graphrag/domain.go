package graphrag

// AttackPattern represents a MITRE ATT&CK pattern with similarity scoring.
// This is a domain-specific type for security queries returning attack patterns.
type AttackPattern struct {
	// TechniqueID is the MITRE ATT&CK technique ID (e.g., "T1566", "T1566.001").
	TechniqueID string `json:"technique_id"`

	// Name is the human-readable name of the attack pattern.
	Name string `json:"name"`

	// Description provides details about the attack pattern.
	Description string `json:"description"`

	// Tactics are the MITRE ATT&CK tactics this pattern belongs to
	// (e.g., ["Initial Access", "Execution"]).
	Tactics []string `json:"tactics"`

	// Platforms are the target platforms for this attack pattern
	// (e.g., ["Windows", "Linux", "macOS"]).
	Platforms []string `json:"platforms"`

	// Similarity is the similarity score from vector search (0.0 to 1.0).
	// Higher values indicate stronger semantic similarity to the query.
	Similarity float64 `json:"similarity"`
}

// FindingNode represents a security finding with similarity scoring.
// This is a domain-specific type for security queries returning findings.
type FindingNode struct {
	// ID is the unique identifier for the finding.
	ID string `json:"id"`

	// Title is a brief summary of the finding.
	Title string `json:"title"`

	// Description provides detailed information about the finding.
	Description string `json:"description"`

	// Severity indicates the severity level (e.g., "Critical", "High", "Medium", "Low").
	Severity string `json:"severity"`

	// Category is the finding category (e.g., "Vulnerability", "Configuration", "Secret").
	Category string `json:"category"`

	// Confidence is the confidence score for this finding (0.0 to 1.0).
	// Higher values indicate higher confidence in the finding.
	Confidence float64 `json:"confidence"`

	// Similarity is the similarity score from vector search (0.0 to 1.0).
	// Higher values indicate stronger semantic similarity to the query.
	Similarity float64 `json:"similarity"`
}

// AttackChain represents a sequence of attack steps forming an attack chain.
// This is a domain-specific type for security queries analyzing attack patterns.
type AttackChain struct {
	// ID is the unique identifier for the attack chain.
	ID string `json:"id"`

	// Name is a descriptive name for the attack chain.
	Name string `json:"name"`

	// Severity indicates the overall severity of the attack chain
	// (e.g., "Critical", "High", "Medium", "Low").
	Severity string `json:"severity"`

	// Steps are the ordered attack steps in this chain.
	Steps []AttackStep `json:"steps"`
}

// AttackStep represents a single step in an attack chain.
// Each step links a MITRE technique to a finding node with contextual information.
type AttackStep struct {
	// Order is the position of this step in the attack chain (1-based).
	Order int `json:"order"`

	// TechniqueID is the MITRE ATT&CK technique ID for this step
	// (e.g., "T1566.001").
	TechniqueID string `json:"technique_id"`

	// NodeID is the finding or entity node ID associated with this step.
	NodeID string `json:"node_id"`

	// Description provides context about this step in the attack chain.
	Description string `json:"description"`

	// Confidence is the confidence score for this step (0.0 to 1.0).
	// Higher values indicate higher confidence in the step's validity.
	Confidence float64 `json:"confidence"`
}
