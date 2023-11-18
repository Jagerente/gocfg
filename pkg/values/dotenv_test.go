package values

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func createTempEnvFile(content string, path ...string) (string, error) {
	var (
		tmpFile *os.File
	)

	if len(path) < 1 {
		tmpFile, _ = os.CreateTemp(".", "test_env_*.env")
	} else {
		tmpFile, _ = os.Create(path[0])
	}
	defer func() { _ = tmpFile.Close() }()

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func Test_NewDotEnvProvider(t *testing.T) {
	var (
		content = `VAR1=value1
VAR2=value2`
	)

	envFilePath, err := createTempEnvFile(content)
	assert.NoError(t, err)
	defer func() { _ = os.Remove(envFilePath) }()

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

	envFilePath, err := createTempEnvFile(content, defaultEnvFile)
	assert.NoError(t, err)
	defer func() { _ = os.Remove(envFilePath) }()

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

	envFilePath1, _ := createTempEnvFile(content1)
	defer func() { _ = os.Remove(envFilePath1) }()

	envFilePath2, err := createTempEnvFile(content2)
	assert.NoError(t, err)
	defer func() { _ = os.Remove(envFilePath2) }()

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
	envFilePath, _ := createTempEnvFile("!@#$%^&*()_+=-", "invalid_env_file.env")
	defer func() { _ = os.Remove(envFilePath) }()

	_, err := NewDotEnvProvider(envFilePath)
	assert.Error(t, err)
}

func Test_NonExistentFile(t *testing.T) {
	_, err := NewDotEnvProvider("!@#$%^&*()_")
	assert.Error(t, err)
}
