package secrets

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	"errors"
	"fmt"
	"github.com/mousybusiness/googlecloudgo/pkg/proj"
	errs "github.com/pkg/errors"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type Secret string

// get secret from Google Cloud Secret Manager
func GetSecret(secretId string) (Secret, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", errs.Wrap(err, "failed to setup client")
	}

	// AppEngine environment variable only exposes projectID and not projectNumber which is required by secrets manager
	projectNumber, err := proj.GetProjectNumber()
	if err != nil {
		return "", err
	}
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%d/secrets/%s/versions/%d", projectNumber, secretId, 1),
	}

	// get secret
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", errs.Wrap(err, "failed to access secret version")
	}

	if string(result.Payload.Data) == "" {
		return "", errors.New("secret empty")
	}

	return Secret(result.Payload.Data), nil
}
