package database

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	applog "github.com/stevensuki/ledgerline-backend/pkg/logger"
)

// slogLogger pipes GORM logs into slog, using the context logger so queries carry request_id.
type slogLogger struct {
	log           *slog.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func newSlogLogger(log *slog.Logger, slowThreshold time.Duration, debug bool) gormlogger.Interface {
	level := gormlogger.Warn
	if debug {
		level = gormlogger.Info
	}
	return &slogLogger{log: log, level: level, slowThreshold: slowThreshold}
}

func (l *slogLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

func (l *slogLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Info {
		l.from(ctx).Info(msg, slog.Any("args", args))
	}
}

func (l *slogLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Warn {
		l.from(ctx).Warn(msg, slog.Any("args", args))
	}
}

func (l *slogLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Error {
		l.from(ctx).Error(msg, slog.Any("args", args))
	}
}

// Trace is called by GORM after every query completes.
func (l *slogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	attrs := []any{
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed),
	}

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		l.from(ctx).Error("query failed", append(attrs, slog.Any("error", err))...)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold:
		l.from(ctx).Warn("slow query", attrs...)
	case l.level >= gormlogger.Info:
		l.from(ctx).Debug("query", attrs...)
	}
}

func (l *slogLogger) from(ctx context.Context) *slog.Logger {
	if ctxLog := applog.FromContext(ctx); ctxLog != nil {
		return ctxLog
	}
	return l.log
}
