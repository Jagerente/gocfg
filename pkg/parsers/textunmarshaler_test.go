package parsers

import (
	"errors"
	"math/big"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_TextUnmarshalerParserProvider(t *testing.T) {
	provider := NewTextUnmarshalerParserProvider()

	t.Run("time.Time", func(t *testing.T) {
		var target time.Time

		parser, ok := provider.Get(reflect.ValueOf(&target).Elem())
		require.True(t, ok)

		parsed, err := parser("2024-01-02T03:04:05Z")
		require.NoError(t, err)

		assert.Equal(t, time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), parsed.(time.Time).UTC())
		assert.Equal(t, time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), target.UTC(),
			"UnmarshalText writes straight through the pointer")
	})

	t.Run("net.IP", func(t *testing.T) {
		var target net.IP

		parser, ok := provider.Get(reflect.ValueOf(&target).Elem())
		require.True(t, ok)

		parsed, err := parser("10.0.0.1")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.1", parsed.(net.IP).String())
	})

	t.Run("big.Int", func(t *testing.T) {
		var target big.Int

		parser, ok := provider.Get(reflect.ValueOf(&target).Elem())
		require.True(t, ok)

		parsed, err := parser("42")
		require.NoError(t, err)

		result := parsed.(big.Int)
		assert.Equal(t, "42", result.String())
	})

	t.Run("propagates the unmarshalling error", func(t *testing.T) {
		var target time.Time

		parser, ok := provider.Get(reflect.ValueOf(&target).Elem())
		require.True(t, ok)

		_, err := parser("not a timestamp")
		assert.Error(t, err)
	})

	t.Run("declines types that cannot decode themselves", func(t *testing.T) {
		var target struct{ Field int }

		_, ok := provider.Get(reflect.ValueOf(&target).Elem())
		assert.False(t, ok)
	})

	t.Run("declines values that cannot be addressed", func(t *testing.T) {
		_, ok := provider.Get(reflect.ValueOf(time.Time{}))
		assert.False(t, ok)
	})
}

type valueReceiverUnmarshaler struct {
	Called bool
}

func (v *valueReceiverUnmarshaler) UnmarshalText(text []byte) error {
	if string(text) == "boom" {
		return errors.New("boom")
	}

	v.Called = true

	return nil
}

func Test_TextUnmarshalerParserProvider_CustomType(t *testing.T) {
	provider := NewTextUnmarshalerParserProvider()

	var target valueReceiverUnmarshaler

	parser, ok := provider.Get(reflect.ValueOf(&target).Elem())
	require.True(t, ok)

	parsed, err := parser("anything")
	require.NoError(t, err)
	assert.True(t, parsed.(valueReceiverUnmarshaler).Called)

	_, err = parser("boom")
	assert.EqualError(t, err, "boom")
}

func Test_ImplementsTextUnmarshaler(t *testing.T) {
	assert.True(t, implementsTextUnmarshaler(reflect.TypeOf(time.Time{})))
	assert.True(t, implementsTextUnmarshaler(reflect.TypeOf(net.IP{})))
	assert.False(t, implementsTextUnmarshaler(reflect.TypeOf("")))
	assert.False(t, implementsTextUnmarshaler(reflect.TypeOf([]string(nil))))
}
