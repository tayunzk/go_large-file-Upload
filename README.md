# go_large-file-Upload
Large File Upload Utility for Golang with MinIO
This utility is designed to help you upload large files in Golang using MinIO. The tool supports file chunking, resuming uploads, and visualizing progress during file upload and download.

## **Features**  
**Chunked Upload**: Upload large files in parts.  
**Resume Upload**: Support for resuming uploads in case of interruption.  
**Progress Visualizatio**n: Real-time progress update during upload and download.  
**MinIO Integration**: Works with MinIO, a high-performance object storage service.  


## **Setup Instructions**  
**Prerequisites**:  
MinIO server: Set up and run a MinIO server. MinIO Setup Guide
Golang: Ensure Go is installed in your system.
**Configuration**:
In the uploadUtil.go file, configure the following variables with your MinIO server details:
```
endpoint = "" // MinIO server address
accessKeyID = "" // MinIO Access Key
secretAccessKey = "" // MinIO Secret Key
```
## File Upload:
+ The tool supports uploading files in chunks. The progress of each upload is shown as a percentage.
+ If the upload is interrupted, it can be resumed from the last successfully uploaded chunk.
## File Download:
+ You can also use the tool to download files from your MinIO server with progress visualization.
Example Usage: see util_test.go 
## Progress Visualization:
  The progress during the upload will be displayed like this:
  ```
  Uploaded 11831.97 of 11831.97 MBs (100.00%)
  File uploaded successfully.
  Total uploaded: 11831.97 MBs
  Total upload time: 988.31 s

  ```
# Golang大文件上传工具包（基于MinIO）
该工具包旨在帮助您使用Golang和MinIO上传大文件。该工具支持文件分片上传、断点续传，并在上传和下载过程中提供实时进度可视化。

## 特性
+ 分片上传：将大文件分块上传。
+ 断点续传：支持上传中断后的续传。
+ 进度可视化：上传和下载过程中实时更新进度。
+ MinIO集成：与MinIO高性能对象存储服务兼容。
## 安装与配置
### 前提条件：
+ MinIO服务器：安装并启动MinIO服务器。可参考MinIO安装指南。
+ Go语言：确保您的系统已安装Go。
### 配置：
+ 在uploadUtil.go文件中，配置以下变量值以连接您的MinIO服务器：
```azure
endpoint = "" // MinIO服务器地址
accessKeyID = "" // MinIO Access Key
secretAccessKey = "" // MinIO Secret Key

```
### 文件上传：
+ 工具支持分块上传大文件，上传过程中显示实时的进度百分比。
+ 如果上传被中断，可以从上次上传成功的分片继续上传。
### 文件下载：
+ 该工具也支持从MinIO服务器下载文件，并显示下载进度。
使用示例：见util_test.go

### 上传进度可视化：
上传过程中，进度将以如下方式显示：
```azure
Uploaded 11831.97 of 11831.97 MBs (100.00%)
文件上传成功。
总上传量: 11831.97 MBs
总上传时间: 988.31 s

```