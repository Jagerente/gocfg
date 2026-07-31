package values

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultEnvFile = ".env"
)

type DotEnvProvider struct {
	values map[string]string
}

// NewDotEnvProvider reads the given .env files, defaulting to ".env" when no
// path is given. Files are merged in order and the first file that defines a
// key wins.
func NewDotEnvProvider(paths ...string) (*DotEnvProvider, error) {
	provider := &DotEnvProvider{
		values: make(map[string]string),
	}

	if len(paths) < 1 {
		paths = []string{defaultEnvFile}
	}

	for _, path := range paths {
		if err := provider.load(path); err != nil {
			return nil, err
		}
	}

	return provider, nil
}

func (p *DotEnvProvider) load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}

	defer func() { _ = file.Close() }()

	values, err := godotenv.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}

	for key, value := range values {
		if _, ok := p.values[key]; ok {
			continue
		}

		p.values[key] = value
	}

	return nil
}

func (p *DotEnvProvider) Get(key string) string {
	return p.values[key]
}

// Lookup reports the value stored for the key and whether the key was present
// in any of the parsed files. It implements gocfg.LookupValueProvider.
func (p *DotEnvProvider) Lookup(key string) (string, bool) {
	value, ok := p.values[key]
	return value, ok
}
