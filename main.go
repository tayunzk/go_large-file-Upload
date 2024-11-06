package main

import (
	uploadUtil "Leetcode/util"
	"context"
	"log"
)

const (
	bucketName       = "test"
	objectName       = "mysql.tar.gz"
	uploadFilePath   = "/Users/tayun/DockerImages/mysql.tar.gz"
	downloadFilePath = "/Users/tayun/DockerImages/Downloads/mysql.tar.gz"
)

func main() {
	// 创建 Minio 客户端
	minioClient, err := uploadUtil.NewMinioCoreClient()
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	// 上传文件
	//err = minioClient.UploadFileWithResume(bucketName, uploadFilePath, objectName)
	//if err != nil {
	//	log.Fatalln(err)
	//}

	// 下载文件
	err = minioClient.DownloadFile(ctx, bucketName, objectName, downloadFilePath)
	if err != nil {
		log.Fatalln(err)
	}
}
