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

	r.Static("/uploads", "./uploads")

	auth := r.Group("/api/auth")
	{
		auth.POST("login", ctrl.AuthCtrl.Login)
		auth.POST("register", ctrl.AuthCtrl.Register)
		auth.POST("refreshTokens", ctrl.AuthCtrl.RefreshTokens)
	}
	apiRouter := r.Group("/api/v1")
	{
		apiRouter.GET("videos/:id", ctrl.VideoCtrl.FindVideoByID)
		apiRouter.GET("videos", ctrl.VideoCtrl.GetVideos)
		apiRouter.POST("upload", ctrl.UploadCtrl.UploadFile)
		apiRouter.GET("users/:id", ctrl.AuthCtrl.GetUserProfileById)
		apiRouter.GET("videos/:id/comments", ctrl.CommentCtrl.GetCommentsByVideoId)
	}
	apiRouter.Use(middlewares.AuthMiddleware())
	{
		apiRouter.POST("videos", ctrl.VideoCtrl.AddNewVideo)
		apiRouter.PUT("videos/:id", ctrl.VideoCtrl.UpdateVideo)
		apiRouter.PUT("videos/:id/like", ctrl.VideoCtrl.UpdateVideoLike)
		apiRouter.DELETE("videos/:id", ctrl.VideoCtrl.DeleteVideoByID)

		apiRouter.GET("users/me", ctrl.AuthCtrl.GetMyUser)
		apiRouter.PUT("users/me", ctrl.AuthCtrl.UpdateMyUser)
		apiRouter.PUT("users/me/psw", ctrl.AuthCtrl.Changepassword)
		apiRouter.GET("videos/search", ctrl.VideoCtrl.SearchVideoByTitle)

		apiRouter.POST("videos/:id/comments", ctrl.CommentCtrl.AddNewComment)
		apiRouter.PUT("comments/:id", ctrl.CommentCtrl.UpdateComment)
		apiRouter.DELETE("comments/:id", ctrl.CommentCtrl.DeleteCommentByID)
	}
	return r
}
