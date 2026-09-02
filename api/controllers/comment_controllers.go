package controllers

import (
	"go_bili/api/service"
	"go_bili/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentController struct {
	commentService service.CommentService
}

func NewCommentController(commentService service.CommentService) *CommentController {
	return &CommentController{commentService: commentService}
}

func (c *CommentController) AddNewComment(ctx *gin.Context) {
	var commentInput models.CommentInput
	userId, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	err := ctx.ShouldBind(&commentInput)
	if err != nil {
		log.Println("绑定评论参数失败:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	videoId, err := strconv.Atoi(ctx.Param("id"))
	commentInfo := models.Comments{
		AuthorId: userId.(uint),
		VideoId:  uint(videoId),
		Content:  commentInput.Content,
		ParentId: commentInput.ParentId,
	}
	err = c.commentService.AddNewComment(ctx, &commentInfo)
	if err != nil {
		log.Println("添加新评论失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "评论添加成功"})
}

func (c *CommentController) DeleteCommentByID(ctx *gin.Context) {
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
	comment, err := c.commentService.FindCommentByID(ctx, uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if userId.(uint) != comment.AuthorId {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "无权限删除该评论"})
		return
	}
	err = c.commentService.DeleteComment(ctx, uint(id))
	if err != nil {
		log.Println("删除评论失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "评论删除成功"})
}

func (c *CommentController) GetCommentsByVideoId(ctx *gin.Context) {
	videoId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	comments, err := c.commentService.GetCommentsById(ctx, uint(videoId), offset, limit)
	if err != nil {
		log.Println("获取评论列表失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"comments": *comments})
}

func (c *CommentController) UpdateComment(ctx *gin.Context) {
	var commentInput models.CommentInput
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
	comment, err := c.commentService.FindCommentByID(ctx, uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if userId.(uint) != comment.AuthorId {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "无权限更新该评论"})
		return
	}
	err = ctx.ShouldBind(&commentInput)
	if err != nil {
		log.Println("绑定评论参数失败:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = c.commentService.UpdateComment(ctx, uint(id), commentInput.Content)
	if err != nil {
		log.Println("更新评论失败:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "评论更新成功"})
}
