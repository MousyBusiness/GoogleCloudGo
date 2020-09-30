package cache

import (
	redisman "cloud.google.com/go/redis/apiv1beta1"
	"context"
	"errors"
	"fmt"
	"github.com/mousybusiness/googlecloudgo/pkg/bq"
	redispb "google.golang.org/genproto/googleapis/cloud/redis/v1beta1"
	"log"
	"os"
	"strings"
)

var (
	redisVersion = os.Getenv("REDIS_VERSION")
	projectId    = os.Getenv("GOOGLE_CLOUD_PROJECT")
	region       = os.Getenv("REGION")
	instanceId   = os.Getenv("MEMORYSTORE_INSTANCE")
)

// creates the Memorystore Redis instance
func CreateCacheInstance() error {
	ctx := context.Background()
	client, err := redisman.NewCloudRedisClient(ctx)
	if err != nil {
		return err
	}

	template := redispb.Instance{
		Name:         fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, region, instanceId),
		DisplayName:  instanceId,
		RedisVersion: redisVersion,
		Tier:         redispb.Instance_BASIC,
		MemorySizeGb: 1,
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", projectId, region)
	log.Println("redis instance parent")

	cio, err := client.CreateInstance(ctx, &redispb.CreateInstanceRequest{
		Parent:     parent,
		InstanceId: instanceId,
		Instance:   &template,
	})
	if err != nil {
		log.Println("error while trying to create cache", err)
		return err
	}

	instance, err := cio.Wait(ctx)
	if err != nil {
		log.Println("error while waiting for create", err)
		return err
	}

	log.Println("REDIS", instance.GetHost())

	return nil
}

// deletes the Memorystore Redis instance
func DeleteCacheIstance() error {
	ctx := context.Background()
	client, err := redisman.NewCloudRedisClient(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, region, instanceId)
	log.Println("instance name", name)
	dio, err := client.DeleteInstance(ctx, &redispb.DeleteInstanceRequest{
		Name: name,
	})

	if err != nil {
		log.Println("error while trying to delete cache:", err)
		return err
	}

	if err := dio.Wait(ctx); err != nil {
		log.Println("error while waiting for delete:", err)
		return err
	}

	return nil
}

// using redis sdk try and find instance details i.e. host ip and port (instance might not exist!)
func FindCacheInstance() (*redispb.Instance, error) {
	ctx := context.Background()
	client, err := redisman.NewCloudRedisClient(ctx)
	if err != nil {
		return nil, err
	}

	instances := client.ListInstances(ctx, &redispb.ListInstancesRequest{
		Parent:   fmt.Sprintf("projects/%s/locations/%s", projectId, region),
		PageSize: 1,
	})

	if instances == nil {
		return nil, errors.New("instances is nil")
	}

	next, err := instances.Next()
	if err != nil {
		return nil, err
	}

	log.Println(next)

	if next.GetDisplayName() != instanceId {
		log.Println("instance id doesnt match required, received:", next.GetDisplayName(), "wanted:", instanceId)
		return nil, errors.New("instance id doesnt match required")
	}

	return next, nil
}

// query BigQuery for all rows in table and put into redis (preloading the cache)
// keyColumn is the column number which you want to use for your redis key for each row
func PreloadCache(keyColumn int) error {
	ValidateMemoryStoreAndCreatePool() // duplicated from client - cant find a good way to unify without creating an external helper

	rows, err := bq.QueryBQ(bq.BuildQuery(bq.DefaultFrom(), nil, ""))
	if err != nil {
		log.Println("error while querying BQ", err)
		return err
	}

	if err := putInCache(keyColumn, rows); err != nil {
		log.Println("Putting in cache failed", err)
		return err
	}

	return nil
}

// load redis cache with all the rows received from BQ
func putInCache(keyColumn int, rows []string) error {
	log.Println("executing putInCache")

	if cacheFailed {
		log.Println("previous cache attempts have failed")
		return errors.New("cache isn't reachable")
	}

	// could be improved in speed by using MSET
	// could be improved by using gob encoding
	for _, v := range rows {
		if err := Set(strings.Split(v, ",")[keyColumn], []byte(v)); err != nil {
			return errors.New("FAILED TO SET TO REDIS")
		}
	}

	return nil
}
