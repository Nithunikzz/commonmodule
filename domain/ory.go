package domain

type Ory string

type OAuthClientConfig struct {
	ClientName              string
	RedirectURIs            []string
	GrantTypes              []string
	ResponseTypes           []string
	TokenEndpointAuthMethod string
	Scope                   string
}
type OAuthClientOption func(*OAuthClientConfig)

type ClientCredentialConfig struct {
	ClientName              string
	GrantTypes              []string
	TokenEndpointAuthMethod string
	Scope                   string
}
