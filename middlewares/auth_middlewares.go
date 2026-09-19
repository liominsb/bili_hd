package middlewares // Package middlewares 中间件

import (
	"fmt"
	"go_bili/global"
	"go_bili/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParseAuthMiddleware 解析 token：合法就写入 ID，没有/无效/被踢都放行（永不 401）
// 全组挂载——公开路由靠它拿到"可选的登录态"
func ParseAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token != "" && strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		} else if queryToken := ctx.Query("token"); queryToken != "" {
			token = queryToken
		} else {
			ctx.Next() // 没带 token：游客，放行
			return
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			ctx.Next() // token 无效：视同游客
			return
		}

		// 防踢校验：session 对不上视同游客，不写 ID
		redisKey := fmt.Sprintf("auth:account:%d", claims.AccountID)
		activeSessionID, err := global.RedisDB.HGet(ctx.Request.Context(), redisKey, "session_id").Result()
		if err != nil || claims.SessionID != activeSessionID {
			ctx.Next()
			return
		}

		ctx.Set("ID", claims.AccountID)
		ctx.Set("Username", claims.Username)
		ctx.Next()
	}
}

// RequireAuthMiddleware 硬认证：前面没人写入 ID 就 401
// 只检查不解析——token 的活全在 Parse 里
func RequireAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if _, ok := ctx.Get("ID"); !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}
		ctx.Next()
	}
}
