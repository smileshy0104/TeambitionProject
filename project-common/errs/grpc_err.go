package errs

import (
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	common "project-common"
)

// GrpcError 将自定义错误 (*BError) 转换为 gRPC 错误。
func GrpcError(err *BError) error {
	return status.Error(codes.Code(err.Code), err.Msg)
}

// ParseGrpcError 解析 gRPC 错误，提取业务码和消息。
func ParseGrpcError(err error) (common.BusinessCode, string) {
	fromError, _ := status.FromError(err)
	return common.BusinessCode(fromError.Code()), fromError.Message()
}
