package log

import (
	"base-server/internal/conf"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ log.Logger = (*ZapLogger)(nil)

// ZapLogger is a logger impl.
type ZapLogger struct {
	log  *zap.Logger
	Sync func() error
}

// NewZapLogger return a zap logger.
func NewZapLogger(bc *conf.Logger, opts ...zap.Option) *ZapLogger {
	level := zapcore.DebugLevel
	switch bc.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// 配置 lumberjack 实现日志切割
	lumberJackLogger := &lumberjack.Logger{
		Filename:   bc.Filename,        // 日志路径
		MaxSize:    int(bc.MaxSize),    // 单文件最大大小（MB）
		MaxBackups: int(bc.MaxBackups), // 保留旧文件数量
		MaxAge:     int(bc.MaxAge),     // 保留天数
		Compress:   bc.Compress,        // 压缩旧文件
	}

	encoder := zapcore.EncoderConfig{
		TimeKey:   "ts",
		LevelKey:  "level",
		NameKey:   "logger",
		CallerKey: "caller",
		//MessageKey:     "msg",
		StacktraceKey: "stack",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02T15:04:05.999Z07:00"))
		},
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	opts = append(opts,
		zap.AddStacktrace(
			zap.NewAtomicLevelAt(zapcore.ErrorLevel)),
		zap.AddCaller(),
		zap.AddCallerSkip(2),
		zap.Development(),
	)

	core := zapcore.NewCore(
		//zapcore.NewConsoleEncoder(encoder),
		zapcore.NewJSONEncoder(encoder),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(lumberJackLogger), // 文件输出
			zapcore.AddSync(os.Stdout),        // 控制台输出
		), zap.NewAtomicLevelAt(level))
	zapLogger := zap.New(core, opts...)
	return &ZapLogger{log: zapLogger, Sync: zapLogger.Sync}
}

// Log Implementation of logger interface.
func (l *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 || len(keyvals)%2 != 0 {
		l.log.Warn(fmt.Sprint("Keyvalues must appear in pairs: ", keyvals))
		return nil
	}
	// Zap.Field is used when keyvals pairs appear
	var data []zap.Field
	for i := 0; i < len(keyvals); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), fmt.Sprint(keyvals[i+1])))
	}
	switch level {
	case log.LevelDebug:
		l.log.Debug("aaa", data...)
	case log.LevelInfo:
		l.log.Info("aaa", data...)
	case log.LevelWarn:
		l.log.Warn("aaa", data...)
	case log.LevelError:
		l.log.Error("aaa", data...)
	}
	return nil
}
