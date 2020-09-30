package proj

import (
	"context"
	"google.golang.org/api/cloudresourcemanager/v1"
	"log"
	"os"
)

var (
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
)

// used to get secrets from secret manager - for some reason projectId doesnt work, might be a security thing so resource id cant be guessed
// expecting env var GOOGLE_CLOUD_PROJECT - AppEngine runtime variable
func GetProjectNumber() (int64, error) {
	cloudresourcemanagerService, err := cloudresourcemanager.NewService(context.Background())
	if err != nil {
		//log.Fatalf("NewService: %v", err)
		return 0, err
	}

	project, err := cloudresourcemanagerService.Projects.Get(projectID).Do()
	if err != nil {
		//log.Fatalf("Get project: %v", err)
		return 0, err
	}

	log.Printf("Project number for project %s: %d\n", projectID, project.ProjectNumber)
	return project.ProjectNumber, nil
}
