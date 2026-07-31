package values

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvProvider_GetWithValue(t *testing.T) {
	var (
		key   = "EXISTING_KEY"
		value = "existing_value"
	)

	t.Setenv(key, value)

	provider := NewEnvProvider()

	result := provider.Get(key)
	assert.Equal(t, value, result)
}

func TestEnvProvider_GetWithoutValue(t *testing.T) {
	var (
		key = "NON_EXISTING_KEY"
	)

	require.NoError(t, os.Unsetenv(key))

	provider := NewEnvProvider()

	result := provider.Get(key)
	assert.Equal(t, "", result)
}

func TestEnvProvider_Lookup(t *testing.T) {
	provider := NewEnvProvider()

	t.Run("set", func(t *testing.T) {
		t.Setenv("LOOKUP_SET_KEY", "value")

		value, found := provider.Lookup("LOOKUP_SET_KEY")
		assert.True(t, found)
		assert.Equal(t, "value", value)
	})

	t.Run("set to an empty value", func(t *testing.T) {
		t.Setenv("LOOKUP_EMPTY_KEY", "")

		value, found := provider.Lookup("LOOKUP_EMPTY_KEY")
		assert.True(t, found, "an explicitly empty variable is still set")
		assert.Equal(t, "", value)
	})

	t.Run("unset", func(t *testing.T) {
		require.NoError(t, os.Unsetenv("LOOKUP_MISSING_KEY"))

		value, found := provider.Lookup("LOOKUP_MISSING_KEY")
		assert.False(t, found)
		assert.Equal(t, "", value)
	})
}
