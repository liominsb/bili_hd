package config

import (
	"context"
	"go_bili/global"
	"log"

	"github.com/redis/go-redis/v9"
)

const maxTotal = 100
const minIdle = maxTotal * 0.5

func initRedis(ctx context.Context) {
	RedisCilnet := redis.NewClient(&redis.Options{
		Addr:         Appconf.Database.Addr,
		Password:     Appconf.Database.Password, // no password set
		DB:           0,                         // use default DB
		MinIdleConns: minIdle,                   //设置最小空闲连接数为20
		PoolSize:     maxTotal,                  //设置连接池大小为100
	})

	_, err := RedisCilnet.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	global.RedisDB = RedisCilnet
}
