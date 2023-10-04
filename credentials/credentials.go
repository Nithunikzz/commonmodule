package credential

import (
	"context"

	"github.com/intelops/go-common/credentials"
	"github.com/pkg/errors"
)

const serviceClientOAuthEntityName = "service-client-oauth"

func GetServiceUserCredential(ctx context.Context, svcEntity, userName string) (cred credentials.ServiceCredential, err error) {
	credReader, err := credentials.NewCredentialReader(ctx)
	if err != nil {
		err = errors.WithMessage(err, "error in initializing credential reader")
		return
	}

	cred, err = credReader.GetServiceUserCredential(context.Background(), svcEntity, userName)
	if err != nil {
		err = errors.WithMessagef(err, "error in reading credential for %s/%s", svcEntity, userName)
	}
	return
}

func StoreServiceOAuthCredential(ctx context.Context, serviceName, clientId, clientSecret string) error {
	credAdmin, err := credentials.NewCredentialAdmin(ctx)
	if err != nil {
		return errors.WithMessage(err, "error in initializing credential admin")
	}

	cred := map[string]string{
		"CLIENT_ID":     clientId,
		"CLIENT_SECRET": clientSecret,
	}

	err = credAdmin.PutCredential(ctx, credentials.GenericCredentialType,
		serviceClientOAuthEntityName, serviceName, cred)
	if err != nil {
		return errors.WithMessagef(err, "error in stroing credential for %s/%s", serviceClientOAuthEntityName, serviceName)
	}
	return nil
}
