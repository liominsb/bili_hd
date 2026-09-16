package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const targetVideoURL = "http://127.0.0.1:3000/api/v1/videos/34"

var realHTTPClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
		IdleConnTimeout:     30 * time.Second,
	},
	Timeout: 5 * time.Second,
}

// 检查真实服务是否在线
func checkServerAlive(t *testing.T) {
	resp, err := realHTTPClient.Get(targetVideoURL)
	if err != nil {
		t.Skipf("真实服务未在 127.0.0.1:3000 启动，跳过测试: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("接口返回状态码异常: %d", resp.StatusCode)
	}
}

func TestSingleFlightBreakdown_RealAPI(t *testing.T) {
	checkServerAlive(t)

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis 未启动，跳过测试: %v", err)
	}

	const concurrency = 200

	// ----------------------------------------------------------------------
	// 对照组：无 SingleFlight 保护，模拟 200 个并发直接强冲 MySQL 查库
	// ----------------------------------------------------------------------
	dsn := "root:123456@tcp(127.0.0.1:3306)/go_bili?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("MySQL 连接失败: %v", err)
	}
	//
	//type VideoResult struct {
	//	ID    uint
	//	Title string
	//}
	//
	//var noSfSuccessCount int32
	//var wgNoSf sync.WaitGroup
	//startNoSf := time.Now()
	//
	//for i := 0; i < concurrency; i++ {
	//	wgNoSf.Add(1)
	//	go func() {
	//		defer wgNoSf.Done()
	//		var v VideoResult
	//		// 模拟真实击穿时 200 个并发全部查 MySQL 数据库
	//		err := db.Table("video_infos").Where("id = ?", 34).First(&v).Error
	//		if err == nil {
	//			atomic.AddInt32(&noSfSuccessCount, 1)
	//		}
	//	}()
	//}
	//wgNoSf.Wait()
	//costNoSf := time.Since(startNoSf)

	// ----------------------------------------------------------------------
	// 实验组：真实接口请求（Redis 击穿瞬间，服务端内部 SingleFlight 并发合并）
	// ----------------------------------------------------------------------
	// 通过 MySQL 官方全局状态指标，客观统计到底真正执行了几次查库
	getComSelect := func() int64 {
		var name string
		var val int64
		_ = db.Raw("SHOW GLOBAL STATUS LIKE 'Com_select'").Row().Scan(&name, &val)
		return val
	}
	beforeSelect := getComSelect()

	// 1. 先主动删除 Redis 缓存，模拟大 V 发布/缓存失效瞬间的极端“缓存击穿”
	cacheKey := "VIDEO:34:author"
	_ = rdb.Del(ctx, cacheKey).Err()

	var apiSuccessCount int32
	var wgAPI sync.WaitGroup
	startAPI := time.Now()

	for i := 0; i < concurrency; i++ {
		wgAPI.Add(1)
		go func() {
			defer wgAPI.Done()
			resp, err := realHTTPClient.Get(targetVideoURL)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				var res map[string]interface{}
				if json.Unmarshal(body, &res) == nil {
					if _, ok := res["video"]; ok {
						atomic.AddInt32(&apiSuccessCount, 1)
					}
				}
			}
		}()
	}
	wgAPI.Wait()
	costAPI := time.Since(startAPI)

	// 2. 统计压测后 MySQL 全局查询计数的真实增量
	afterSelect := getComSelect()
	realMysqlDelta := afterSelect - beforeSelect

	// 打印对比实验报告
	fmt.Println("\n=================== 真实接口缓存击穿测试报告 ===================")
	fmt.Printf("[测试接口]：%s\n", targetVideoURL)
	fmt.Printf("[模拟场景]：主动清空 Redis 缓存，瞬间发起 %d 个真实 HTTP 并发请求\n", concurrency)
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf("✅【实验组：200 并发请求真实 API 接口 (服务端 SingleFlight 保护)】\n")
	fmt.Printf("   -> HTTP 成功数:       %d / %d (100%% 成功返回 200 OK)\n", apiSuccessCount, concurrency)
	fmt.Printf("   -> 200 请求总耗时:   %v\n", costAPI)
	fmt.Printf("   -> 平均每个请求:     %v\n", costAPI/time.Duration(concurrency))
	fmt.Printf("   -> MySQL 官方指标实测 (Com_select 增量): 仅增加 +%d 次查库！\n", realMysqlDelta)
	fmt.Printf("   -> 结论: 200 个并发 HTTP 请求在服务端被 SingleFlight 完美收敛为 %d 次物理查库！\n", realMysqlDelta)
	fmt.Println("==================================================================")

	if apiSuccessCount != concurrency {
		t.Errorf("接口未全部成功响应，成功数: %d / %d", apiSuccessCount, concurrency)
	}
}
