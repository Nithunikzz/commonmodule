package cerbos

import (
	"context"

	"github.com/cerbos/cerbos/client"
	cerbosclient "github.com/cerbos/cerbos/client"
	credential "github.com/common/commonmodule/credentials"
	"github.com/common/commonmodule/domain"
	"github.com/intelops/go-common/logging"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type CerbosEnv struct {
	CerbosUrl        string `envconfig:"CERBOS_URL" required:"true"`
	CerbosUsername   string `envconfig:"CERBOS_USERNAME" required:"true"`
	CerbosEntityName string `envconfig:"CERBOS_ENTITY_NAME" required:"true"`
}

type Client struct {
	Client            cerbosclient.Client
	CerbosAdminClient cerbosclient.AdminClient
	log               logging.Logger
}

func GetCerbosEnv() (*CerbosEnv, error) {
	cfg := &CerbosEnv{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewCerbosClient(log logging.Logger) (*Client, error) {
	cfg, err := GetCerbosEnv()
	if err != nil {
		return nil, err
	}
	cli, err := cerbosclient.New(cfg.CerbosUrl, cerbosclient.WithPlaintext())
	if err != nil {
		return nil, errors.WithMessage(err, "unable to create cerbos client")
	}

	serviceCredential, err := credential.GetServiceUserCredential(context.Background(),
		cfg.CerbosEntityName, cfg.CerbosUsername)
	if err != nil {
		return nil, err
	}

	admincli, err := cerbosclient.NewAdminClientWithCredentials(cfg.CerbosUrl,
		cfg.CerbosUsername, serviceCredential.Password, cerbosclient.WithPlaintext())
	if err != nil {
		return nil, errors.WithMessage(err, "unable to create cerbos adminclient")
	}
	return &Client{
		Client:            cli,
		CerbosAdminClient: admincli,
		log:               log,
	}, nil

}

func (c *Client) ListPolicies(ctx context.Context) (*[]string, error) {
	policies, err := c.CerbosAdminClient.ListPolicies(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to get the policies lists")
	}

	return &policies, nil
}

func (c *Client) AddOrUpdatePolicy(ctx context.Context, data *domain.CerbosPolicy) error {

	for _, data := range data.Policies {
		for _, policy := range data.ResourcePolicy.Rules {
			ps := cerbosclient.PolicySet{}
			actions := policy.Actions
			rr1 := client.NewAllowResourceRule(actions...).WithRoles(policy.Roles)
			resource := data.ResourcePolicy.Resource
			resourcePolicy := client.NewResourcePolicy(resource, "default").AddResourceRules(rr1)
			policySet := ps.AddResourcePolicies(resourcePolicy)
			err := c.CerbosAdminClient.AddOrUpdatePolicy(ctx, policySet)
			if err != nil {
				return errors.WithMessage(err, "Failed to add or update policy")
			}
		}
	}

	return nil
}
