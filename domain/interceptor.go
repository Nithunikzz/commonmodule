package domain

// Config contains the configuration for the interceptor.
type Config struct {
	Exclude               []string `yaml:"exclude"`
	Authenticate          []string `yaml:"authenticate"`
	AuthenticateAuthorize []string `yaml:"authenticate-authorize"`
}
