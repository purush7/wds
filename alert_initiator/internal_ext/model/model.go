package model

import (
	"alert_system/alert_initiator/internal_ext/entities"
	"alert_system/constants"
	"alert_system/infra/redis"
	"log"
)

func WriteIntoDb(id, filepath string) {
	DbStore, err := redis.RetrieveMapFromRedis(constants.RedisDbStoreKey)
	if err != nil && err != redis.Nil {
		log.Println("got error: %s while retrieving the map from db", err)
	}
	if err == redis.Nil {
		DbStore = make(map[string]interface{}, 0)
	}
	DbStore[id] = filepath
	log.Println("write into db: ", id, filepath)
	err = redis.StoreMapInRedis(constants.RedisDbStoreKey, DbStore)
	if err != nil {
		log.Printf("got error: %s while storing the map from db\n", err)
	}
}

func GetBatchFromDb() []entities.BatcherPayload {
	DbStore, err := redis.RetrieveMapFromRedis(constants.RedisDbStoreKey)
	if err != nil && err != redis.Nil {
		log.Println("got error: %s while retrieving the map from db", err)
	}
	index := int64(0)
	batchPayload := make([]entities.BatcherPayload, 0, constants.BatchLimit)
	if err == redis.Nil {
		return nil
	}
	for k := range DbStore {
		if index == constants.BatchLimit {
			break
		}
		batchPayload = append(batchPayload, entities.BatcherPayload{
			Id:       k,
			FilePath: DbStore[k].(string),
		})
		log.Println("write batch: ", index, k, DbStore[k])
		index++
	}
	return batchPayload
}

func Delete(id string) {
	DbStore, err := redis.RetrieveMapFromRedis(constants.RedisDbStoreKey)
	if err != nil && err != redis.Nil {
		log.Println("got error: %s while retrieving the map from db", err)
	}
	if err == redis.Nil {
		return
	}
	delete(DbStore, id)
	log.Println("dbstore: ", DbStore)
	err = redis.StoreMapInRedis(constants.RedisDbStoreKey, DbStore)
	if err != nil {
		log.Printf("got error: %s while storing the map from db\n", err)
	}
}
