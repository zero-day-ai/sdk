package domain

// Finding represents a security vulnerability or issue discovered during testing.
// This type will be fully implemented in subsequent phases.

type Finding struct{ ID string }

func (f *Finding) NodeType() string                      { return "finding" }
func (f *Finding) IdentifyingProperties() map[string]any { return map[string]any{"id": f.ID} }
func (f *Finding) Properties() map[string]any            { return f.IdentifyingProperties() }
func (f *Finding) ParentRef() *NodeRef                   { return nil }
func (f *Finding) RelationshipType() string              { return "" }
