package secrets

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	"fmt"
	"github.com/mousybusiness/googlecloudgo/pkg/proj"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"log"
)

// get secret from Google Cloud Secret Manager
func GetSecret(secretId string) (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Println("failed to setup client:", err)
		return "", err
	}

	// AppEngine environment variable only exposes projectID and not projectNumber which is required by secrets manager
	projectNumber, err := proj.GetProjectNumber()
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%d/secrets/%s/versions/%d", projectNumber, secretId, 1),
	}

	// get secret
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		log.Println("failed to access secret version:", err)
		return "", err
	}

	return string(result.Payload.Data), nil
}
