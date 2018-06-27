package config

import (
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/robfig/cron"
	"log"
	"io"
	"os"
	"go.uber.org/zap"
)

var LOG *zap.Logger
var logCron *cron.Cron

type Logger struct {
	Filename    string
	Level       string
	Encoding    string
	MaxSize     int `toml:"max_size"`
	MaxAge      int `toml:"max_age"`
	Development bool
}

//log 配置
func StartLogger() {
	var lc = Prop.Logger

	if lc.MaxAge == 0 {
		lc.MaxAge = 10
	}
	if lc.MaxSize == 0 {
		lc.MaxSize = 500
	}

	var zapLevel zap.AtomicLevel
	if lc.Level == "info" {
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else {
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	wc := zap.WrapCore(lc.writeFile)
	zc := zap.Config{
		Level:       zapLevel,
		Development: lc.Development,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         lc.Encoding,
		EncoderConfig:    newEncoderConfig(),
		OutputPaths:      []string{},
		ErrorOutputPaths: []string{},
	}
	logCron = cron.New()
	var err error
	LOG, err = zc.Build(wc)
	if err != nil {
		panic(err)
	}
}

//使用lumberjack处理分割文件问题
//The other problem with rotating every day is that then the amount of space your backups will take up is non-deterministic. You can have a max number of backups of 5, and a max size of 100MB... and your total size might be anywhere from 0 (if you had 5 days of rotated logs with no logging) or 500MB.
//Rotating per day really just does not fit well with a sized-based model. Note that if you want to implement it yourself on top of lumberjack, there is the Rotate() method that will cause an immediate rotation... it's easy enough to write a little goroutine to rotate once per day:
//https://github.com/natefinch/lumberjack/issues/17
//go func() {
//	for {
//		<-time.After(time.Hour * 30)
//		lj.Rotate()
//	}
//}()
func (lc *Logger) writeFile(c zapcore.Core) zapcore.Core {
	lj := &lumberjack.Logger{
		Filename:   lc.Filename,
		MaxSize:    lc.MaxSize, //megabytes,
		MaxBackups: 0,
		MaxAge:     lc.MaxAge,
		LocalTime:  true,
	}
	//每天生成一个文件
	logCron.AddFunc("@daily", func() {
		lj.Rotate()
	})

	//写到文件
	log.SetOutput(io.MultiWriter(lj, os.Stdout))
	w := zapcore.AddSync(lj)
	w1 := zapcore.AddSync(os.Stdout)
	ww := zap.CombineWriteSyncers(w1, w)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		ww,
		zap.InfoLevel,
	)
	return core
}

//日志内容格式配置
func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
