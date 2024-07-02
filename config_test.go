package gocfg

import (
	"errors"
	"github.com/Jagerente/gocfg/pkg/values"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
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

		EmptyField string `env:"EMPTY_FIELD,omitempty"`
	}

	_ = os.Setenv("BOOL_FIELD", "true")

	_ = os.Setenv("STRING_FIELD", "test")

	_ = os.Setenv("INT_FIELD", "-2147483648")
	_ = os.Setenv("INT8_FIELD", "-128")
	_ = os.Setenv("INT16_FIELD", "-32768")
	_ = os.Setenv("INT32_FIELD", "-2147483648")
	_ = os.Setenv("INT64_FIELD", "-9223372036854775808")

	_ = os.Setenv("UINT_FIELD", "4294967295")
	_ = os.Setenv("UINT8_FIELD", "255")
	_ = os.Setenv("UINT16_FIELD", "65535")
	_ = os.Setenv("UINT32_FIELD", "4294967295")
	_ = os.Setenv("UINT64_FIELD", "18446744073709551615")

	_ = os.Setenv("FLOAT32_FIELD", "3.14")
	_ = os.Setenv("FLOAT64_FIELD", "3.14159265359")

	_ = os.Setenv("TIME_FIELD", "5s")
	_ = os.Setenv("BYTE_SLICE_FIELD", "test")
	_ = os.Setenv("STRING_SLICE_FIELD", "test1,test2,test3")

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
STRING_SLICE_FIELD=test1,test2,test3`
	)

	tmpFile, _ := os.CreateTemp(".", "test_env_*.env")
	envFilePath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(envFilePath)
	}()

	_, _ = tmpFile.WriteString(envContent)

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
}

func Test_EmptyField(t *testing.T) {
	type TestConfig struct {
		BoolField   bool   `env:"BOOL_FIELD"`
		StringField string `env:"STRING_FIELD"`
	}

	_ = os.Setenv("BOOL_FIELD", "true")
	_ = os.Setenv("STRING_FIELD", "")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "STRING_FIELD cannot be empty")
}

func Test_InvalidType(t *testing.T) {
	type TestConfig struct {
		BoolField bool `env:"BOOL_FIELD"`
		IntField  int  `env:"INT_FIELD"`
	}

	_ = os.Setenv("BOOL_FIELD", "true")
	_ = os.Setenv("INT_FIELD", "invalid")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse INT_FIELD")
}

func Test_StructField(t *testing.T) {
	type TestConfig struct {
		StructField struct {
			NestedField int `env:"NESTED_FIELD"`
		}
	}

	_ = os.Setenv("NESTED_FIELD", "123")

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
	}

	_ = os.Setenv("BOOL_FIELD", "")
	_ = os.Setenv("STRING_FIELD", "")
	_ = os.Setenv("INT_FIELD", "")
	_ = os.Setenv("INT8_FIELD", "")
	_ = os.Setenv("INT16_FIELD", "")
	_ = os.Setenv("INT32_FIELD", "")
	_ = os.Setenv("INT64_FIELD", "")
	_ = os.Setenv("UINT_FIELD", "")
	_ = os.Setenv("UINT8_FIELD", "")
	_ = os.Setenv("UINT16_FIELD", "")
	_ = os.Setenv("UINT32_FIELD", "")
	_ = os.Setenv("UINT64_FIELD", "")
	_ = os.Setenv("FLOAT32_FIELD", "")
	_ = os.Setenv("FLOAT64_FIELD", "")
	_ = os.Setenv("TIME_FIELD", "")
	_ = os.Setenv("BYTE_SLICE_FIELD", "")
	_ = os.Setenv("STRING_SLICE_FIELD", "")

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
}

func Test_DefaultValues(t *testing.T) {
	type TestConfig struct {
		BoolField         bool          `env:"BOOL_FIELD" default:"true"`
		StringField       string        `env:"STRING_FIELD" default:"default"`
		IntField          int           `env:"INT_FIELD" default:"42"`
		Float64Field      float64       `env:"FLOAT64_FIELD" default:"3.14"`
		TimeDurationField time.Duration `env:"TIME_DURATION_FIELD" default:"1h"`
	}

	_ = os.Setenv("BOOL_FIELD", "")
	_ = os.Setenv("STRING_FIELD", "")
	_ = os.Setenv("INT_FIELD", "")
	_ = os.Setenv("FLOAT64_FIELD", "")
	_ = os.Setenv("TIME_DURATION_FIELD", "")

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

	_ = os.Setenv("BOOL_FIELD", "false")
	_ = os.Setenv("STRING_FIELD", "not default")
	_ = os.Setenv("INT_FIELD", "83")
	_ = os.Setenv("FLOAT64_FIELD", "8.3")
	_ = os.Setenv("TIME_DURATION_FIELD", "5s")

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

	_ = os.Setenv("BOOL_FIELD", "true")
	_ = os.Setenv("STRING_FIELD", "value")
	_ = os.Setenv("INT_FIELD", "83")

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

	_ = os.Setenv("NESTED_FIELD", "invalid")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse StructField")
}

func Test_GetParserUnsupportedField(t *testing.T) {
	type TestConfig struct {
		UnsupportedField complex128 `env:"UNSUPPORTED_FIELD"`
	}

	_ = os.Setenv("UNSUPPORTED_FIELD", "value")

	cfg := new(TestConfig)
	cfgManager := NewDefault()
	err := cfgManager.Unmarshal(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get parser for UNSUPPORTED_FIELD: unsupported")
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
	assert.Len(t, mockDocGenerator.GeneratedDoc.Fields, 3)
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
	cfgManager.parseDocGroup(docGroup, cfg)

	assert.NotNil(t, docGroup)
	assert.Equal(t, "", docGroup.Title)
	assert.Len(t, docGroup.Fields, 3)
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
