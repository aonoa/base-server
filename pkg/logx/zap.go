package logx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ log.Logger = (*ZapLogger)(nil)

// ZapLogger adapts zap to the Kratos logger interface.
type ZapLogger struct {
	log *zap.Logger
}

// Config is the logger configuration contract shared by generated proto configs.
type Config interface {
	GetLevel() string
	GetFilename() string
	GetMaxSize() int32
	GetMaxBackups() int32
	GetMaxAge() int32
	GetCompress() bool
}

// New builds a shared structured logger for all services.
func New(serviceID, serviceName, serviceVersion string, cfg Config) (log.Logger, func()) {
	encoderConfig := zapcore.EncoderConfig{
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
	if cfg != nil && strings.TrimSpace(cfg.GetFilename()) != "" {
		filename := strings.TrimSpace(cfg.GetFilename())
		if dir := filepath.Dir(filename); dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0o755)
		}
		syncers = append(syncers, zapcore.AddSync(&lumberjack.Logger{
			Filename:   filename,
			MaxSize:    int(defaultInt32(cfg.GetMaxSize(), 100)),
			MaxBackups: int(defaultInt32(cfg.GetMaxBackups(), 10)),
			MaxAge:     int(defaultInt32(cfg.GetMaxAge(), 30)),
			Compress:   cfg.GetCompress(),
		}))
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(syncers...),
		zap.NewAtomicLevelAt(parseLevel(envOrConfigLevel(cfg))),
	)

	base := &ZapLogger{
		log: zap.New(
			core,
			zap.AddCaller(),
			zap.AddCallerSkip(2),
			zap.AddStacktrace(zapcore.ErrorLevel),
		),
	}

	logger := log.With(
		base,
		"service.id", serviceID,
		"service.name", serviceName,
		"service.version", serviceVersion,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	return logger, func() {
		_ = base.log.Sync()
	}
}

// Log implements the Kratos logger interface.
func (l *ZapLogger) Log(level log.Level, keyvals ...any) error {
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

func envOrConfigLevel(cfg Config) string {
	if level := strings.TrimSpace(os.Getenv("LOG_LEVEL")); level != "" {
		return level
	}
	if cfg == nil {
		return ""
	}
	return cfg.GetLevel()
}

func defaultInt32(value, fallback int32) int32 {
	if value <= 0 {
		return fallback
	}
	return value
}
