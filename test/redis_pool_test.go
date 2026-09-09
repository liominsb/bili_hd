package test

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 测试真实 HTTP 接口在不同并发 Worker 压力下的表现
func TestRealAPIConcurrencyLevels(t *testing.T) {
	checkServerAlive(t)

	runHTTPLevelTest := func(concurrency int, totalRequests int) (float64, time.Duration, int32) {
		ch := make(chan struct{}, concurrency)
		var wg sync.WaitGroup
		var successCount int32

		start := time.Now()
		for i := 0; i < totalRequests; i++ {
			wg.Add(1)
			ch <- struct{}{}
			go func() {
				defer func() {
					<-ch
					wg.Done()
				}()
				resp, err := realHTTPClient.Get(targetVideoURL)
				if err != nil {
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					atomic.AddInt32(&successCount, 1)
				}
			}()
		}
		wg.Wait()
		cost := time.Since(start)
		qps := float64(totalRequests) / cost.Seconds()
		return qps, cost, successCount
	}

	const requestsPerLevel = 2000

	fmt.Println("\n=================== 真实 HTTP 接口并发承载力梯度测试 ===================")
	fmt.Printf("[测试接口]：%s\n", targetVideoURL)
	fmt.Printf("每个梯度发送 %d 次真实 HTTP GET 请求\n", requestsPerLevel)
	fmt.Println("------------------------------------------------------------------------")

	// 梯度 1：50 并发
	qps50, cost50, ok50 := runHTTPLevelTest(50, requestsPerLevel)
	fmt.Printf("【梯度一：50 并发 Worker】  -> 耗时: %-10v, QPS: %-10.2f req/s, 成功: %d/%d\n",
		cost50, qps50, ok50, requestsPerLevel)

	// 梯度 2：100 并发
	qps100, cost100, ok100 := runHTTPLevelTest(100, requestsPerLevel)
	fmt.Printf("【梯度二：100 并发 Worker】 -> 耗时: %-10v, QPS: %-10.2f req/s, 成功: %d/%d\n",
		cost100, qps100, ok100, requestsPerLevel)

	// 梯度 3：200 并发
	qps200, cost200, ok200 := runHTTPLevelTest(200, requestsPerLevel)
	fmt.Printf("【梯度三：200 并发 Worker】 -> 耗时: %-10v, QPS: %-10.2f req/s, 成功: %d/%d\n",
		cost200, qps200, ok200, requestsPerLevel)

	fmt.Println("========================================================================")
}