package ds

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/pkg/errors"
	"log"
	"os"
)

type DatastoreClient interface {
	Create(ctx context.Context, entity Entity) (*datastore.Key, error)
	Get(ctx context.Context, id int64, entity Entity) error
	Delete(ctx context.Context, kind string, id int64) error
}

type Client struct {
	kind string
	ds   *datastore.Client
}

// value must be a pointer to a struct
type Entity interface {
	GetValue() interface{}
}

func ConnectToDatastore(ctx context.Context, kind string) (DatastoreClient, error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT") // environment variable provided by app engine

	// Creates a client.
	c, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new datastore client")
	}

	log.Println("datastore client created")
	return Client{kind: kind, ds: c}, nil
}

func (client Client) Create(ctx context.Context, entity Entity) (*datastore.Key, error) {
	key := datastore.IncompleteKey(client.kind, nil)
	key, err := client.ds.Put(ctx, key, entity.GetValue())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create datastore entity")
	}
	return key, nil
}

func (client Client) Get(ctx context.Context, id int64, entity Entity) error {
	taskKey := datastore.IDKey(client.kind, id, nil)
	err := client.ds.Get(ctx, taskKey, entity)
	if err != nil {
		return errors.Wrap(err, "failed to get datastore entity")
	}

	return nil
}

func (client Client) Delete(ctx context.Context, kind string, id int64) error {
	key := datastore.IDKey(client.kind, id, nil)
	err := client.ds.Delete(ctx, key)
	if err != nil {
		return errors.Wrap(err, "failed to delete datastore entity")
	}

	return nil
}
