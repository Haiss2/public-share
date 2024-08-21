package main

import (
	"math"
	"os"
	"path"
	"runtime"
	"strconv"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func SetupLogger() *zap.SugaredLogger {
	pConf := zap.NewProductionEncoderConfig()
	pConf.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewConsoleEncoder(pConf)
	level := zap.NewAtomicLevelAt(zap.DebugLevel)
	l := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level), zap.AddCaller())
	zap.ReplaceGlobals(l)
	return zap.S()
}

// RoundingNumberDown ...
func RoundDown(n float64, precision int) float64 {
	rounding := math.Pow10(precision)
	return math.Floor(n*rounding) / rounding
}

func round(raw float64) float64 {
	if raw < 1e-5 {
		return RoundDown(raw, 6)
	}
	if raw < 1e-4 {
		return RoundDown(raw, 5)
	}
	if raw < 1e-3 {
		return RoundDown(raw, 4)
	}
	if raw < 1e-2 {
		return RoundDown(raw, 3)
	}
	if raw < 1e-1 {
		return RoundDown(raw, 2)
	}
	if raw < 1 {
		return RoundDown(raw, 1)
	}

	return RoundDown(raw, 0)
}

func StringToFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		pc, file, line, _ := runtime.Caller(1)
		zap.S().Errorw(
			"Failed to ParseStringToFloat!",
			"err", err, "raw", s,
			"funcName", runtime.FuncForPC(pc).Name(),
			"file", path.Base(file), "line", line,
		)
	}
	return f
}

func FloatToString(f float64) string {
	df := decimal.NewFromFloat(f)
	return df.String()
}

func IntToString(d int64) string {
	return strconv.FormatInt(d, 10)
}
