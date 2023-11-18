package gocfg

import (
	"fmt"
	"github.com/Jagerente/gocfg/pkg/parsers"
	"github.com/Jagerente/gocfg/pkg/values"
	"log"
	"reflect"
	"strings"
)

// Default tags for struct field annotations
const (
	structKeyTag        = "env"
	structDefaultTag    = "default"
	structAllowEmptyTag = "omitempty"
)

// ValueProvider defines the interface for retrieving values based on keys
type ValueProvider interface {
	Get(key string) string
}

// ParserProvider defines the interface for retrieving parsers for struct fields
type ParserProvider interface {
	Get(reflect.Value) (func(v string) (interface{}, error), bool)
}

// ConfigManager represents the configuration manager
type ConfigManager struct {
	structKeyTag        string
	structDefaultTag    string
	structAllowEmptyTag string
	parserProviders     []ParserProvider
	valueProviders      []ValueProvider
	useDefaults         bool
	forceDefaults       bool
}

// NewEmpty creates a new ConfigManager instance with default tags and empty providers
func NewEmpty() *ConfigManager {
	return &ConfigManager{
		structKeyTag:        structKeyTag,
		structDefaultTag:    structDefaultTag,
		structAllowEmptyTag: structAllowEmptyTag,
		parserProviders:     make([]ParserProvider, 0),
		valueProviders:      make([]ValueProvider, 0),
	}
}

// NewDefault creates a new ConfigManager instance with default tags, default parser, and environment value provider
func NewDefault() *ConfigManager {
	cfg := NewEmpty().
		AddParserProviders(parsers.NewDefaultParserProvider()).
		AddValueProviders(values.NewEnvProvider()).
		UseDefaults()
	return cfg
}

// AddParserProviders adds parser providers to the ConfigManager instance, with higher priority for the providers added first.
// Which means second provider's result will not overwrite the first providers' result.
func (c *ConfigManager) AddParserProviders(provider ...ParserProvider) *ConfigManager {
	for _, p := range provider {
		c.parserProviders = append(c.parserProviders, p)
	}
	return c
}

// AddValueProviders adds value providers to the Config instance, with higher priority for the providers added first.
// Which means second provider's result will not overwrite the first provider's result.
func (c *ConfigManager) AddValueProviders(providers ...ValueProvider) *ConfigManager {
	for _, p := range providers {
		c.valueProviders = append(c.valueProviders, p)
	}
	return c
}

// UseDefaults enables the use of default values during the configuration process.
func (c *ConfigManager) UseDefaults() *ConfigManager {
	c.useDefaults = true
	return c
}

// ForceDefaults enables the use of default values even when a value is provided
func (c *ConfigManager) ForceDefaults() *ConfigManager {
	c.useDefaults = true
	c.forceDefaults = true
	return c
}

// UseCustomKeyTag sets a custom key tag for struct field annotations
func (c *ConfigManager) UseCustomKeyTag(tag string) *ConfigManager {
	c.structKeyTag = tag
	return c
}

// Unmarshal looking for environment variables and assigns their values
// to corresponding fields of a structure using "env" tags.
//
// The 'cfg' argument must be a pointer to the structure where values will be placed.
// The function recursively traverses the fields of the structure and its nested structures,
// which means structure can contain as much nested structures as you want.
//
// You may use omitempty tags to allow empty fields. STRINGS ONLY!
//
// Example:
//
//	type TestConfig struct {
//		BoolField				bool			`env:"BOOL_FIELD"`
//		StringField				string			`env:"STRING_FIELD"`
//		IntField				int				`env:"INT_FIELD"`
//		Int8Field				int8			`env:"INT8_FIELD"`
//		Int16Field				int16			`env:"INT16_FIELD"`
//		Int32Field				int32			`env:"INT32_FIELD"`
//		Int64Field				int64			`env:"INT64_FIELD"`
//		UintField				uint			`env:"UINT_FIELD"`
//		Uint8Field				uint8			`env:"UINT8_FIELD"`
//		Uint16Field				uint16			`env:"UINT16_FIELD"`
//		Uint32Field				uint32			`env:"UINT32_FIELD"`
//		Uint64Field				uint64			`env:"UINT64_FIELD"`
//		Float32Field			float32			`env:"FLOAT32_FIELD"`
//		Float64Field			float64			`env:"FLOAT64_FIELD"`
//		TimeDurationField 		time.Duration 	`env:"TIME_DURATION_FIELD"`
//		EmptyField				string			`env:"EMPTY_FIELD,omitempty"`
//		WithDefaultField		string			`env:"WITH_DEFAULT_FIELD" default:"ave"`
//	}
func (c *ConfigManager) Unmarshal(cfg interface{}) error {
	val := reflect.ValueOf(cfg).Elem()

	for i := 0; i < val.NumField(); i++ {
		var (
			field        = val.Field(i)
			tag          = val.Type().Field(i).Tag.Get(c.structKeyTag)
			key          = strings.Split(tag, ",")[0]
			allowEmpty   = strings.Contains(tag, c.structAllowEmptyTag)
			defaultValue = val.Type().Field(i).Tag.Get(c.structDefaultTag)
		)

		if field.Kind() == reflect.Struct {
			if err := c.Unmarshal(field.Addr().Interface()); err != nil {
				return fmt.Errorf("failed to parse %s: %w", val.Type().Field(i).Name, err)
			}
			continue
		}

		var value string
		if !c.forceDefaults {
			value = c.getValue(key)
		}

		if !allowEmpty && value == "" && (defaultValue == "" || !c.useDefaults) {
			return fmt.Errorf("%s cannot be empty", key)
		}

		parser, ok := c.getParser(field)
		if !ok {
			return fmt.Errorf("failed to get parser for %s: unsupported", key)
		}

		if (value == "" && c.useDefaults) || c.forceDefaults {
			if !c.forceDefaults {
				log.Printf("WARNING: value for %s not found, using default value: %s", key, defaultValue)
			}

			value = defaultValue
		}

		v, err := parser(value)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", key, err)
		}

		field.Set(reflect.ValueOf(v).Convert(field.Type()))
	}

	return nil
}

// getValue retrieves the value for a key from registered value providers
func (c *ConfigManager) getValue(key string) string {
	for _, p := range c.valueProviders {
		if value := p.Get(key); value != "" {
			return value
		}
	}
	return ""
}

// getParser retrieves the parser function for a field from registered parser providers
func (c *ConfigManager) getParser(field reflect.Value) (parser func(v string) (interface{}, error), ok bool) {
	for _, provider := range c.parserProviders {
		if parser, ok = provider.Get(field); ok {
			return
		}
	}
	return
}
