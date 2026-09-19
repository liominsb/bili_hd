package router

import (
	"go_bili/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()
	//r := gin.New()
	//r.Use(gin.Recovery())
	ctrl := Inject()

	r.Static("/uploads", "./uploads")
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
	auth := r.Group("/api/auth")
	{
		auth.POST("login", ctrl.AuthCtrl.Login)
		auth.POST("register", ctrl.AuthCtrl.Register)
		auth.POST("refreshTokens", ctrl.AuthCtrl.RefreshTokens)

		// GitHub OAuth：login 负责跳到 GitHub，callback 是 GitHub 授权后回跳的落点
		auth.GET("github/login", ctrl.OAuthCtrl.GitHubLogin)
		auth.GET("github/callback", ctrl.OAuthCtrl.GitHubCallback)
	}
	apiRouter := r.Group("/api/v1")
	apiRouter.Use(middlewares.ParseAuthMiddleware())
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

		// 收藏数公开可看，is_favorite 只对登录用户有意义
		apiRouter.GET("videos/:id/favorite/stats", ctrl.FavoriteCtrl.GetFavoriteStats)
	}
	apiRouter.Use(middlewares.RequireAuthMiddleware())
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

		// 播放历史：me 只认 JWT 里的 userID，全部需要登录
		apiRouter.PUT("videos/:id/history", ctrl.HistoryCtrl.ReportHistory)
		apiRouter.DELETE("videos/:id/history", ctrl.HistoryCtrl.DeleteHistory)
		apiRouter.GET("users/me/history", ctrl.HistoryCtrl.ListHistory)
		apiRouter.DELETE("users/me/history", ctrl.HistoryCtrl.ClearHistory)

		// 收藏/取消收藏 + 我的收藏列表：全部需要登录
		apiRouter.PUT("videos/:id/favorite", ctrl.FavoriteCtrl.Favorite)
		apiRouter.DELETE("videos/:id/favorite", ctrl.FavoriteCtrl.Unfavorite)
		apiRouter.GET("users/me/favorites", ctrl.FavoriteCtrl.ListMyFavorites)
	}
	return r
}
