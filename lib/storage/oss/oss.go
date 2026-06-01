package oss

import (
	"bytes"
	"fmt"
	"log"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/oldfritter/lucy/base"
)

var (
	OSSConfig map[string]string
	OssClient *oss.Client
)

func init() {
	if len(OSSConfig) == 0 {
		RemoteInitOss()
	}
}

func RemoteInitOss() {
	config := base.GetConfig()
	OSSConfig = map[string]string{
		"Asset":           config.Get("oss.Asset", ""),
		"Endpoint":        config.Get("oss.Endpoint", ""),
		"AccessKeyId":     config.Get("oss.AccessKeyId", ""),
		"AccessKeySecret": config.Get("oss.AccessKeySecret", ""),
		"AssetsBucket":    config.Get("oss.AssetsBucket", ""),
	}
	OssClient, _ = oss.New(
		OSSConfig["Endpoint"],
		OSSConfig["AccessKeyId"],
		OSSConfig["AccessKeySecret"],
	)
}

func OssAssetsBucket() string {
	return OSSConfig["AssetsBucket"]
}

func OssAsset() string {
	return OSSConfig["Asset"]
}

func PutObject(key string, b *[]byte) (url string, err error) {
	if bucket, err := OssClient.Bucket(OssAssetsBucket()); err != nil {
		log.Println(err)
	} else {
		if err = bucket.PutObject(
			key,
			bytes.NewBuffer(*b),
			oss.ObjectStorageClass(oss.StorageStandard),
			oss.ObjectACL(oss.ACLPrivate),
		); err != nil {
			log.Println("Error:", err)
			panic(err)
		} else {
			url = fmt.Sprintf("%s/%s", OSSConfig["Asset"], key)
		}
	}
	return
}

func DeleteObject(key string) error {
	bucket, err := OssClient.Bucket(OssAssetsBucket())
	if err != nil {
		fmt.Println(err)
		log.Println(err)
	}
	return bucket.DeleteObject(key)
}

// GetObjectURL 生成带签名的临时下载链接，expireSec 为有效期（秒）
func GetObjectURL(key string, expireSec int64) (string, error) {
	bucket, err := OssClient.Bucket(OssAssetsBucket())
	if err != nil {
		return "", err
	}
	return bucket.SignURL(key, oss.HTTPGet, expireSec)
}
