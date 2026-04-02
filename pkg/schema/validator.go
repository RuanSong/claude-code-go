package schema

import (
	"encoding/json"
	"fmt"
)

// Schema represents a JSON Schema for tool input validation
type Schema interface {
	Validate(value json.RawMessage) error
	Type() string
}

type Object struct {
	Properties map[string]Schema
	Required   []string
	Optional   map[string]Schema
}

func (o Object) Type() string { return "object" }

func (o Object) Validate(data json.RawMessage) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	requiredSet := make(map[string]bool)
	for _, k := range o.Required {
		requiredSet[k] = true
	}

	for k := range m {
		if _, exists := o.Properties[k]; !exists {
			if o.Optional == nil {
				return fmt.Errorf("unknown property: %s", k)
			}
			if _, optExists := o.Optional[k]; !optExists {
				return fmt.Errorf("unknown property: %s", k)
			}
		}
	}

	for k := range o.Properties {
		if requiredSet[k] {
			if v, exists := m[k]; !exists {
				return fmt.Errorf("required property missing: %s", k)
			} else if err := o.Properties[k].Validate(v); err != nil {
				return fmt.Errorf("property %s: %w", k, err)
			}
		}
	}

	return nil
}

type String struct{}

func (s String) Type() string { return "string" }

func (s String) Validate(data json.RawMessage) error {
	var val string
	if err := json.Unmarshal(data, &val); err != nil {
		return fmt.Errorf("expected string: %w", err)
	}
	return nil
}

type Integer struct {
	Min *int64
	Max *int64
}

func (i Integer) Type() string { return "integer" }

func (i Integer) Validate(data json.RawMessage) error {
	var val int64
	if err := json.Unmarshal(data, &val); err != nil {
		return fmt.Errorf("expected integer: %w", err)
	}
	if i.Min != nil && val < *i.Min {
		return fmt.Errorf("value %d is less than minimum %d", val, *i.Min)
	}
	if i.Max != nil && val > *i.Max {
		return fmt.Errorf("value %d is greater than maximum %d", val, *i.Max)
	}
	return nil
}

type Boolean struct{}

func (b Boolean) Type() string { return "boolean" }

func (b Boolean) Validate(data json.RawMessage) error {
	var val bool
	if err := json.Unmarshal(data, &val); err != nil {
		return fmt.Errorf("expected boolean: %w", err)
	}
	return nil
}

type Float struct{}

func (f Float) Type() string { return "number" }

func (f Float) Validate(data json.RawMessage) error {
	var val float64
	if err := json.Unmarshal(data, &val); err != nil {
		return fmt.Errorf("expected number: %w", err)
	}
	return nil
}

type Array struct {
	Items Schema
}

func (a Array) Type() string { return "array" }

func (a Array) Validate(data json.RawMessage) error {
	var val []json.RawMessage
	if err := json.Unmarshal(data, &val); err != nil {
		return fmt.Errorf("expected array: %w", err)
	}
	for i, item := range val {
		if err := a.Items.Validate(item); err != nil {
			return fmt.Errorf("item %d: %w", i, err)
		}
	}
	return nil
}

type Optional struct {
	Schema Schema
}

func (o Optional) Type() string { return o.Schema.Type() }

func (o Optional) Validate(data json.RawMessage) error {
	if data == nil || string(data) == "null" {
		return nil
	}
	return o.Schema.Validate(data)
}

func ObjectProperty(key string, schema Schema) map[string]Schema {
	return map[string]Schema{key: schema}
}

func Required(props ...string) []string {
	return props
}
