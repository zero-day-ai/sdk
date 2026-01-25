package graphrag

import (
	"github.com/zero-day-ai/sdk/api/gen/graphragpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ExtractDiscovery attempts to extract a DiscoveryResult from a proto message.
// It handles two cases:
//  1. The message is a DiscoveryResult itself - return it directly
//  2. The message has a field named "discovery" - extract and return it
//
// Returns nil if no discovery data is found.
//
// This function is used by the harness to automatically extract discovery data
// from tool responses and persist it to the knowledge graph.
func ExtractDiscovery(msg proto.Message) *graphragpb.DiscoveryResult {
	if msg == nil {
		return nil
	}

	// Case 1: Check if the message itself is a DiscoveryResult
	if discovery, ok := msg.(*graphragpb.DiscoveryResult); ok {
		return discovery
	}

	// Case 2: Use reflection to find the "discovery" field
	refl := msg.ProtoReflect()
	fields := refl.Descriptor().Fields()

	// Look for a field named "discovery"
	discoveryField := fields.ByName("discovery")
	if discoveryField == nil {
		// No discovery field in this message type
		return nil
	}

	// Check if the field is set
	if !refl.Has(discoveryField) {
		// Field exists but is not set
		return nil
	}

	// Get the field value
	fieldValue := refl.Get(discoveryField)

	// The field should be a message type
	if discoveryField.Kind() != protoreflect.MessageKind {
		return nil
	}

	// Extract the message
	discoveryMsg := fieldValue.Message().Interface()

	// Type assert to DiscoveryResult
	if discovery, ok := discoveryMsg.(*graphragpb.DiscoveryResult); ok {
		return discovery
	}

	return nil
}
