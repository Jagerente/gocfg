package values

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempEnvFile(t *testing.T, content string, name ...string) string {
	t.Helper()

	fileName := "test_env.env"
	if len(name) > 0 {
		fileName = name[0]
	}

	path := filepath.Join(t.TempDir(), fileName)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func chdir(t *testing.T, dir string) {
	t.Helper()

	previous, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))

	t.Cleanup(func() { _ = os.Chdir(previous) })
}

func Test_NewDotEnvProvider(t *testing.T) {
	var (
		content = `VAR1=value1
VAR2=value2`
	)

	envFilePath := createTempEnvFile(t, content)

	provider, err := NewDotEnvProvider(envFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	assert.Equal(t, "value1", provider.Get("VAR1"))
	assert.Equal(t, "value2", provider.Get("VAR2"))

	assert.Equal(t, "", provider.Get("NON_EXISTING_KEY"))
}

func Test_NewDotEnvProviderDefaultFile(t *testing.T) {
	var (
		content = `VAR1=value1
VAR2=value2`
	)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, defaultEnvFile), []byte(content), 0o600))
	chdir(t, dir)

	provider, err := NewDotEnvProvider()
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	assert.Equal(t, "value1", provider.Get("VAR1"))
	assert.Equal(t, "value2", provider.Get("VAR2"))

	assert.Equal(t, "", provider.Get("NON_EXISTING_KEY"))
}

func Test_DotEnvProviderMultipleFiles(t *testing.T) {
	var (
		content1 = `VAR1=value1
VAR2=value2`
		content2 = `VAR2=value_that_should_not_be_set
VAR3=value3
VAR4=value4`
	)

	envFilePath1 := createTempEnvFile(t, content1, "first.env")
	envFilePath2 := createTempEnvFile(t, content2, "second.env")

	provider, err := NewDotEnvProvider(envFilePath1, envFilePath2)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	assert.Equal(t, "value1", provider.Get("VAR1"))
	assert.Equal(t, "value2", provider.Get("VAR2"))
	assert.Equal(t, "value3", provider.Get("VAR3"))
	assert.Equal(t, "value4", provider.Get("VAR4"))

	assert.Equal(t, "", provider.Get("NON_EXISTING_KEY"))
}

func Test_InvalidFileContent(t *testing.T) {
	envFilePath := createTempEnvFile(t, "!@#$%^&*()_+=-", "invalid_env_file.env")

	_, err := NewDotEnvProvider(envFilePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), envFilePath, "the failing file must be named in the error")
}

func Test_NonExistentFile(t *testing.T) {
	_, err := NewDotEnvProvider("!@#$%^&*()_")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "!@#$%^&*()_", "the missing file must be named in the error")
}

func Test_DotEnvProviderLookup(t *testing.T) {
	envFilePath := createTempEnvFile(t, "SET_VAR=value\nEMPTY_VAR=")

	provider, err := NewDotEnvProvider(envFilePath)
	require.NoError(t, err)

	value, found := provider.Lookup("SET_VAR")
	assert.True(t, found)
	assert.Equal(t, "value", value)

	value, found = provider.Lookup("EMPTY_VAR")
	assert.True(t, found, "a key present with an empty value is still present")
	assert.Equal(t, "", value)

	value, found = provider.Lookup("MISSING_VAR")
	assert.False(t, found)
	assert.Equal(t, "", value)
}

func Test_DotEnvProviderClosesFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.env")
	require.NoError(t, os.WriteFile(path, []byte("!@#$%^&*()_+=-"), 0o600))

	_, err := NewDotEnvProvider(path)
	require.Error(t, err)

	assert.NoError(t, os.Remove(path), "the file must not be held open after a parse failure")
}
