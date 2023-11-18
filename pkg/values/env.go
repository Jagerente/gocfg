package values

import "os"

type EnvProvider struct {
}

func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

func (p *EnvProvider) Get(key string) string {
	return os.Getenv(key)
}
