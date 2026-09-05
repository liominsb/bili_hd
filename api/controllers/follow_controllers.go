package controllers

import (
	"go_bili/api/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FollowController struct {
	followService service.FollowService
}

func NewFollowController(followService service.FollowService) *FollowController {
	return &FollowController{followService: followService}
}

// Follow 关注某个用户
func (c *FollowController) Follow(ctx *gin.Context) {
	userID, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	targetID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}
	if err := c.followService.Follow(ctx.Request.Context(), userID.(uint), uint(targetID)); err != nil {
		log.Println("关注失败:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "关注成功"})
}

// Unfollow 取消关注
func (c *FollowController) Unfollow(ctx *gin.Context) {
	userID, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	targetID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}
	if err := c.followService.Unfollow(ctx.Request.Context(), userID.(uint), uint(targetID)); err != nil {
		log.Println("取关失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "已取消关注"})
}

// GetFollowStats 关注数、粉丝数、是否已关注
func (c *FollowController) GetFollowStats(ctx *gin.Context) {
	targetID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 未登录也能看计数，is_following 只对登录用户有意义，未登录记为 false
	var isFollowing bool
	if uid, ok := ctx.Get("ID"); ok {
		isFollowing, err = c.followService.IsFollowing(ctx.Request.Context(), uid.(uint), uint(targetID))
		if err != nil {
			log.Println("查询关注状态失败:", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	followingCount, err := c.followService.CountFollowing(ctx.Request.Context(), uint(targetID))
	if err != nil {
		log.Println("统计关注数失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	followerCount, err := c.followService.CountFollowers(ctx.Request.Context(), uint(targetID))
	if err != nil {
		log.Println("统计粉丝数失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"following_count": followingCount,
			"follower_count":  followerCount,
			"is_following":    isFollowing,
		},
	})
}

// ListFollowers 粉丝列表
func (c *FollowController) ListFollowers(ctx *gin.Context) {
	targetID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
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
	users, err := c.followService.ListFollowers(ctx.Request.Context(), uint(targetID), offset, limit)
	if err != nil {
		log.Println("获取粉丝列表失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"users": *users})
}

// ListFollowing 关注列表
func (c *FollowController) ListFollowing(ctx *gin.Context) {
	targetID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
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
	users, err := c.followService.ListFollowing(ctx.Request.Context(), uint(targetID), offset, limit)
	if err != nil {
		log.Println("获取关注列表失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"users": *users})
}
