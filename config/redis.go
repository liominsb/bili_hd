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
		MinIdleConns: 1,                         //设置最小空闲连接数为3
		PoolSize:     10,                        //设置连接池大小为10
	})

	_, err := RedisCilnet.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	global.RedisDB = RedisCilnet
}
