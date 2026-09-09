package test

import (
	"go_bili/utils"
	"io"
	"regexp"
	"testing"
)

const targetVideoListURL = "http://127.0.0.1:3000/api/v1/videos?offset=0&limit=10"

// 1. 基准测试：视频详情真实 HTTP API 接口吞吐
func BenchmarkVideoDetail_RealAPI(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := realHTTPClient.Get(targetVideoURL)
			if err == nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
			}
		}
	})
}

// 2. 基准测试：视频列表真实 HTTP API 接口吞吐
func BenchmarkVideoList_RealAPI(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := realHTTPClient.Get(targetVideoListURL)
			if err == nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
			}
		}
	})
}

// 3. 对照基准测试：你写的 FilterSymbolsFast（预分配 strings.Builder）vs 普通正则表达式
var testRawString = "Hello, 世界！@#$%^&*()_+ 123456 Go-Bilibili 视频弹幕测试！！！"
var reg = regexp.MustCompile(`[^a-zA-Z0-9\s]+`)

func filterSymbolsRegex(s string) string {
	return reg.ReplaceAllString(s, "")
}

func BenchmarkFilterSymbols_Regex(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = filterSymbolsRegex(testRawString)
	}
}

func BenchmarkFilterSymbols_Fast(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = utils.FilterSymbolsFast(testRawString)
	}
}