package ds

import (
	"cloud.google.com/go/datastore"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
)

type DatastoreClient interface {
	Create(ctx context.Context, entity Entity) (*datastore.Key, error)
	Get(ctx context.Context, id int64, entity Entity) error
	Delete(ctx context.Context, kind string, id int64) error
	QGet(ctx context.Context, kind string, property string, value string, entity Entity) (*datastore.Key, error)
}

type Client struct {
	ds *datastore.Client
}

// value must be a pointer to a struct
type Entity interface {
	GetKind() string
	GetValue() interface{}
}

func ConnectToDatastore(ctx context.Context) (DatastoreClient, error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT") // environment variable provided by app engine

	// Creates a client.
	c, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new datastore client")
	}

	log.Println("datastore client created")
	return Client{ds: c}, nil
}

func (client Client) Create(ctx context.Context, entity Entity) (*datastore.Key, error) {
	key := datastore.IncompleteKey(entity.GetKind(), nil)
	key, err := client.ds.Put(ctx, key, entity.GetValue())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create datastore entity")
	}
	return key, nil
}

func (client Client) Get(ctx context.Context, id int64, entity Entity) error {
	taskKey := datastore.IDKey(entity.GetKind(), id, nil)
	err := client.ds.Get(ctx, taskKey, entity)
	if err != nil {
		return errors.Wrap(err, "failed to get datastore entity")
	}

	return nil
}

func (client Client) Delete(ctx context.Context, kind string, id int64) error {
	key := datastore.IDKey(kind, id, nil)
	err := client.ds.Delete(ctx, key)
	if err != nil {
		return errors.Wrap(err, "failed to delete datastore entity")
	}

	return nil
}

func (client Client) QGet(ctx context.Context, kind string, property string, value string, entity Entity) (*datastore.Key, error) {
	query := datastore.NewQuery(kind).Filter(fmt.Sprintf("%s =", property), value)
	it := client.ds.Run(ctx, query)
	key, err := it.Next(entity)
	if err != nil {
		return nil, err
	}
	return key, nil
}
