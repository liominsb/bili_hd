package controllers

import (
	"go_bili/api/service"
	"go_bili/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService service.AuthService
}

// 构造函数：在外部（比如 main.go 的路由配置里）初始化时传入依赖
func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

type AuthInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 注册
func (c *AuthController) Register(ctx *gin.Context) {
	var input AuthInput
	if err := ctx.ShouldBind(&input); err != nil {
		log.Println("注册参数错误:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "注册参数错误:" + err.Error()})
		return
	}

	token, refreshToken, err := c.authService.Register(ctx.Request.Context(), input.Username, input.Password)
	if err != nil {
		log.Println("注册失败:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败:" + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token":        token,
		"refreshToken": refreshToken})
}

// Login 登录
func (c *AuthController) Login(ctx *gin.Context) {
	var input AuthInput
	if err := ctx.ShouldBind(&input); err != nil {
		log.Println("登录参数错误:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "登录参数错误:" + err.Error()})
		return
	}

	token, refreshToken, err := c.authService.Login(ctx.Request.Context(), input.Username, input.Password)
	if err != nil {
		log.Println("登录失败:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败:" + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token":        token,
		"refreshToken": refreshToken,
	})
}

// GetMyUser 查当前登录的用户信息
func (c *AuthController) GetMyUser(ctx *gin.Context) {
	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ID"})
		return
	}

	user, err := c.authService.GetMyUser(ctx.Request.Context(), id.(uint))
	if err != nil {
		log.Println("获取用户信息失败:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

// GetUserProfileById 通过动态 ID 获取其他用户的信息
func (c *AuthController) GetUserProfileById(ctx *gin.Context) {
	// 从 URL 参数中获取 ID (例如请求 /api/v1/user/5，这里拿到的就是 "5")
	idStr := ctx.Param("id")

	// 把字符串转换成整数
	targetUserID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID格式"})
		return
	}

	user, err := c.authService.GetUserProfileById(ctx.Request.Context(), uint(targetUserID))
	if err != nil {
		log.Println("获取用户信息失败:", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "找不到该用户"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

// Changepassword 更改密码
func (c *AuthController) Changepassword(ctx *gin.Context) {
	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ID"})
		return
	}

	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.authService.ChangePassword(ctx.Request.Context(), id.(uint), input.OldPassword, input.NewPassword); err != nil {
		log.Println("修改密码失败:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// UpdateMyUser 更新个人资料（用户名）
func (c *AuthController) UpdateMyUser(ctx *gin.Context) {
	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ID"})
		return
	}

	var input struct {
		Username string `json:"username" binding:"required"`
		Image    string `json:"image"`
		Bio      string `json:"bio"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.authService.UpdateProfile(ctx.Request.Context(), id.(uint), input.Username, input.Image, input.Bio); err != nil {
		log.Println("更新资料失败:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回更新后的用户信息
	user, err := c.authService.GetMyUser(ctx.Request.Context(), id.(uint))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

// RefreshTokens 刷新令牌
func (c *AuthController) RefreshTokens(ctx *gin.Context) {

	var input struct {
		AccessToken  string `json:"access_token" binding:"required"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := ctx.ShouldBind(&input); err != nil {
		log.Println("登录参数错误:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := utils.ParseTokenAllowExpired(input.AccessToken)
	if err != nil || claims == nil {
		log.Println("解析Token失败:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的访问令牌"})
		return
	}

	token, refreshToken, err := c.authService.RefreshTokens(ctx.Request.Context(), claims.AccountID, input.RefreshToken, claims.Username)
	if err != nil {
		log.Println("登录失败:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token,
		"refreshToken": refreshToken})
}
