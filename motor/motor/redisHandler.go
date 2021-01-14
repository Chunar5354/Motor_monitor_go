package motor

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background() // used in redis

type RedisInfo struct {
	password string
	port     string
	db       int
}

var redisInfo RedisInfo

// set redis information by user
func RedisInit(password, port string, db int) {
	redisInfo.password = password
	redisInfo.port = port
	redisInfo.db = db
}

// fetch data from redis
func redisFetch(rdb *redis.Client, serialNumber, para, createTime string) []string {
	key := serialNumber + "_" + para + "_" + createTime
	data, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil
	}
	return strings.Split(data, "/")
}

func redisUpload(req map[string]string, serial_number, create_time string) error {
	// set for redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:" + redisInfo.port,
		Password: redisInfo.password, // no password set
		DB:       redisInfo.db,       // use default DB
	})
	defer rdb.Close()
	err := rdb.Set(ctx, serial_number+"lastTime", create_time, 0).Err()
	if err != nil {
		return err
	}
	for para, data := range req {
		key := serial_number + "_" + para + "_" + create_time
		err := rdb.Set(ctx, key, data[:len(data)-1], 2*time.Minute).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// create redis connection and call detchData function to get data from redis
func haldleRedis(serialNumber, startTime string, parameters []string) string {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:" + redisInfo.port,
		Password: redisInfo.password, // no password set
		DB:       redisInfo.db,       // use default DB
	})
	createTime := startTime
	if createTime == "" {
		// if doesn't give a time, get the newest data, firstly get last time
		lastTime, err := rdb.Get(ctx, serialNumber+"lastTime").Result()
		if err == redis.Nil {
			log.Println("No result in redis")
			return ""
		}
		createTime = lastTime
	}
	m := make(map[string]interface{})
	m["create_time"] = createTime
	for _, para := range parameters {
		redisResult := redisFetch(rdb, serialNumber, para, createTime)
		if redisResult == nil {
			log.Println("No result in redis:", para)
			return ""
		}
		m[para] = redisResult
	}
	response, err := json.Marshal(m)
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	return string(response)
}
