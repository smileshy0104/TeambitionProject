package minio

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"strconv"
)

// MinioClient 是一个封装了Minio客户端的结构体
type MinioClient struct {
	c *minio.Client
}

// Put 将数据上传到Minio服务器
// 参数:
//
//	ctx: 上下文，用于传递请求相关的配置和元数据
//	bucketName: 存储桶名称
//	fileName: 文件名
//	data: 要上传的数据
//	size: 数据的大小
//	contentType: 数据的MIME类型
//
// 返回值:
//
//	minio.UploadInfo: 上传信息，包含上传对象的元数据
//	error: 错误信息，如果上传过程中发生错误
func (c *MinioClient) Put(ctx context.Context, bucketName string, fileName string,
	data []byte, size int64, contentType string) (minio.UploadInfo, error) {
	// 调用Minio客户端的PutObject方法将数据上传到Minio服务器。
	object, err := c.c.PutObject(
		ctx,                   // 上下文，用于传递请求相关的配置和元数据
		bucketName,            // 存储桶名称
		fileName,              // 文件名
		bytes.NewBuffer(data), // 数据
		size,                  // 数据的大小
		minio.PutObjectOptions{ContentType: contentType}, // 数据的MIME类型
	)
	return object, err
}

// Compose 将多个对象合并成一个对象
// 参数:
//
//	ctx: 上下文，用于传递请求相关的配置和元数据
//	bucketName: 存储桶名称
//	fileName: 合并后的文件名
//	totalChunks: 总共要合并的分块数
//
// 返回值:
//
//	minio.UploadInfo: 合并后的对象信息
//	error: 错误信息，如果合并过程中发生错误
func (c *MinioClient) Compose(ctx context.Context, bucketName string, fileName string, totalChunks int) (minio.UploadInfo, error) {
	// 创建一个CopyDestOptions实例，用于指定合并后的对象存储桶和名称。
	dst := minio.CopyDestOptions{
		Bucket: bucketName,
		Object: fileName,
	}
	// 创建一个CopySrcOptions实例，用于指定每个分块的存储桶和名称。
	var srcs []minio.CopySrcOptions
	// 循环创建CopySrcOptions实例，并添加到srcs切片中。
	for i := 1; i <= totalChunks; i++ {
		formatInt := strconv.FormatInt(int64(i), 10)
		src := minio.CopySrcOptions{
			Bucket: bucketName,
			Object: fileName + "_" + formatInt,
		}
		// 添加CopySrcOptions实例到srcs切片中。
		srcs = append(srcs, src)
	}
	// 调用Minio客户端的ComposeObject方法将分块合并成一个对象。
	object, err := c.c.ComposeObject(
		ctx,
		dst,
		srcs...,
	)
	return object, err
}

// New 创建并初始化一个新的Minio客户端
// 参数:
//
//	endpoint: Minio服务器的地址
//	accessKey: 访问密钥
//	secretKey: 秘密密钥
//	useSSL: 是否使用SSL连接
//
// 返回值:
//
//	*MinioClient: 初始化后的Minio客户端
//	error: 错误信息，如果初始化过程中发生错误
func New(endpoint, accessKey, secretKey string, useSSL bool) (*MinioClient, error) {
	// 创建并初始化一个新的Minio客户端。
	minioClient, err := minio.New(endpoint, &minio.Options{
		// 使用静态凭证创建Minio客户端。
		Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
		// 指定是否使用SSL连接。
		Secure: useSSL,
	})
	// 返回初始化后的Minio客户端和错误信息。
	return &MinioClient{c: minioClient}, err
}
