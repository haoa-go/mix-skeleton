package jsonRpcHelper

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func NewError(code int32, message string) *Error {
	data := struct{}{}
	return NewErrorData(code, message, data)
}

func NewErrorData(code int32, message string, data any) *Error {
	if message == "" {
		message = GetMsgByCode(code)
	}
	return &Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
