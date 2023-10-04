package ory

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	credential "github.com/common/commonmodule/credentials"
	"github.com/common/commonmodule/domain"
	"github.com/pkg/errors"

	"github.com/AlekSi/pointer"
	"github.com/intelops/go-common/logging"

	//"gitlab.com/tariandev_intelops/iam/credential"

	"github.com/kelseyhightower/envconfig"
	ory "github.com/ory/client-go"

	//sqldb "gitlab.com/tariandev_intelops/iam/internal/adapter/db"

	// "gitlab.com/tariandev_intelops/iam/internal/application/api"
	// "gitlab.com/tariandev_intelops/iam/internal/application/domain"
	// "gitlab.com/tariandev_intelops/iam/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Config will have the configuration details
// Config will have the configuration details
type OryEnv struct {
	RootUser      string `envconfig:"ORY_ROOT_USER" required:"true"`
	OrySchema     string `envconfig:"ORY_SCHEMA" required:"true"`
	OryEntityName string `envconfig:"ORY_ENTITY_NAME" required:"true"`
}

type OryClient interface {
	// PatchIdentity(ctx context.Context, oryid, firstName, lastName string) error
	// ListOAuth2Clients(ctx context.Context, serviceName string) ([]ory.OAuth2Client, error)
	// DeleteOAuthClient(ctx context.Context, clientID string) error
	CreateOAuthClientCredential(context.Context, *domain.ClientCredentialConfig) (*ory.OAuth2Client, error)
	NewClientCredentialsConfig(string) *domain.ClientCredentialConfig
	// InitializeOry(dbAdapter *sqldb.Adapter) (string, error)
	// CreateUserInOry(email, firstName, lastName string) (string, error)
	// RecoveryLink(id string) (*ory.RecoveryLinkForIdentity, error)
	// DeleteInOry(id string) error
	GetTokenFromContext(ctx context.Context) (string, error)
	Authorize(ctx context.Context, accessToken string) (context.Context, error)
	GetOryIDFromContext(ctx context.Context) (string, error)
	GetOrgIdFromContext(ctx context.Context) (string, error)
	CreateOAuthClient(ctx context.Context, config *domain.OAuthClientConfig) (*ory.OAuth2Client, error)
	NewOAuthClientConfig(clientName string, redirectURI string, opts ...domain.OAuthClientOption) *domain.OAuthClientConfig
	WithGrantTypes(grantTypes []string) domain.OAuthClientOption
	WithResponseTypes(responseTypes []string) domain.OAuthClientOption
	WithTokenEndpointAuthMethod(tokenEndpointAuthMethod string) domain.OAuthClientOption
	WithScope(scope string) domain.OAuthClientOption
	GetOryTokenUrl() string
	IntrospectToken(context.Context, string) (bool, error)
	GetOauthTokenFromContext(ctx context.Context) (string, error)
}

type Client struct {
	config          *OryEnv
	conn            *ory.APIClient
	identity        *ory.Identity
	oryPAT          string
	oryURL          string
	oryRootPassword string
	log             logging.Logger
}

func NewClient(log logging.Logger) (OryClient, error) {
	cfg, err := GetOryEnv()
	if err != nil {
		return nil, err
	}

	serviceCredential, err := credential.GetServiceUserCredential(context.Background(),
		cfg.OryEntityName, cfg.RootUser)
	if err != nil {
		return nil, err
	}
	oryPAT := serviceCredential.AdditionalData["ORY_PAT"]
	oryURL := serviceCredential.AdditionalData["ORY_URL"]

	oryConfig := NewOrySdk(cfg, oryURL)
	return &Client{
		config:          cfg,
		conn:            oryConfig,
		log:             log,
		oryPAT:          oryPAT,
		oryURL:          oryURL,
		oryRootPassword: serviceCredential.Password,
	}, nil
}

func GetOryEnv() (*OryEnv, error) {
	cfg := &OryEnv{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewOrySdk(conf *OryEnv, oryURL string) *ory.APIClient {
	log.Println("creating a ory client")
	config := ory.NewConfiguration()
	config.Servers = ory.ServerConfigurations{{
		URL: oryURL,
	}}

	return ory.NewAPIClient(config)
}
func (c *Client) RecoveryLink(id string) (*ory.RecoveryLinkForIdentity, error) {
	oryAuthedContext := context.WithValue(context.Background(), ory.ContextAccessToken, c.oryPAT)
	recovery, _, err := c.conn.IdentityApi.CreateRecoveryLinkForIdentity(oryAuthedContext).
		CreateRecoveryLinkForIdentityBody(*ory.NewCreateRecoveryLinkForIdentityBody(id)).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to create the recovery link %v \n", err)
	}
	return recovery, nil
}

// newIdentityWithCredentials was to include a password in identity body
func newIdentityWithCredentials(password *string) *ory.IdentityWithCredentials {
	return &ory.IdentityWithCredentials{
		Password: &ory.IdentityWithCredentialsPassword{
			Config: &ory.IdentityWithCredentialsPasswordConfig{
				Password: password,
			},
		},
	}
}

func (c *Client) CreateRootUser() (*ory.Identity, error) {
	rootCredentials := newIdentityWithCredentials(pointer.ToString(c.oryRootPassword))

	oryAuthedContext := context.WithValue(context.Background(), ory.ContextAccessToken, c.oryPAT)
	rootIdentity, _, err := c.conn.IdentityApi.CreateIdentity(oryAuthedContext).
		CreateIdentityBody(ory.CreateIdentityBody{
			MetadataAdmin: map[string]interface{}{
				"Root": "Yes",
			},
			Credentials: rootCredentials,
			SchemaId:    c.config.OrySchema,
			Traits: map[string]interface{}{
				"email": c.config.RootUser,
			},
		}).Execute()
	if err != nil {
		return nil, err
	}
	if rootIdentity == nil {
		return nil, fmt.Errorf("Root identity creation failed.")
	}

	return rootIdentity, nil
}

// func (c *Client) CreateUserInOry(email string) (string, error) {
// 	oryAuthedContext := context.WithValue(context.Background(), ory.ContextAccessToken, c.oryPAT)
// 	createIdentityBody := *ory.NewCreateIdentityBody(
// 		"preset://email",
// 		map[string]interface{}{
// 			"email": email,
// 		},
// 	)
// 	createdIdentity, _, err := c.conn.IdentityApi.CreateIdentity(oryAuthedContext).
// 		CreateIdentityBody(createIdentityBody).Execute()
// 	if err != nil {
// 		if err.Error() == "409 Conflict" {
// 			c.log.Errorf("Identity already exists %v ", err)
// 			return "", err
// 		}
// 		c.log.Errorf("Identity creation in ory failed %v ", err)
// 		return "", err
// 	}
// 	return createdIdentity.Id, nil
// }

func (c *Client) CreateUserInOry(email, firstName, lastName string) (string, error) {
	oryAuthedContext := context.WithValue(context.Background(), ory.ContextAccessToken, c.oryPAT)
	createIdentityBody := *ory.NewCreateIdentityBody(
		"70720c2328babf93141ffbcc8f42b764b451cfc417530f3547e258d8e1f1290110553b5389002083f37cfb944322d42700893a83e095f9771e44d9a3dc547009",
		map[string]interface{}{
			"email":     email,
			"firstName": firstName,
			"lastName":  lastName,
		},
	)
	createdIdentity, _, err := c.conn.IdentityApi.CreateIdentity(oryAuthedContext).
		CreateIdentityBody(createIdentityBody).Execute()
	if err != nil {
		if err.Error() == "409 Conflict" {
			c.log.Errorf("Identity already exists %v ", err)
			return "", err
		}
		c.log.Errorf("Identity creation in ory failed %v ", err)
		return "", err
	}
	return createdIdentity.Id, nil
}

func (c *Client) DeleteInOry(id string) error {
	oryAuthedContext := context.WithValue(context.Background(), ory.ContextAccessToken, c.oryPAT)
	_, err := c.conn.IdentityApi.DeleteIdentity(oryAuthedContext, id).Execute()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Authorize(ctx context.Context, accessToken string) (context.Context, error) {
	expand := []string{
		"Identity",
	}
	ctx = context.WithValue(ctx, ory.ContextAccessToken, c.oryPAT)
	sessionInfo, _, err := c.conn.IdentityApi.GetSession(ctx, accessToken).Expand(expand).Execute()
	if err != nil {
		c.log.Error("Error occured while getting session info for session id - "+accessToken+"+\nError - ", err.Error())
		return ctx, status.Errorf(codes.Unauthenticated, "Failed to introspect session id - %v", err)
	}
	log.Println("session id: ", sessionInfo.Id)
	if !sessionInfo.GetActive() {
		c.log.Error("Error occured while getting session info for session id - "+accessToken+"+\nError - ", err.Error())
		return ctx, status.Error(codes.Unauthenticated, "session id is not active")
	}
	ctx = context.WithValue(ctx, domain.Ory("SESSION_ID"), sessionInfo.Id)
	ctx = context.WithValue(ctx, domain.Ory("ORY_ID"), sessionInfo.GetIdentity().Id)
	return ctx, nil
}
func (c *Client) GetTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "Failed to get metadata from context")
	}
	bearerToken := md.Get("authorization")
	if len(bearerToken) == 0 || len(strings.Split(bearerToken[0], " ")) != 2 {
		return "", status.Error(codes.Unauthenticated, "No access token provided")
	}
	accessToken := bearerToken[0]
	if len(accessToken) < 8 || accessToken[:7] != "Bearer " {
		return "", status.Error(codes.Unauthenticated, "Invalid access token")
	}
	return accessToken[7:], nil
}

func (c *Client) GetOryIDFromContext(ctx context.Context) (string, error) {
	oryID, ok := ctx.Value(domain.Ory("ORY_ID")).(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "Failed to get ory id from context")
	}
	return oryID, nil
}

func (c *Client) GetOrgIdFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "Failed to get metadata from context")
	}
	orgid := md.Get("organisationid")
	if len(orgid) == 0 {
		return "", status.Error(codes.Unauthenticated, "No organisation id provided")
	}
	return orgid[0], nil
}
func (c *Client) NewOAuthClientConfig(clientName string, redirectURI string, opts ...domain.OAuthClientOption) *domain.OAuthClientConfig {
	config := &domain.OAuthClientConfig{
		ClientName:              clientName,
		RedirectURIs:            []string{redirectURI},
		GrantTypes:              []string{"authorization_code", "refresh_token", "client_credentials"},
		ResponseTypes:           []string{"code", "id_token", "token"},
		TokenEndpointAuthMethod: "client_secret_post",
		Scope:                   "openid email offline",
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
func (c *Client) WithGrantTypes(grantTypes []string) domain.OAuthClientOption {
	return func(config *domain.OAuthClientConfig) {
		config.GrantTypes = grantTypes
	}
}

func (c *Client) WithResponseTypes(responseTypes []string) domain.OAuthClientOption {
	return func(config *domain.OAuthClientConfig) {
		config.ResponseTypes = responseTypes
	}
}

func (c *Client) WithTokenEndpointAuthMethod(tokenEndpointAuthMethod string) domain.OAuthClientOption {
	return func(config *domain.OAuthClientConfig) {
		config.TokenEndpointAuthMethod = tokenEndpointAuthMethod
	}
}

func (c *Client) WithScope(scope string) domain.OAuthClientOption {
	return func(config *domain.OAuthClientConfig) {
		config.Scope = scope
	}
}
func (c *Client) CreateOAuthClient(ctx context.Context, config *domain.OAuthClientConfig) (*ory.OAuth2Client, error) {
	oAuth2Client := ory.OAuth2Client{
		ClientName:              c.stringToPtr(config.ClientName),
		RedirectUris:            config.RedirectURIs,
		GrantTypes:              config.GrantTypes,
		ResponseTypes:           config.ResponseTypes,
		TokenEndpointAuthMethod: c.stringToPtr(config.TokenEndpointAuthMethod),
		Scope:                   c.stringToPtr(config.Scope),
	}

	oryAuthedContext := context.WithValue(ctx, ory.ContextAccessToken, c.oryPAT)
	resp, r, err := c.conn.OAuth2Api.CreateOAuth2Client(oryAuthedContext).OAuth2Client(oAuth2Client).Execute()
	if err != nil {
		if r != nil {
			switch r.StatusCode {
			case http.StatusConflict:
				return nil, status.Error(codes.AlreadyExists, "OAuth client already exists")
			default:
				return nil, status.Error(codes.Internal, fmt.Sprintf("Error when calling `OAuth2Api.CreateOAuth2Client`: %v. Full HTTP response: %v", err, r))
			}
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unexpected error: %v", err))
	}

	return resp, nil
}

func (c *Client) stringToPtr(s string) *string {
	return &s
}
func (c *Client) GetOryTokenUrl() string {
	tokenUrl := c.oryURL + "/oauth2/token"
	return tokenUrl
}

func (c *Client) IntrospectToken(ctx context.Context, token string) (bool, error) {
	oryAuthedContext := context.WithValue(context.Background(), ory.ContextAccessToken, c.oryPAT)
	introspect, _, err := c.conn.OAuth2Api.IntrospectOAuth2Token(oryAuthedContext).Token(token).Scope("").Execute()
	if err != nil {
		return false, err
	}
	return introspect.Active, nil
}
func (c *Client) GetOauthTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "Failed to get metadata from context")
	}
	OauthToken := md.Get("oauth_token")
	if len(OauthToken) == 0 {
		return "", status.Error(codes.Unauthenticated, "No oauth token provided")
	}
	return OauthToken[0], nil
}

func (c *Client) NewClientCredentialsConfig(clientName string) *domain.ClientCredentialConfig {
	config := &domain.ClientCredentialConfig{
		ClientName:              clientName,
		GrantTypes:              []string{"client_credentials"},
		TokenEndpointAuthMethod: "client_secret_post",
		Scope:                   "openid email offline",
	}
	return config
}

func (c *Client) CreateOAuthClientCredential(ctx context.Context, config *domain.ClientCredentialConfig) (*ory.OAuth2Client, error) {
	oAuth2Client := ory.OAuth2Client{
		ClientName:              c.stringToPtr(config.ClientName),
		GrantTypes:              config.GrantTypes,
		TokenEndpointAuthMethod: c.stringToPtr(config.TokenEndpointAuthMethod),
		Scope:                   c.stringToPtr(config.Scope),
	}

	oryAuthedContext := context.WithValue(ctx, ory.ContextAccessToken, c.oryPAT)
	resp, r, err := c.conn.OAuth2Api.CreateOAuth2Client(oryAuthedContext).OAuth2Client(oAuth2Client).Execute()
	if err != nil {
		if r != nil {
			switch r.StatusCode {
			case http.StatusConflict:
				return nil, status.Error(codes.AlreadyExists, "OAuth client credential with the given name already exists")
			default:
				return nil, status.Error(codes.Internal, fmt.Sprintf("Error when calling `OAuth2Api.CreateOAuth2Client`: %v. Full HTTP response: %v", err, r))
			}
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unexpected error: %v", err))
	}

	return resp, nil
}

func (c *Client) DeleteOAuthClient(ctx context.Context, clientID string) error {
	c.log.Debug("DeleteOAuthClient func invoked...")
	oryAuthedContext := context.WithValue(ctx, ory.ContextAccessToken, c.oryPAT)
	_, err := c.conn.OAuth2Api.DeleteOAuth2Client(oryAuthedContext, clientID).Execute()
	if err != nil {
		return errors.WithMessage(err, "error occured while deleting Oauth client from ory")
	}
	return nil
}

func (c *Client) ListOAuth2Clients(ctx context.Context, serviceName string) ([]ory.OAuth2Client, error) {
	oryAuthedContext := context.WithValue(ctx, ory.ContextAccessToken, c.oryPAT)
	clients, _, err := c.conn.OAuth2Api.ListOAuth2Clients(oryAuthedContext).ClientName(serviceName).Execute()
	if err != nil {
		return nil, errors.WithMessage(err, "error occured while listing oauth clients")
	}
	return clients, nil
}

func (c *Client) PatchIdentity(ctx context.Context, oryid, firstName, lastName string) error {
	oryAuthedContext := context.WithValue(ctx, ory.ContextAccessToken, c.oryPAT)

	_, _, err := c.conn.IdentityApi.
		PatchIdentity(oryAuthedContext, oryid).
		JsonPatch([]ory.JsonPatch{{Op: "replace", Path: "/traits/firstName", Value: firstName}}).Execute()
	if err != nil {
		return err
	}

	_, _, err = c.conn.IdentityApi.
		PatchIdentity(oryAuthedContext, oryid).
		JsonPatch([]ory.JsonPatch{{Op: "replace", Path: "/traits/lastName", Value: lastName}}).Execute()
	if err != nil {
		return err
	}

	return nil
}
