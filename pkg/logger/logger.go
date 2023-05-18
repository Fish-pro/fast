package logger

import (
	"os"
	"sync"
	"time"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	Log              *zap.Logger
	customTimeFormat string
	onceInit         sync.Once
)

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(customTimeFormat))
}

func InitLogger(lvl int, timeFormat string) {
	onceInit.Do(func() {
		globalLevel := zapcore.Level(lvl)

		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})

		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= globalLevel && lvl < zapcore.ErrorLevel
		})

		consoleInfos := zapcore.Lock(os.Stdout)
		consoleErrors := zapcore.Lock(os.Stderr)

		var useCustomTimeFormat bool
		ecfg := zap.NewProductionEncoderConfig()
		if len(timeFormat) > 0 {
			customTimeFormat = timeFormat
			ecfg.EncodeTime = customTimeEncoder
			useCustomTimeFormat = true
		}
		consoleEncoder := zapcore.NewJSONEncoder(ecfg)
		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
			zapcore.NewCore(consoleEncoder, consoleInfos, lowPriority),
		)
		Log = zap.New(core)
		zap.RedirectStdLog(Log)

		if !useCustomTimeFormat {
			Log.Warn("timeFormat is error---use zap default")
		}
	})
}

func codeToLevel(code codes.Code) zapcore.Level {
	if code == codes.OK {
		return zap.DebugLevel
	}
	return grpczap.DefaultClientCodeToLevel(code)
}

func AddLogging(opts []grpc.ServerOption) {
	logger := Log
	o := []grpczap.Option{
		grpczap.WithLevels(codeToLevel),
	}
	grpczap.ReplaceGrpcLoggerV2(logger)

	opts = append(opts, grpcmiddleware.WithUnaryServerChain(
		grpcctxtags.UnaryServerInterceptor(grpcctxtags.WithFieldExtractor(grpcctxtags.CodeGenRequestFieldExtractor)),
		grpczap.UnaryServerInterceptor(logger, o...),
	))

	opts = append(opts, grpcmiddleware.WithStreamServerChain(
		grpcctxtags.StreamServerInterceptor(grpcctxtags.WithFieldExtractor(grpcctxtags.CodeGenRequestFieldExtractor)),
		grpczap.StreamServerInterceptor(logger, o...),
	))
}
