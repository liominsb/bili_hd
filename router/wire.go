package router

import (
	"go_bili/api/controllers"
	"go_bili/api/repository"
	"go_bili/api/service"
	"go_bili/global"
)

// Controllers 所有控制器的集合
type Controllers struct {
	AuthCtrl   *controllers.AuthController
	VideoCtrl  *controllers.VideoController
	UploadCtrl *controllers.UploadController
}

// Inject 依赖注入装配，像乐高积木一样一层层组装
func Inject() *Controllers {
	// Repository 层
	authRepo := repository.NewAuthRepository(global.Db)
	videoRepo := repository.NewVideoRepository(global.Db)

	// Service 层：拿到 Repo 和 Redis
	authService := service.NewAuthService(authRepo, global.RedisDB)
	videoService := service.NewVideoService(videoRepo, global.RedisDB)

	// Controller 层：拿到 Service
	return &Controllers{
		AuthCtrl:   controllers.NewAuthController(authService),
		VideoCtrl:  controllers.NewVideoController(videoService),
		UploadCtrl: controllers.NewUploadController(),
	}
}
