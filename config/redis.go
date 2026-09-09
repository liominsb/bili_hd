package config

import (
	"context"
	"go_bili/global"
	"log"

	"github.com/redis/go-redis/v9"
)

func initRedis(ctx context.Context) {
	RedisCilnet := redis.NewClient(&redis.Options{
		Addr:         Appconf.Database.Addr,
		Password:     Appconf.Database.Password, // no password set
		DB:           0,                         // use default DB
		MinIdleConns: 10,                        //设置最小空闲连接数为20
		PoolSize:     20,                        //设置连接池大小为100
	})

	_, err := RedisCilnet.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	global.RedisDB = RedisCilnet
}
