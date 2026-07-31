package gocfg

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jagerente/gocfg/pkg/parsers"
	"github.com/Jagerente/gocfg/pkg/values"
)

func Test_UnmarshalFromEnv(t *testing.T) {
	type TestConfig struct {
		BoolField bool `env:"BOOL_FIELD"`

		StringField string `env:"STRING_FIELD"`

		IntField   int   `env:"INT_FIELD"`
		Int8Field  int8  `env:"INT8_FIELD"`
		Int16Field int16 `env:"INT16_FIELD"`
		Int32Field int32 `env:"INT32_FIELD"`
		Int64Field int64 `env:"INT64_FIELD"`

		UintField   uint   `env:"UINT_FIELD"`
		Uint8Field  uint8  `env:"UINT8_FIELD"`
		Uint16Field uint16 `env:"UINT16_FIELD"`
		Uint32Field uint32 `env:"UINT32_FIELD"`
		Uint64Field uint64 `env:"UINT64_FIELD"`

		Float32Field float32 `env:"FLOAT32_FIELD"`
		Float64Field float64 `env:"FLOAT64_FIELD"`

		TimeField time.Duration `env:"TIME_FIELD"`

		ByteSliceField   []byte   `env:"BYTE_SLICE_FIELD"`
		StringSliceField []string `env:"STRING_SLICE_FIELD"`
		IntSliceField    []int    `env:"INT_SLICE_FIELD"`

		EmptyField string `env:"EMPTY_FIELD,omitempty"`
	}

	t.Setenv("BOOL_FIELD", "true")

	t.Setenv("STRING_FIELD", "test")

	t.Setenv("INT_FIELD", "-2147483648")
	t.Setenv("INT8_FIELD", "-128")
	t.Setenv("INT16_FIELD", "-32768")
	t.Setenv("INT32_FIELD", "-2147483648")
	t.Setenv("INT64_FIELD", "-9223372036854775808")

	t.Setenv("UINT_FIELD", "4294967295")
	t.Setenv("UINT8_FIELD", "255")
	t.Setenv("UINT16_FIELD", "65535")
	t.Setenv("UINT32_FIELD", "4294967295")
	t.Setenv("UINT64_FIELD", "18446744073709551615")

	t.Setenv("FLOAT32_FIELD", "3.14")
	t.Setenv("FLOAT64_FIELD", "3.14159265359")

	t.Setenv("TIME_FIELD", "5s")
	t.Setenv("BYTE_SLICE_FIELD", "test")
	t.Setenv("STRING_SLICE_FIELD", "test1,test2,test3")
	t.Setenv("INT_SLICE_FIELD", "3,2,1,0,-1,-2,-3")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.NoError(t, err)
	assert.Equal(t, true, cfg.BoolField)
	assert.Equal(t, "test", cfg.StringField)
	assert.Equal(t, -2147483648, cfg.IntField)
	assert.Equal(t, int8(-128), cfg.Int8Field)
	assert.Equal(t, int16(-32768), cfg.Int16Field)
	assert.Equal(t, int32(-2147483648), cfg.Int32Field)
	assert.Equal(t, int64(-9223372036854775808), cfg.Int64Field)
	assert.Equal(t, uint(4294967295), cfg.UintField)
	assert.Equal(t, uint8(255), cfg.Uint8Field)
	assert.Equal(t, uint16(65535), cfg.Uint16Field)
	assert.Equal(t, uint32(4294967295), cfg.Uint32Field)
	assert.Equal(t, uint64(18446744073709551615), cfg.Uint64Field)
	assert.Equal(t, float32(3.14), cfg.Float32Field)
	assert.Equal(t, float64(3.14159265359), cfg.Float64Field)
	assert.Equal(t, 5*time.Second, cfg.TimeField)
	assert.Equal(t, []byte("test"), cfg.ByteSliceField)
	assert.Equal(t, []string{"test1", "test2", "test3"}, cfg.StringSliceField)
	assert.Equal(t, []int{3, 2, 1, 0, -1, -2, -3}, cfg.IntSliceField)
}

// Test_PlatformSizedIntegers guards against int and uint being parsed as 32-bit
// values on 64-bit platforms, which used to reject half of their valid range.
func Test_PlatformSizedIntegers(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("test is only meaningful on 64-bit platforms")
	}

	type TestConfig struct {
		IntField  int  `env:"BIG_INT_FIELD"`
		UintField uint `env:"BIG_UINT_FIELD"`
	}

	t.Setenv("BIG_INT_FIELD", strconv.FormatInt(math.MaxInt64, 10))
	t.Setenv("BIG_UINT_FIELD", strconv.FormatUint(math.MaxUint64, 10))

	cfg := new(TestConfig)
	err := NewDefault().Unmarshal(cfg)

	assert.NoError(t, err)
	assert.Equal(t, math.MaxInt64, cfg.IntField)
	assert.Equal(t, uint(math.MaxUint64), cfg.UintField)
}

func Test_UnmarshalFromDotEnv(t *testing.T) {
	var (
		envContent = `BOOL_FIELD=true
STRING_FIELD=test
INT_FIELD=-2147483648
INT8_FIELD=-128
INT16_FIELD=-32768
INT32_FIELD=-2147483648
INT64_FIELD=-9223372036854775808
UINT_FIELD=4294967295
UINT8_FIELD=255
UINT16_FIELD=65535
UINT32_FIELD=4294967295
UINT64_FIELD=18446744073709551615
FLOAT32_FIELD=3.14
FLOAT64_FIELD=3.14159265359
TIME_FIELD=5s
BYTE_SLICE_FIELD=test
STRING_SLICE_FIELD=test1,test2,test3
INT_SLICE_FIELD=3,2,1,0,-1,-2,-3`
	)

	envFilePath := filepath.Join(t.TempDir(), "test.env")
	require.NoError(t, os.WriteFile(envFilePath, []byte(envContent), 0o600))

	type TestConfig struct {
		BoolField         bool          `env:"BOOL_FIELD"`
		StringField       string        `env:"STRING_FIELD"`
		IntField          int           `env:"INT_FIELD"`
		Int8Field         int8          `env:"INT8_FIELD"`
		Int16Field        int16         `env:"INT16_FIELD"`
		Int32Field        int32         `env:"INT32_FIELD"`
		Int64Field        int64         `env:"INT64_FIELD"`
		UintField         uint          `env:"UINT_FIELD"`
		Uint8Field        uint8         `env:"UINT8_FIELD"`
		Uint16Field       uint16        `env:"UINT16_FIELD"`
		Uint32Field       uint32        `env:"UINT32_FIELD"`
		Uint64Field       uint64        `env:"UINT64_FIELD"`
		Float32Field      float32       `env:"FLOAT32_FIELD"`
		Float64Field      float64       `env:"FLOAT64_FIELD"`
		TimeDurationField time.Duration `env:"TIME_FIELD"`
		ByteSliceField    []byte        `env:"BYTE_SLICE_FIELD"`
		StringSliceField  []string      `env:"STRING_SLICE_FIELD"`
		IntSliceField     []int         `env:"INT_SLICE_FIELD"`
	}

	dotEnvProvider, err := values.NewDotEnvProvider(envFilePath)
	assert.NoError(t, err)

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	cfgManager.AddValueProviders(dotEnvProvider)

	err = cfgManager.Unmarshal(cfg)

	assert.NoError(t, err)
	assert.Equal(t, true, cfg.BoolField)
	assert.Equal(t, "test", cfg.StringField)
	assert.Equal(t, -2147483648, cfg.IntField)
	assert.Equal(t, int8(-128), cfg.Int8Field)
	assert.Equal(t, int16(-32768), cfg.Int16Field)
	assert.Equal(t, int32(-2147483648), cfg.Int32Field)
	assert.Equal(t, int64(-9223372036854775808), cfg.Int64Field)
	assert.Equal(t, uint(4294967295), cfg.UintField)
	assert.Equal(t, uint8(255), cfg.Uint8Field)
	assert.Equal(t, uint16(65535), cfg.Uint16Field)
	assert.Equal(t, uint32(4294967295), cfg.Uint32Field)
	assert.Equal(t, uint64(18446744073709551615), cfg.Uint64Field)
	assert.Equal(t, float32(3.14), cfg.Float32Field)
	assert.Equal(t, float64(3.14159265359), cfg.Float64Field)
	assert.Equal(t, 5*time.Second, cfg.TimeDurationField)
	assert.Equal(t, []byte("test"), cfg.ByteSliceField)
	assert.Equal(t, []string{"test1", "test2", "test3"}, cfg.StringSliceField)
	assert.Equal(t, []int{3, 2, 1, 0, -1, -2, -3}, cfg.IntSliceField)
}

func Test_EmptyField(t *testing.T) {
	type TestConfig struct {
		BoolField   bool   `env:"BOOL_FIELD"`
		StringField string `env:"STRING_FIELD"`
	}

	t.Setenv("BOOL_FIELD", "true")
	t.Setenv("STRING_FIELD", "")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "STRING_FIELD cannot be empty")
	assert.ErrorIs(t, err, ErrRequired)
}

func Test_InvalidType(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		type TestConfig struct {
			IntField int `env:"INT_FIELD"`
		}

		t.Setenv("INT_FIELD", "invalid")

		cfg := new(TestConfig)
		cfgManager := NewDefault()
		err := cfgManager.Unmarshal(cfg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse INT_FIELD")
		assert.ErrorIs(t, err, ErrParse)
		assert.ErrorIs(t, err, strconv.ErrSyntax)
	})

	t.Run("int slice", func(t *testing.T) {
		type TestConfig struct {
			IntField []int `env:"INT_SLICE_FIELD"`
		}

		t.Setenv("INT_SLICE_FIELD", "invalid,1,2")

		cfg := new(TestConfig)
		cfgManager := NewDefault()
		err := cfgManager.Unmarshal(cfg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse INT_SLICE_FIELD")
		assert.ErrorIs(t, err, ErrParse)
	})
}

func Test_StructField(t *testing.T) {
	type TestConfig struct {
		StructField struct {
			NestedField int `env:"NESTED_FIELD"`
		}
	}

	t.Setenv("NESTED_FIELD", "123")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.NoError(t, err)
	assert.Equal(t, 123, cfg.StructField.NestedField)
}

func Test_OmitEmpty(t *testing.T) {
	type TestConfig struct {
		BoolField         bool          `env:"BOOL_FIELD,omitempty"`
		StringField       string        `env:"STRING_FIELD,omitempty"`
		IntField          int           `env:"INT_FIELD,omitempty"`
		Int8Field         int8          `env:"INT8_FIELD,omitempty"`
		Int16Field        int16         `env:"INT16_FIELD,omitempty"`
		Int32Field        int32         `env:"INT32_FIELD,omitempty"`
		Int64Field        int64         `env:"INT64_FIELD,omitempty"`
		UintField         uint          `env:"UINT_FIELD,omitempty"`
		Uint8Field        uint8         `env:"UINT8_FIELD,omitempty"`
		Uint16Field       uint16        `env:"UINT16_FIELD,omitempty"`
		Uint32Field       uint32        `env:"UINT32_FIELD,omitempty"`
		Uint64Field       uint64        `env:"UINT64_FIELD,omitempty"`
		Float32Field      float32       `env:"FLOAT32_FIELD,omitempty"`
		Float64Field      float64       `env:"FLOAT64_FIELD,omitempty"`
		TimeDurationField time.Duration `env:"TIME_FIELD,omitempty"`
		ByteSliceField    []byte        `env:"BYTE_SLICE_FIELD,omitempty"`
		StringSliceField  []string      `env:"STRING_SLICE_FIELD,omitempty"`
		IntSliceField     []int         `env:"INT_SLICE_FIELD,omitempty"`
	}

	t.Setenv("BOOL_FIELD", "")
	t.Setenv("STRING_FIELD", "")
	t.Setenv("INT_FIELD", "")
	t.Setenv("INT8_FIELD", "")
	t.Setenv("INT16_FIELD", "")
	t.Setenv("INT32_FIELD", "")
	t.Setenv("INT64_FIELD", "")
	t.Setenv("UINT_FIELD", "")
	t.Setenv("UINT8_FIELD", "")
	t.Setenv("UINT16_FIELD", "")
	t.Setenv("UINT32_FIELD", "")
	t.Setenv("UINT64_FIELD", "")
	t.Setenv("FLOAT32_FIELD", "")
	t.Setenv("FLOAT64_FIELD", "")
	t.Setenv("TIME_FIELD", "")
	t.Setenv("BYTE_SLICE_FIELD", "")
	t.Setenv("STRING_SLICE_FIELD", "")
	t.Setenv("INT_SLICE_FIELD", "")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.NoError(t, err)
	assert.Equal(t, false, cfg.BoolField)
	assert.Equal(t, "", cfg.StringField)
	assert.Equal(t, 0, cfg.IntField)
	assert.Equal(t, int8(0), cfg.Int8Field)
	assert.Equal(t, int16(0), cfg.Int16Field)
	assert.Equal(t, int32(0), cfg.Int32Field)
	assert.Equal(t, int64(0), cfg.Int64Field)
	assert.Equal(t, uint(0), cfg.UintField)
	assert.Equal(t, uint8(0), cfg.Uint8Field)
	assert.Equal(t, uint16(0), cfg.Uint16Field)
	assert.Equal(t, uint32(0), cfg.Uint32Field)
	assert.Equal(t, uint64(0), cfg.Uint64Field)
	assert.Equal(t, float32(0), cfg.Float32Field)
	assert.Equal(t, float64(0), cfg.Float64Field)
	assert.Equal(t, time.Duration(0), cfg.TimeDurationField)
	assert.Equal(t, []byte(nil), cfg.ByteSliceField)
	assert.Equal(t, []string(nil), cfg.StringSliceField)
	assert.Equal(t, []int(nil), cfg.IntSliceField)
}

func Test_OmitEmptyIsAnExactOption(t *testing.T) {
	type TestConfig struct {
		Field string `env:"KEY_omitempty_SUFFIX"`
	}

	require.NoError(t, os.Unsetenv("KEY_omitempty_SUFFIX"))

	err := NewDefault().Unmarshal(new(TestConfig))

	assert.ErrorIs(t, err, ErrRequired)
	assert.Contains(t, err.Error(), "KEY_omitempty_SUFFIX cannot be empty")
}

func Test_DefaultValues(t *testing.T) {
	type TestConfig struct {
		BoolField         bool          `env:"BOOL_FIELD" default:"true"`
		StringField       string        `env:"STRING_FIELD" default:"default"`
		IntField          int           `env:"INT_FIELD" default:"42"`
		Float64Field      float64       `env:"FLOAT64_FIELD" default:"3.14"`
		TimeDurationField time.Duration `env:"TIME_DURATION_FIELD" default:"1h"`
	}

	t.Setenv("BOOL_FIELD", "")
	t.Setenv("STRING_FIELD", "")
	t.Setenv("INT_FIELD", "")
	t.Setenv("FLOAT64_FIELD", "")
	t.Setenv("TIME_DURATION_FIELD", "")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.NoError(t, err)
	assert.Equal(t, true, cfg.BoolField)
	assert.Equal(t, "default", cfg.StringField)
	assert.Equal(t, 42, cfg.IntField)
	assert.Equal(t, 3.14, cfg.Float64Field)
	assert.Equal(t, time.Hour, cfg.TimeDurationField)
}

func Test_ForceDefaults(t *testing.T) {
	type TestConfig struct {
		BoolField         bool          `env:"BOOL_FIELD" default:"true"`
		StringField       string        `env:"STRING_FIELD" default:"default"`
		IntField          int           `env:"INT_FIELD" default:"42"`
		Float64Field      float64       `env:"FLOAT64_FIELD" default:"3.14"`
		TimeDurationField time.Duration `env:"TIME_DURATION_FIELD" default:"1h"`
	}

	t.Setenv("BOOL_FIELD", "false")
	t.Setenv("STRING_FIELD", "not default")
	t.Setenv("INT_FIELD", "83")
	t.Setenv("FLOAT64_FIELD", "8.3")
	t.Setenv("TIME_DURATION_FIELD", "5s")

	cfg := new(TestConfig)
	cfgManager := NewDefault().ForceDefaults()
	err := cfgManager.Unmarshal(cfg)

	assert.NoError(t, err)
	assert.Equal(t, true, cfg.BoolField)
	assert.Equal(t, "default", cfg.StringField)
	assert.Equal(t, 42, cfg.IntField)
	assert.Equal(t, 3.14, cfg.Float64Field)
	assert.Equal(t, time.Hour, cfg.TimeDurationField)
}

func Test_UseCustomKeyTag(t *testing.T) {
	type TestConfig struct {
		BoolField   bool   `mapstructure:"BOOL_FIELD"`
		StringField string `mapstructure:"STRING_FIELD"`
		IntField    int    `mapstructure:"INT_FIELD"`
	}

	t.Setenv("BOOL_FIELD", "true")
	t.Setenv("STRING_FIELD", "value")
	t.Setenv("INT_FIELD", "83")

	cfg := new(TestConfig)
	cfgManager := NewDefault().
		UseCustomKeyTag("mapstructure")
	assert.Equal(t, cfgManager.structKeyTag, "mapstructure")

	err := cfgManager.Unmarshal(cfg)

	assert.NoError(t, err)
	assert.Equal(t, true, cfg.BoolField)
	assert.Equal(t, "value", cfg.StringField)
	assert.Equal(t, 83, cfg.IntField)
}

func Test_UnmarshalErrorInNestedStruct(t *testing.T) {
	type TestConfig struct {
		StructField struct {
			NestedField int `env:"NESTED_FIELD"`
		}
	}

	t.Setenv("NESTED_FIELD", "invalid")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse StructField")

	var fieldErr *FieldError
	require.ErrorAs(t, err, &fieldErr)
	assert.Equal(t, "StructField.NestedField", fieldErr.Path)
	assert.Equal(t, "NESTED_FIELD", fieldErr.Key)
}

func Test_GetParserUnsupportedField(t *testing.T) {
	type TestConfig struct {
		UnsupportedField complex128 `env:"UNSUPPORTED_FIELD"`
	}

	t.Setenv("UNSUPPORTED_FIELD", "value")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get parser for UNSUPPORTED_FIELD: unsupported")
	assert.ErrorIs(t, err, ErrUnsupportedType)
}

func Test_FieldsWithoutKeyAreIgnored(t *testing.T) {
	type TestConfig struct {
		Tagged   string `env:"TAGGED_FIELD"`
		Untagged string
		Skipped  string `env:"-"`
		Empty    string `env:""`
	}

	t.Setenv("TAGGED_FIELD", "value")

	cfg := &TestConfig{Untagged: "untouched", Skipped: "untouched", Empty: "untouched"}
	err := NewDefault().Unmarshal(cfg)

	assert.NoError(t, err)
	assert.Equal(t, "value", cfg.Tagged)
	assert.Equal(t, "untouched", cfg.Untagged)
	assert.Equal(t, "untouched", cfg.Skipped)
	assert.Equal(t, "untouched", cfg.Empty)
}

func Test_UnexportedFieldsAreIgnored(t *testing.T) {
	type nested struct {
		Field string `env:"UNEXPORTED_NESTED_FIELD"`
	}

	type TestConfig struct {
		Exported string `env:"EXPORTED_FIELD"`
		//nolint:unused
		unexported string `env:"EXPORTED_FIELD"`
		//nolint:unused
		unexportedNested nested
	}

	t.Setenv("EXPORTED_FIELD", "value")

	cfg := new(TestConfig)

	assert.NotPanics(t, func() {
		assert.NoError(t, NewDefault().Unmarshal(cfg))
	})
	assert.Equal(t, "value", cfg.Exported)
}

func Test_SkippedNestedStructIsNotTraversed(t *testing.T) {
	type Nested struct {
		Field string `env:"MUST_NOT_BE_READ"`
	}

	type TestConfig struct {
		Nested Nested `env:"-"`
	}

	require.NoError(t, os.Unsetenv("MUST_NOT_BE_READ"))

	cfg := new(TestConfig)

	assert.NoError(t, NewDefault().Unmarshal(cfg))
	assert.Equal(t, "", cfg.Nested.Field)
}

func Test_InvalidTarget(t *testing.T) {
	type TestConfig struct {
		Field string `env:"FIELD"`
	}

	var nilPointer *TestConfig
	notAStruct := "hello"

	tests := map[string]interface{}{
		"untyped nil":         nil,
		"non-pointer struct":  TestConfig{},
		"nil typed pointer":   nilPointer,
		"pointer to string":   &notAStruct,
		"non-pointer string":  notAStruct,
		"pointer to pointer":  &nilPointer,
		"nil map as a target": map[string]string(nil),
	}

	for name, target := range tests {
		target := target
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				err := NewDefault().Unmarshal(target)
				assert.ErrorIs(t, err, ErrInvalidTarget)
			})

			assert.NotPanics(t, func() {
				err := NewEmpty().GenerateDocumentation(target, &MockDocGenerator{})
				assert.ErrorIs(t, err, ErrInvalidTarget)
			})
		})
	}
}

type structParserProvider struct{}

type Point struct {
	X, Y int
}

func (structParserProvider) Get(value reflect.Value) (Parser, bool) {
	if value.Type() != reflect.TypeOf(Point{}) {
		return nil, false
	}

	return func(v string) (interface{}, error) {
		var p Point
		if _, err := fmt.Sscanf(v, "%d:%d", &p.X, &p.Y); err != nil {
			return nil, err
		}
		return p, nil
	}, true
}

func Test_CustomParserForStructType(t *testing.T) {
	type TestConfig struct {
		Origin Point `env:"ORIGIN"`
	}

	t.Setenv("ORIGIN", "3:4")

	cfg := new(TestConfig)
	err := NewDefault().AddParserProviders(structParserProvider{}).Unmarshal(cfg)

	assert.NoError(t, err)
	assert.Equal(t, Point{X: 3, Y: 4}, cfg.Origin)
}

func Test_CustomParserForSliceOfCustomType(t *testing.T) {
	type TestConfig struct {
		Single Point   `env:"SINGLE_POINT"`
		Many   []Point `env:"MANY_POINTS"`
	}

	t.Setenv("SINGLE_POINT", "3:4")
	t.Setenv("MANY_POINTS", "3:4, 5:6")

	cfg := new(TestConfig)
	err := NewDefault().AddParserProviders(structParserProvider{}).Unmarshal(cfg)

	require.NoError(t, err)
	assert.Equal(t, Point{X: 3, Y: 4}, cfg.Single)
	assert.Equal(t, []Point{{X: 3, Y: 4}, {X: 5, Y: 6}}, cfg.Many)
}

func Test_SliceOfTextUnmarshalerType(t *testing.T) {
	type TestConfig struct {
		Times []time.Time `env:"MANY_TIMES"`
		IPs   []net.IP    `env:"MANY_IPS" envSeparator:";"`
	}

	t.Setenv("MANY_TIMES", "2024-01-02T03:04:05Z,2025-01-02T03:04:05Z")
	t.Setenv("MANY_IPS", "10.0.0.1;10.0.0.2")

	cfg := new(TestConfig)
	err := NewDefault().Unmarshal(cfg)

	require.NoError(t, err)
	require.Len(t, cfg.Times, 2)
	assert.Equal(t, time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), cfg.Times[0].UTC())
	assert.Equal(t, time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC), cfg.Times[1].UTC())
	require.Len(t, cfg.IPs, 2)
	assert.Equal(t, "10.0.0.1", cfg.IPs[0].String())
	assert.Equal(t, "10.0.0.2", cfg.IPs[1].String())
}

type Level int

type levelParserProvider struct{}

func (levelParserProvider) Get(value reflect.Value) (Parser, bool) {
	if value.Type() != reflect.TypeOf(Level(0)) {
		return nil, false
	}

	return func(v string) (interface{}, error) {
		switch v {
		case "debug":
			return Level(1), nil
		case "info":
			return Level(2), nil
		}
		return nil, fmt.Errorf("unknown level %q", v)
	}, true
}

func Test_ParserProviderPriorityForNamedTypes(t *testing.T) {
	type TestConfig struct {
		Single Level   `env:"SINGLE_LEVEL"`
		Many   []Level `env:"MANY_LEVELS"`
	}

	t.Setenv("SINGLE_LEVEL", "info")
	t.Setenv("MANY_LEVELS", "info,debug")

	t.Run("registered after the default provider: shadowed", func(t *testing.T) {
		err := NewDefault().AddParserProviders(levelParserProvider{}).Unmarshal(new(TestConfig))

		assert.ErrorIs(t, err, ErrParse, "the built-in int parser wins, as documented")
	})

	t.Run("registered before the default provider: used", func(t *testing.T) {
		cfg := new(TestConfig)
		err := NewEmpty().
			UseDefaults().
			AddParserProviders(
				levelParserProvider{},
				parsers.NewDefaultParserProvider(),
				parsers.NewTextUnmarshalerParserProvider(),
			).
			AddValueProviders(values.NewEnvProvider()).
			Unmarshal(cfg)

		require.NoError(t, err)
		assert.Equal(t, Level(2), cfg.Single)
		assert.Equal(t, []Level{2, 1}, cfg.Many)
	})
}

// accumulator decodes in place and appends, which exposes state leaking from
// one slice element to the next if the scratch element is not reset.
type accumulator struct {
	Seen []string
}

func (a *accumulator) UnmarshalText(text []byte) error {
	a.Seen = append(a.Seen, string(text))
	return nil
}

func Test_SliceElementsDoNotShareState(t *testing.T) {
	type TestConfig struct {
		Items []accumulator `env:"ACCUMULATED"`
	}

	t.Setenv("ACCUMULATED", "a,b,c")

	cfg := new(TestConfig)
	require.NoError(t, NewDefault().Unmarshal(cfg))

	require.Len(t, cfg.Items, 3)
	assert.Equal(t, []string{"a"}, cfg.Items[0].Seen)
	assert.Equal(t, []string{"b"}, cfg.Items[1].Seen)
	assert.Equal(t, []string{"c"}, cfg.Items[2].Seen)
}

type brokenElementParserProvider struct{}

func (brokenElementParserProvider) Get(value reflect.Value) (Parser, bool) {
	if value.Type() != reflect.TypeOf(Point{}) {
		return nil, false
	}

	return func(string) (interface{}, error) { return "not a Point", nil }, true
}

func Test_ComposedSliceParserRejectsWrongElementType(t *testing.T) {
	type TestConfig struct {
		Many []Point `env:"BROKEN_POINTS"`
	}

	t.Setenv("BROKEN_POINTS", "1:2")

	assert.NotPanics(t, func() {
		err := NewDefault().AddParserProviders(brokenElementParserProvider{}).Unmarshal(new(TestConfig))

		assert.ErrorIs(t, err, ErrParse)
		assert.Contains(t, err.Error(), "cannot be assigned to")
	})
}

type brokenParserProvider struct {
	result interface{}
}

func (p brokenParserProvider) Get(value reflect.Value) (Parser, bool) {
	if value.Kind() != reflect.Int {
		return nil, false
	}

	return func(string) (interface{}, error) { return p.result, nil }, true
}

func Test_MisbehavingParser(t *testing.T) {
	type TestConfig struct {
		Field int `env:"BROKEN_FIELD"`
	}

	t.Setenv("BROKEN_FIELD", "1")

	for name, result := range map[string]interface{}{
		"wrong type": "not an int",
		"nil":        nil,
	} {
		result := result
		t.Run(name, func(t *testing.T) {
			cfg := new(TestConfig)

			assert.NotPanics(t, func() {
				err := NewEmpty().
					AddParserProviders(brokenParserProvider{result: result}).
					AddValueProviders(values.NewEnvProvider()).
					Unmarshal(cfg)

				assert.ErrorIs(t, err, ErrInvalidParserResult)
			})
		})
	}
}

func Test_TextUnmarshalerTypes(t *testing.T) {
	type TestConfig struct {
		Time    time.Time `env:"TIME_VALUE"`
		IP      net.IP    `env:"IP_VALUE"`
		Big     big.Int   `env:"BIG_VALUE"`
		Ignored time.Time `env:"MISSING_TIME,omitempty"`
	}

	t.Setenv("TIME_VALUE", "2024-01-02T03:04:05Z")
	t.Setenv("IP_VALUE", "192.168.1.10")
	t.Setenv("BIG_VALUE", "123456789012345678901234567890")
	require.NoError(t, os.Unsetenv("MISSING_TIME"))

	cfg := new(TestConfig)
	err := NewDefault().Unmarshal(cfg)

	require.NoError(t, err)
	assert.Equal(t, time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), cfg.Time.UTC())
	assert.Equal(t, "192.168.1.10", cfg.IP.String())
	assert.Equal(t, "123456789012345678901234567890", cfg.Big.String())
	assert.True(t, cfg.Ignored.IsZero())
}

func Test_TextUnmarshalerError(t *testing.T) {
	type TestConfig struct {
		Time time.Time `env:"TIME_VALUE"`
	}

	t.Setenv("TIME_VALUE", "not a timestamp")

	err := NewDefault().Unmarshal(new(TestConfig))

	assert.ErrorIs(t, err, ErrParse)
	assert.Contains(t, err.Error(), "failed to parse TIME_VALUE")
}

func Test_GenericSliceParsing(t *testing.T) {
	type TestConfig struct {
		Floats    []float64       `env:"FLOAT_SLICE"`
		Bools     []bool          `env:"BOOL_SLICE"`
		Durations []time.Duration `env:"DURATION_SLICE"`
		Int8s     []int8          `env:"INT8_SLICE"`
	}

	t.Setenv("FLOAT_SLICE", "1.5, 2.5,3")
	t.Setenv("BOOL_SLICE", "true,false, true")
	t.Setenv("DURATION_SLICE", "1s, 2m")
	t.Setenv("INT8_SLICE", "-128,127")

	cfg := new(TestConfig)
	err := NewDefault().Unmarshal(cfg)

	require.NoError(t, err)
	assert.Equal(t, []float64{1.5, 2.5, 3}, cfg.Floats)
	assert.Equal(t, []bool{true, false, true}, cfg.Bools)
	assert.Equal(t, []time.Duration{time.Second, 2 * time.Minute}, cfg.Durations)
	assert.Equal(t, []int8{-128, 127}, cfg.Int8s)
}

func Test_CustomSeparator(t *testing.T) {
	type TestConfig struct {
		Hosts   []string `env:"HOSTS" envSeparator:";"`
		Ports   []int    `env:"PORTS" envSeparator:"|"`
		Default []string `env:"DEFAULTS"`
		// []byte is the raw value and must ignore the separator entirely.
		Raw []byte `env:"RAW" envSeparator:";"`
	}

	t.Setenv("HOSTS", "a.example.com;b.example.com")
	t.Setenv("PORTS", "80|443")
	t.Setenv("DEFAULTS", "x,y")
	t.Setenv("RAW", "a;b,c")

	cfg := new(TestConfig)
	err := NewDefault().Unmarshal(cfg)

	require.NoError(t, err)
	assert.Equal(t, []string{"a.example.com", "b.example.com"}, cfg.Hosts)
	assert.Equal(t, []int{80, 443}, cfg.Ports)
	assert.Equal(t, []string{"x", "y"}, cfg.Default)
	assert.Equal(t, []byte("a;b,c"), cfg.Raw)
}

func Test_UnsupportedSliceElement(t *testing.T) {
	type TestConfig struct {
		Field []complex128 `env:"UNSUPPORTED_SLICE"`
	}

	t.Setenv("UNSUPPORTED_SLICE", "1,2")

	err := NewDefault().Unmarshal(new(TestConfig))

	assert.ErrorIs(t, err, ErrUnsupportedType)
	assert.Contains(t, err.Error(), "failed to get parser for UNSUPPORTED_SLICE: unsupported")
}

// Test_SliceElementsAreTrimmed pins the uniform behaviour of every slice type:
// elements are trimmed, so a value laid out for readability parses as intended.
func Test_SliceElementsAreTrimmed(t *testing.T) {
	type TestConfig struct {
		Strings []string `env:"TRIMMED_STRING_SLICE"`
		Ints    []int    `env:"TRIMMED_INT_SLICE"`
		// []byte is the raw value: never split, never trimmed.
		Raw []byte `env:"RAW_BYTES"`
	}

	t.Setenv("TRIMMED_STRING_SLICE", "a, b ,c")
	t.Setenv("TRIMMED_INT_SLICE", " 1, 2 ,3")
	t.Setenv("RAW_BYTES", "a, b ,c")

	cfg := new(TestConfig)
	require.NoError(t, NewDefault().Unmarshal(cfg))

	assert.Equal(t, []string{"a", "b", "c"}, cfg.Strings)
	assert.Equal(t, []int{1, 2, 3}, cfg.Ints)
	assert.Equal(t, []byte("a, b ,c"), cfg.Raw)
}

func Test_EnvPrefix(t *testing.T) {
	type Credentials struct {
		User     string `env:"USER"`
		Password string `env:"PASSWORD"`
	}

	type DBConfig struct {
		Host        string `env:"HOST"`
		Credentials Credentials
	}

	type TestConfig struct {
		Primary DBConfig `envPrefix:"PRIMARY_"`
		Replica DBConfig `envPrefix:"REPLICA_"`
		Plain   string   `env:"PLAIN"`
	}

	t.Setenv("PRIMARY_HOST", "primary.example.com")
	t.Setenv("PRIMARY_USER", "primary-user")
	t.Setenv("PRIMARY_PASSWORD", "primary-pass")
	t.Setenv("REPLICA_HOST", "replica.example.com")
	t.Setenv("REPLICA_USER", "replica-user")
	t.Setenv("REPLICA_PASSWORD", "replica-pass")
	t.Setenv("PLAIN", "plain")

	cfg := new(TestConfig)
	err := NewDefault().Unmarshal(cfg)

	require.NoError(t, err)
	assert.Equal(t, "primary.example.com", cfg.Primary.Host)
	assert.Equal(t, "primary-user", cfg.Primary.Credentials.User)
	assert.Equal(t, "replica.example.com", cfg.Replica.Host)
	assert.Equal(t, "replica-pass", cfg.Replica.Credentials.Password)
	assert.Equal(t, "plain", cfg.Plain)
}

func Test_EnvPrefixNesting(t *testing.T) {
	type Inner struct {
		Field string `env:"FIELD"`
	}

	type Middle struct {
		Inner Inner `envPrefix:"INNER_"`
	}

	type TestConfig struct {
		Outer Middle `envPrefix:"OUTER_"`
	}

	t.Setenv("OUTER_INNER_FIELD", "nested")

	cfg := new(TestConfig)

	require.NoError(t, NewDefault().Unmarshal(cfg))
	assert.Equal(t, "nested", cfg.Outer.Inner.Field)
}

// Test_MissingKeyVsEmptyValue pins the whole matrix of the rule an empty value
// only counts as a value for a field marked omitempty. Everywhere else an empty
// value is indistinguishable from a missing key.
func Test_MissingKeyVsEmptyValue(t *testing.T) {
	const key = "MATRIX_FIELD"

	tests := []struct {
		name string
		// unmarshal runs the matching struct shape and returns a printable result.
		unmarshal func(t *testing.T) (interface{}, error)
		whenUnset interface{}
		whenEmpty interface{}
		// errUnset and errEmpty name the expected sentinel, nil when none.
		errUnset error
		errEmpty error
	}{
		{
			name: `env:"KEY"`,
			unmarshal: func(t *testing.T) (interface{}, error) {
				type C struct {
					F string `env:"MATRIX_FIELD"`
				}
				cfg := new(C)
				return cfg.F, NewDefault().Unmarshal(cfg)
			},
			errUnset: ErrRequired,
			errEmpty: ErrRequired,
		},
		{
			name: `env:"KEY" default:"x"`,
			unmarshal: func(t *testing.T) (interface{}, error) {
				type C struct {
					F string `env:"MATRIX_FIELD" default:"x"`
				}
				cfg := new(C)
				return cfg.F, NewDefault().Unmarshal(cfg)
			},
			whenUnset: "x",
			whenEmpty: "x",
		},
		{
			name: `env:"KEY,omitempty"`,
			unmarshal: func(t *testing.T) (interface{}, error) {
				type C struct {
					F string `env:"MATRIX_FIELD,omitempty"`
				}
				cfg := new(C)
				return cfg.F, NewDefault().Unmarshal(cfg)
			},
			whenUnset: "",
			whenEmpty: "",
		},
		{
			name: `env:"KEY,omitempty" default:"x"`,
			unmarshal: func(t *testing.T) (interface{}, error) {
				type C struct {
					F string `env:"MATRIX_FIELD,omitempty" default:"x"`
				}
				cfg := new(C)
				return cfg.F, NewDefault().Unmarshal(cfg)
			},
			whenUnset: "x",
			whenEmpty: "",
		},
		{
			name: `env:"KEY" default:"8080" on an int`,
			unmarshal: func(t *testing.T) (interface{}, error) {
				type C struct {
					F int `env:"MATRIX_FIELD" default:"8080"`
				}
				cfg := new(C)
				return cfg.F, NewDefault().Unmarshal(cfg)
			},
			whenUnset: 8080,
			whenEmpty: 8080,
		},
		{
			name: `env:"KEY,omitempty" default:"8080" on an int`,
			unmarshal: func(t *testing.T) (interface{}, error) {
				type C struct {
					F int `env:"MATRIX_FIELD,omitempty" default:"8080"`
				}
				cfg := new(C)
				return cfg.F, NewDefault().Unmarshal(cfg)
			},
			whenUnset: 8080,
			whenEmpty: 0,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Run("key unset", func(t *testing.T) {
				require.NoError(t, os.Unsetenv(key))

				actual, err := test.unmarshal(t)

				if test.errUnset != nil {
					assert.ErrorIs(t, err, test.errUnset)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, test.whenUnset, actual)
			})

			t.Run("key set to an empty value", func(t *testing.T) {
				t.Setenv(key, "")

				actual, err := test.unmarshal(t)

				if test.errEmpty != nil {
					assert.ErrorIs(t, err, test.errEmpty)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, test.whenEmpty, actual)
			})
		})
	}
}

type staticProvider struct {
	key   string
	value string
}

func (p *staticProvider) Get(key string) string {
	if key == p.key {
		return p.value
	}
	return ""
}

type lookupProvider struct {
	values map[string]string
}

func (p *lookupProvider) Get(key string) string {
	return p.values[key]
}

func (p *lookupProvider) Lookup(key string) (string, bool) {
	value, ok := p.values[key]
	return value, ok
}

func Test_ProviderChain(t *testing.T) {
	// An optional field asks providers whether they know the key at all, so the
	// first provider that has it ends the lookup even with an empty value.
	t.Run("optional field: an explicitly empty value ends the lookup", func(t *testing.T) {
		type TestConfig struct {
			Field string `env:"PRIORITY_FIELD,omitempty"`
		}

		t.Setenv("PRIORITY_FIELD", "")

		cfg := new(TestConfig)
		err := NewDefault().
			AddValueProviders(&staticProvider{key: "PRIORITY_FIELD", value: "from-fallback"}).
			Unmarshal(cfg)

		require.NoError(t, err)
		assert.Equal(t, "", cfg.Field)
	})

	t.Run("optional field: an unknown key moves on to the next provider", func(t *testing.T) {
		type TestConfig struct {
			Field string `env:"FALLTHROUGH_FIELD,omitempty"`
		}

		require.NoError(t, os.Unsetenv("FALLTHROUGH_FIELD"))

		cfg := new(TestConfig)
		err := NewEmpty().
			AddParserProviders(parsers.NewDefaultParserProvider()).
			AddValueProviders(
				&lookupProvider{values: map[string]string{"OTHER_FIELD": "ignored"}},
				&lookupProvider{values: map[string]string{"FALLTHROUGH_FIELD": "found"}},
			).
			Unmarshal(cfg)

		require.NoError(t, err)
		assert.Equal(t, "found", cfg.Field)
	})

	// A field that cannot be empty treats an empty value as no value at all, so
	// the lookup keeps going regardless of what the provider reports.
	t.Run("required field: empty values are skipped", func(t *testing.T) {
		type TestConfig struct {
			Field string `env:"PLAIN_PROVIDER_FIELD"`
		}

		require.NoError(t, os.Unsetenv("PLAIN_PROVIDER_FIELD"))

		for name, providers := range map[string][]ValueProvider{
			"provider exposing only Get": {
				&staticProvider{key: "PLAIN_PROVIDER_FIELD", value: ""},
				&staticProvider{key: "PLAIN_PROVIDER_FIELD", value: "from-fallback"},
			},
			"provider exposing Lookup": {
				&lookupProvider{values: map[string]string{"PLAIN_PROVIDER_FIELD": ""}},
				&lookupProvider{values: map[string]string{"PLAIN_PROVIDER_FIELD": "from-fallback"}},
			},
		} {
			providers := providers
			t.Run(name, func(t *testing.T) {
				cfg := new(TestConfig)
				err := NewEmpty().
					AddParserProviders(parsers.NewDefaultParserProvider()).
					AddValueProviders(providers...).
					Unmarshal(cfg)

				require.NoError(t, err)
				assert.Equal(t, "from-fallback", cfg.Field)
			})
		}
	})
}

func Test_CollectAllErrors(t *testing.T) {
	type Nested struct {
		Third string `env:"THIRD_MISSING"`
	}

	type TestConfig struct {
		First  string `env:"FIRST_MISSING"`
		Second int    `env:"SECOND_INVALID"`
		Nested Nested
	}

	require.NoError(t, os.Unsetenv("FIRST_MISSING"))
	require.NoError(t, os.Unsetenv("THIRD_MISSING"))
	t.Setenv("SECOND_INVALID", "not a number")

	t.Run("disabled: stops at the first error", func(t *testing.T) {
		err := NewDefault().Unmarshal(new(TestConfig))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "FIRST_MISSING")
		assert.NotContains(t, err.Error(), "SECOND_INVALID")
	})

	t.Run("enabled: reports every field", func(t *testing.T) {
		err := NewDefault().CollectAllErrors().Unmarshal(new(TestConfig))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "FIRST_MISSING")
		assert.Contains(t, err.Error(), "SECOND_INVALID")
		assert.Contains(t, err.Error(), "THIRD_MISSING")
		assert.ErrorIs(t, err, ErrRequired)
		assert.ErrorIs(t, err, ErrParse)
	})
}

type recordingLogger struct {
	messages []string
}

func (l *recordingLogger) Printf(format string, args ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(format, args...))
}

func Test_LoggerIsSilentByDefault(t *testing.T) {
	type TestConfig struct {
		Field string `env:"LOGGED_FIELD" default:"value"`
	}

	require.NoError(t, os.Unsetenv("LOGGED_FIELD"))

	cfgManager := NewDefault()
	require.Nil(t, cfgManager.logger, "no logger must be configured by default")

	assert.NotPanics(t, func() {
		assert.NoError(t, cfgManager.Unmarshal(new(TestConfig)))
	})
}

func Test_WithLogger(t *testing.T) {
	type TestConfig struct {
		Field    string `env:"LOGGED_FIELD" default:"value"`
		Provided string `env:"PROVIDED_FIELD" default:"unused"`
	}

	require.NoError(t, os.Unsetenv("LOGGED_FIELD"))
	t.Setenv("PROVIDED_FIELD", "provided")

	logger := &recordingLogger{}

	cfg := new(TestConfig)
	require.NoError(t, NewDefault().WithLogger(logger).Unmarshal(cfg))

	assert.Equal(t, "value", cfg.Field)
	assert.Equal(t, "provided", cfg.Provided)

	require.Len(t, logger.messages, 1, "only the field falling back to its default is reported")
	assert.Contains(t, logger.messages[0], "LOGGED_FIELD")
	assert.Contains(t, logger.messages[0], "value")
}

func Test_LoggerFunc(t *testing.T) {
	type TestConfig struct {
		Field string `env:"LOGGED_FIELD" default:"value"`
	}

	require.NoError(t, os.Unsetenv("LOGGED_FIELD"))

	var captured []string
	logger := LoggerFunc(func(format string, args ...interface{}) {
		captured = append(captured, fmt.Sprintf(format, args...))
	})

	require.NoError(t, NewDefault().WithLogger(logger).Unmarshal(new(TestConfig)))
	assert.Len(t, captured, 1)
}

func Test_ForceDefaultsDoesNotLog(t *testing.T) {
	type TestConfig struct {
		Field string `env:"LOGGED_FIELD" default:"value"`
	}

	t.Setenv("LOGGED_FIELD", "provided")

	logger := &recordingLogger{}
	require.NoError(t, NewDefault().ForceDefaults().WithLogger(logger).Unmarshal(new(TestConfig)))

	assert.Empty(t, logger.messages)
}

type MockDocGenerator struct {
	GeneratedDoc *DocTree
	WithErr      bool
}

func (m *MockDocGenerator) GenerateDoc(doc *DocTree) error {
	if m.WithErr {
		return errors.New("failed to generate doc")
	}

	m.GeneratedDoc = doc
	return nil
}

func Test_GenerateDocumentation(t *testing.T) {
	type Nested struct {
		BoolField bool `env:"NESTED_BOOL_FIELD" description:"Description for Nested BoolField"`
	}

	type TestConfig struct {
		StringField   string `env:"STRING_FIELD" description:"Description for StringField"`
		IntField      int    `env:"INT_FIELD" description:"Description for IntField"`
		NestedStruct  Nested `title:"Nested Struct Config"`
		WithoutEnvTag string `description:"Description for WithoutEnvTag"`
	}

	cfg := new(TestConfig)
	mockDocGenerator := &MockDocGenerator{}

	cfgManager := NewEmpty()
	err := cfgManager.GenerateDocumentation(cfg, mockDocGenerator)

	assert.NoError(t, err)
	assert.NotNil(t, mockDocGenerator.GeneratedDoc)
	assert.Equal(t, "", mockDocGenerator.GeneratedDoc.Title)
	assert.Len(t, mockDocGenerator.GeneratedDoc.Fields, 2, "fields without a key tag are not documented")
	assert.Len(t, mockDocGenerator.GeneratedDoc.Groups, 1)
	assert.Equal(t, "STRING_FIELD", mockDocGenerator.GeneratedDoc.Fields[0].Key)
	assert.Equal(t, "Description for StringField", mockDocGenerator.GeneratedDoc.Fields[0].Description)
	assert.Equal(t, "INT_FIELD", mockDocGenerator.GeneratedDoc.Fields[1].Key)
	assert.Equal(t, "Description for IntField", mockDocGenerator.GeneratedDoc.Fields[1].Description)
	assert.Equal(t, "Nested Struct Config", mockDocGenerator.GeneratedDoc.Groups[0].Title)
	assert.Len(t, mockDocGenerator.GeneratedDoc.Groups[0].Fields, 1)
	assert.Equal(t, "NESTED_BOOL_FIELD", mockDocGenerator.GeneratedDoc.Groups[0].Fields[0].Key)
	assert.Equal(t, "Description for Nested BoolField", mockDocGenerator.GeneratedDoc.Groups[0].Fields[0].Description)
}

func Test_GenerateDocumentation_WithError(t *testing.T) {
	type TestConfig struct {
	}

	cfg := new(TestConfig)
	mockDocGenerator := &MockDocGenerator{
		WithErr: true,
	}

	cfgManager := NewEmpty()
	err := cfgManager.GenerateDocumentation(cfg, mockDocGenerator)

	assert.NotNil(t, err)
}

func Test_parseDocGroup(t *testing.T) {
	type Nested struct {
		BoolField bool `env:"NESTED_BOOL_FIELD" description:"Description for Nested BoolField"`
	}

	type TestConfig struct {
		StringField   string `env:"STRING_FIELD" description:"Description for StringField"`
		IntField      int    `env:"INT_FIELD" description:"Description for IntField"`
		NestedStruct  Nested `title:"Nested Struct Config"`
		WithoutEnvTag string `description:"Description for WithoutEnvTag"`
	}

	cfg := new(TestConfig)

	docGroup := NewDoc()

	cfgManager := NewEmpty()
	require.NoError(t, cfgManager.parseDocGroup(docGroup, cfg))

	assert.NotNil(t, docGroup)
	assert.Equal(t, "", docGroup.Title)
	assert.Len(t, docGroup.Fields, 2, "fields without a key tag are not documented")
	assert.Len(t, docGroup.Groups, 1)
	assert.Equal(t, "STRING_FIELD", docGroup.Fields[0].Key)
	assert.Equal(t, "Description for StringField", docGroup.Fields[0].Description)
	assert.Equal(t, "INT_FIELD", docGroup.Fields[1].Key)
	assert.Equal(t, "Description for IntField", docGroup.Fields[1].Description)
	assert.Equal(t, "Nested Struct Config", docGroup.Groups[0].Title)
	assert.Len(t, docGroup.Groups[0].Fields, 1)
	assert.Equal(t, "NESTED_BOOL_FIELD", docGroup.Groups[0].Fields[0].Key)
	assert.Equal(t, "Description for Nested BoolField", docGroup.Groups[0].Fields[0].Description)
}

func Test_parseDocGroup_WithExampleTag(t *testing.T) {
	type TestConfig struct {
		StringField      string `env:"STRING_FIELD" description:"Description for StringField" default:"default_value" example:"example_value"`
		IntField         int    `env:"INT_FIELD" description:"Description for IntField" example:"42"`
		OnlyDefaultField string `env:"ONLY_DEFAULT_FIELD" default:"only_default"`
	}

	cfg := new(TestConfig)

	docGroup := NewDoc()

	cfgManager := NewEmpty()
	require.NoError(t, cfgManager.parseDocGroup(docGroup, cfg))

	assert.NotNil(t, docGroup)
	assert.Len(t, docGroup.Fields, 3)

	assert.Equal(t, "STRING_FIELD", docGroup.Fields[0].Key)
	assert.Equal(t, "Description for StringField", docGroup.Fields[0].Description)
	assert.Equal(t, "default_value", docGroup.Fields[0].DefaultValue)
	assert.Equal(t, "example_value", docGroup.Fields[0].ExampleValue)

	assert.Equal(t, "INT_FIELD", docGroup.Fields[1].Key)
	assert.Equal(t, "Description for IntField", docGroup.Fields[1].Description)
	assert.Equal(t, "", docGroup.Fields[1].DefaultValue)
	assert.Equal(t, "42", docGroup.Fields[1].ExampleValue)

	assert.Equal(t, "ONLY_DEFAULT_FIELD", docGroup.Fields[2].Key)
	assert.Equal(t, "only_default", docGroup.Fields[2].DefaultValue)
	assert.Equal(t, "", docGroup.Fields[2].ExampleValue)
}

func Test_parseDocGroup_WithPrefix(t *testing.T) {
	type Credentials struct {
		User     string `env:"USER" default:"admin"`
		Password string `env:"PASSWORD"`
	}

	type TestConfig struct {
		Primary Credentials `envPrefix:"PRIMARY_" title:"Primary"`
	}

	docGroup := NewDoc()
	require.NoError(t, NewEmpty().parseDocGroup(docGroup, new(TestConfig)))

	require.Len(t, docGroup.Groups, 1)
	group := docGroup.Groups[0]

	assert.Equal(t, "Primary", group.Title)
	require.Len(t, group.Fields, 2)

	assert.Equal(t, "PRIMARY_USER", group.Fields[0].Key)
	assert.Equal(t, "admin", group.Fields[0].DefaultValue)

	assert.Equal(t, "PRIMARY_PASSWORD", group.Fields[1].Key)
}

func Test_parseDocGroup_IgnoredFields(t *testing.T) {
	type Nested struct {
		Field string `env:"MUST_NOT_BE_DOCUMENTED"`
	}

	type TestConfig struct {
		Documented string `env:"DOCUMENTED_FIELD"`
		Excluded   string `env:"-" description:"Excluded from documentation"`
		Nested     Nested `env:"-"`
		//nolint:unused // the point of the test is that gocfg skips it
		unexported string `env:"UNEXPORTED_FIELD"`
	}

	docGroup := NewDoc()
	require.NoError(t, NewEmpty().parseDocGroup(docGroup, new(TestConfig)))

	require.Len(t, docGroup.Fields, 1)
	assert.Empty(t, docGroup.Groups)
	assert.Equal(t, "DOCUMENTED_FIELD", docGroup.Fields[0].Key)
}

func Test_GenerateDocumentation_TextUnmarshalerIsAValue(t *testing.T) {
	type TestConfig struct {
		StartedAt time.Time `env:"STARTED_AT" example:"2024-01-02T03:04:05Z"`
	}

	docGroup := NewDoc()
	require.NoError(t, NewDefault().parseDocGroup(docGroup, new(TestConfig)))

	require.Len(t, docGroup.Fields, 1)
	assert.Empty(t, docGroup.Groups)
	assert.Equal(t, "STARTED_AT", docGroup.Fields[0].Key)
}
