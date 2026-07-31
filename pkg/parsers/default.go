package parsers

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Parser converts a raw configuration value into a Go value.
type Parser = func(v string) (interface{}, error)

// DefaultSeparator splits slice values when no envSeparator tag is present.
const DefaultSeparator = ","

var (
	byteSliceType = reflect.TypeOf([]byte(nil))
	durationType  = reflect.TypeOf(time.Duration(0))
)

// byteSliceParser keeps []byte as the raw content of the value: byte slices are
// never split on the separator.
var byteSliceParser Parser = func(v string) (interface{}, error) {
	return []byte(v), nil
}

var (
	defaultTypeParsers = map[reflect.Type]Parser{
		durationType: func(v string) (interface{}, error) {
			return time.ParseDuration(v)
		},
	}

	defaultKindParsers = map[reflect.Kind]Parser{
		reflect.Bool: func(v string) (interface{}, error) {
			return strconv.ParseBool(v)
		},
		reflect.String: func(v string) (interface{}, error) {
			return v, nil
		},
		reflect.Int: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 0)
			return int(i), err
		},
		reflect.Int16: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 16)
			return int16(i), err
		},
		reflect.Int32: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int32(i), err
		},
		reflect.Int64: func(v string) (interface{}, error) {
			return strconv.ParseInt(v, 10, 64)
		},
		reflect.Int8: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 8)
			return int8(i), err
		},
		reflect.Uint: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 0)
			return uint(i), err
		},
		reflect.Uint16: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 16)
			return uint16(i), err
		},
		reflect.Uint32: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint32(i), err
		},
		reflect.Uint64: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 64)
			return i, err
		},
		reflect.Uint8: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 8)
			return uint8(i), err
		},
		reflect.Float64: func(v string) (interface{}, error) {
			return strconv.ParseFloat(v, 64)
		},
		reflect.Float32: func(v string) (interface{}, error) {
			f, err := strconv.ParseFloat(v, 32)
			return float32(f), err
		},
	}
)

type DefaultParserProvider struct {
}

func NewDefaultParserProvider() *DefaultParserProvider {
	return &DefaultParserProvider{}
}

// Get returns a parser for the type of value.
//
// Every slice type but []byte is declined on purpose: gocfg composes slices
// from a parser for their element type, which lets any provider in the chain
// contribute that element parser.
func (p *DefaultParserProvider) Get(value reflect.Value) (Parser, bool) {
	return p.parserFor(value.Type())
}

func (p *DefaultParserProvider) parserFor(t reflect.Type) (Parser, bool) {
	if parser, ok := defaultTypeParsers[t]; ok {
		return parser, true
	}

	if t.Kind() == reflect.Slice {
		// []byte is the raw value and is never split.
		if t == byteSliceType {
			return byteSliceParser, true
		}

		return nil, false
	}

	if parser, ok := defaultKindParsers[t.Kind()]; ok {
		return parser, true
	}

	return nil, false
}

// NewSliceParser builds a parser for sliceType out of a parser for its element
// type. The value is split on separator and every element is trimmed before
// being handed to elemParser.
//
// It is exported so that gocfg can compose slice parsers from element parsers
// coming from any registered provider, not just from this package.
func NewSliceParser(sliceType reflect.Type, elemParser Parser, separator string) Parser {
	if separator == "" {
		separator = DefaultSeparator
	}

	elemType := sliceType.Elem()

	return func(v string) (interface{}, error) {
		parts := strings.Split(v, separator)
		result := reflect.MakeSlice(sliceType, len(parts), len(parts))

		for i, part := range parts {
			parsed, err := elemParser(strings.TrimSpace(part))
			if err != nil {
				return nil, fmt.Errorf("element %d: %w", i, err)
			}

			parsedValue := reflect.ValueOf(parsed)
			if !parsedValue.IsValid() || !parsedValue.CanConvert(elemType) {
				return nil, fmt.Errorf("element %d: parser returned %T which cannot be assigned to %s",
					i, parsed, elemType)
			}

			result.Index(i).Set(parsedValue.Convert(elemType))
		}

		return result.Interface(), nil
	}
}
