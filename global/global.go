package global

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	Db      *gorm.DB //database 数据库
	RedisDB *redis.Client
	MQConn  *amqp091.Connection
)
