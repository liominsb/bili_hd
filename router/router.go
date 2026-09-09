package router

import (
	"go_bili/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()
	ctrl := Inject()

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
		apiRouter.GET("videos/search", ctrl.VideoCtrl.SearchVideoByTitle)

		// 关注系统：stats / 粉丝列表 / 关注列表 公开，未登录也能看
		apiRouter.GET("users/:id/follow/stats", ctrl.FollowCtrl.GetFollowStats)
		apiRouter.GET("users/:id/followers", ctrl.FollowCtrl.ListFollowers)
		apiRouter.GET("users/:id/following", ctrl.FollowCtrl.ListFollowing)
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

		apiRouter.POST("videos/:id/comments", ctrl.CommentCtrl.AddNewComment)
		apiRouter.PUT("comments/:id", ctrl.CommentCtrl.UpdateComment)
		apiRouter.DELETE("comments/:id", ctrl.CommentCtrl.DeleteCommentByID)

		// 关注/取关需要登录
		apiRouter.PUT("users/:id/follow", ctrl.FollowCtrl.Follow)
		apiRouter.DELETE("users/:id/follow", ctrl.FollowCtrl.Unfollow)
	}
	return r
}
