package service

import (
	"context"
	"errors"
	"fmt"
	"go_bili/api/repository"
	"go_bili/models"
	"go_bili/utils"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthService interface {
	Register(ctx context.Context, Username string, Password string) (string, string, error)
	Login(ctx context.Context, Username string, Password string) (string, string, error)
	GetMyUser(ctx context.Context, userID uint) (*models.User, error)
	GetUserProfileById(ctx context.Context, targetUserID uint) (*models.User, error)
	ChangePassword(ctx context.Context, userID uint, oldPassword string, newPassword string) error
	UpdateProfile(ctx context.Context, userID uint, username string, image string, bio string) error
	RefreshTokens(ctx context.Context, accountID uint, incomingRT string, username string) (string, string, error)
	IssueSession(ctx context.Context, user *models.User) (string, string, error) // 密码登录/注册/第三方登录共用的会话签发入口
}

type authServiceImpl struct {
	authRepo    repository.AuthRepository
	redisClient *redis.Client
}

func NewAuthService(authRepo repository.AuthRepository, redisClient *redis.Client) AuthService {
	return &authServiceImpl{authRepo: authRepo, redisClient: redisClient}
}

func userRedisKey(userID uint) string {
	return fmt.Sprintf("USER:%d", userID)
}

// IssueSession 为一个身份已确认的用户开新会话并签发双 token。
// 注册、密码登录、第三方登录三个入口都走这里，保证签发逻辑只有一份。
func (s *authServiceImpl) IssueSession(ctx context.Context, user *models.User) (string, string, error) {
	// 1. 生成唯一会话标识
	sessionID := uuid.New().String()

	token, err := utils.GenerateToken(user.ID, user.Username, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("生成 token 失败: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("生成 refresh token 失败: %w", err)
	}

	redisKey := fmt.Sprintf("auth:account:%d", user.ID)

	// HSet + Expire 用事务管道一起提交：
	// 否则 HSet 成功、Expire 失败时这个 key 会永不过期 —— 会话永久有效、踢下线彻底失效
	pipe := s.redisClient.TxPipeline()
	pipe.HSet(ctx, redisKey,
		"session_id", sessionID,
		"refresh_token", refreshToken,
	)
	pipe.Expire(ctx, redisKey, 7*24*time.Hour)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", "", fmt.Errorf("写入会话失败: %w", err)
	}

	return token, refreshToken, nil
}

func (s *authServiceImpl) Register(ctx context.Context, Username string, Password string) (string, string, error) {
	var user models.User
	hashedPwd, err := utils.HashPassword(Password)
	if err != nil {
		return "", "", err
	}

	user.Username = Username
	user.Password = hashedPwd
	user.Bio = "这个人什么也没说"
	user.Image = "https://i0.hdslb.com/bfs/face/d413e1275b3a6e97dc9c812afc21d7b46f231f50.jpg@96w_96h_1c_1s_!web-avatar.avif"

	if err := s.authRepo.Register(ctx, &user); err != nil {
		return "", "", err
	}

	return s.IssueSession(ctx, &user)
}

func (s *authServiceImpl) Login(ctx context.Context, Username string, Password string) (string, string, error) {
	var user models.User
	if err := s.authRepo.GetUserByUsername(ctx, &user, Username); err != nil {
		return "", "", errors.New("账号或密码错误")
	}
	if !utils.CheckPassword(Password, user.Password) {
		return "", "", errors.New("账号或密码错误")
	}

	return s.IssueSession(ctx, &user)
}

func (s *authServiceImpl) GetMyUser(ctx context.Context, userID uint) (*models.User, error) {
	cacheKey := userRedisKey(userID)

	result, err := utils.GetCacheOrQuery(ctx, s.redisClient, cacheKey, func() (*models.User, error) {
		var user models.User
		if err := s.authRepo.GetUserByID(ctx, &user, userID); err != nil {
			return nil, err
		}
		return &user, nil
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	// 返回用户信息（不包含密码）
	result.Password = ""
	return result, nil
}

func (s *authServiceImpl) GetUserProfileById(ctx context.Context, targetUserID uint) (*models.User, error) {
	cacheKey := userRedisKey(targetUserID)

	result, err := utils.GetCacheOrQuery(ctx, s.redisClient, cacheKey, func() (*models.User, error) {
		var user models.User
		if err := s.authRepo.GetUserByID(ctx, &user, targetUserID); err != nil {
			return nil, err
		}
		return &user, nil
	})
	if err != nil {
		return nil, err
	}

	// 返回用户信息（不包含密码）
	if result == nil {
		return nil, nil
	}
	result.Password = ""
	return result, nil
}

func (s *authServiceImpl) ChangePassword(ctx context.Context, userID uint, oldPassword string, newPassword string) error {
	// 获取用户信息
	var user models.User
	if err := s.authRepo.GetUserByID(ctx, &user, userID); err != nil {
		return err
	}

	// 验证旧密码
	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("旧密码不正确")
	}

	// 检查新密码是否与旧密码相同
	if oldPassword == newPassword {
		return errors.New("新密码不能与旧密码相同")
	}

	// 加密新密码
	hashedPwd, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	if err := s.authRepo.UpdatePassword(ctx, userID, hashedPwd); err != nil {
		return err
	}
	redisKey := fmt.Sprintf("auth:account:%d", user.ID)
	// 删除缓存
	cacheKey := userRedisKey(userID)
	if err := s.redisClient.Del(ctx, cacheKey, redisKey).Err(); err != nil {
		fmt.Printf("删除缓存失败 UserID: %d, err: %v\n", userID, err)
	}

	return nil
}

func (s *authServiceImpl) UpdateProfile(ctx context.Context, userID uint, username string, image string, bio string) error {
	if err := s.authRepo.UpdateProfile(ctx, userID, username, image, bio); err != nil {
		return err
	}
	// 删除缓存
	cacheKey := userRedisKey(userID)
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		fmt.Printf("删除缓存失败 UserID: %d, err: %v\n", userID, err)
	}
	return nil
}

// RefreshTokens 使用 Refresh Token 换取新的双 Token
func (s *authServiceImpl) RefreshTokens(ctx context.Context, accountID uint, incomingRT string, username string) (string, string, error) {
	redisKey := fmt.Sprintf("auth:account:%d", accountID)
	prevKey := redisKey + ":prev"
	// 1. 验证传入的 RT 是否与 Redis 中记录的当前合法 RT 一致
	storedRT, err := s.redisClient.HGet(ctx, redisKey, "refresh_token").Result()
	if err != nil || storedRT != incomingRT {
		// RT 不匹配或已失效，强制要求重新走密码登录
		prevRT, _ := s.redisClient.Get(ctx, prevKey).Result()
		if prevRT != "" && prevRT == incomingRT {
			sessionID, _ := s.redisClient.HGet(ctx, redisKey, "session_id").Result()
			if token, e := utils.GenerateToken(accountID, username, sessionID); e == nil {
				return token, storedRT, nil // 沿用当前 session 和 RT，只补发 access token
			}
		}
		return "", "", errors.New("RT 不匹配或已失效，强制要求重新走密码登录")
	}

	sessionID, err := s.redisClient.HGet(ctx, redisKey, "session_id").Result()
	if err != nil {
		return "", "", errors.New("会话已失效，请重新登录")
	}

	token, err := utils.GenerateToken(accountID, username, sessionID)
	if err != nil {
		return "", "", errors.New("生成 token 失败")
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	s.redisClient.Set(ctx, prevKey, storedRT, 30*time.Second)
	// 使用 Redis Hash 存储当前合法的 Session 和 RT，设置 7 天过期
	err = s.redisClient.HSet(ctx, redisKey,
		"session_id", sessionID,
		"refresh_token", refreshToken,
	).Err()
	if err != nil {
		return "", "", err
	}
	s.redisClient.Expire(ctx, redisKey, 7*24*time.Hour)

	return token, refreshToken, nil
}
