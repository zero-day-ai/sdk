package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
)

// ParseJSONLines parses newline-delimited JSON output using generics
func ParseJSONLines[T any](data []byte) ([]T, error) {
	var results []T
	scanner := bufio.NewScanner(bytes.NewReader(data))

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		// Skip empty lines
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var item T
		if err := json.Unmarshal(line, &item); err != nil {
			return nil, fmt.Errorf("failed to parse JSON at line %d: %w", lineNum, err)
		}
		results = append(results, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading JSON lines: %w", err)
	}

	return results, nil
}

// ParseJSON parses single JSON object using generics
func ParseJSON[T any](data []byte) (*T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &result, nil
}

// ParseJSONArray parses a JSON array into a slice using generics
func ParseJSONArray[T any](data []byte) ([]T, error) {
	var results []T
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array: %w", err)
	}
	return results, nil
}
