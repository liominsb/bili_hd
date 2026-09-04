package config

import (
	"go_bili/global"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func InitRabbitMQ() {
	conn, err := amqp.Dial(Appconf.RabbitMQ.Url)
	if err != nil {
		log.Fatalf("RabbitMQ 连接失败: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("RabbitMQ 打开 Channel 失败: %v", err)
	}
	// 声明持久化队列
	_, err = ch.QueueDeclare(
		"video_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	ch.Close()
	if err != nil {
		log.Fatalf("队列声明失败: %v", err)
	}

	global.MQConn = conn
}
