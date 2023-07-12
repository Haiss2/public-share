package main

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L = SetupLogger()

const (
	debug = true
)

func SetupLogger() *zap.SugaredLogger {
	pConf := zap.NewProductionEncoderConfig()
	pConf.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewConsoleEncoder(pConf)
	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	if debug {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	l := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level), zap.AddCaller())
	zap.ReplaceGlobals(l)
	return zap.S()
}

func Mean(a []int64) float64 {
	if len(a) == 0 {
		return 0.0
	}

	sum := float64(0)
	for _, n := range a {
		sum += float64(n)
	}
	return sum / float64(len(a))
}
