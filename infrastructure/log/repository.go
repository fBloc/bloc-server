package log

type Logger interface {
	SetName(name string) // ‘文件’名
	Infof(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	ForceUpload() // 强制“提交/刷盘...”、避免在内存/缓冲里放太久等...
}
