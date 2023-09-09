package httpHeper

const (
	// 成功
	CodeSuccess = 0

	// 通用报错
	CodeMsgError = 4000

	// 服务器错误
	CodeServerError = 5000
)

func GetMsgByCode(code int32) string {
	list := map[int32]string{
		CodeServerError: "服务器错误",
		CodeSuccess:     "success",
	}
	if value, ok := list[code]; ok {
		return value
	}
	return ""
}
