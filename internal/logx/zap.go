package logx

import (
	"base-server/internal/conf"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ log.Logger = (*ZapLogger)(nil)

// ZapLogger is a logger impl.
type ZapLogger struct {
	log  *zap.Logger
	Sync func() error
}

// NewZapLogger returns a zap logger.
func NewZapLogger(bc *conf.Logger, opts ...zap.Option) *ZapLogger {
	level := zapcore.InfoLevel
	filename := ""
	maxSize := uint32(100)
	maxBackups := uint32(10)
	maxAge := uint32(30)
	compress := false

	if levelOverride := strings.TrimSpace(os.Getenv("LOG_LEVEL")); levelOverride != "" {
		level = parseLevel(levelOverride)
	}
	if bc != nil {
		if value := strings.TrimSpace(bc.GetLevel()); value != "" {
			level = parseLevel(value)
		}
		filename = strings.TrimSpace(bc.GetFilename())
		maxSize = defaultUint32(bc.GetMaxSize(), maxSize)
		maxBackups = defaultUint32(bc.GetMaxBackups(), maxBackups)
		maxAge = defaultUint32(bc.GetMaxAge(), maxAge)
		compress = bc.GetCompress()
	}

	encoder := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339Nano),
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	syncers := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
	if filename != "" {
		if dir := filepath.Dir(filename); dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0o755)
		}
		syncers = append(syncers, zapcore.AddSync(&lumberjack.Logger{
			Filename:   filename,
			MaxSize:    int(maxSize),
			MaxBackups: int(maxBackups),
			MaxAge:     int(maxAge),
			Compress:   compress,
		}))
	}

	opts = append(opts,
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCaller(),
		zap.AddCallerSkip(2),
	)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder),
		zapcore.NewMultiWriteSyncer(syncers...),
		zap.NewAtomicLevelAt(level),
	)

	zapLogger := zap.New(core, opts...)
	return &ZapLogger{log: zapLogger, Sync: zapLogger.Sync}
}

// Log Implementation of logger interface.
func (l *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	fields := make([]zap.Field, 0, len(keyvals)/2+1)
	msg := "log"

	limit := len(keyvals)
	if limit%2 != 0 {
		limit--
		fields = append(fields, zap.Any("invalid.key", keyvals[len(keyvals)-1]))
	}

	for i := 0; i < limit; i += 2 {
		key := fmt.Sprint(keyvals[i])
		value := keyvals[i+1]
		if key == "" {
			key = fmt.Sprintf("field.%d", i/2)
		}
		if key == log.DefaultMessageKey {
			msg = fmt.Sprint(value)
			continue
		}
		if err, ok := value.(error); ok {
			fields = append(fields, zap.NamedError(key, err))
			continue
		}
		fields = append(fields, zap.Any(key, value))
	}

	switch level {
	case log.LevelDebug:
		l.log.Debug(msg, fields...)
	case log.LevelInfo:
		l.log.Info(msg, fields...)
	case log.LevelWarn:
		l.log.Warn(msg, fields...)
	case log.LevelError:
		l.log.Error(msg, fields...)
	case log.LevelFatal:
		l.log.Fatal(msg, fields...)
	default:
		l.log.Info(msg, fields...)
	}

	return nil
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func defaultUint32(value, fallback uint32) uint32 {
	if value == 0 {
		return fallback
	}
	return value
}
