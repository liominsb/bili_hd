package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go_bili/api/repository"
	"go_bili/config"
	"go_bili/models"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	oauthProviderGitHub = "github"
	githubAuthorizeURL  = "https://github.com/login/oauth/authorize"
	githubTokenURL      = "https://github.com/login/oauth/access_token"
	githubUserAPI       = "https://api.github.com/user"
	oauthStateTTL       = 5 * time.Minute
)

// 单独一个带超时的 client：调外网接口绝不能让它无限等，
// 否则对方不响应时 Goroutine 会一直挂在那里
var oauthHTTPClient = &http.Client{Timeout: 10 * time.Second}

type OAuthService interface {
	// GitHubLoginURL 生成 GitHub 授权地址，并把 state 存进 Redis（防 CSRF）
	GitHubLoginURL(ctx context.Context) (string, error)
	// GitHubCallback 处理回调：校验 state → code 换 token → 取用户 → 找/建本地用户 → 签发本地双 token
	GitHubCallback(ctx context.Context, code string, state string) (string, string, error)
}

type oauthServiceImpl struct {
	oauthRepo   repository.OAuthRepository
	authRepo    repository.AuthRepository
	authService AuthService
	redisClient *redis.Client
}

// 依赖里多了一个 authService：OAuth 走完之后要复用它的 IssueSession 签发本地会话
func NewOAuthService(oauthRepo repository.OAuthRepository, authRepo repository.AuthRepository, authService AuthService, redisClient *redis.Client) OAuthService {
	return &oauthServiceImpl{
		oauthRepo:   oauthRepo,
		authRepo:    authRepo,
		authService: authService,
		redisClient: redisClient,
	}
}

// ---------- 第一步：让浏览器跳到 GitHub ----------

func (s *oauthServiceImpl) GitHubLoginURL(ctx context.Context) (string, error) {
	cfg := config.Appconf.GitHub
	if cfg.ClientID == "" {
		return "", errors.New("GitHub OAuth 未配置 ClientID")
	}

	// state 是一次性随机串：GitHub 原样带回来，我们靠它确认「这次回调确实是我发起的」。
	state := uuid.New().String()
	if err := s.redisClient.Set(ctx, "oauth:state:"+state, oauthProviderGitHub, oauthStateTTL).Err(); err != nil {
		return "", fmt.Errorf("保存 state 失败: %w", err)
	}

	params := url.Values{}
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", cfg.RedirectURI)
	params.Set("scope", "read:user")
	params.Set("state", state)

	return githubAuthorizeURL + "?" + params.Encode(), nil
}

// ---------- 第二步：处理 GitHub 回调 ----------

func (s *oauthServiceImpl) GitHubCallback(ctx context.Context, code string, state string) (string, string, error) {
	// 1. 校验 state（一次性，用完即删）
	if err := s.consumeState(ctx, state); err != nil {
		return "", "", err
	}
	if code == "" {
		return "", "", errors.New("缺少 code 参数")
	}

	// 2. 用 code 换 GitHub 的 access_token（必须在后端做，因为要带 client_secret）
	accessToken, err := s.exchangeCode(ctx, code)
	if err != nil {
		return "", "", err
	}

	// 3. 拿 GitHub 用户信息
	ghUser, err := s.fetchGitHubUser(ctx, accessToken)
	if err != nil {
		return "", "", err
	}

	// 4. 认人：只认 GitHub 的数字 id，不认 login
	//    GitHub 用户名可以改名，旧名会被释放，别人抢注同名就能登进你的账号
	providerUID := strconv.FormatInt(ghUser.ID, 10)

	var binding models.OAuthBinding
	err = s.oauthRepo.GetBinding(ctx, &binding, oauthProviderGitHub, providerUID)

	var user models.User
	switch {
	case err == nil:
		// 之前绑过 → 找出对应的本地用户
		if err := s.authRepo.GetUserByID(ctx, &user, binding.UserID); err != nil {
			return "", "", fmt.Errorf("绑定存在但本地用户查不到: %w", err)
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		// 第一次用 GitHub 登录 → 建本地用户 + 建绑定
		user, err = s.createUserWithBinding(ctx, ghUser, providerUID)
		if err != nil {
			return "", "", err
		}
	default:
		return "", "", err
	}

	// 5. 复用密码登录那套签发逻辑，出来的双 token 和账号密码登录完全一样，
	//    所以前端的请求拦截器、刷新逻辑一行都不用改
	return s.authService.IssueSession(ctx, &user)
}

// consumeState 校验并消费 state，同一个 state 不能用第二次（防重放）
func (s *oauthServiceImpl) consumeState(ctx context.Context, state string) error {
	if state == "" {
		return errors.New("缺少 state 参数")
	}
	key := "oauth:state:" + state
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil || val == "" {
		return errors.New("state 无效或已过期，请重新发起登录")
	}
	s.redisClient.Del(ctx, key)
	return nil
}

// exchangeCode 用授权码换 access_token：这是唯一需要 client_secret 的一步，只能在后端做
func (s *oauthServiceImpl) exchangeCode(ctx context.Context, code string) (string, error) {
	cfg := config.Appconf.GitHub

	form := url.Values{}
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", cfg.RedirectURI) // 必须和授权时带的一模一样

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 不加这个，GitHub 默认返回 form-encoded 文本，还得自己手工解析
	req.Header.Set("Accept", "application/json")

	resp, err := oauthHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 GitHub 换取 token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 限 1MB，防意外大响应
	if err != nil {
		return "", err
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析 GitHub token 响应失败: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("GitHub 返回错误: %s (%s)", result.Error, result.ErrorDesc)
	}
	if result.AccessToken == "" {
		return "", errors.New("GitHub 未返回 access_token")
	}
	return result.AccessToken, nil
}

// githubUser 只取我们真正要用的三个字段
type githubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

func (s *oauthServiceImpl) fetchGitHubUser(ctx context.Context, accessToken string) (*githubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := oauthHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub 用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub 用户接口返回 %d: %s", resp.StatusCode, string(body))
	}

	var u githubUser
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, fmt.Errorf("解析 GitHub 用户信息失败: %w", err)
	}
	if u.ID == 0 || u.Login == "" {
		return nil, errors.New("GitHub 返回的用户信息不完整")
	}
	return &u, nil
}

// createUserWithBinding 首次用第三方登录时，建一个本地用户并记下绑定关系
func (s *oauthServiceImpl) createUserWithBinding(ctx context.Context, ghUser *githubUser, providerUID string) (models.User, error) {
	var user models.User
	user.Username = s.uniqueUsername(ctx, ghUser.Login)
	// 第三方登录没有本地密码。空串存进去是安全的：
	// CheckPassword 拿空 hash 去 bcrypt 比对必然失败，不会出现「空密码能登进来」
	user.Password = ""
	user.Image = ghUser.AvatarURL

	if err := s.authRepo.Register(ctx, &user); err != nil {
		return models.User{}, err
	}

	binding := models.OAuthBinding{
		UserID:      user.ID,
		Provider:    oauthProviderGitHub,
		ProviderUID: providerUID,
		Nickname:    ghUser.Login,
	}
	if err := s.oauthRepo.CreateBinding(ctx, &binding); err != nil {
		// 用户建好了但绑定没建成，会留下一个没有第三方关系的空账号。
		// 不影响之后登录（下次会重新建），属于极端情况，先不为它引入跨 repo 事务。
		// 数据库那边的 idx_provider_uid 唯一索引会给并发建绑定兜底。
		return models.User{}, fmt.Errorf("创建绑定失败: %w", err)
	}

	return user, nil
}

// uniqueUsername GitHub 用户名可能和站内已有用户名撞车，撞了就加 _gh1、_gh2 后缀
func (s *oauthServiceImpl) uniqueUsername(ctx context.Context, base string) string {
	// users.username 是 size:50，GitHub 用户名最长 39，留出后缀空间
	if len(base) > 40 {
		base = base[:40]
	}

	for i := 0; i < 20; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s_gh%d", base, i)
		}
		var u models.User
		// GetUserByUsername 内部是 First：查到才返回 nil，
		// 所以只要返回了 err 就说明这个名字当前没被占用（gorm.ErrRecordNotFound）
		if err := s.authRepo.GetUserByUsername(ctx, &u, candidate); err != nil {
			return candidate
		}
	}

	// 兜底：20 次都撞（几乎不可能），加短 uuid 保证一定能创建
	return base + "_" + uuid.New().String()[:8]
}
