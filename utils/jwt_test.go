package utils // 注意：和源码同包，才能直接调用未导出的 jwtSecret()

import (
	"os"
	"testing"
	"time"

	"go_bili/config"

	"github.com/golang-jwt/jwt/v5"
)

// TestMain 包内所有测试跑之前执行一次，做初始化
func TestMain(m *testing.M) {
	config.Appconf = &config.Config{}
	config.Appconf.JWT.Key = "test-secret-key" // 固定密钥，保证签发和校验用同一把
	os.Exit(m.Run())
}

// 测试1：签发 → 解析，字段能 round-trip 对得上
func TestGenerateTokenRoundTrip(t *testing.T) {
	token, err := GenerateToken(42, "bili", "session-abc")
	if err != nil {
		t.Fatalf("GenerateToken 失败: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken 失败: %v", err)
	}

	if claims.AccountID != 42 {
		t.Errorf("AccountID 期望 42，实际 %d", claims.AccountID)
	}
	if claims.Username != "bili" {
		t.Errorf("Username 期望 bili，实际 %s", claims.Username)
	}
	if claims.SessionID != "session-abc" {
		t.Errorf("SessionID 期望 session-abc，实际 %s", claims.SessionID)
	}
}

// 测试2：被篡改的 token 必须解析失败（安全底线）
func TestParseTokenRejectsTampered(t *testing.T) {
	token, _ := GenerateToken(1, "u", "s")
	tampered := token + "x" // 改掉最后一位，签名就对不上了

	if _, err := ParseToken(tampered); err == nil {
		t.Errorf("被篡改的 token 居然解析通过了，这是安全漏洞")
	}
}

// 测试3：过期 token，严格版拒绝、宽容版仍能取出身份（对应你今天学的那个知识点）
func TestParseTokenAllowExpiredAcceptsExpired(t *testing.T) {
	// 白盒测试：同包能直接调 jwtSecret()，手工造一个已过期的 token
	claims := Claims{
		AccountID: 7,
		Username:  "old-user",
		SessionID: "old-session",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)), // 1 分钟前已过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expired, err := tk.SignedString(jwtSecret())
	if err != nil {
		t.Fatalf("构造过期 token 失败: %v", err)
	}

	// 严格版：必须拒绝过期的
	if _, err := ParseToken(expired); err == nil {
		t.Errorf("过期的 token 应该被 ParseToken 拒绝")
	}

	// 宽容版：应能取回身份
	got, err := ParseTokenAllowExpired(expired)
	if err != nil {
		t.Fatalf("ParseTokenAllowExpired 失败: %v", err)
	}
	if got.AccountID != 7 {
		t.Errorf("AccountID 期望 7，实际 %d", got.AccountID)
	}
}
