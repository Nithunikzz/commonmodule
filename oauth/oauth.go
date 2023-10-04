package outh

import (
	"context"

	"github.com/common/commonmodule/domain"
	oauthpb "github.com/common/commonmodule/proto"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s Adapter) CreateOauthClient(ctx context.Context, in *oauthpb.OauthClientRequest) (*oauthpb.OauthClientResponse, error) {
	// Check for empty values in the request and return an error if required fields are not provided
	clientName := in.GetClientName()
	if clientName == "" {
		return nil, status.Error(codes.InvalidArgument, "client_name is required")
	}
	redirectURIs := in.GetRedirectUris()
	if len(redirectURIs) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one redirect_uri is required")
	}
	// Create a new OAuth client using the provided information
	opts := []domain.OAuthClientOption{}
	if grantTypes := in.GetGrantTypes(); len(grantTypes) > 0 {
		opts = append(opts, s.oryClient.WithGrantTypes(grantTypes))
	}
	if responseTypes := in.GetResponseTypes(); len(responseTypes) > 0 {
		opts = append(opts, s.oryClient.WithResponseTypes(responseTypes))
	}
	if tokenEndpointAuthMethod := in.GetTokenEndpointAuthMethod(); tokenEndpointAuthMethod != "" {
		opts = append(opts, s.oryClient.WithTokenEndpointAuthMethod(tokenEndpointAuthMethod))
	}
	if scope := in.GetScope(); scope != "" {
		opts = append(opts, s.oryClient.WithScope(scope))
	}
	config := s.oryClient.NewOAuthClientConfig(clientName, redirectURIs[0], opts...)
	oauthClient, err := s.oryClient.CreateOAuthClient(ctx, config) // Passing the context
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.AlreadyExists:
			return nil, status.Error(codes.AlreadyExists, "OAuth client with the given name already exists")
		case codes.Internal:
			return nil, status.Error(codes.Internal, "Internal error occurred while creating the OAuth client")
		default:
			return nil, err
		}
	}
	return &oauthpb.OauthClientResponse{
		ClientId:     *oauthClient.ClientId,
		ClientSecret: *oauthClient.ClientSecret,
	}, nil
}

func (s Adapter) GetOauthToken(ctx context.Context, in *oauthpb.OauthTokenRequest) (*oauthpb.OauthTokenResponse, error) {
	conf := &clientcredentials.Config{
		ClientID:     in.GetClientId(),
		ClientSecret: in.GetClientSecret(),
		Scopes:       []string{"openid email offline"},
		TokenURL:     s.oryClient.GetOryTokenUrl(),
	}
	at, err := conf.Token(ctx)
	if err != nil {
		s.log.Error(err.Error())
	}
	return &oauthpb.OauthTokenResponse{
		OauthToken:   at.AccessToken,
		RefreshToken: at.RefreshToken,
	}, nil
}

// ValidateOauthToken validates the provided OAuth token using the ORY client's introspection method.
// It checks the authenticity and validity of the token and returns a response indicating its status.
func (s Adapter) ValidateOauthToken(ctx context.Context, in *oauthpb.ValidateOauthTokenRequest) (*oauthpb.ValidateOauthTokenResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "Received nil request")
	}
	defer s.log.Debug("Exiting ValidateOauthToken RPC")
	s.log.Debugf("validating oauth token: %s", in.GetOauthToken())
	active, err := s.oryClient.IntrospectToken(ctx, in.GetOauthToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Error introspecting token: "+err.Error())
	}
	if !active {
		return &oauthpb.ValidateOauthTokenResponse{Valid: "no"}, nil
	}
	return &oauthpb.ValidateOauthTokenResponse{Valid: "yes"}, nil
}

func (s Adapter) CreateClientCredentialsClient(ctx context.Context, in *oauthpb.CreateClientCredentialsClientRequest) (*oauthpb.CreateClientCredentialsClientResponse, error) {
	s.log.Debug("CreateClientCredentialsClient RPC invoked...")
	defer s.log.Debug("CreateClientCredentialsClient RPC Exited...")

	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "Received nil request")
	}
	if in.ClientName == "" {
		return nil, status.Error(codes.InvalidArgument, "Provide a client name")
	}
	cc := s.oryClient.NewClientCredentialsConfig(in.ClientName)
	client, err := s.oryClient.CreateOAuthClientCredential(ctx, cc) // Passing the context
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.AlreadyExists:
			return nil, status.Error(codes.AlreadyExists, "OAuth client credential with the given name already exists")
		case codes.Internal:
			return nil, status.Error(codes.Internal, "Internal error occurred while creating the client credential OAuth client")
		default:
			return nil, err
		}
	}
	if client == nil || client.ClientId == nil || client.ClientSecret == nil {
		return nil, status.Error(codes.Internal, "Received incomplete client data from oryClient")
	}
	return &oauthpb.CreateClientCredentialsClientResponse{
		ClientId:     *client.ClientId,
		ClientSecret: *client.ClientSecret,
	}, nil
}
