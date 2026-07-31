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

// Lookup reports the value of the environment variable and whether it is set at
// all, which lets gocfg tell `KEY=` apart from a missing key when strict lookup
// is enabled. It implements gocfg.LookupValueProvider.
func (p *EnvProvider) Lookup(key string) (string, bool) {
	return os.LookupEnv(key)
}
