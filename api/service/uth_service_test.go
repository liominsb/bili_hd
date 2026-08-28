package service

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"go_bili/config"
	"go_bili/global"
	"go_bili/models"
	"go_bili/utils"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var testUser models.User

func TestMain(m *testing.M) {
	// 给 jwtSecret() 一个固定密钥，否则它随机生成，签发和校验对不上
	config.Appconf = &config.Config{}
	config.Appconf.JWT.Key = "test-secret-key"

	mr, err := miniredis.Run() // 内存 Redis，不用装任何东西
	if err != nil {
		log.Fatalf("启动 miniredis 失败: %v", err)
	}
	defer mr.Close()

	hashed, _ := utils.HashPassword("password123")
	testUser = models.User{ID: 1, Username: "testuser", Password: hashed, Bio: "hi", Image: ""}

	global.RedisDB = redis.NewClient(&redis.Options{Addr: mr.Addr()})
	os.Exit(m.Run())
}

// 假数据库：只有 GetUserByUsername/GetUserByID 有内容，其余直接返回 nil
type fakeAuthRepo struct {
	user models.User
}

func (f *fakeAuthRepo) Register(ctx context.Context, user *models.User) error { return nil }

func (f *fakeAuthRepo) GetUserByUsername(ctx context.Context, user *models.User, username string) error {
	if username != f.user.Username {
		return gorm.ErrRecordNotFound
	}
	*user = f.user
	return nil
}

func (f *fakeAuthRepo) GetUserByID(ctx context.Context, user *models.User, userID uint) error {
	if userID != f.user.ID {
		return gorm.ErrRecordNotFound
	}
	*user = f.user
	return nil
}

func (f *fakeAuthRepo) UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error {
	return nil
}
func (f *fakeAuthRepo) UpdateProfile(ctx context.Context, userID uint, username string, image string, bio string) error {
	return nil
}

// 唯一的测试：登录 → 刷新 → 断言新 token 是裸 token 且能被解析
func TestRefreshTokenReturnsBareToken(t *testing.T) {
	svc := NewAuthService(&fakeAuthRepo{user: testUser}, global.RedisDB)

	// 1. 登录，拿到 access token 和 refresh token
	access, rt, err := svc.Login(context.Background(), "testuser", "password123")
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if strings.HasPrefix(access, "Bearer ") {
		t.Fatalf("登录返回的 token 不应带 Bearer 前缀: %q", access)
	}

	// 2. 刷新，拿新 access token
	newAccess, _, err := svc.RefreshTokens(context.Background(), testUser.ID, rt, testUser.Username)
	if err != nil {
		t.Fatalf("刷新失败: %v", err)
	}
	// 3. 这就是今天的 bug：如果带 "Bearer " 前缀，这条立刻红
	if strings.HasPrefix(newAccess, "Bearer ") {
		t.Fatalf("刷新返回的 token 不应带 Bearer 前缀（今天的 bug！）: %q", newAccess)
	}

	// 4. 顺带验证：新 token 能被正常解析（真的是张能用的票）
	if _, err := utils.ParseToken(newAccess); err != nil {
		t.Fatalf("新 token 解析失败: %v", err)
	}
}
