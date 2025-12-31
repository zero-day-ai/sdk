package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
)

// LineParser parses text output line by line with regex patterns
type LineParser struct {
	patterns map[string]*regexp.Regexp
}

// NewLineParser creates a new line parser with patterns
// patterns is a map of pattern names to regex strings
func NewLineParser(patterns map[string]string) (*LineParser, error) {
	compiled := make(map[string]*regexp.Regexp)

	for name, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile pattern %q: %w", name, err)
		}
		compiled[name] = re
	}

	return &LineParser{
		patterns: compiled,
	}, nil
}

// Parse processes text and extracts matched data
// Returns a slice of maps where each map contains the named captures from matched lines
func (p *LineParser) Parse(data []byte) ([]map[string]string, error) {
	var results []map[string]string
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()

		// Try each pattern
		for patternName, re := range p.patterns {
			if re.MatchString(line) {
				match := re.FindStringSubmatch(line)
				if len(match) > 0 {
					result := make(map[string]string)
					result["_pattern"] = patternName
					result["_line"] = line

					// Extract named groups
					for i, name := range re.SubexpNames() {
						if i > 0 && i < len(match) && name != "" {
							result[name] = match[i]
						}
					}

					results = append(results, result)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading text: %w", err)
	}

	return results, nil
}

// ParseWithPattern parses text with a single regex pattern
func ParseWithPattern(data []byte, pattern string) ([]map[string]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile pattern: %w", err)
	}

	var results []map[string]string
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()

		if re.MatchString(line) {
			match := re.FindStringSubmatch(line)
			if len(match) > 0 {
				result := make(map[string]string)
				result["_line"] = line

				// Extract named groups
				for i, name := range re.SubexpNames() {
					if i > 0 && i < len(match) && name != "" {
						result[name] = match[i]
					}
				}

				results = append(results, result)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading text: %w", err)
	}

	return results, nil
}

// ExtractAll extracts all matches of a pattern from text
func ExtractAll(data []byte, pattern string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile pattern: %w", err)
	}

	matches := re.FindAllString(string(data), -1)
	return matches, nil
}

// ExtractGroups extracts all submatch groups from text
func ExtractGroups(data []byte, pattern string) ([][]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile pattern: %w", err)
	}

	matches := re.FindAllStringSubmatch(string(data), -1)
	return matches, nil
}
