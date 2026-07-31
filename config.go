package gocfg

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Jagerente/gocfg/pkg/parsers"
	"github.com/Jagerente/gocfg/pkg/values"
)

// Default tags for struct field annotations
const (
	structKeyTag         = "env"
	structDefaultTag     = "default"
	structExampleTag     = "example"
	structAllowEmptyTag  = "omitempty"
	structDescriptionTag = "description"
	structTitleTag       = "title"
	structPrefixTag      = "envPrefix"
	structSeparatorTag   = "envSeparator"
)

// skipTagValue marks a field that gocfg must ignore entirely, e.g. `env:"-"`.
const skipTagValue = "-"

// Parser converts a raw configuration value into a Go value.
type Parser = func(v string) (interface{}, error)

// ValueProvider defines the interface for retrieving values based on keys
type ValueProvider interface {
	Get(key string) string
}

// LookupValueProvider is an optional interface a ValueProvider may implement to
// distinguish "the key is not set" from "the key is set to an empty string".
//
// It is consulted for fields marked with the omitempty option, the only ones for
// which an empty value is a value of its own. There, `KEY=` stops the lookup at
// that provider and keeps the default tag from being applied. Both
// values.EnvProvider and values.DotEnvProvider implement it; a provider exposing
// only Get is never able to report an empty value as present.
type LookupValueProvider interface {
	Lookup(key string) (value string, found bool)
}

// ParserProvider defines the interface for retrieving parsers for struct fields
type ParserProvider interface {
	Get(reflect.Value) (Parser, bool)
}

// DocGenerator defines the interface for generating documentation for struct fields
type DocGenerator interface {
	GenerateDoc(*DocTree) error
}

// ConfigManager represents the configuration manager
type ConfigManager struct {
	structKeyTag         string
	structDefaultTag     string
	structExampleTag     string
	structAllowEmptyTag  string
	parserProviders      []ParserProvider
	valueProviders       []ValueProvider
	useDefaults          bool
	forceDefaults        bool
	structDescriptionTag string
	structTitleTag       string
	structPrefixTag      string
	structSeparatorTag   string
	collectAllErrors     bool
	logger               Logger
}

// NewEmpty creates a new ConfigManager instance with default tags and empty providers
func NewEmpty() *ConfigManager {
	return &ConfigManager{
		structKeyTag:         structKeyTag,
		structDefaultTag:     structDefaultTag,
		structExampleTag:     structExampleTag,
		structAllowEmptyTag:  structAllowEmptyTag,
		structDescriptionTag: structDescriptionTag,
		structTitleTag:       structTitleTag,
		structPrefixTag:      structPrefixTag,
		structSeparatorTag:   structSeparatorTag,
		parserProviders:      make([]ParserProvider, 0),
		valueProviders:       make([]ValueProvider, 0),
	}
}

// NewDefault creates a new ConfigManager instance with default tags, default parsers, and environment value provider.
//
// The default parsers handle every built-in type listed in the package
// documentation. Types that none of them recognise are handed over to
// parsers.TextUnmarshalerParserProvider, which covers any type whose pointer
// implements encoding.TextUnmarshaler (time.Time, net.IP, netip.Addr, ...).
func NewDefault() *ConfigManager {
	cfg := NewEmpty().
		AddParserProviders(
			parsers.NewDefaultParserProvider(),
			parsers.NewTextUnmarshalerParserProvider(),
		).
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

// CollectAllErrors makes Unmarshal report every invalid field at once instead of
// stopping at the first one. The returned error joins all field errors, so
// errors.Is and errors.As keep working on it.
func (c *ConfigManager) CollectAllErrors() *ConfigManager {
	c.collectAllErrors = true
	return c
}

// WithLogger sets the logger used for non-fatal diagnostics, currently emitted
// when a field falls back to its default value.
//
// Passing nil disables logging, which is the default: gocfg writes nothing
// anywhere unless a logger is configured. Note that a nil pointer stored in a
// non-nil interface (for example a nil *log.Logger) is not detected and will
// panic when used.
//
//	cfg.WithLogger(log.Default())                  // standard library
//	cfg.WithLogger(gocfg.LoggerFunc(sugar.Infof))  // zap
func (c *ConfigManager) WithLogger(logger Logger) *ConfigManager {
	c.logger = logger
	return c
}

// Unmarshal looking for environment variables and assigns their values
// to corresponding fields of a structure using "env" tags.
//
// The 'cfg' argument must be a non-nil pointer to a structure where values will
// be placed; anything else is reported as ErrInvalidTarget.
// The function recursively traverses the fields of the structure and its nested structures,
// which means structure can contain as much nested structures as you want.
//
// Fields without a key tag, fields tagged `env:"-"` and unexported fields are
// ignored. Nested structures are still traversed when they carry no key tag.
//
// You may use omitempty tags to allow fields to be empty.
// If both the parsed value and the default value are empty, the field will be set to the zero value for its type in Go.
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
//		TimeField 				time.Time	 	`env:"TIME_FIELD"`
//		ByteSliceField 			[]byte 			`env:"BYTE_SLICE_FIELD"`
//		StringSliceField 		[]string		`env:"STRING_SLICE_FIELD"`
//		FloatSliceField 		[]float64		`env:"FLOAT_SLICE_FIELD" envSeparator:";"`
//		EmptyField				string			`env:"EMPTY_FIELD,omitempty"`
//		WithDefaultField		string			`env:"WITH_DEFAULT_FIELD" default:"ave"`
//		IgnoredField			string			`env:"-"`
//		Nested					NestedConfig	`envPrefix:"NESTED_"`
//	}
func (c *ConfigManager) Unmarshal(cfg interface{}) error {
	val, err := structValue(cfg)
	if err != nil {
		return err
	}

	return c.unmarshal(val, "", "")
}

// structValue validates that cfg is a non-nil pointer to a struct and returns
// the struct value it points at.
func structValue(cfg interface{}) (reflect.Value, error) {
	val := reflect.ValueOf(cfg)

	if !val.IsValid() {
		return reflect.Value{}, fmt.Errorf("%w: expected a non-nil pointer to a struct, got nil", ErrInvalidTarget)
	}

	if val.Kind() != reflect.Pointer {
		return reflect.Value{}, fmt.Errorf("%w: expected a pointer to a struct, got %s", ErrInvalidTarget, val.Type())
	}

	if val.IsNil() {
		return reflect.Value{}, fmt.Errorf("%w: expected a non-nil pointer to a struct, got a nil %s", ErrInvalidTarget, val.Type())
	}

	if val.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("%w: expected a pointer to a struct, got %s", ErrInvalidTarget, val.Type())
	}

	return val.Elem(), nil
}

func (c *ConfigManager) unmarshal(val reflect.Value, prefix, path string) error {
	var errs []error

	for i := 0; i < val.NumField(); i++ {
		err := c.unmarshalField(val, i, prefix, path)
		if err == nil {
			continue
		}

		if !c.collectAllErrors {
			return err
		}

		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (c *ConfigManager) unmarshalField(val reflect.Value, i int, prefix, path string) error {
	var (
		structField = val.Type().Field(i)
		field       = val.Field(i)
		meta        = c.fieldMeta(structField, prefix)
		fieldPath   = joinPath(path, structField.Name)
	)

	if meta.Ignored {
		return nil
	}

	// The parser is resolved before descending into a struct so that struct
	// types with a dedicated parser (time.Time, decimal.Decimal, ...) are
	// treated as values instead of groups of nested fields.
	parser, hasParser := c.getParser(field, meta.Separator)

	if !hasParser && field.Kind() == reflect.Struct {
		if err := c.unmarshal(field, prefix+meta.Prefix, fieldPath); err != nil {
			return fmt.Errorf("failed to parse %s: %w", structField.Name, err)
		}
		return nil
	}

	if !meta.HasKey {
		return nil
	}

	var (
		value    string
		hasValue bool
	)

	if !c.forceDefaults {
		value, hasValue = c.lookupValue(meta.Key, meta.OmitEmpty)
	}

	if !hasValue && meta.OmitEmpty && meta.Default == "" {
		return nil
	}

	if !hasValue && !meta.OmitEmpty && (meta.Default == "" || !c.useDefaults) {
		return newFieldError(fieldPath, meta.Key, ErrRequired, "%s cannot be empty", meta.Key)
	}

	if !hasParser {
		return newFieldError(fieldPath, meta.Key, ErrUnsupportedType,
			"failed to get parser for %s: unsupported", meta.Key)
	}

	if (!hasValue && c.useDefaults) || c.forceDefaults {
		if !c.forceDefaults {
			c.logf("WARNING: value for %s not found, using default value: %s", meta.Key, meta.Default)
		}

		value = meta.Default
	}

	// An empty value on a field that allows it: leave the zero value in place
	// rather than handing an empty string to a parser that cannot use it.
	if value == "" && meta.OmitEmpty {
		return nil
	}

	parsed, err := parser(value)
	if err != nil {
		return newFieldError(fieldPath, meta.Key, errors.Join(ErrParse, err),
			"failed to parse %s: %v", meta.Key, err)
	}

	parsedValue := reflect.ValueOf(parsed)
	if !parsedValue.IsValid() || !parsedValue.CanConvert(field.Type()) {
		return newFieldError(fieldPath, meta.Key, ErrInvalidParserResult,
			"failed to parse %s: parser returned %T which cannot be assigned to %s",
			meta.Key, parsed, field.Type())
	}

	// The field is always settable here: fieldMeta marks every unexported field
	// as ignored, and everything else is reached through an addressable struct.
	field.Set(parsedValue.Convert(field.Type()))

	return nil
}

// GenerateDocumentation walks cfg and feeds the resulting documentation tree to
// docGen. The 'cfg' argument must be a non-nil pointer to a struct.
func (c *ConfigManager) GenerateDocumentation(cfg interface{}, docGen DocGenerator) error {
	doc := NewDoc()

	if err := c.parseDocGroup(doc, cfg); err != nil {
		return err
	}

	if err := docGen.GenerateDoc(doc); err != nil {
		return err
	}

	return nil
}

func (c *ConfigManager) parseDocGroup(docGroup *DocTree, cfg interface{}) error {
	val, err := structValue(cfg)
	if err != nil {
		return err
	}

	c.parseDocFields(docGroup, val, "")

	return nil
}

func (c *ConfigManager) parseDocFields(docGroup *DocTree, val reflect.Value, prefix string) {
	for i := 0; i < val.NumField(); i++ {
		var (
			structField = val.Type().Field(i)
			field       = val.Field(i)
			meta        = c.fieldMeta(structField, prefix)
		)

		if meta.Ignored {
			continue
		}

		_, hasParser := c.getParser(field, meta.Separator)

		if !hasParser && field.Kind() == reflect.Struct {
			c.parseDocFields(docGroup.AddGroup(meta.Title), field, prefix+meta.Prefix)
			continue
		}

		if !meta.HasKey {
			continue
		}

		docGroup.AddField(&DocField{
			Key:          meta.Key,
			OmitEmpty:    meta.OmitEmpty,
			Description:  meta.Description,
			DefaultValue: meta.Default,
			ExampleValue: meta.Example,
		})
	}
}

// fieldMeta holds everything gocfg knows about a single struct field, decoded
// from its tags exactly once so that Unmarshal and documentation generation
// cannot drift apart.
type fieldMeta struct {
	// Key is the fully qualified configuration key, prefix included.
	Key string
	// HasKey reports whether the field carries a usable key tag.
	HasKey bool
	// Ignored reports whether gocfg must skip the field entirely, either
	// because it is unexported or because it is tagged `env:"-"`.
	Ignored bool

	OmitEmpty   bool
	Default     string
	Example     string
	Description string
	Title       string
	Prefix      string
	Separator   string
}

func (c *ConfigManager) fieldMeta(structField reflect.StructField, prefix string) fieldMeta {
	meta := fieldMeta{
		Default:     structField.Tag.Get(c.structDefaultTag),
		Example:     structField.Tag.Get(c.structExampleTag),
		Description: structField.Tag.Get(c.structDescriptionTag),
		Title:       structField.Tag.Get(c.structTitleTag),
		Prefix:      structField.Tag.Get(c.structPrefixTag),
		Separator:   structField.Tag.Get(c.structSeparatorTag),
	}

	// Unexported fields cannot be populated through reflection.
	if structField.PkgPath != "" {
		meta.Ignored = true
		return meta
	}

	tag, ok := structField.Tag.Lookup(c.structKeyTag)
	if !ok {
		// No key tag at all: either a nested group or a field gocfg leaves alone.
		return meta
	}

	parts := strings.Split(tag, ",")
	key := strings.TrimSpace(parts[0])

	for _, option := range parts[1:] {
		if strings.TrimSpace(option) == c.structAllowEmptyTag {
			meta.OmitEmpty = true
		}
	}

	if key == "" || key == skipTagValue {
		meta.Ignored = true
		return meta
	}

	meta.Key = prefix + key
	meta.HasKey = true

	return meta
}

// lookupValue retrieves the value for a key from registered value providers and
// reports whether any of them supplied one.
//
// An empty value only counts as a value when the field allows being empty. For
// every other field an empty value is indistinguishable from a missing key, so
// the lookup moves on to the next provider and, failing that, to the default
// tag: an environment variable that expanded to nothing cannot quietly override
// a value the field is not allowed to hold anyway.
func (c *ConfigManager) lookupValue(key string, omitEmpty bool) (string, bool) {
	for _, p := range c.valueProviders {
		if omitEmpty {
			if lookupProvider, ok := p.(LookupValueProvider); ok {
				if value, found := lookupProvider.Lookup(key); found {
					return value, true
				}

				continue
			}
		}

		if value := p.Get(key); value != "" {
			return value, true
		}
	}

	return "", false
}

// getParser retrieves the parser function for a field from registered parser
// providers.
//
// The chain is walked twice: first looking for a provider handling the field
// type as a whole, then for slice fields only looking for one handling its
// element type, in which case the slice parser is composed from it. Whole-type
// parsers therefore always win, which keeps self-decoding slice types such as
// net.IP intact, while []T still works as soon as T does.
func (c *ConfigManager) getParser(field reflect.Value, separator string) (Parser, bool) {
	if parser, ok := c.getProvidedParser(field); ok {
		return parser, true
	}

	if field.Kind() != reflect.Slice {
		return nil, false
	}

	// A scratch element, addressable so that providers needing a pointer
	// receiver encoding.TextUnmarshaler above all can handle it.
	element := reflect.New(field.Type().Elem()).Elem()

	elemParser, ok := c.getProvidedParser(element)
	if !ok {
		return nil, false
	}

	return parsers.NewSliceParser(field.Type(), resettingParser(element, elemParser), separator), true
}

// getProvidedParser asks every registered provider, in priority order, for a
// parser handling the value type as a whole.
func (c *ConfigManager) getProvidedParser(value reflect.Value) (Parser, bool) {
	for _, provider := range c.parserProviders {
		if parser, ok := provider.Get(value); ok {
			return parser, true
		}
	}

	return nil, false
}

// resettingParser zeroes the shared scratch element before every call, so that
// parsers decoding in place cannot leak state from one slice element to the next.
func resettingParser(element reflect.Value, parser Parser) Parser {
	zero := reflect.Zero(element.Type())

	return func(v string) (interface{}, error) {
		element.Set(zero)
		return parser(v)
	}
}

func joinPath(parent, name string) string {
	if parent == "" {
		return name
	}

	return parent + "." + name
}
