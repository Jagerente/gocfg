package gocfg_test

import (
	"fmt"
	"os"
	"time"

	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/docgens"
)

func Example() {
	type AppConfig struct {
		Host     string        `env:"APP_HOST" default:"localhost"`
		Port     uint16        `env:"APP_PORT" default:"8080"`
		Timeout  time.Duration `env:"APP_TIMEOUT" default:"5s"`
		Optional string        `env:"APP_OPTIONAL,omitempty"`
	}

	_ = os.Setenv("APP_PORT", "3000")
	defer func() { _ = os.Unsetenv("APP_PORT") }()

	cfg := new(AppConfig)
	if err := gocfg.NewDefault().Unmarshal(cfg); err != nil {
		panic(err)
	}

	fmt.Printf("%s:%d timeout=%s optional=%q\n", cfg.Host, cfg.Port, cfg.Timeout, cfg.Optional)
	// Output: localhost:3000 timeout=5s optional=""
}

// ExampleConfigManager_WithLogger shows how to observe fields falling back to
// their default value. Without a logger, gocfg stays completely silent.
func ExampleConfigManager_WithLogger() {
	type AppConfig struct {
		Host string `env:"LOGGED_HOST" default:"localhost"`
	}

	_ = os.Unsetenv("LOGGED_HOST")

	logger := gocfg.LoggerFunc(func(format string, args ...interface{}) {
		fmt.Printf(format+"\n", args...)
	})

	if err := gocfg.NewDefault().WithLogger(logger).Unmarshal(new(AppConfig)); err != nil {
		panic(err)
	}

	// Output: WARNING: value for LOGGED_HOST not found, using default value: localhost
}

// ExampleConfigManager_CollectAllErrors reports every invalid field in one go
// instead of stopping at the first one.
func ExampleConfigManager_CollectAllErrors() {
	type AppConfig struct {
		Host string `env:"COLLECTED_HOST"`
		Port int    `env:"COLLECTED_PORT"`
	}

	_ = os.Unsetenv("COLLECTED_HOST")
	_ = os.Setenv("COLLECTED_PORT", "not a number")
	defer func() { _ = os.Unsetenv("COLLECTED_PORT") }()

	err := gocfg.NewDefault().CollectAllErrors().Unmarshal(new(AppConfig))
	fmt.Println(err)

	// Output:
	// COLLECTED_HOST cannot be empty
	// failed to parse COLLECTED_PORT: strconv.ParseInt: parsing "not a number": invalid syntax
}

// ExampleConfigManager_Unmarshal_envPrefix reuses one struct for several
// instances of the same dependency.
func ExampleConfigManager_Unmarshal_envPrefix() {
	type DBConfig struct {
		Host string `env:"HOST"`
		Port uint16 `env:"PORT" default:"5432"`
	}

	type AppConfig struct {
		Primary DBConfig `envPrefix:"PRIMARY_"`
		Replica DBConfig `envPrefix:"REPLICA_"`
	}

	_ = os.Setenv("PRIMARY_HOST", "primary.example.com")
	_ = os.Setenv("REPLICA_HOST", "replica.example.com")
	_ = os.Setenv("REPLICA_PORT", "5433")
	defer func() {
		_ = os.Unsetenv("PRIMARY_HOST")
		_ = os.Unsetenv("REPLICA_HOST")
		_ = os.Unsetenv("REPLICA_PORT")
	}()

	cfg := new(AppConfig)
	if err := gocfg.NewDefault().Unmarshal(cfg); err != nil {
		panic(err)
	}

	fmt.Printf("%s:%d\n%s:%d\n", cfg.Primary.Host, cfg.Primary.Port, cfg.Replica.Host, cfg.Replica.Port)
	// Output:
	// primary.example.com:5432
	// replica.example.com:5433
}

func ExampleConfigManager_GenerateDocumentation() {
	type LoggerConfig struct {
		Level string `env:"LOG_LEVEL" default:"info" example:"debug" description:"Verbosity of the application log"`
	}

	type AppConfig struct {
		Host   string       `env:"APP_HOST" default:"localhost" example:"0.0.0.0"`
		Logger LoggerConfig `title:"Logger"`
	}

	if err := gocfg.NewDefault().GenerateDocumentation(new(AppConfig), docgens.NewEnvDocGenerator(os.Stdout)); err != nil {
		panic(err)
	}

	// Output:
	// # Auto-generated config
	//
	// # Default: `localhost`
	// APP_HOST=0.0.0.0
	//
	// #############################
	// # Logger
	// #############################
	//
	// # Description:
	// #  Verbosity of the application log
	// #
	// # Default: `info`
	// LOG_LEVEL=debug
}
