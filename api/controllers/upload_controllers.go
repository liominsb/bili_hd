package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"go_bili/global"
	"go_bili/models"
	"go_bili/utils"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	uploadDir     = "uploads"
	maxUploadSize = 10000 << 20 // 10000MB
)

// 允许的文件后缀（白名单）
var allowExt = map[string]string{
	".mp4":  "video/mp4",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
}

type UploadController struct {
}

func NewUploadController() *UploadController {
	return &UploadController{}
}

func (c *UploadController) UploadFile(ctx *gin.Context) {
	// 1. 给请求体封顶，防止超大文件吃满内存
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxUploadSize)

	// 2. 取出前端传来的文件
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	// 3. 校验后缀白名单，顺带拿到规范化的后缀（小写、带点）
	ext, err := checkExt(header.Filename)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. 起个不重复的文件名，存进 uploads 目录后存入cos
	fName := newFileName(ext)
	savePath := uploadDir + "/" + fName
	path, err := saveFile(ctx, header, savePath, fName, allowExt[ext])
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if ext == ".mp4" {
		// 2. 发送 MQ 消息
		body, _ := json.Marshal(models.VideoTranscodeMsg{
			FName:    fName,
			FilePath: savePath, // 本地磁盘相对路径
		})
		ch, err := global.MQConn.Channel()
		if err != nil {
			log.Println("创建MQ通道失败:", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer ch.Close()
		ctxMQ, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = ch.PublishWithContext(ctxMQ,
			"",            // default exchange
			"video_queue", // routing key / queue name
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent, // 消息持久化
				ContentType:  "application/json",
				Body:         body,
			},
		)
		if err != nil {
			log.Println("发送MQ消息失败:", err)
			if delErr := utils.Delete(fName); delErr != nil {
				// 删除失败只记日志，不要改变返回给用户的错误
				log.Println("回滚失败，COS 上可能留下孤儿对象:", delErr)
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// 5. 把可访问地址返回给前端
	ctx.JSON(http.StatusOK, gin.H{"url": path})
}

func checkExt(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename)) // 转小写，防止 .MP4 绕过白名单
	if _, ok := allowExt[ext]; !ok {
		return "", fmt.Errorf("不支持的文件类型：%s", ext)
	}
	return ext, nil
}

func newFileName(ext string) string {
	// 时间戳当文件名，避开中文名和重名覆盖
	return fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
}

func saveFile(ctx *gin.Context, header *multipart.FileHeader, savePath string, fName string, contentType string) (string, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil { // 目录不存在就建，已存在不报错
		return "", err
	}
	err := ctx.SaveUploadedFile(header, savePath) // Gin 封装：内部就是建文件 + 拷内容
	if err != nil {
		return "", err
	}
	path, err := utils.Upload(fName, savePath, contentType)
	if err != nil {
		return "", err
	}
	return path, nil
}
