package exception

type ErrorException struct {
	Code int32
	Msg  string
	Data any
}

func NewError(msg string, code int32, data any) ErrorException {
	return ErrorException{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func NewErrorEmptyData(msg string, code int32) ErrorException {
	return ErrorException{
		Code: code,
		Msg:  msg,
		Data: struct{}{},
	}
}
