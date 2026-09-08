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
	"sync"
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

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		utils.Worker(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctrl := router.Inject()
		//展示滞后被缓存放大：
		//详情走 GetCacheOrQuery，
		//整个 VideoInfo 连 view_count 一起缓存 10~70 分钟。
		//所以页面上看到的播放数最多滞后一小时，10 秒同步的及时性全被这层缓存吃了。
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ctrl.VideoCtrl.SyncViewCounts(ctx)
			}
		}
	}()

	defer cancel()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt) //监听系统中断信号
	<-quit
	log.Println("服务器正在关闭...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("服务器强行关闭或超时异常: ", err)
	}
	cancel()
	log.Println("已通知后台任务停止，等待 HTTP 请求处理完成...")
	wg.Wait()
	log.Println("服务器已成功优雅退出")
}
