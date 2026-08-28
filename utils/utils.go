package utils // Package utils 实用性

import (
	"context"
	"encoding/json"
	"errors"
	"go_bili/global"
	"log"
	"math/rand/v2"
	"strings"
	"time"
	"unicode"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(hash), err
}

// 检查密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Setcache 设置缓存
func Setcache(ctx context.Context, key string, value interface{}) error {
	valueJSON, err := json.Marshal(value)

	if err != nil {
		return err
	}
	a := time.Duration(rand.IntN(5) + 10)
	if err := global.RedisDB.Set(ctx, key, valueJSON, a*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

// GetCacheOrQuery 通用缓存函数：先查Redis，缓存未命中或解析失败则降级到数据库查询
// T: 数据类型
// 参数：
//
//	ctx: 上下文
//	redisClient: Redis客户端
//	cacheKey: 缓存键
//	queryFunc: 查询函数，当缓存未命中时调用，返回数据指针和错误
//
// 返回：
//
//	数据指针和错误
func GetCacheOrQuery[T any](ctx context.Context, redisClient *redis.Client, cacheKey string, queryFunc func() (*T, error)) (*T, error) {
	cacheData, err := redisClient.Get(ctx, cacheKey).Result()

	if err == nil {
		var data T
		if err := json.Unmarshal([]byte(cacheData), &data); err == nil {
			return &data, nil
		}
		log.Println("缓存解析失败，降级到数据库查询:", cacheKey)
	} else if !errors.Is(err, redis.Nil) {
		log.Println("Redis查询错误，降级到数据库查询:", err)
	}

	data, err := queryFunc()
	if err != nil {
		return nil, err
	}

	if data != nil {
		if err := Setcache(ctx, cacheKey, data); err != nil {
			log.Println("缓存数据失败:", err)
		}
	}

	return data, nil
}

// RandomExpiration 传入一个基础过期时间，返回增加 0~59 秒随机抖动后的时间
func RandomExpiration(baseTime time.Duration) time.Duration {
	// 使用 rand.Intn(60) 生成 0-59 的随机数，更加标准和易读
	jitter := time.Duration(rand.IntN(60)) * time.Second
	return baseTime + jitter
}

// 快速过滤掉字符串中的所有“非字母、非数字、非空格”的符号，只保留字母、数字和空白字符
func FilterSymbolsFast(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
