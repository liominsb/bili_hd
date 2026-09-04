package utils

import (
	"encoding/json"
	"go_bili/global"
	"go_bili/models"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Worker() {
	ch, err := global.MQConn.Channel()
	if err != nil {
		log.Fatalf("Worker 创建信道失败: %v", err)
	}
	defer ch.Close()

	_ = ch.Qos(2, 0, false) // 限制并发为 2
	msgs, err := ch.Consume(
		"video_queue",
		"",
		false, // 手动 ACK
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Worker 启动监听失败: %v", err)
	}

	for d := range msgs {
		go func(d amqp.Delivery) {
			var task models.VideoTranscodeMsg
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				log.Printf("消息反序列化失败，直接丢弃坏数据: %v", err)
				_ = d.Nack(false, false)
				return
			}
			log.Printf("收到视频 %s 转码任务，文件路径: %s", task.FName, task.FilePath)
			err = Faststart(task.FilePath)
			if err != nil {
				d.Nack(false, false)
				return
			}
			d.Ack(false)
		}(d)
	}
}
