package main

import (
	"encoding/json"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func setupLogger(debug bool) *zap.SugaredLogger {
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

func saveData(data interface{}, filename string) error {
	f, err := os.Create(*storagePath + filename)
	if err != nil {
		L.Error("Fail to create file", "filename", filename, "error", err)
		return err
	}

	err = json.NewEncoder(f).Encode(data)
	if err != nil {
		L.Error("Fail to write data to file", "error", err)
		return err
	}

	return nil
}
