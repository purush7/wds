package redis

import (
	"context"
	"fmt"
	"time"

	"log"

	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rdb *redis.Client
	Nil = redis.Nil
)

func init() {
	endpoint := "redis:6379"
	log.Printf("Endpoint:%s \n", endpoint)

	rdb = redis.NewClient(&redis.Options{
		Addr:         endpoint,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     20,
		PoolTimeout:  30 * time.Second,
		DB:           2,
	})
}

func GetClient() *redis.Client {
	return rdb
}

func formatCacheKey(key string) string {
	return fmt.Sprintf("%v-swilly", key)
}

func Set(key string, value string, ttl time.Duration) error {
	key = formatCacheKey(key)
	err := rdb.Set(ctx, key, value, ttl).Err()

	if err != nil {
		return err
	}

	return nil
}

func Incr(key string) error {
	key = formatCacheKey(key)
	err := rdb.Incr(ctx, key).Err()
	return err
}

func Get(key string) (string, error) {
	key = formatCacheKey(key)
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func Delete(key string) error {
	key = formatCacheKey(key)
	_, err := rdb.Del(ctx, key).Result()
	if err != nil {
		// Alert
		log.Printf("alert error while deleting the key : %v\n", err)
		return err
	}
	return nil
}

// StoreMapInRedis stores a map in a Redis Hash
func StoreMapInRedis(hashKey string, data map[string]interface{}) error {
	ctx := context.Background()
	// Convert the map to redis.X structure
	// Use the HSet command to set multiple field-value pairs in the Hash
	rdb.HSet(ctx, hashKey, data).Err()
	return nil
}

// RetrieveMapFromRedis retrieves a map from a Redis Hash
func RetrieveMapFromRedis(hashKey string) (map[string]interface{}, error) {
	ctx := context.Background()

	// Use the HGetAll command to get all field-value pairs from the Hash
	result, err := rdb.HGetAll(ctx, hashKey).Result()
	if err != nil {
		return nil, err
	}

	log.Println("results ", result)
	// Convert the result to a map[string]string
	retrievedMap := make(map[string]interface{})
	for key, value := range result {
		retrievedMap[key] = value
	}

	return retrievedMap, nil
}
