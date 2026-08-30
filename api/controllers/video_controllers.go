package controllers

import (
	"go_bili/api/service"
	"go_bili/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VideoController struct {
	videoService service.VideoService
}

func NewVideoController(videoService service.VideoService) *VideoController {
	return &VideoController{videoService: videoService}
}

func (c *VideoController) AddNewVideo(ctx *gin.Context) {
	var videoInput models.VideoInput
	var videoInfo models.VideoInfo
	userId, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "未授权"})
		return
	}
	err := ctx.ShouldBind(&videoInput)
	if err != nil {
		log.Println("绑定视频参数失败:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	videoInfo.VideoUrl = videoInput.VideoUrl
	videoInfo.Pic = videoInput.Pic
	videoInfo.Title = videoInput.Title
	videoInfo.AuthorId = userId.(uint)
	err = c.videoService.AddNewVideo(ctx, &videoInfo)
	if err != nil {
		log.Println("添加新视频失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "视频添加成功"})
}

func (c *VideoController) UpdateVideo(ctx *gin.Context) {
	var videoInput models.VideoInput
	err := ctx.ShouldBind(&videoInput)
	if err != nil {
		log.Println("绑定视频参数失败:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	video, err := c.videoService.FindVideoByID(ctx, uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userId.(uint) != video.AuthorId {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "无权限更新该视频"})
		return
	}
	video.VideoUrl = videoInput.VideoUrl
	video.Pic = videoInput.Pic
	video.Title = videoInput.Title
	err = c.videoService.UpdateVideo(ctx, video)
	if err != nil {
		log.Println("更新视频失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "视频更新成功"})
}

func (c *VideoController) FindVideoByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}
	video, err := c.videoService.FindVideoByID(ctx, uint(id))
	if err != nil {
		log.Println("查找视频失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"video": video})
}

func (c *VideoController) GetVideos(ctx *gin.Context) {
	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}
	videos, err := c.videoService.GetVideos(ctx, offset, limit)
	if err != nil {
		log.Println("获取视频列表失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"videos": videos})
}

func (c *VideoController) DeleteVideoByID(ctx *gin.Context) {
	userId, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}
	video, err := c.videoService.FindVideoByID(ctx, uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if userId.(uint) != video.AuthorId {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "无权限删除该视频"})
		return
	}
	err = c.videoService.DeleteVideo(ctx, uint(id))
	if err != nil {
		log.Println("删除视频失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "视频删除成功"})
}
