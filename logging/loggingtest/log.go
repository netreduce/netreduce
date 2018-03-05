package loggingtest

type NopLog struct {}

func (l NopLog) Error(...interface{}) {}
