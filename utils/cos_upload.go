package utils

import (
	"context"
	"go_bili/global"
	"os"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func Upload(key string, filePath string, contentType string) (string, error) {
	// 对象键（Key）是对象在存储桶中的唯一标识。
	// 例如，在对象的访问域名 `examplebucket-1250000000.cos.COS_REGION.myqcloud.com/test/objectPut.go` 中，对象键为 test/objectPut.go

	// 通过本地文件上传对象
	_, err := global.CosClient.Object.PutFromFile(context.Background(), key, filePath, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	})
	if err != nil {
		return "", err
	}
	os.Remove(filePath)
	return global.CosBaseURL + "/" + key, nil
}

func Download(key string, Path string) error {

	// 2.获取对象到本地文件
	_, err := global.CosClient.Object.GetToFile(context.Background(), key, Path, nil)
	return err
}

func Delete(key string) error {
	_, err := global.CosClient.Object.Delete(context.Background(), key)
	if err != nil {
		return err
	}
	return nil
}
