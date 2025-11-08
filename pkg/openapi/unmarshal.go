package openapi

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML implements custom YAML unmarshaling for Schema
// This handles the type field which can be either a string or array
func (s *Schema) UnmarshalYAML(node *yaml.Node) error {
	// Unmarshal to a map first to get raw values
	var raw map[string]any
	if err := node.Decode(&raw); err != nil {
		return err
	}

	// Handle type field specially
	if typeVal, ok := raw["type"]; ok {
		switch v := typeVal.(type) {
		case string:
			s.Type = []string{v}
		case []any:
			types := make([]string, len(v))
			for i, t := range v {
				if str, ok := t.(string); ok {
					types[i] = str
				} else {
					return fmt.Errorf("type array must contain strings")
				}
			}
			s.Type = types
		default:
			return fmt.Errorf("type must be string or array of strings")
		}
		// Remove type from raw to avoid processing it again
		delete(raw, "type")
	}

	// Marshal back to YAML and unmarshal to struct for all other fields
	yamlData, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}

	// Use type alias to avoid infinite recursion
	type schemaAlias Schema
	return yaml.Unmarshal(yamlData, (*schemaAlias)(s))
}

// UnmarshalJSON implements custom JSON unmarshaling for Schema
func (s *Schema) UnmarshalJSON(data []byte) error {
	// Create a map to parse the raw data
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Handle type field specially first
	if typeRaw, ok := raw["type"]; ok {
		if err := handleTypeField(typeRaw, s); err != nil {
			return err
		}
		// Remove type from raw so it doesn't get processed again
		delete(raw, "type")
	}

	// Marshal the remaining fields back and unmarshal into schema
	remaining, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	// Use a type alias to avoid infinite recursion
	type schemaAlias Schema
	return json.Unmarshal(remaining, (*schemaAlias)(s))
}

// handleTypeField processes the type field which can be string or array
func handleTypeField(data json.RawMessage, schema *Schema) error {
	// Try as array first
	var typeArray []string
	if err := json.Unmarshal(data, &typeArray); err == nil {
		schema.Type = typeArray
		return nil
	}

	// Try as string
	var typeString string
	if err := json.Unmarshal(data, &typeString); err == nil {
		schema.Type = []string{typeString}
		return nil
	}

	return fmt.Errorf("type field must be a string or array of strings")
}
