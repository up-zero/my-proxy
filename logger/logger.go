package logger

import (
	"fmt"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"runtime"
	"strings"
	"time"
)

var LOGGER *zap.Logger

func NewLogger() {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   util.GetLogPath(), // 文件基础名字
		MaxSize:    50,                // 单个日志文件的大小 (MB)
		MaxAge:     30,                // 日志的保存时间 (天)
		MaxBackups: 50,                // 最多保存多少个备份
		LocalTime:  true,              // 是否使用本地时间
		Compress:   true,              // 是否压缩
	}
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "level",
		TimeKey:       "time",
		NameKey:       "logger",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Local().Format(util.DateTimeWithMilliLayout))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberjackLogger), zapcore.AddSync(os.Stdout)),
		zap.DebugLevel)
	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	dev := zap.Development()

	LOGGER = zap.New(core, caller, dev)
}

func Info(msg string, fields ...zap.Field) {
	LOGGER.Info(msg, appendFileLine(fields)...)
}

func Warn(msg string, fields ...zap.Field) {
	LOGGER.Warn(msg, appendFileLine(fields)...)
}

func Error(msg string, fields ...zap.Field) {
	LOGGER.Error(msg, appendFileLine(fields)...)
}

func appendFileLine(fields []zap.Field) []zap.Field {
	_, file, line, _ := runtime.Caller(2)
	arr := strings.Split(file, util.AppName)
	if len(arr) > 1 {
		file = strings.Join(arr[1:], "")
	}
	fields = append(fields, zap.String("file", fmt.Sprintf("%s:%d", file, line)))
	return fields
}
