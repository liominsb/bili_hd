package main

import (
	"context"
	"errors"
	"go_bili/config"
	"go_bili/router"
	"go_bili/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	config.InitConfig()

	r := router.SetupRouter()

	port := config.Appconf.App.Port
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	go func() {
		log.Printf("服务器正在%s端口运行 \n", port)
		// ErrServerClosed 是调用 Shutdown 后的正常返回，需要过滤掉以免报 err
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	go utils.Worker()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt) //监听系统中断信号
	<-quit
	log.Println("服务器正在关闭...")
	cancel()
	log.Println("已通知后台任务停止，等待 HTTP 请求处理完成...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("服务器强行关闭或超时异常: ", err)
	}

	log.Println("服务器已成功优雅退出")
}
