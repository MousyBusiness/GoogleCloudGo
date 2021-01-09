package ds

import (
	"context"
	"errors"
	dstore "google.golang.org/api/datastore/v1beta1"
	"time"
)

func Export(ctx context.Context, timeout int, project string, bucket string, kinds ...string) ([]byte, error) {
	ctxWithDeadline, cancel := context.WithTimeout(ctx, time.Minute*time.Duration(timeout))
	defer cancel()

	if len(kinds) == 0 {
		return nil, errors.New("require kind")
	}

	service, err := dstore.NewService(ctxWithDeadline)
	if err != nil {
		return nil, err
	}

	o, err := service.Projects.Export(project, &dstore.GoogleDatastoreAdminV1beta1ExportEntitiesRequest{
		EntityFilter: &dstore.GoogleDatastoreAdminV1beta1EntityFilter{
			NamespaceIds: []string{},
			Kinds:        kinds,
		},
		OutputUrlPrefix: bucket,
	}).Do()
	if err != nil {
		return nil, err
	}

	json, err := o.Response.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return json, nil
}
