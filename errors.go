package gocfg

import (
	"errors"
	"fmt"
)

// Sentinel errors returned by Unmarshal and GenerateDocumentation.
// Use errors.Is to test for them:
//
//	if err := cfg.Unmarshal(appConfig); errors.Is(err, gocfg.ErrRequired) {
//		// at least one mandatory value is missing
//	}
var (
	// ErrInvalidTarget is returned when the value passed to Unmarshal or
	// GenerateDocumentation is not a non-nil pointer to a struct.
	ErrInvalidTarget = errors.New("gocfg: invalid target")

	// ErrRequired is returned when a field has no value and no usable default,
	// and is not marked with the omitempty option.
	ErrRequired = errors.New("gocfg: required value is missing")

	// ErrUnsupportedType is returned when none of the registered
	// ParserProviders can handle the type of a field.
	ErrUnsupportedType = errors.New("gocfg: unsupported field type")

	// ErrParse is returned when a parser rejects the raw value. The underlying
	// parser error is wrapped as well, so errors.Is keeps working for it too.
	ErrParse = errors.New("gocfg: parse error")

	// ErrInvalidParserResult is returned when a ParserProvider returns a value
	// that cannot be assigned to the field it was requested for.
	ErrInvalidParserResult = errors.New("gocfg: invalid parser result")
)

// FieldError reports a failure to populate a single configuration field.
// It is always produced by gocfg itself; use errors.As to inspect it:
//
//	var fieldErr *gocfg.FieldError
//	if errors.As(err, &fieldErr) {
//		log.Printf("%s (%s) is broken", fieldErr.Path, fieldErr.Key)
//	}
type FieldError struct {
	// Path is the Go field path inside the configuration struct,
	// for example "RedisConfig.RedisPort".
	Path string

	// Key is the configuration key that was looked up, prefix included,
	// for example "REDIS_PORT".
	Key string

	// Err is the underlying cause. It always matches one of the sentinel
	// errors above through errors.Is, and additionally wraps the error
	// returned by the parser, when there is one.
	Err error

	msg string
}

// Error implements the error interface.
func (e *FieldError) Error() string {
	if e.msg != "" {
		return e.msg
	}

	return fmt.Sprintf("%s: %v", e.Key, e.Err)
}

// Unwrap exposes the underlying cause to errors.Is and errors.As.
func (e *FieldError) Unwrap() error {
	return e.Err
}

func newFieldError(path, key string, cause error, format string, args ...interface{}) *FieldError {
	return &FieldError{
		Path: path,
		Key:  key,
		Err:  cause,
		msg:  fmt.Sprintf(format, args...),
	}
}
