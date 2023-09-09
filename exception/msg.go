package exception

type MsgException struct {
	Code int32
	Msg  string
	Data any
}

func NewMsg(msg string, code int32, data any) MsgException {
	return MsgException{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func NewMsgEmptyData(msg string, code int32) MsgException {
	return MsgException{
		Code: code,
		Msg:  msg,
		Data: struct {
		}{},
	}
}
