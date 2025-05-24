// 包kafka提供了日志结构和方法，用于生成和编码日志消息。
package kafka

// 引入必要的包
import (
	"encoding/json"
	"project-common/tms"
	"time"
)

// FieldMap是一个键值对集合，用于存储日志中的自定义字段。
type FieldMap map[string]any

// KafkaLog定义了日志消息的结构。
// 它包含了日志的类型、操作、时间、消息内容、自定义字段和函数名称。
type KafkaLog struct {
	Type     string   // 日志类型
	Action   string   // 操作类型
	Time     string   // 时间戳
	Msg      string   // 消息内容
	Field    FieldMap // 自定义字段
	FuncName string   // 函数名称
}

// Error生成一个错误类型的日志消息。
// 参数err是错误类型，代表发生了错误。
// 参数funcName是字符串类型，代表错误发生的函数名称。
// 参数fieldMap是FieldMap类型，包含与日志相关的自定义字段。
// 返回值是编码后的日志消息的字节切片。
func Error(err error, funcName string, fieldMap FieldMap) []byte {
	// 创建KafkaLog结构体实例
	kl := KafkaLog{
		Type:     "error",
		Action:   "click",
		Time:     tms.Format(time.Now()),
		Msg:      err.Error(),
		Field:    fieldMap,
		FuncName: funcName,
	}
	// 将KafkaLog结构体编码为JSON格式
	bytes, _ := json.Marshal(kl)
	return bytes
}

// Info生成一个信息类型的日志消息。
// 参数msg是字符串类型，代表日志的消息内容。
// 参数funcName是字符串类型，代表日志发生的函数名称。
// 参数fieldMap是FieldMap类型，包含与日志相关的自定义字段。
// 返回值是编码后的日志消息的字节切片。
func Info(msg string, funcName string, fieldMap FieldMap) []byte {
	// 创建KafkaLog结构体实例
	kl := KafkaLog{
		Type:     "info",
		Action:   "click",
		Time:     tms.Format(time.Now()),
		Msg:      msg,
		Field:    fieldMap,
		FuncName: funcName,
	}
	// 将KafkaLog结构体编码为JSON格式
	bytes, _ := json.Marshal(kl)
	return bytes
}
