package utils // Package utils 实用性

import (
	"context"
	"encoding/json"
	"errors"
	"go_bili/global"
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/singleflight"
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

// 全局并发合并器
var sfGroup singleflight.Group

// 空值占位符，防止缓存穿透
const nullCachePlaceholder = "{}"

// Setcache 设置缓存
func Setcache(ctx context.Context, key string, value interface{}) error {
	valueJSON, err := json.Marshal(value)

	if err != nil {
		return err
	}
	a := time.Duration(rand.IntN(60) + 10) //防雪崩
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
//
// 缓存防击穿、防穿透
func GetCacheOrQuery[T any](ctx context.Context, redisClient *redis.Client, cacheKey string, queryFunc func() (*T, error)) (*T, error) {
	// 1. 查缓存
	cacheData, err := redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// 命中空值占位符，说明数据库也是空的，直接返回空指针（防穿透）
		if cacheData == nullCachePlaceholder {
			return nil, nil
		}
		var data T
		if err := json.Unmarshal([]byte(cacheData), &data); err == nil {
			return &data, nil
		}
		log.Println("缓存反序列化失败，降级查库:", cacheKey)
	} else if !errors.Is(err, redis.Nil) {
		log.Println("Redis查询异常，降级查库:", err)
	}

	// 2. 缓存未命中：使用 singleflight 拦截并发（防击穿）
	// 同一个 cacheKey 的并发请求，仅放行一个去执行 DB 查询，其余等待共享结果
	result, err, _ := sfGroup.Do(cacheKey, func() (interface{}, error) {
		// 双重检查：拿锁进来的 Goroutine 先再探查一次缓存
		val, getErr := redisClient.Get(ctx, cacheKey).Result()
		if getErr == nil {
			if val == nullCachePlaceholder {
				return nil, nil
			}
			var d T
			if json.Unmarshal([]byte(val), &d) == nil {
				return &d, nil
			}
		}

		// 真正查数据库
		dbData, dbErr := queryFunc()
		if dbErr != nil {
			return nil, dbErr
		}

		// 3. 查不到数据：写空缓存，TTL 设为极短的 1 分钟（防穿透）
		if dbData == nil {
			_ = redisClient.Set(ctx, cacheKey, nullCachePlaceholder, 1*time.Minute).Err()
			return nil, nil
		}

		// 查到了数据：正常写缓存
		_ = Setcache(ctx, cacheKey, dbData)
		return dbData, nil
	})

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*T), nil
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

func Faststart(path string) error {
	tmp := path + ".tmp"
	cmd := exec.Command("ffmpeg", "-y",
		"-i", path,
		"-c", "copy",
		"-movflags", "+faststart",
		tmp,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("faststart 失败(不影响上传): %v, %s", err, out)
		return err
	}
	if err := os.Remove(path); err != nil {
		log.Println("faststart 删除原文件失败:", err)
		return err
	}
	return os.Rename(tmp, path)
}
