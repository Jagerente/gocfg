package values

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestEnvProvider_GetWithValue(t *testing.T) {
	var (
		key   = "EXISTING_KEY"
		value = "existing_value"
	)

	_ = os.Setenv(key, value)

	provider := NewEnvProvider()

	result := provider.Get(key)
	assert.Equal(t, value, result)
}

func TestEnvProvider_GetWithoutValue(t *testing.T) {
	var (
		key = "NON_EXISTING_KEY"
	)

	_ = os.Unsetenv(key)

	provider := NewEnvProvider()

	result := provider.Get(key)
	assert.Equal(t, "", result)
}
