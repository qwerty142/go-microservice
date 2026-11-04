package logger

import (
	"context"
	"fmt"
	"sync"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerCtx string

const (
	loggerCtxValue loggerCtx = "logger"
	TraceIdKey     string    = "trace_id"
	SpanIdKey      string    = "span_id"
)

type Logger struct {
	logger *otelzap.SugaredLogger
}

var global *Logger

func init() {
	_, err := New()
	if err != nil {
		panic(err)
	}
}

func New() (*Logger, error) {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level.SetLevel(zapcore.DebugLevel)
	loggerConfig.ErrorOutputPaths = []string{"stdout"}

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("loggerConfig.Build: %w", err)
	}

	once := sync.Once{}
	once.Do(func() {
		global = &Logger{logger: otelzap.New(logger).Sugar().With("service", "loms")}
	})

	return &Logger{logger: otelzap.New(logger).Sugar()}, nil
}

func ToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerCtxValue, logger)
}

func ContextWithSpanValues(ctx context.Context, span trace.Span) context.Context {
	ctx = context.WithValue(ctx, TraceIdKey, span.SpanContext().TraceID())
	ctx = context.WithValue(ctx, SpanIdKey, span.SpanContext().SpanID())

	return ctx
}

func Infow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	keysAndValues = parseContext(ctx, keysAndValues)

	if loggerC, ok := ctx.Value(loggerCtxValue).(*Logger); ok {
		loggerC.logger.Infow(msg, keysAndValues...)

		return
	}

	global.logger.Infow(msg, keysAndValues...)
}

func Errorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	keysAndValues = parseContext(ctx, keysAndValues)

	if loggerC, ok := ctx.Value(loggerCtxValue).(*Logger); ok {
		loggerC.logger.Errorw(msg, keysAndValues...)

		return
	}

	global.logger.Errorw(msg, keysAndValues...)
}

func Fatalw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	keysAndValues = parseContext(ctx, keysAndValues)

	if loggerC, ok := ctx.Value(loggerCtxValue).(*Logger); ok {
		loggerC.logger.Fatalw(msg, keysAndValues...)

		return
	}

	global.logger.Fatalw(msg, keysAndValues...)
}

func Panicw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	keysAndValues = parseContext(ctx, keysAndValues)

	if loggerC, ok := ctx.Value(loggerCtxValue).(*Logger); ok {
		loggerC.logger.Panicw(msg, keysAndValues...)

		return
	}

	global.logger.Panicw(msg, keysAndValues...)
}

func With(args ...interface{}) *Logger {
	return &Logger{global.logger.With(args...)}
}

func parseContext(ctx context.Context, keysAndValues ...interface{}) []interface{} {
	if ctx.Value(TraceIdKey) != nil {
		keysAndValues = append(keysAndValues, TraceIdKey, ctx.Value(TraceIdKey))
	}
	if ctx.Value(SpanIdKey) != nil {
		keysAndValues = append(keysAndValues, SpanIdKey, ctx.Value(SpanIdKey))
	}

	return keysAndValues
}
