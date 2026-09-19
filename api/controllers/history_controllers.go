package controllers

import (
	"go_bili/api/service"
	"go_bili/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HistoryController struct {
	historyService service.HistoryService
}

func NewHistoryController(historyService service.HistoryService) *HistoryController {
	return &HistoryController{historyService: historyService}
}

// ReportHistory 上报播放进度 PUT /api/v1/videos/:id/history
func (c *HistoryController) ReportHistory(ctx *gin.Context) {
	userID, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	videoID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的视频ID"})
		return
	}
	var input models.HistoryInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		log.Println("绑定播放进度参数失败:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.historyService.Report(ctx, userID.(uint), uint(videoID), input.Progress); err != nil {
		log.Println("上报播放历史失败:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// DeleteHistory 删除单条 DELETE /api/v1/videos/:id/history
func (c *HistoryController) DeleteHistory(ctx *gin.Context) {
	userID, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	videoID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的视频ID"})
		return
	}
	if err := c.historyService.Delete(ctx, userID.(uint), uint(videoID)); err != nil {
		log.Println("删除播放历史失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "已删除该记录"})
}

// ClearHistory 清空全部 DELETE /api/v1/users/me/history
func (c *HistoryController) ClearHistory(ctx *gin.Context) {
	userID, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	if err := c.historyService.Clear(ctx, userID.(uint)); err != nil {
		log.Println("清空播放历史失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "已清空播放历史"})
}

// ListHistory 我的历史 GET /api/v1/users/me/history
// me 只认 JWT 里的 userID，不接受路径传参，天然没有越权问题
func (c *HistoryController) ListHistory(ctx *gin.Context) {
	userID, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	list, err := c.historyService.List(ctx, userID.(uint), offset, limit)
	if err != nil {
		log.Println("获取播放历史失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"history": *list})
}
