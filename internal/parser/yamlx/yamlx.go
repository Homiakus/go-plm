// Package yamlx provides typed YAML encode/decode with line/column error reporting.
package yamlx

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Decode unmarshals YAML data into a typed value.
func Decode[T any](data []byte) (T, error) {
	var v T
	if err := yaml.Unmarshal(data, &v); err != nil {
		return v, fmt.Errorf("yamlx: %w", err)
	}
	return v, nil
}

// Encode marshals a value to YAML.
func Encode(v any) ([]byte, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("yamlx: %w", err)
	}
	return data, nil
}

// StrictDecode unmarshals with strict mode (unknown fields are errors).
func StrictDecode[T any](data []byte) (T, error) {
	var v T
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&v); err != nil {
		return v, fmt.Errorf("yamlx strict: %w", err)
	}
	return v, nil
}
