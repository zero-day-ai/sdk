package parser

import (
	"encoding/xml"
	"fmt"
)

// ParseXML parses XML output into a structured type using generics
func ParseXML[T any](data []byte) (*T, error) {
	var result T
	if err := xml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}
	return &result, nil
}
