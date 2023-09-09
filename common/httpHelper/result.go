package httpHeper

type Result struct {
	Code int32
	Msg  string
	Data any
}

func Error(msg string, code int32, data any) Result {
	if code == 0 {
		code = CodeMsgError
	}
	return Result{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func Success(msg string, code int32, data any) Result {
	if code == 0 {
		code = CodeSuccess
	}
	if msg == "" {
		msg = GetMsgByCode(code)
	}
	return Result{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}
