package cache

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

var (
	pool        *redis.Pool
	cacheFailed = true
)

// get value from redis
func Get(key string) ([]byte, error) {
	conn := pool.Get()
	defer conn.Close()

	var data []byte

	r, err := redis.DoWithTimeout(conn, time.Millisecond*100, "GET", key)
	if err != nil {
		return nil, err
	}

	data, err2 := redis.Bytes(r, err)
	if err2 != nil {
		return data, fmt.Errorf("error getting key %s: %v", key, err2)
	}

	return data, nil
}

// set value to redis
func Set(key string, value []byte) error {
	conn := pool.Get()
	defer conn.Close()

	r, err := redis.DoWithTimeout(conn, time.Millisecond*100, "SET", key, value)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		return fmt.Errorf("error setting key %s to %s: %v", key, v, err)
	}

	log.Println("redis response", r)
	return err
}

func Ping(c redis.Conn) error {
	s, err := redis.String(c.Do("PING"))
	if err != nil {
		return err
	}

	log.Println("ping result", s)
	return nil
}

func IncrementCounter(key string) (int, error) {
	if cacheFailed {
		log.Println("cache has already failed, skipping")
		return -1, errors.New("cache isn't up")
	}

	conn := pool.Get()
	defer conn.Close()

	counter, err := redis.Int(redis.DoWithTimeout(conn, time.Millisecond*50, "INCR", key))
	if err != nil {
		log.Println("error while doing with timeout", err)
		cacheFailed = true
		return -1, err
	}

	return counter, nil
}

// if cache has failed (or first run) poll for redis instance information
func ValidateMemoryStoreAndCreatePool() {
	if cacheFailed {
		instance, err := FindCacheInstance()
		if err != nil || instance.Host == "" || instance.Port == 0 {
			log.Println("FAILED TO GET REDIS!", err)
			cacheFailed = true
		} else {
			log.Println("REDIS IS LIVE!")
			cacheFailed = false
			redisAddr := fmt.Sprintf("%s:%d", instance.Host, instance.Port)
			log.Println("using redis address:", redisAddr)
			pool = NewPool(redisAddr)
		}
	}
}

// create redis connection pools
func NewPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr, redis.DialConnectTimeout(time.Millisecond*100))
		},
	}
}
