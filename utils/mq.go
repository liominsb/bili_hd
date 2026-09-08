package utils

import (
	"context"
	"encoding/json"
	"go_bili/global"
	"go_bili/models"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Worker(ctx context.Context) {
	ch, err := global.MQConn.Channel()
	if err != nil {
		log.Printf("Worker 创建信道失败: %v", err)
		return
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
		log.Printf("Worker 启动监听失败: %v", err)
		return
	}

	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			log.Println("收到退出信号，停止消费新消息，等待转码任务收尾...")
			wg.Wait() // 等已经启动的转码任务跑完 remove/rename + ack
			return
		case d, ok := <-msgs:
			if !ok {
				wg.Wait()
				return
			}
			wg.Add(1)
			go func(d amqp.Delivery) {
				defer wg.Done()
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

}
