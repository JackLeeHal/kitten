package log

import "go.uber.org/zap"

var (
	Logger *zap.SugaredLogger
)

func init() {
	l, _ := zap.NewProduction()
	defer l.Sync() // flushes buffer, if any
	Logger = l.Sugar()
}
