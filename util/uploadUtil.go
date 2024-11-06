// Package uploadUtil
// @Description: 关于文件上传的工具包
package uploadUtil

import (
	"context"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"os"
	"time"
)

const (
	endpoint        = ""    // MinIO 服务器地址
	accessKeyID     = ""    // MinIO Access Key
	secretAccessKey = ""    // MinIO Secret Key
	useSSL          = false // 是否使用 SSL
)

type MinioCoreClient struct {
	Client *minio.Core
}

type MinioClient struct {
	Client *minio.Client
}

// NewMinioCoreClient
//
//	@Description: 创建新的Minio客户端
//	@return *MinioCoreClient
//	@return error
func NewMinioCoreClient() (*MinioCoreClient, error) {
	minioClient, err := minio.NewCore(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &MinioCoreClient{Client: minioClient}, nil
}

func NewMinioClient() (*MinioClient, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &MinioClient{Client: minioClient}, nil
}

// ProgressReader 包装一个 io.Reader 以跟踪读取进度
type ProgressReader struct {
	io.Reader
	Total    int64
	Current  int64
	Callback func(int64, int64)
}

// Read 实现 io.Reader 接口，读取数据并更新进度
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if n > 0 {
		pr.Current += int64(n)
		pr.Callback(pr.Current, pr.Total)
	}
	return n, err
}

// UploadFile
//
//	@Description:上传文件到指定的存储桶，不支持断点续传
//	@receiver cli
//	@param bucketName
//	@param filePath
//	@param objectName
//	@return error
func (cli *MinioClient) UploadFile(bucketName, filePath, objectName string) error {
	ctx := context.Background()

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// 设置上传选项，特别是分块大小
	partSize := uint64(30 * 1024 * 1024) // 每块 500MB，这里可以根据需要调整，至少5MB
	opts := minio.PutObjectOptions{
		PartSize: partSize,
	}

	// 设置进度回调函数
	progressCallback := func(current, total int64) {
		fmt.Printf("\rUploaded %d of %d MBs (%.2f%%)", current/1024/1024, total/1024/1024, float64(current)/float64(total)*100)
	}

	progressReader := &ProgressReader{
		Reader:   file,
		Total:    fileInfo.Size(),
		Callback: progressCallback,
	}

	// 记录上传开始时间
	startTime := time.Now()

	// 上传文件
	n, err := cli.Client.PutObject(ctx, bucketName, objectName, progressReader, fileInfo.Size(), opts)
	if err != nil {
		return err
	}

	// 记录上传结束时间并计算总耗时
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime).Seconds()

	// 确保最终显示 100% 进度
	progressCallback(fileInfo.Size(), fileInfo.Size())

	// 显示上传结果
	fmt.Printf("\nFile uploaded successfully. Total uploaded: %d MBs\n", n.Size/1024/1024)
	fmt.Printf("Total upload time: %.2f s\n", elapsedTime)

	return nil
}

// UploadFileWithResume
//
//	@Description:  支持断点续传的上传文件
//	@receiver cli
//	@param bucketName
//	@param filePath
//	@param objectName
//	@return error
func (cli *MinioCoreClient) UploadFileWithResume(bucketName, filePath, objectName string) error {
	ctx := context.Background()
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	partSize := int64(40 * 1024 * 1024) // 设置分片大小
	totalSize := fileInfo.Size()

	// 设置进度回调函数
	progressCallback := func(current, total int64) {
		fmt.Printf("\rUploaded %.2f of %.2f MBs (%.2f%%)", float64(current)/1024/1024, float64(total)/1024/1024, float64(current)/float64(total)*100)
	}

	// 记录上传开始时间
	startTime := time.Now()

	// 查询是否存在未完成的上传
	uploads := cli.Client.ListIncompleteUploads(ctx, bucketName, objectName, false)

	var uploadID string
	for upload := range uploads {
		if upload.Key == objectName {
			uploadID = upload.UploadID
			break
		}
	}

	// 如果没有找到未完成的上传，创建新的上传ID
	if uploadID == "" {
		uploadID, err = cli.Client.NewMultipartUpload(ctx, bucketName, objectName, minio.PutObjectOptions{})
		if err != nil {
			return err
		}
	}

	// 获取已上传的分片信息
	partsInfo, err := cli.Client.ListObjectParts(ctx, bucketName, objectName, uploadID, 0, 1000)
	if err != nil {
		return err
	}

	// 创建一个映射来跟踪已上传的分片
	uploadedParts := make(map[int]struct{})
	for _, part := range partsInfo.ObjectParts {
		uploadedParts[part.PartNumber] = struct{}{}
	}

	var partNumber int
	offset := int64(0)
	completedParts := make([]minio.CompletePart, 0, len(uploadedParts))

	// 上传剩余的分片
	for offset < totalSize {
		partNumber++
		// 如果该分片已存在，跳过
		exist, size := partExists(partsInfo, partNumber, &completedParts)
		if exist {
			offset += size
			continue
		}

		// 计算当前分片的实际大小
		currentPartSize := partSize
		if offset+partSize > totalSize {
			currentPartSize = totalSize - offset // 最后一个分片的大小
		}

		// 读取分片数据
		// 读取分片数据并包装进度读取器
		sectionReader := io.NewSectionReader(file, offset, partSize)
		progressReader := &ProgressReader{
			Reader:   sectionReader,
			Total:    totalSize,
			Current:  offset,
			Callback: progressCallback,
		}

		// 上传分片
		partUpload, err := cli.Client.PutObjectPart(ctx, bucketName, objectName, uploadID, partNumber, progressReader, currentPartSize, minio.PutObjectPartOptions{})
		if err != nil {
			return err
		}

		offset += currentPartSize
		uploadedParts[partNumber] = struct{}{}
		completedParts = append(completedParts, minio.CompletePart{PartNumber: partNumber, ETag: partUpload.ETag})
	}

	// 收集已完成的分片
	if len(completedParts) == partNumber {
		_, err = cli.Client.CompleteMultipartUpload(ctx, bucketName, objectName, uploadID, completedParts, minio.PutObjectOptions{})
		if err != nil {
			return err
		}
	} else {
		return errors.New("文件未完全上传，请重试")
	}
	// 记录上传结束时间并计算总耗时
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime).Seconds()

	// 确保最终显示 100% 进度
	progressCallback(fileInfo.Size(), fileInfo.Size())
	// 显示上传结果
	fmt.Printf("\nFile uploaded successfully. Total uploaded: %.2f MBs\n", float64(fileInfo.Size())/1024/1024)
	minutes := int(elapsedTime) / 60
	seconds := int(elapsedTime) % 60
	fmt.Printf("Total upload time: %d min %d s\n", minutes, seconds)
	return nil
}

// partExists
//
//	@Description: 检查分片是否已经存在
//	@param parts
//	@param partNumber
//	@param completedParts
//	@return bool
//	@return int64
func partExists(parts minio.ListObjectPartsResult, partNumber int, completedParts *[]minio.CompletePart) (bool, int64) {
	for _, part := range parts.ObjectParts {
		if part.PartNumber == partNumber {
			*completedParts = append(*completedParts, minio.CompletePart{PartNumber: part.PartNumber, ETag: part.ETag})
			return true, part.Size
		}
	}
	return false, -1
}

type ProgressWriter struct {
	writer     io.Writer
	totalBytes int64
	downloaded int64
	startTime  time.Time
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	// 写入数据
	n, err = pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	// 更新已下载的字节数
	pw.downloaded += int64(n)

	// 计算已下载百分比
	percent := float64(pw.downloaded) / float64(pw.totalBytes) * 100

	// 转换为MB
	downloadedMB := float64(pw.downloaded) / (1024 * 1024)
	totalMB := float64(pw.totalBytes) / (1024 * 1024)

	// 打印进度
	fmt.Printf("\rDownloaded: %.2f MB of %.2f MB (%.2f%%)", downloadedMB, totalMB, percent)
	return n, err
}

// DownloadFile
//
//	@Description: 下载文件
//	@receiver cli
//	@param ctx
//	@param bucketName
//	@param objectName
//	@param filePath
//	@return error
func (cli *MinioCoreClient) DownloadFile(ctx context.Context, bucketName, objectName, filePath string) error {
	// 获取对象
	object, objectInfo, _, err := cli.Client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println("Error getting object:", err)
		return err
	}
	defer object.Close()

	// 创建本地文件
	localFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer localFile.Close()

	// 创建进度写入器
	pw := &ProgressWriter{
		writer:     localFile,
		totalBytes: objectInfo.Size, // 设置总字节数
		startTime:  time.Now(),
	}

	// 将对象内容复制到本地文件，同时跟踪进度
	_, err = io.Copy(pw, object)
	if err != nil {
		return fmt.Errorf("failed to copy object content to local file: %v", err)
	}
	duration := time.Since(pw.startTime)
	// 计算分钟和秒
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	fmt.Printf("\nDownload completed in %d minutes and %d seconds.\n", minutes, seconds)
	fmt.Println("\nDownload completed successfully!")
	return nil
}
