package parsers

import (
	"errors"
	"math"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseInto resolves a parser for the type of T and runs it against raw.
func parseInto[T any](t *testing.T, raw string) (interface{}, error) {
	t.Helper()

	var target T

	provider := NewDefaultParserProvider()
	parser, ok := provider.Get(reflect.ValueOf(&target).Elem())
	require.True(t, ok, "no parser for %T", target)

	return parser(raw)
}

func Test_DefaultParserProvider_Scalars(t *testing.T) {
	tests := []struct {
		name     string
		parse    func(t *testing.T, raw string) (interface{}, error)
		raw      string
		expected interface{}
	}{
		{"bool true", parseInto[bool], "true", true},
		{"bool false", parseInto[bool], "false", false},
		{"string", parseInto[string], "  spaced  ", "  spaced  "},
		{"int min", parseInto[int], strconv.Itoa(math.MinInt32), math.MinInt32},
		{"int8 min", parseInto[int8], "-128", int8(-128)},
		{"int8 max", parseInto[int8], "127", int8(127)},
		{"int16", parseInto[int16], "-32768", int16(-32768)},
		{"int32", parseInto[int32], "-2147483648", int32(-2147483648)},
		{"int64 min", parseInto[int64], "-9223372036854775808", int64(math.MinInt64)},
		{"uint8 max", parseInto[uint8], "255", uint8(255)},
		{"uint16 max", parseInto[uint16], "65535", uint16(65535)},
		{"uint32 max", parseInto[uint32], "4294967295", uint32(4294967295)},
		{"uint64 max", parseInto[uint64], "18446744073709551615", uint64(math.MaxUint64)},
		{"float32", parseInto[float32], "3.14", float32(3.14)},
		{"float64", parseInto[float64], "3.14159265359", 3.14159265359},
		{"duration", parseInto[time.Duration], "1h30m", 90 * time.Minute},
		{"byte slice keeps raw content", parseInto[[]byte], "a,b c", []byte("a,b c")},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			actual, err := test.parse(t, test.raw)

			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func Test_PlatformSizedIntegers(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("test is only meaningful on 64-bit platforms")
	}

	parsed, err := parseInto[int](t, strconv.FormatInt(math.MaxInt64, 10))
	require.NoError(t, err)
	assert.Equal(t, math.MaxInt64, parsed)

	parsed, err = parseInto[uint](t, strconv.FormatUint(math.MaxUint64, 10))
	require.NoError(t, err)
	assert.Equal(t, uint(math.MaxUint64), parsed)
}

func Test_DefaultParserProvider_OutOfRange(t *testing.T) {
	tests := []struct {
		name  string
		parse func(t *testing.T, raw string) (interface{}, error)
		raw   string
	}{
		{"int8 overflow", parseInto[int8], "128"},
		{"int8 underflow", parseInto[int8], "-129"},
		{"int16 overflow", parseInto[int16], "32768"},
		{"int32 overflow", parseInto[int32], "2147483648"},
		{"uint8 overflow", parseInto[uint8], "256"},
		{"uint negative", parseInto[uint], "-1"},
		{"bool garbage", parseInto[bool], "yes"},
		{"float garbage", parseInto[float64], "3,14"},
		{"duration garbage", parseInto[time.Duration], "5 seconds"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			_, err := test.parse(t, test.raw)
			assert.Error(t, err)
		})
	}
}

// Test_DefaultParserProvider_Slices covers the only slice type this provider
// handles: []byte. Every other slice is declined so that the caller composes it
// from an element parser taken from the full provider chain.
func Test_DefaultParserProvider_Slices(t *testing.T) {
	t.Run("byte slice keeps the raw value", func(t *testing.T) {
		parsed, err := parseInto[[]byte](t, "a, b ,c")

		require.NoError(t, err)
		assert.Equal(t, []byte("a, b ,c"), parsed)
	})

	t.Run("every other slice type is left to the caller", func(t *testing.T) {
		provider := NewDefaultParserProvider()

		type Tags []string

		var (
			strings   []string
			ints      []int
			floats    []float64
			bools     []bool
			durations []time.Duration
			ips       net.IP
			complexes []complex128
			tags      Tags
		)

		for name, target := range map[string]interface{}{
			"[]string":        &strings,
			"[]int":           &ints,
			"[]float64":       &floats,
			"[]bool":          &bools,
			"[]time.Duration": &durations,
			"net.IP":          &ips,
			"[]complex128":    &complexes,
			"named []string":  &tags,
		} {
			target := target
			t.Run(name, func(t *testing.T) {
				_, ok := provider.Get(reflect.ValueOf(target).Elem())
				assert.False(t, ok)
			})
		}
	})

	t.Run("self-referencing slice type does not recurse", func(t *testing.T) {
		type recursive []recursive

		var target recursive

		assert.NotPanics(t, func() {
			_, ok := NewDefaultParserProvider().Get(reflect.ValueOf(&target).Elem())
			assert.False(t, ok)
		})
	})
}

func Test_NewSliceParser(t *testing.T) {
	var (
		stringSliceType = reflect.TypeOf([]string(nil))
		intSliceType    = reflect.TypeOf([]int(nil))
	)

	elemParser := func(v string) (interface{}, error) {
		return strings.ToUpper(v), nil
	}

	t.Run("splits, trims and converts", func(t *testing.T) {
		type Tag string

		parser := NewSliceParser(reflect.TypeOf([]Tag(nil)), elemParser, ";")

		parsed, err := parser("a; b ;c")
		require.NoError(t, err)
		assert.Equal(t, []Tag{"A", "B", "C"}, parsed)
	})

	t.Run("empty separator falls back to the default", func(t *testing.T) {
		parser := NewSliceParser(stringSliceType, elemParser, "")

		parsed, err := parser("a,b")
		require.NoError(t, err)
		assert.Equal(t, []string{"A", "B"}, parsed)
	})

	t.Run("reports the failing element index", func(t *testing.T) {
		failing := func(v string) (interface{}, error) {
			if v == "bad" {
				return nil, errors.New("nope")
			}
			return v, nil
		}

		parser := NewSliceParser(stringSliceType, failing, DefaultSeparator)

		_, err := parser("ok,bad")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "element 1")
		assert.ErrorContains(t, err, "nope")
	})

	t.Run("rejects an element parser returning the wrong type", func(t *testing.T) {
		wrongType := func(string) (interface{}, error) { return struct{}{}, nil }

		parser := NewSliceParser(intSliceType, wrongType, DefaultSeparator)

		assert.NotPanics(t, func() {
			_, err := parser("1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be assigned to int")
		})
	})

	t.Run("rejects an element parser returning nil", func(t *testing.T) {
		nilResult := func(string) (interface{}, error) { return nil, nil }

		parser := NewSliceParser(intSliceType, nilResult, DefaultSeparator)

		assert.NotPanics(t, func() {
			_, err := parser("1")
			assert.Error(t, err)
		})
	})
}

func Test_DefaultParserProvider_UnsupportedKinds(t *testing.T) {
	provider := NewDefaultParserProvider()

	var (
		complexTarget complex128
		mapTarget     map[string]string
		chanTarget    chan int
		pointerTarget *int
		structTarget  struct{ Field int }
	)

	for name, target := range map[string]interface{}{
		"complex": &complexTarget,
		"map":     &mapTarget,
		"chan":    &chanTarget,
		"pointer": &pointerTarget,
		"struct":  &structTarget,
	} {
		target := target
		t.Run(name, func(t *testing.T) {
			_, ok := provider.Get(reflect.ValueOf(target).Elem())
			assert.False(t, ok)
		})
	}
}
