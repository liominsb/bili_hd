package router

import (
	"go_bili/middlewares"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()
	ctrl := Inject()
	r.Use(cors.New(cors.Config{
		// 开发环境：允许所有 localhost 来源（避免每次改端口都要改配置）
		AllowOriginFunc: func(origin string) bool {
			return true // 开发阶段允许所有来源
		},
		// 允许前端使用哪些危险方法？(解决预检请求问题)
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// 允许前端携带哪些特殊的请求头？
		AllowHeaders: []string{"Origin", "Authorization", "Content-Type"},
		// 允许前端读取哪些额外的响应头？
		ExposeHeaders: []string{"Content-Length"},
		// 是否允许携带 Cookie 等凭证？
		AllowCredentials: true,
		// 预检请求的结果缓存多久？(12小时内，同一个请求就不用再发 OPTIONS 探路了)
		MaxAge: 12 * time.Hour,
	}))
	//r.Static("/uploads", "./uploads")
	//
	auth := r.Group("/api/auth")
	{
		auth.POST("login", ctrl.AuthCtrl.Login)
		auth.POST("register", ctrl.AuthCtrl.Register)
		auth.POST("refreshTokens", ctrl.AuthCtrl.RefreshTokens)
	}
	apiRouter := r.Group("/api/v1")
	{
		apiRouter.GET("videos/:id", ctrl.VideoCtrl.FindVideoByID)
		apiRouter.GET("videos/:offset/:limit", ctrl.VideoCtrl.GetVideos)
	}
	apiRouter.Use(middlewares.AuthMiddleware())
	{
		// 写操作放在鉴权之后，controller 才能从 ctx 拿到当前登录用户 ID
		apiRouter.POST("videos", ctrl.VideoCtrl.AddNewVideo)
		apiRouter.PUT("videos", ctrl.VideoCtrl.UpdateVideo)
		apiRouter.DELETE("videos/:id", ctrl.VideoCtrl.DeleteVideoByID)

		apiRouter.GET("users/me", ctrl.AuthCtrl.GetMyUser)
		apiRouter.PUT("users/me", ctrl.AuthCtrl.UpdateMyUser)
		apiRouter.PUT("users/me/psw", ctrl.AuthCtrl.Changepassword)

		apiRouter.GET("users/:id", ctrl.AuthCtrl.GetUserProfileById)

	}
	return r
}
