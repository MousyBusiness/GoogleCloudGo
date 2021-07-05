// The ds pgk creates a simple wrapper around common
// Datastore functions to facilitate unit testing
package ds

import (
	"cloud.google.com/go/datastore"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"log"
	"os"
	"reflect"
)

type DatastoreClient interface {
	Client() *datastore.Client
	Create(ctx context.Context, parent *datastore.Key, entity Entity) (*datastore.Key, error)
	Update(ctx context.Context, parent *datastore.Key, id int64, entity Entity) (*datastore.Key, error)
	Get(ctx context.Context, id int64, parent *datastore.Key, entity Entity) error
	Delete(ctx context.Context, kind string, id int64, parent *datastore.Key) error
	CreateNamed(ctx context.Context, name string, parent *datastore.Key, entity Entity) (*datastore.Key, error)
	GetNamed(ctx context.Context, name string, parent *datastore.Key, entity Entity) error
	DeleteNamed(ctx context.Context, kind string, name string, parent *datastore.Key) error
	QGet(ctx context.Context, kind string, property string, value string, entity Entity) (*datastore.Key, error)
	QueryParent(ctx context.Context, kind string, parent *datastore.Key, entitySlicePtr interface{}) ([]*datastore.Key, error)
}

type Client struct {
	ds *datastore.Client
}

// value must be a pointer to a struct
type Entity interface {
	GetKind() string
	GetValue() interface{}
}

// ConnectToDatastore will establish a connection to Datastore and wrap the
// client in a DatastoreClient instance
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

func (client Client) Client() *datastore.Client {
	return client.ds
}

// Create will create a single Entity and return a generated key
func (client Client) Create(ctx context.Context, parent *datastore.Key, entity Entity) (*datastore.Key, error) {
	key := datastore.IncompleteKey(entity.GetKind(), parent)
	key, err := client.ds.Put(ctx, key, entity.GetValue())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create datastore entity")
	}
	return key, nil
}

// Update will update a single Entity using its generated key and parent (if exists)
func (client Client) Update(ctx context.Context, parent *datastore.Key, id int64, entity Entity) (*datastore.Key, error) {
	key := datastore.IDKey(entity.GetKind(), id, parent)

	key, err := client.ds.Put(ctx, key, entity.GetValue())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create datastore entity")
	}
	return key, nil
}

// Get will get a single Entity using its generated key and parent (if exists)
func (client Client) Get(ctx context.Context, id int64, parent *datastore.Key, entity Entity) error {
	if entity.GetValue() == nil {
		return errors.New("entity.GetValue cannot return nil")
	}
	key := datastore.IDKey(entity.GetKind(), id, parent)
	err := client.ds.Get(ctx, key, entity.GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to get datastore entity")
	}

	return nil
}

// Delete will delete a single Entity using its generated key and parent (if exists)
func (client Client) Delete(ctx context.Context, kind string, id int64, parent *datastore.Key) error {
	key := datastore.IDKey(kind, id, parent)
	err := client.ds.Delete(ctx, key)
	if err != nil {
		return errors.Wrap(err, "failed to delete datastore entity")
	}

	return nil
}

// CreateNamed will create a single Entity using its name key and parent (if exists)
func (client Client) CreateNamed(ctx context.Context, name string, parent *datastore.Key, entity Entity) (*datastore.Key, error) {
	key := datastore.NameKey(entity.GetKind(), name, parent)
	key, err := client.ds.Put(ctx, key, entity.GetValue())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create named datastore entity")
	}
	return key, nil
}

// GetNamed will get a single Entity using its name key and parent (if exists)
func (client Client) GetNamed(ctx context.Context, name string, parent *datastore.Key, entity Entity) error {
	if entity.GetValue() == nil {
		return errors.New("entity.GetValue cannot return nil")
	}
	key := datastore.NameKey(entity.GetKind(), name, parent)
	err := client.ds.Get(ctx, key, entity.GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to get named datastore entity")
	}

	return nil
}

// DeleteNamed will delete a single named Entity using its name key and parent (if exists)
func (client Client) DeleteNamed(ctx context.Context, kind string, name string, parent *datastore.Key) error {
	key := datastore.NameKey(kind, name, parent)
	err := client.ds.Delete(ctx, key)
	if err != nil {
		return errors.Wrap(err, "failed to delete named datastore entity")
	}

	return nil
}

// QGet returns a single Entity using a property and value combination
func (client Client) QGet(ctx context.Context, kind string, property string, value string, entity Entity) (*datastore.Key, error) {
	query := datastore.NewQuery(kind).Filter(fmt.Sprintf("%s =", property), value)
	it := client.ds.Run(ctx, query)
	key, err := it.Next(entity)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// QueryParent will get all Entities of a Kind with the same parent
// entitySlicePtr must be a pointer to a slice of structs e.g. &[]struct{}
// returns a slice of keys which relate to the returned entities
func (client Client) QueryParent(ctx context.Context, kind string, parent *datastore.Key, entitySlicePtr interface{}) ([]*datastore.Key, error) {
	p := reflect.ValueOf(entitySlicePtr)
	slice := p.Elem()
	elemType := slice.Type().Elem()

	query := datastore.NewQuery(kind).Ancestor(parent)
	it := client.ds.Run(ctx, query)

	var keys []*datastore.Key
	for {
		v := reflect.New(elemType)
		key, err := it.Next(v.Interface())
		if err != nil {
			if err == iterator.Done {
				return keys, nil
			}
			return keys, err
		}

		keys = append(keys, key)
		slice.Set(reflect.Append(slice, v.Elem()))
	}
}
