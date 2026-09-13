package controllers

import (
	"go_bili/api/service"
	"go_bili/config"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type OAuthController struct {
	oauthService service.OAuthService
}

func NewOAuthController(oauthService service.OAuthService) *OAuthController {
	return &OAuthController{oauthService: oauthService}
}

// GitHubLogin 浏览器访问它就 302 跳到 GitHub 授权页
func (c *OAuthController) GitHubLogin(ctx *gin.Context) {
	authURL, err := c.oauthService.GitHubLoginURL(ctx.Request.Context())
	if err != nil {
		log.Println("生成 GitHub 授权地址失败:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub 登录暂时不可用"})
		return
	}
	ctx.Redirect(http.StatusFound, authURL)
}

// GitHubCallback 处理 GitHub 回调。
// 成功失败都重定向回前端 —— 请求是浏览器跳过来的，返回 JSON 用户也看不懂。
func (c *OAuthController) GitHubCallback(ctx *gin.Context) {
	frontendURL := config.Appconf.GitHub.FrontendURL

	// 用户在 GitHub 页面上点了「取消」，会带回 error 参数
	if e := ctx.Query("error"); e != "" {
		log.Println("GitHub 授权被拒绝:", e, ctx.Query("error_description"))
		ctx.Redirect(http.StatusFound, frontendURL+"?error="+url.QueryEscape("你取消了 GitHub 授权"))
		return
	}

	token, refreshToken, err := c.oauthService.GitHubCallback(ctx.Request.Context(), ctx.Query("code"), ctx.Query("state"))
	if err != nil {
		// 具体原因只写日志，不带给前端：避免把内部细节（含 GitHub 返回的原文）暴露到 URL 上
		log.Println("GitHub 登录失败:", err)
		ctx.Redirect(http.StatusFound, frontendURL+"?error="+url.QueryEscape("GitHub 登录失败，请重试"))
		return
	}

	// 双 token 拼在前端回调地址上，前端读取后存进 localStorage
	ctx.Redirect(http.StatusFound, frontendURL+
		"?token="+url.QueryEscape(token)+
		"&refreshToken="+url.QueryEscape(refreshToken))
}
