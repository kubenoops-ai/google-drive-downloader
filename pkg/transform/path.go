package transform

import (
	"fmt"
	"regexp"
	"strings"
)

// PathTransformer handles path transformation using regex patterns and format strings
type PathTransformer struct {
	pattern *regexp.Regexp
	format  string
}

// NewPathTransformer creates a new PathTransformer with the given pattern and format
func NewPathTransformer(pattern, format string) (*PathTransformer, error) {
	if pattern == "" || format == "" {
		return nil, fmt.Errorf("both pattern and format must be non-empty")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	// Validate that format string uses at least one named group from pattern
	subexpNames := re.SubexpNames()
	hasNamedGroup := false
	for _, name := range subexpNames {
		if name != "" && strings.Contains(format, "${"+name+"}") {
			hasNamedGroup = true
			break
		}
	}
	if !hasNamedGroup {
		return nil, fmt.Errorf("format string does not use any captured variables")
	}

	return &PathTransformer{
		pattern: re,
		format:  format,
	}, nil
}

// Transform applies the transformation to the given path
func (t *PathTransformer) Transform(path string) (string, error) {
	// Find named submatches in the path
	match := t.pattern.FindStringSubmatch(path)
	if match == nil {
		return "", fmt.Errorf("path does not match pattern: %s", path)
	}

	// Get the names of the capturing groups
	subexpNames := t.pattern.SubexpNames()

	// Create a map of named captures
	captures := make(map[string]string)
	for i, name := range subexpNames {
		if i != 0 && name != "" { // Skip the first empty submatch and unnamed groups
			captures[name] = match[i]
		}
	}

	// Replace placeholders in the format string
	result := t.format
	for name, value := range captures {
		placeholder := "${" + name + "}"
		result = strings.Replace(result, placeholder, value, -1)
	}

	// Check if any placeholders remain unreplaced
	if strings.Contains(result, "${") && strings.Contains(result, "}") {
		return "", fmt.Errorf("some placeholders in format string were not replaced: %s", result)
	}

	return result, nil
}
