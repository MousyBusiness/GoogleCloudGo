package secrets

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	"fmt"
	"github.com/mousybusiness/googlecloudgo/pkg/proj"
	errs "github.com/pkg/errors"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// get secret from Google Cloud Secret Manager
func GetSecret(secretId string) (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", errs.Wrap(err, "failed to setup client")
	}

	// AppEngine environment variable only exposes projectID and not projectNumber which is required by secrets manager
	projectNumber, err := proj.GetProjectNumber()
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%d/secrets/%s/versions/%d", projectNumber, secretId, 1),
	}

	// get secret
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", errs.Wrap(err, "failed to access secret version")
	}

	return string(result.Payload.Data), nil
}
