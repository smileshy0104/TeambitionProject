package errs

import "fmt"

// ErrorCode 是错误代码的类型，用于标识不同的错误类型。
type ErrorCode int

// BError 是自定义错误类型的结构体，包含了错误代码和错误消息。
type BError struct {
	Code ErrorCode // 错误代码，用于标识错误类型
	Msg  string    // 错误消息，描述错误的详细信息
}

// Error 实现了 error 接口，返回包含错误代码和错误消息的字符串。
func (e *BError) Error() string {
	return fmt.Sprintf("code:%v,msg:%s", e.Code, e.Msg)
}

// NewError 创建并返回一个新的 BError 实例。
// 参数 code 是错误代码，msg 是错误消息。
func NewError(code ErrorCode, msg string) *BError {
	return &BError{
		Code: code,
		Msg:  msg,
	}
}
