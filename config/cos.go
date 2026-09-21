package config

import (
	"go_bili/global"
	"net/http"
	"net/url"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func InitCOSClient() {
	c_cos := Appconf.Cos

	// 将 examplebucket-1250000000 和 COS_REGION 修改为用户真实的信息
	// 存储桶名称，由 bucketname-appid 组成，appid 必须填入，可以在 COS 控制台查看存储桶名称。https://console.cloud.tencent.com/cos5/bucket
	// COS_REGION 可以在控制台查看，https://console.cloud.tencent.com/cos5/bucket, 关于地域的详情见 https://cloud.tencent.com/document/product/436/6224
	u, _ := url.Parse(Appconf.Cos.BucketURL)
	// 用于 Get Service 查询，默认全地域 service.cos.myqcloud.com
	su, _ := url.Parse("https://service.cos.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}
	// 1.永久密钥
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  c_cos.SecretID,
			SecretKey: c_cos.SecretKey,
		},
	})
	global.CosClient = client
	global.CosBaseURL = Appconf.Cos.BucketURL
}
