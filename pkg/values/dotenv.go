package values

import (
	"github.com/joho/godotenv"
	"os"
)

const (
	defaultEnvFile = ".env"
)

type DotEnvProvider struct {
	path   string
	values map[string]string
}

func NewDotEnvProvider(paths ...string) (*DotEnvProvider, error) {
	provider := &DotEnvProvider{
		values: make(map[string]string),
	}

	if len(paths) < 1 {
		paths = []string{defaultEnvFile}
	}

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		values, err := godotenv.Parse(file)
		if err != nil {
			return nil, err
		}

		for key, value := range values {
			if _, ok := provider.values[key]; ok {
				continue
			}

			provider.values[key] = value
		}

		_ = file.Close()
	}

	return provider, nil
}

func (p *DotEnvProvider) Get(key string) string {
	if _, ok := p.values[key]; ok {
		return p.values[key]
	}

	return ""
}
