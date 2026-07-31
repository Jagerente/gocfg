package parsers

import (
	"encoding"
	"reflect"
)

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

// implementsTextUnmarshaler reports whether a pointer to t can decode itself
// from text.
func implementsTextUnmarshaler(t reflect.Type) bool {
	return reflect.PointerTo(t).Implements(textUnmarshalerType)
}

// TextUnmarshalerParserProvider parses every field whose pointer type implements
// encoding.TextUnmarshaler. That single rule covers time.Time, net.IP,
// netip.Addr, big.Int, uuid.UUID and any user-defined type that follows the same
// standard library convention.
//
// It is registered by NewDefault after DefaultParserProvider, so it only ever
// handles types the built-in parsers do not recognise. Register it first if you
// want UnmarshalText to take precedence over the built-in parsing of named
// scalar types:
//
//	gocfg.NewEmpty().
//		AddParserProviders(
//			parsers.NewTextUnmarshalerParserProvider(),
//			parsers.NewDefaultParserProvider(),
//		)
type TextUnmarshalerParserProvider struct {
}

func NewTextUnmarshalerParserProvider() *TextUnmarshalerParserProvider {
	return &TextUnmarshalerParserProvider{}
}

func (p *TextUnmarshalerParserProvider) Get(value reflect.Value) (Parser, bool) {
	if !value.CanAddr() || !value.CanInterface() {
		return nil, false
	}

	if !implementsTextUnmarshaler(value.Type()) {
		return nil, false
	}

	pointer := value.Addr()

	return func(v string) (interface{}, error) {
		// The assertion cannot fail: the pointer type was checked above.
		unmarshaler := pointer.Interface().(encoding.TextUnmarshaler)

		if err := unmarshaler.UnmarshalText([]byte(v)); err != nil {
			return nil, err
		}

		// UnmarshalText writes through the pointer, so the field already holds
		// the result; returning it keeps the ParserProvider contract intact.
		return value.Interface(), nil
	}, true
}
