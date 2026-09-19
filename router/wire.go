package router

import (
	"go_bili/api/controllers"
	"go_bili/api/repository"
	"go_bili/api/service"
	"go_bili/global"
)

// Controllers 所有控制器的集合
type Controllers struct {
	AuthCtrl    *controllers.AuthController
	VideoCtrl   *controllers.VideoController
	UploadCtrl  *controllers.UploadController
	CommentCtrl *controllers.CommentController
	FollowCtrl  *controllers.FollowController
	OAuthCtrl   *controllers.OAuthController
	HistoryCtrl *controllers.HistoryController
}

// Inject 依赖注入装配，像乐高积木一样一层层组装
func Inject() *Controllers {
	// Repository 层
	authRepo := repository.NewAuthRepository(global.Db)
	videoRepo := repository.NewVideoRepository(global.Db)
	commentRepo := repository.NewCommentRepository(global.Db)
	followRepo := repository.NewFollowRepository(global.Db)
	oauthRepo := repository.NewOAuthRepository(global.Db)
	historyRepo := repository.NewHistoryRepository(global.Db)

	// Service 层：拿到 Repo 和 Redis
	authService := service.NewAuthService(authRepo, global.RedisDB)
	videoService := service.NewVideoService(videoRepo, global.RedisDB)
	commentService := service.NewCommentService(commentRepo, global.RedisDB)
	followService := service.NewFollowService(followRepo, authRepo)
	// OAuth 要复用 authService 的签发逻辑（IssueSession），所以把 authService 也注入进去
	oauthService := service.NewOAuthService(oauthRepo, authRepo, authService, global.RedisDB)
	historyService := service.NewHistoryService(historyRepo)

	// Controller 层：拿到 Service
	return &Controllers{
		AuthCtrl:    controllers.NewAuthController(authService),
		VideoCtrl:   controllers.NewVideoController(videoService),
		UploadCtrl:  controllers.NewUploadController(),
		CommentCtrl: controllers.NewCommentController(commentService),
		FollowCtrl:  controllers.NewFollowController(followService),
		OAuthCtrl:   controllers.NewOAuthController(oauthService),
		HistoryCtrl: controllers.NewHistoryController(historyService),
	}
}
