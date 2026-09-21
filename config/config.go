package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name string
		Port string
	}
	Database struct {
		Dsn          string
		MaxIdleConns int
		MaxOpenConns int
		Addr         string
		Password     string
		SubSwitch    bool //是否开启redis分布式锁，默认为false
	}
	JWT struct {
		Key string
	}
	RabbitMQ struct {
		Url string
	}
	GitHub struct {
		ClientID     string
		ClientSecret string
		RedirectURI  string
		FrontendURL  string
	}
	Cos struct {
		SecretID  string
		SecretKey string
		BucketURL string
	}
}

var Appconf *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	Appconf = &Config{}

	if err := viper.Unmarshal(Appconf); err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}

	// 生产环境的 OAuth secret 不进版本库（仓库是公开的），
	// 由 docker-compose.prod.yml 的 environment 注入，这里用环境变量覆盖。
	// 必须放在 Unmarshal 之后：Appconf 是在上面才被 &Config{} 赋值的，
	// 放在赋值之前访问它就是 nil 指针 panic
	if secret := os.Getenv("GITHUB_CLIENT_SECRET"); secret != "" {
		Appconf.GitHub.ClientSecret = secret
	}
	// COS 的密钥同理，环境变量名加上 COS_ 前缀避免和宿主机上别的变量撞名
	if id := os.Getenv("COS_SECRET_ID"); id != "" {
		Appconf.Cos.SecretID = id
	}
	if key := os.Getenv("COS_SECRET_KEY"); key != "" {
		Appconf.Cos.SecretKey = key
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	initDB()
	initRedis(ctx)
	InitRabbitMQ()
	InitCOSClient()
}
