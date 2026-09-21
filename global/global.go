package global

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/gorm"
)

var (
	Db         *gorm.DB //database 数据库
	RedisDB    *redis.Client
	MQConn     *amqp091.Connection
	CosClient  *cos.Client
	CosBaseURL string
)
