package gocfg

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FieldError(t *testing.T) {
	t.Run("message produced by gocfg", func(t *testing.T) {
		err := newFieldError("Redis.Port", "REDIS_PORT", ErrRequired, "%s cannot be empty", "REDIS_PORT")

		assert.Equal(t, "REDIS_PORT cannot be empty", err.Error())
		assert.Equal(t, "Redis.Port", err.Path)
		assert.Equal(t, "REDIS_PORT", err.Key)
		assert.ErrorIs(t, err, ErrRequired)
	})

	// A FieldError built by hand carries no rendered message, so Error falls
	// back to the key and the cause.
	t.Run("built by hand", func(t *testing.T) {
		cause := errors.New("boom")
		err := &FieldError{Path: "Redis.Port", Key: "REDIS_PORT", Err: cause}

		assert.Equal(t, "REDIS_PORT: boom", err.Error())
		assert.ErrorIs(t, err, cause)
	})

	t.Run("wraps both the sentinel and the parser error", func(t *testing.T) {
		cause := errors.New("invalid syntax")
		err := newFieldError("Port", "PORT", errors.Join(ErrParse, cause), "failed to parse %s", "PORT")

		assert.ErrorIs(t, err, ErrParse)
		assert.ErrorIs(t, err, cause)
	})

	t.Run("errors.As finds it through wrapping", func(t *testing.T) {
		wrapped := errors.Join(
			errors.New("unrelated"),
			newFieldError("A.B", "KEY", ErrRequired, "KEY cannot be empty"),
		)

		var fieldErr *FieldError
		require.ErrorAs(t, wrapped, &fieldErr)
		assert.Equal(t, "A.B", fieldErr.Path)
	})
}
