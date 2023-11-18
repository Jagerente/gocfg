[![CI](https://github.com/Jagerente/gocfg/actions/workflows/ci.yml/badge.svg)](https://github.com/Jagerente/gocfg/actions/workflows/ci.yml)
[![CodeQL](https://github.com/Jagerente/gocfg/workflows/CodeQL/badge.svg)](https://github.com/Jagerente/gocfg/actions?query=workflow%3ACodeQL)
[![Go Report Card](https://goreportcard.com/badge/github.com/Jagerente/gocfg)](https://goreportcard.com/report/github.com/Jagerente/gocfg)
[![codecov](https://codecov.io/gh/Jagerente/gocfg/graph/badge.svg?token=7M88UL4ZG4)](https://codecov.io/gh/Jagerente/gocfg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Jagerente/gocfg.svg)](https://pkg.go.dev/github.com/Jagerente/gocfg)

## GoCfg

## Key Features

- Unmarshal from **Environment Variables**, **.env** and any other sources right to your structs.
- Set default values for each field using tags.
- Easy to inject as much custom parsers as you need.
- Easy to inject your own values providers as much as you need and use them all at once with priority.

## Quick start

### Install package:

```bash
go get -u github.com/Jagerente/gocfg
```

### Basic usage:

It will use environment variables and default values defined in tags.

```go
package main

import (
	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/parsers"
	"github.com/Jagerente/gocfg/pkg/values"
	"time"
)

type LoggerConfig struct {
	LogLevel string `env:"LOG_LEVEL" default:"debug"`
}

type RedisConfig struct {
	RedisHost     string `env:"REDIS_HOST" default:"localhost"`
	RedisPort     uint16 `env:"REDIS_PORT" default:"6379"`
	RedisUser     string `env:"REDIS_USER,omitempty"`
	RedisPassword string `env:"REDIS_PASS"`
	RedisDatabase string `env:"REDIS_DATABASE"`
}

type AppConfig struct {
	// Supported Tags:
	// - env: Specifies the environment variable name.
	// - default: Specifies the default value for the field.
	// - omitempty: Allows empty fields. FOR STRINGS ONLY!

	LogLevel          LoggerConfig
	RedisConfig       RedisConfig
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
	TimeDurationField time.Duration `env:"TIME_DURATION_FIELD"`
	EmptyField        string        `env:"EMPTY_FIELD,omitempty"`
	WithDefaultField  string        `env:"WITH_DEFAULT_FIELD" default:"ave"`
}

func main() {
	cfg := gocfg.NewDefault()

	// Equals to
	cfg = gocfg.NewEmpty().
		UseDefaults().
		AddParserProviders(parsers.NewDefaultParserProvider()).
		AddValueProviders(values.NewEnvProvider())

	appConfig := new(AppConfig)
	if err := cfg.Unmarshal(appConfig); err != nil {
		panic(err)
	}
}

```

### Default Type Parsers

> The following types are supported by default parsers:

- time.Duration
- bool
- string
- int, int8, int16, int32, int64
- uint, uint8, uint16, uint32, uint64
- float32, float64

### .env file

```go
package main

import (
	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/parsers"
	"github.com/Jagerente/gocfg/pkg/values"
)

type AppConfig struct {
	BoolField   bool   `env:"BOOL_FIELD"`
	StringField string `env:"STRING_FIELD"`
	IntField    int    `env:"INT_FIELD"`
}

func main() {
	// With default '.env' file
	dotEnvProvider, _ := values.NewDotEnvProvider()

	// With custom env file path 
	dotEnvProvider, _ = values.NewDotEnvProvider("local.env")

	// With multiple env files
	dotEnvProvider, _ = values.NewDotEnvProvider("local.env", "dev.env")

	cfg := gocfg.NewDefault().
		AddValueProviders(dotEnvProvider)

	// Equals to
	cfg = gocfg.NewEmpty().
		UseDefaults().
		AddParserProviders(parsers.NewDefaultParserProvider()).
		AddValueProviders(
			values.NewEnvProvider(),
			dotEnvProvider,
		)

	appConfig := new(AppConfig)
	if err := cfg.Unmarshal(appConfig); err != nil {
		panic(err)
	}
}
```

### Custom key tag

```go
package main

import (
	"github.com/Jagerente/gocfg"
)

type AppConfig struct {
	BoolField   bool   `mapstructure:"BOOL_FIELD"`
	StringField string `mapstructure:"STRING_FIELD"`
	IntField    int    `mapstructure:"INT_FIELD"`
}

func main() {
	cfg := gocfg.NewDefault().
		UseCustomKeyTag("mapstructure")

	appConfig := new(AppConfig)
	if err := cfg.Unmarshal(appConfig); err != nil {
		panic(err)
	}
}

```

### Custom parser provider

```go 
package main

import (
	"github.com/Jagerente/gocfg"
	"reflect"
	"time"
)

type CustomParserProvider struct {
}

func NewCustomParserProvider() *CustomParserProvider {
	return &CustomParserProvider{}
}

func (p *CustomParserProvider) Get(field reflect.Value) (func(v string) (any, error), bool) {
	switch field.Type() {
	case reflect.TypeOf(time.Duration(83)):
		return func(v string) (any, error) {
			return time.ParseDuration(v)
		}, true
	default:
		return nil, false
	}
}

type AppConfig struct {
	BoolField   bool   `env:"BOOL_FIELD"`
	StringField string `env:"STRING_FIELD"`
	IntField    int    `env:"INT_FIELD"`
}

func main() {
	customParserProvider := NewCustomParserProvider()

	cfg := gocfg.NewDefault().
		AddParserProviders(customParserProvider)

	appConfig := new(AppConfig)
	if err := cfg.Unmarshal(appConfig); err != nil {
		panic(err)
	}
}

```

### Custom value provider

```go 
package main

import (
	"github.com/Jagerente/gocfg"
	"os"
)

type CustomValueProvider struct {
}

func NewCustomValueProvider() *CustomValueProvider {
	return &CustomValueProvider{}
}

func (p *CustomValueProvider) Get(key string) string {
	return os.Getenv("CUSTOM_" + key)
}

type AppConfig struct {
	BoolField   bool   `env:"BOOL_FIELD"`
	StringField string `env:"STRING_FIELD"`
	IntField    int    `env:"INT_FIELD"`
}

func main() {
	customValueProvider := NewCustomValueProvider()

	cfg := gocfg.NewDefault().
		AddValueProviders(customValueProvider)

	appConfig := new(AppConfig)
	if err := cfg.Unmarshal(appConfig); err != nil {
		panic(err)
	}
}

```