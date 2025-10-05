package ports

type ILogger interface {
	MustTrace(msg string)
	MustTraceErr(err error)
	MustDebug(msg string)
	MustDebugErr(err error)
	MustInfo(msg string)
	MustWarn(msg string)
	MustError(e error)
	MustFatal(msg string)
	MustPanic(msg string)
}
