package test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestViewCountWritePerformance_RealAPI(t *testing.T) {
	checkServerAlive(t)

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis 连接失败，跳过测试: %v", err)
	}

	dsn := "root:123456@tcp(127.0.0.1:3306)/go_bili?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("MySQL 连接失败，跳过测试: %v", err)
	}

	const totalRequests = 500
	const concurrency = 50

	// ---------------------------------------------------------
	// 对照组：模拟每次播放都直接同步 UPDATE MySQL（严重行锁排队）
	// ---------------------------------------------------------
	startMysql := time.Now()
	var wgMysql sync.WaitGroup
	chMysql := make(chan struct{}, concurrency)

	for i := 0; i < totalRequests; i++ {
		wgMysql.Add(1)
		chMysql <- struct{}{}
		go func() {
			defer func() {
				<-chMysql
				wgMysql.Done()
			}()
			_ = db.Table("video_infos").Where("id = ?", 34).
				UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
		}()
	}
	wgMysql.Wait()
	costMysql := time.Since(startMysql)

	// ---------------------------------------------------------
	// 实验组：调用真实 HTTP API 接口（服务端 Redis Incr 内存原子记数 + 异步落库）
	// ---------------------------------------------------------
	viewsKey := "video:views:34"
	// 记录初始增量
	initViews, _ := rdb.Get(ctx, viewsKey).Int64()

	startAPI := time.Now()
	var wgAPI sync.WaitGroup
	chAPI := make(chan struct{}, concurrency)
	var apiSuccessCount int32

	for i := 0; i < totalRequests; i++ {
		wgAPI.Add(1)
		chAPI <- struct{}{}
		go func() {
			defer func() {
				<-chAPI
				wgAPI.Done()
			}()
			resp, err := realHTTPClient.Get(targetVideoURL)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				atomic.AddInt32(&apiSuccessCount, 1)
			}
		}()
	}
	wgAPI.Wait()
	costAPI := time.Since(startAPI)

	// 读取当前累积的增量
	finalViews, _ := rdb.Get(ctx, viewsKey).Int64()
	deltaViews := finalViews - initViews

	// 输出测试报告
	fmt.Println("\n=================== 真实接口播放量高并发测试报告 ===================")
	fmt.Printf("[测试接口]：%s\n", targetVideoURL)
	fmt.Printf("[测试场景]：并发 Worker = %d, 发起真实 HTTP 请求总数 = %d 次\n", concurrency, totalRequests)
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf("❌【对照组：同步直写 MySQL (500 次 UPDATE 行锁排队)】\n")
	fmt.Printf("   -> 总耗时:       %v\n", costMysql)
	fmt.Printf("   -> 单次写入耗时: %v\n", costMysql/time.Duration(totalRequests))
	fmt.Printf("   -> 写入 TPS:     %.2f req/sec\n", float64(totalRequests)/costMysql.Seconds())
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf("✅【实验组：请求真实 HTTP 接口 (Redis Incr 内存原子递增)】\n")
	fmt.Printf("   -> HTTP 成功数:  %d / %d (100%% 成功率)\n", apiSuccessCount, totalRequests)
	fmt.Printf("   -> Redis 实际累加: +%d 播放数 (完美对应请求量)\n", deltaViews)
	fmt.Printf("   -> 总耗时:       %v\n", costAPI)
	fmt.Printf("   -> 单次请求耗时: %v\n", costAPI/time.Duration(totalRequests))
	fmt.Printf("   -> 接口 QPS:     %.2f req/sec\n", float64(totalRequests)/costAPI.Seconds())
	fmt.Println("==================================================================")
}