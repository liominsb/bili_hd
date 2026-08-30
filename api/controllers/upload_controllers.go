package controllers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	uploadDir     = "uploads"
	maxUploadSize = 1000 << 20 // 1000MB
)

// 允许的文件后缀（白名单）
var allowExt = map[string]bool{
	".mp4": true, ".png": true, ".jpg": true, ".jpeg": true, ".webp": true,
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

	// 4. 起个不重复的文件名，存进 uploads 目录
	fName := newFileName(ext)
	savePath := uploadDir + "/" + fName
	if err := saveFile(ctx, header, savePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	faststart(savePath)
	// 5. 把可访问地址返回给前端
	ctx.JSON(http.StatusOK, gin.H{"url": "/" + uploadDir + "/" + fName})
}

func checkExt(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename)) // 转小写，防止 .MP4 绕过白名单
	if !allowExt[ext] {
		return "", fmt.Errorf("不支持的文件类型：%s", ext)
	}
	return ext, nil
}

func newFileName(ext string) string {
	// 时间戳当文件名，避开中文名和重名覆盖
	return fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
}

func saveFile(ctx *gin.Context, header *multipart.FileHeader, savePath string) error {
	if err := os.MkdirAll(uploadDir, 0755); err != nil { // 目录不存在就建，已存在不报错
		return err
	}
	return ctx.SaveUploadedFile(header, savePath) // Gin 封装：内部就是建文件 + 拷内容
}

func faststart(path string) {
	tmp := path + ".tmp"
	cmd := exec.Command("ffmpeg", "-y",
		"-i", path,
		"-c", "copy",
		"-movflags", "+faststart",
		tmp,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("faststart 失败(不影响上传): %v, %s", err, out)
		return // 只打日志，不 return error —— 视频已经存好了
	}
	if err := os.Remove(path); err != nil {
		log.Println("faststart 删除原文件失败:", err)
		return
	}
	os.Rename(tmp, path)
}
