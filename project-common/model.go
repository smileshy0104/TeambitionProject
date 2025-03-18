package common

// 定义业务状态码类型，使用整数表示
type BusinessCode int

// 定义统一的响应结果结构体，包含业务状态码、消息和数据
type Result struct {
	Code BusinessCode `json:"code"` // 业务状态码，用于标识响应的状态
	Msg  string       `json:"msg"`  // 响应消息，通常用于描述成功或失败的原因
	Data any          `json:"data"` // 响应数据，可以是任意类型的业务数据
}

// Success 方法用于创建一个成功的响应结果
func (r *Result) Success(data any) *Result {
	r.Code = 200      // 设置业务状态码为 200，表示成功
	r.Msg = "success" // 设置消息为 "success"
	r.Data = data     // 设置需要返回的数据
	return r          // 返回当前 Result 对象
}

// Fail 方法用于创建一个失败的响应结果
func (r *Result) Fail(code BusinessCode, msg string) *Result {
	r.Code = code // 设置业务状态码为指定的错误码
	r.Msg = msg   // 设置错误消息
	return r      // 返回当前 Result 对象
}
