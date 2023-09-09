package jsonRpcHelper

const (
	// Parse error语法解析错误
	// 服务端接收到无效的json。该错误发送于服务器尝试解析json文本
	CodeParseError = -32700

	// 发送的json不是一个有效的请求对象。
	CodeInvalidRequest = -32600

	// 该方法不存在或无效
	CodeMethodNotFound = -32601

	// 无效的方法参数。
	CodeInvalidParams = -32602

	// 服务器错误
	CodeInternalError = -32603
)

func GetMsgByCode(code int32) string {
	list := map[int32]string{
		CodeParseError:     "json解析失败",
		CodeInvalidRequest: "json格式错误",
		CodeMethodNotFound: "方法不存在或无效",
		CodeInternalError:  "服务器错误",
	}
	if value, ok := list[code]; ok {
		return value
	}
	return ""
}
