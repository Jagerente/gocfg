[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![CI](https://github.com/Jagerente/gocfg/actions/workflows/ci.yml/badge.svg)](https://github.com/Jagerente/gocfg/actions/workflows/ci.yml)
[![CodeQL](https://github.com/Jagerente/gocfg/workflows/CodeQL/badge.svg)](https://github.com/Jagerente/gocfg/actions?query=workflow%3ACodeQL)
[![Go Report Card](https://goreportcard.com/badge/github.com/Jagerente/gocfg)](https://goreportcard.com/report/github.com/Jagerente/gocfg)
[![codecov](https://codecov.io/gh/Jagerente/gocfg/graph/badge.svg?token=7M88UL4ZG4)](https://codecov.io/gh/Jagerente/gocfg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Jagerente/gocfg.svg)](https://pkg.go.dev/github.com/Jagerente/gocfg)

## GoCfg

## Key Features

- Unmarshal from **Environment Variables**, **.env** and any other sources right to your structs.
- Set default values for each field using tags.
- Every type implementing `encoding.TextUnmarshaler` works out of the box.
- Easy to inject as much custom parsers as you need.
- Easy to inject your own values providers as much as you need and use them all at once with priority.
- Automatic documentation generator.

Requires Go 1.21 or newer.

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
	"time"

	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/parsers"
	"github.com/Jagerente/gocfg/pkg/values"
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
	TimeField         time.Time     `env:"TIME_FIELD"`
	ByteSliceField    []byte        `env:"BYTE_SLICE_FIELD"`
	StringSliceField  []string      `env:"STRING_SLICE_FIELD" default:"string1,string2,string3"`
	IntSliceField     []int         `env:"INT_SLICE_FIELD" default:"3,2,1,0,-1,-2,-3"`
	HostsField        []string      `env:"HOSTS_FIELD" envSeparator:";"`
	EmptyField        string        `env:"EMPTY_FIELD,omitempty"`
	WithDefaultField  string        `env:"WITH_DEFAULT_FIELD" default:"ave"`
	IgnoredField      string        `env:"-"`
}

func main() {
	cfg := gocfg.NewDefault()

	// Equals to
	cfg = gocfg.NewEmpty().
		UseDefaults().
		AddParserProviders(
			parsers.NewDefaultParserProvider(),
			parsers.NewTextUnmarshalerParserProvider(),
		).
		AddValueProviders(values.NewEnvProvider())

	appConfig := new(AppConfig)
	if err := cfg.Unmarshal(appConfig); err != nil {
		panic(err)
	}
}
```

### Supported tags

| Tag            | Applies to     | Description                                                                                          |
|----------------|----------------|------------------------------------------------------------------------------------------------------|
| `env`          | any field      | Name of the configuration key. `env:"-"` excludes the field entirely.                                |
| `omitempty`    | `env` option   | Allows the field to be empty, and makes an explicit `KEY=` mean exactly that. See below.             |
| `default`      | any field      | Value used when nothing is found.                                                                    |
| `example`      | any field      | Example value used by the documentation generator.                                                   |
| `description`  | any field      | Description rendered as a comment by the documentation generator. Supports `\n`.                     |
| `envSeparator` | slice fields   | Separator used to split the value. Defaults to `,`.                                                  |
| `envPrefix`    | nested structs | Prefix prepended to every key inside the struct, recursively.                                        |
| `title`        | nested structs | Section title used by the documentation generator.                                                   |

Fields without an `env` tag and unexported fields are ignored. Nested structs are
still traversed when they have no `env` tag.

### Default Type Parsers

> The following types are supported by default parsers:

- time.Duration
- bool
- string
- int, int8, int16, int32, int64
- uint, uint8, uint16, uint32, uint64
- float32, float64
- any type whose pointer implements `encoding.TextUnmarshaler`: `time.Time`,
  `net.IP`, `netip.Addr`, `big.Int`, `uuid.UUID`, and your own types
- a slice of anything above, and of anything a custom parser provider handles:
  registering a parser for `T` is enough to make `[]T` work. `[]byte` is taken as
  the raw value and never split

> Note: slice elements are trimmed, so `TAGS="a, b"` yields `["a", "b"]`.
> `[]byte` is the exception: it is the raw value, never split nor trimmed.

### Reusing a struct with envPrefix

```go
type DBConfig struct {
	Host string `env:"HOST"`
	Port uint16 `env:"PORT" default:"5432"`
}

type AppConfig struct {
	Primary DBConfig `envPrefix:"PRIMARY_" title:"Primary database"`
	Replica DBConfig `envPrefix:"REPLICA_" title:"Replica database"`
}
```

Reads `PRIMARY_HOST`, `PRIMARY_PORT`, `REPLICA_HOST` and `REPLICA_PORT`.
Prefixes accumulate through nesting.

### Logging

GoCfg never writes anywhere unless you give it a logger. `Logger` is a single
method interface satisfied by `*log.Logger`; `LoggerFunc` adapts anything else.

```go
cfg := gocfg.NewDefault().WithLogger(log.Default())

// zap
cfg = gocfg.NewDefault().WithLogger(gocfg.LoggerFunc(sugar.Infof))

// testing
cfg = gocfg.NewDefault().WithLogger(gocfg.LoggerFunc(t.Logf))
```

### Error handling

```go
err := gocfg.NewDefault().CollectAllErrors().Unmarshal(appConfig)

if errors.Is(err, gocfg.ErrRequired) {
	// at least one mandatory value is missing
}

var fieldErr *gocfg.FieldError
if errors.As(err, &fieldErr) {
	log.Printf("%s (%s) is broken: %v", fieldErr.Path, fieldErr.Key, fieldErr.Err)
}
```

Without `CollectAllErrors`, `Unmarshal` stops at the first invalid field.
The sentinel errors are `ErrInvalidTarget`, `ErrRequired`, `ErrUnsupportedType`,
`ErrParse` and `ErrInvalidParserResult`.

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
		AddParserProviders(
			parsers.NewDefaultParserProvider(),
			parsers.NewTextUnmarshalerParserProvider(),
		).
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

### Missing keys vs empty values

One rule governs `KEY=` (present but empty):

> **An empty value counts as a value only for a field marked `omitempty`.**
> Everywhere else walking the provider chain and applying the `default` tag
> an empty value is treated as a missing key.

The reasoning: `omitempty` is the field's own declaration that being empty is
meaningful. Without it, empty is not a value the field may hold, so an
environment variable that expanded to nothing is noise rather than intent and
must not override the fallback the author declared.

| Field                                       | `KEY` unset   | `KEY=`        |
|---------------------------------------------|---------------|---------------|
| `env:"KEY"`                                 | `ErrRequired` | `ErrRequired` |
| `env:"KEY" default:"x"`                     | `"x"`         | `"x"`         |
| `env:"KEY,omitempty"`                       | zero value    | zero value    |
| `env:"KEY,omitempty" default:"x"`           | `"x"`         | zero value    |
| `env:"PORT" default:"8080"` on an `int`     | `8080`        | `8080`        |
| `env:"PORT,omitempty" default:"8080"` (int) | `8080`        | `0`           |

So each tag answers exactly one question: `default` says what to use when
nothing was supplied, `omitempty` says whether the field may end up empty.
Marking a field `omitempty` is what makes `KEY=` mean "turn this off", which is
otherwise impossible to express.

The same rule applies to the provider chain: for an optional field the lookup
stops at the first provider that *has* the key, so an empty value in the
environment beats a value in a `.env` file; for every other field, providers
returning an empty value are skipped.

This relies on the optional `LookupValueProvider` interface, which both
`values.EnvProvider` and `values.DotEnvProvider` implement. A provider exposing
only `Get` can never report an empty value as present.

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

A parser provider is consulted before a struct field is traversed as a nested
group, which is what makes it possible to handle whole struct types as a single
value.

Registering a parser for `T` is also enough to make `[]T` work: gocfg first asks
every provider for the slice type itself, and if none of them claims it, asks
them again for the element type and composes the slice parser from the result.

```go 
package main

import (
	"reflect"
	"time"

	"github.com/Jagerente/gocfg"
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

A provider never has to deal with slices or with `envSeparator`: gocfg splits
the value and applies the separator itself when composing the slice. Should you
need the composition elsewhere, `parsers.NewSliceParser` builds a slice parser
out of an element parser.

> Providers added first have higher priority. `NewDefault` registers the built-in
> parsers first, so a custom parser for a named type over a built-in kind
> (`type Level int`) is shadowed by them. To take precedence, build the chain
> yourself with `NewEmpty` and register your provider before
> `parsers.NewDefaultParserProvider()`.

### Custom value provider

```go 
package main

import (
	"os"

	"github.com/Jagerente/gocfg"
)

type CustomValueProvider struct {
}

func NewCustomValueProvider() *CustomValueProvider {
	return &CustomValueProvider{}
}

func (p *CustomValueProvider) Get(key string) string {
	return os.Getenv("CUSTOM_" + key)
}

// Optional: implement Lookup so that an empty value is told apart from a
// missing key. Without it, an empty result means "ask the next provider".
func (p *CustomValueProvider) Lookup(key string) (string, bool) {
	return os.LookupEnv("CUSTOM_" + key)
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

### Documentation generation

1. Let's say you have such config file `/internal/config/config.go`:

```go
package config

import (
	"time"

	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/values"
	cache_factory "your_cool_app/internal/router/cache"
)

type LoggerConfig struct {
	LogLevel     int  `env:"LOG_LEVEL" default:"6" example:"4" description:"https://pkg.go.dev/github.com/sirupsen/logrus@v1.9.3#Level"`
	ReportCaller bool `env:"REPORT_CALLER" default:"true" example:"false"`
	LogFormatter int  `env:"LOG_FORMATTER" default:"0" example:"1"`
}

type CassandraConfig struct {
	CassandraHosts    string `env:"CASSANDRA_HOSTS" default:"127.0.0.1" example:"cassandra.example.com"`
	CassandraKeyspace string `env:"CASSANDRA_KEYSPACE" default:"user_data_service" example:"production_keyspace"`
}

type RouterConfig struct {
	ServerPort               uint16        `env:"SERVER_PORT" default:"8080" example:"3000"`
	Debug                    bool          `env:"ROUTER_DEBUG" default:"true" example:"false"`
	CacheAdapter             string        `env:"CACHE_ADAPTER,omitempty" example:"redis" description:"Leave blank to not use.\nPossible values:\n- redis\n- memcache"`
	CacheAdapterTTL          time.Duration `env:"CACHE_ADAPTER_TTL,omitempty" default:"1m" example:"5m"`
	CacheAdapterNoCacheParam string        `env:"CACHE_ADAPTER_NOCACHE_PARAM,omitempty" default:"no-cache" example:"skip-cache"`
}

type RedisCacheAdapterConfig struct {
	RedisAddr     string `env:"CACHE_ADAPTER_REDIS_ADDR,omitempty" default:":6379"`
	RedisDB       int    `env:"CACHE_ADAPTER_REDIS_DB,omitempty" default:"0"`
	RedisUsername string `env:"CACHE_ADAPTER_REDIS_USERNAME,omitempty"`
	RedisPassword string `env:"CACHE_ADAPTER_REDIS_PASSWORD,omitempty"`
}

type MemcacheCacheAdapterConfig struct {
	Capacity         int                     `env:"CACHE_ADAPTER_MEMCACHE_CAPACITY,omitempty" default:"10000000"`
	CachingAlgorithm cache_factory.Algorithm `env:"CACHE_ADAPTER_MEMCACHE_CACHING_ALGORITHM,omitempty" default:"LRU"`
}

type Config struct {
	LoggerConfig               `title:"Logger configuration"`
	RouterConfig               `title:"Router configuration"`
	RedisCacheAdapterConfig    `title:"Redis Cache Adapter configuration"`
	MemcacheCacheAdapterConfig `title:"Memcache Cache Adapter configuration"`
	CassandraConfig            `title:"Cassandra configuration"`
}

func New() (*Config, error) {
	var cfg = new(Config)

	cfgManager := gocfg.NewDefault()
	if dotEnvProvider, err := values.NewDotEnvProvider(); err == nil {
		cfgManager = cfgManager.AddValueProviders(dotEnvProvider)
	}

	if err := cfgManager.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

```

2. Create new app, for example `/cmd/docs/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/docgens"
	"your_cool_app/internal/config"
)

const outputFile = ".env.dist.generated"

func main() {
	cfg := new(config.Config)

	file, err := os.Create(outputFile)
	if err != nil {
		panic(fmt.Errorf("error creating %s file: %v", outputFile, err))
	}
	defer file.Close()

	cfgManager := gocfg.NewDefault()
	if err := cfgManager.GenerateDocumentation(cfg, docgens.NewEnvDocGenerator(file)); err != nil {
		panic(err)
	}
}

```

3. Run it by executing `go run cmd/docs/main.go`; it will generate the following file `.env.dist.generated`:

```bash
# Auto-generated config

#############################
# Logger configuration
#############################

# Description:
#  https://pkg.go.dev/github.com/sirupsen/logrus@v1.9.3#Level
#
# Default: `6`
LOG_LEVEL=4

# Default: `true`
REPORT_CALLER=false

# Default: `0`
LOG_FORMATTER=1

#############################
# Router configuration
#############################

# Default: `8080`
SERVER_PORT=3000

# Default: `true`
ROUTER_DEBUG=false

# Allowed to be empty
# Description:
#  Leave blank to not use.
#  Possible values:
#  - redis
#  - memcache
CACHE_ADAPTER=redis

# Allowed to be empty
# Default: `1m`
CACHE_ADAPTER_TTL=5m

# Allowed to be empty
# Default: `no-cache`
CACHE_ADAPTER_NOCACHE_PARAM=skip-cache

#############################
# Redis Cache Adapter configuration
#############################

# Allowed to be empty
# Default: `:6379`
CACHE_ADAPTER_REDIS_ADDR=:6379

# Allowed to be empty
# Default: `0`
CACHE_ADAPTER_REDIS_DB=0

# Allowed to be empty
CACHE_ADAPTER_REDIS_USERNAME=

# Allowed to be empty
CACHE_ADAPTER_REDIS_PASSWORD=

#############################
# Memcache Cache Adapter configuration
#############################

# Allowed to be empty
# Default: `10000000`
CACHE_ADAPTER_MEMCACHE_CAPACITY=10000000

# Allowed to be empty
# Default: `LRU`
CACHE_ADAPTER_MEMCACHE_CACHING_ALGORITHM=LRU

#############################
# Cassandra configuration
#############################

# Default: `127.0.0.1`
CASSANDRA_HOSTS=cassandra.example.com

# Default: `user_data_service`
CASSANDRA_KEYSPACE=production_keyspace

```
