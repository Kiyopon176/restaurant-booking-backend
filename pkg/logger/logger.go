package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)

	Warn(msg string, fields ...zap.Field)

	Fatal(msg string, fields ...zap.Field)

	Debug(msg string, fields ...zap.Field)
}

func New(logLevel string, fields ...zapcore.Field) (Logger, error) {
	atom := zap.NewAtomicLevel()
	err := atom.UnmarshalText([]byte(logLevel))
	if err != nil {
		return nil, err
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	l := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	)).With(fields...)

	defer l.Sync()

	return l, nil
}
