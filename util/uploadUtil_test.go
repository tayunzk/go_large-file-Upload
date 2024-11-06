package uploadUtil

import (
	"context"
	"github.com/minio/minio-go/v7"
	"testing"
)

func TestMinioClient_UploadFileWithResume(t *testing.T) {
	type fields struct {
		Client *minio.Core
	}
	type args struct {
		bucketName string
		filePath   string
		objectName string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				bucketName: "test",
				filePath:   "/Users/tayun/DockerImages/test-image.tar",
				objectName: "test-image.tar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minioClient, err := NewMinioCoreClient()
			if err != nil {
				t.Errorf("client failed, error:%v ", err)
			}

			err = minioClient.UploadFileWithResume(tt.args.bucketName, tt.args.filePath, tt.args.objectName)
			if err != nil {
				t.Errorf("Error UploadFileWithResume : %v", err)
			}
		})
	}
}

func TestNewMinioClient(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMinioClient()
			if err != nil {
				t.Errorf("NewMinioClient() error = %v", err)
			}
		})
	}
}

func TestMinioCoreClient_DownloadFile(t *testing.T) {
	type fields struct {
		Client *minio.Core
	}
	type args struct {
		ctx        context.Context
		bucketName string
		objectName string
		filePath   string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				ctx:        context.Background(),
				bucketName: "test",
				objectName: "alpine_latest.tar",
				filePath:   "/Users/tayun/DockerImages/Downloads/alpine_latest.tar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minioClient, err := NewMinioCoreClient()
			if err != nil {
				t.Errorf("client failed, error:%v ", err)
			}

			err = minioClient.DownloadFile(tt.args.ctx, tt.args.bucketName, tt.args.objectName, tt.args.filePath)
			if err != nil {
				t.Errorf("Error DownloadFile : %v", err)
			}
		})
	}
}
