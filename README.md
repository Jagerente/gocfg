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
- Automatic documentation generator.

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
	// - example: Specifies an example value for documentation generation.
	// - omitempty: Allows empty fields. 
	//              If both the parsed value and the default value are empty, 
	//              the field will be set to the zero value for its type in Go.
	// - description: Describes the field for documentation generation.
	// - title: Specifies the title for nested struct documentation.

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
	ByteSliceField    []byte        `env:"BYTE_SLICE_FIELD"`
	StringSliceField  []string      `env:"STRING_SLICE_FIELD" default:"string1,string2,string3"`
	IntSliceField     []int         `env:"INT_SLICE_FIELD" default:"3,2,1,0,-1,-2,-3"`
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
- slices: bytes, strings, ints

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

### Documentation generation

1. Let's say you have such config file `/internal/config/config.go`:

```go
package config

import (
	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/values"
	"time"
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
	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/docgens"
	"os"
	"your_cool_app/internal/config"
)

const outputFile = ".env.dist.generated"

func main() {
	cfg := new(config.Config)

	file, err := os.Create(outputFile)
	if err != nil {
		panic(fmt.Errorf("error creating %s file: %v", outputFile, err))
	}

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
CACHE_ADAPTER_REDIS_ADDR=:6379

# Allowed to be empty
CACHE_ADAPTER_REDIS_DB=0

# Allowed to be empty
CACHE_ADAPTER_REDIS_USERNAME=

# Allowed to be empty
CACHE_ADAPTER_REDIS_PASSWORD=

#############################
# Memcache Cache Adapter configuration
#############################

# Allowed to be empty
CACHE_ADAPTER_MEMCACHE_CAPACITY=10000000

# Allowed to be empty
CACHE_ADAPTER_MEMCACHE_CACHING_ALGORITHM=LRU

#############################
# Cassandra configuration
#############################

# Default: `127.0.0.1`
CASSANDRA_HOSTS=cassandra.example.com

# Default: `user_data_service`
CASSANDRA_KEYSPACE=production_keyspace

```
