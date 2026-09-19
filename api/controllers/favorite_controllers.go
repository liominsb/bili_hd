package controllers

import (
	"go_bili/api/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FavoriteController struct {
	favoriteService service.FavoriteService
}

func NewFavoriteController(favoriteService service.FavoriteService) *FavoriteController {
	return &FavoriteController{favoriteService: favoriteService}
}

// Favorite 收藏视频
func (c *FavoriteController) Favorite(ctx *gin.Context) {
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
	if err := c.favoriteService.Favorite(ctx.Request.Context(), userID.(uint), uint(videoID)); err != nil {
		log.Println("收藏失败:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "收藏成功"})
}

// Unfavorite 取消收藏
func (c *FavoriteController) Unfavorite(ctx *gin.Context) {
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
	if err := c.favoriteService.Unfavorite(ctx.Request.Context(), userID.(uint), uint(videoID)); err != nil {
		log.Println("取消收藏失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "已取消收藏"})
}

// GetFavoriteStats 收藏数 + 是否已收藏（照 follow/stats 抄：未登录也能看计数）
func (c *FavoriteController) GetFavoriteStats(ctx *gin.Context) {
	videoID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的视频ID"})
		return
	}

	var isFavorite bool
	if uid, ok := ctx.Get("ID"); ok {
		isFavorite, err = c.favoriteService.IsFavorite(ctx.Request.Context(), uid.(uint), uint(videoID))
		if err != nil {
			log.Println("查询收藏状态失败:", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	favoriteCount, err := c.favoriteService.CountByVideo(ctx.Request.Context(), uint(videoID))
	if err != nil {
		log.Println("统计收藏数失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"favorite_count": favoriteCount,
			"is_favorite":    isFavorite,
		},
	})
}

// ListMyFavorites 我的收藏列表（offset/limit 分页，同 history）
func (c *FavoriteController) ListMyFavorites(ctx *gin.Context) {
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
	list, err := c.favoriteService.ListMyFavorites(ctx.Request.Context(), userID.(uint), offset, limit)
	if err != nil {
		log.Println("获取收藏列表失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"favorites": *list})
}
