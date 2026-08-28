package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"go_bili/config"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func jwtSecret() []byte {
	//开发测试不用
	//secret := os.Getenv("JWT_SECRET")
	secret := config.Appconf.JWT.Key
	if secret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Println("生成随机 JWT 密钥失败:", err)
			return []byte("default_secret_key")
		}
		secret = hex.EncodeToString(b)
		log.Println("警告: JWT_SECRET 环境变量未设置，使用随机生成的密钥:", secret)
	}
	return []byte(secret)
}

// Claims 自定义 Payload 结构体
type Claims struct {
	AccountID uint   `json:"account_id"`
	Username  string `json:"username"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// GenerateRefreshToken 签发不透明的刷新令牌 (极度安全，无内在状态，全靠 Redis/DB 校验)
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateToken 签发 Access Token (JWT)
func GenerateToken(accountID uint, username string, sessionID string) (string, error) {
	now := time.Now()

	claims := Claims{
		AccountID: accountID,
		Username:  username,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)), // 强制短有效，降低泄露风险
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

// ParseToken 解析并校验 JWT
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// 严格校验签名算法，防止攻击者篡改 Header 将算法改为 None 或对称算法
			if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("意外签名方法")
			}
			return jwtSecret(), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("无效token")
	}

	return claims, nil
}

func ParseTokenAllowExpired(tokenString string) (*Claims, error) {
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[7:]
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("意外签名方法")
			}
			return jwtSecret(), nil
		},
	)

	// 即使token过期，只要签名正确就返回claims
	if errors.Is(err, jwt.ErrTokenExpired) {
		if claims, ok := token.Claims.(*Claims); ok {
			return claims, nil
		}
	}

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("无效token")
	}
	return claims, nil
}

//// GenerateJWT 生成JWT
//func GenerateJWT(id uint) (string, error) { //生成JWT
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
//		"ID":  id,
//		"exp": time.Now().Add(72 * time.Hour).Unix(),
//	})
//	signedToken, err := token.SignedString([]byte(config.Appconf.JWT.Key))
//	return "Bearer " + signedToken, err
//}

//// ParseJWT 解析JWT
//func ParseJWT(tokenString string) (uint, error) { //
//	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
//		tokenString = tokenString[7:]
//	}
//	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
//		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
//		}
//		return []byte(config.Appconf.JWT.Key), nil
//	})
//	if err != nil {
//		return 0, err
//	}
//	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
//		id, ok := claims["ID"].(float64)
//		if !ok {
//			return 0, errors.New("invalid token")
//		}
//		return uint(id), nil
//	}
//	return 0, errors.New("invalid token")
//}
