package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"

	"jww/internal/domain"
)

type Logger struct {
	labels []slog.Attr
}

var (
	level         = new(slog.LevelVar)
	base          = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	configureOnce sync.Once
	configureErr  error
)

func init() {
	level.Set(slog.LevelInfo)
}

func Configure(config Config) error {
	configureOnce.Do(func() {
		configureErr = configure(config)
	})
	return configureErr
}

func configure(config Config) error {
	parsed, err := parseLevel(config.Level)
	if err != nil {
		return err
	}

	level.Set(parsed)

	if config.Pretty {
		// tint renders colorized, human-readable lines for local development.
		// Colors are disabled when stdout isn't a terminal so redirected or
		// piped logs stay plain text.
		base = slog.New(tint.NewTextHandler(os.Stdout, &tint.Options{
			Level:   level,
			NoColor: !isatty.IsTerminal(os.Stdout.Fd()),
		}))
		return nil
	}
	base = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	return nil
}

func New() Logger {
	return Logger{}
}

func With(labels ...slog.Attr) Logger {
	return New().With(labels...)
}

func Slog() *slog.Logger {
	return currentBase()
}

func (l Logger) With(labels ...slog.Attr) Logger {
	next := Logger{labels: make([]slog.Attr, 0, len(l.labels)+len(labels))}
	next.labels = append(next.labels, l.labels...)
	next.labels = append(next.labels, labels...)
	return next
}

func (l Logger) WithComponent(component string) Logger {
	return l.With(slog.String("component", component))
}

func (l Logger) Slog() *slog.Logger {
	return currentBase().With(attrsToArgs(l.labels)...)
}

func (l Logger) WithContext(ctx domain.AppContext) *slog.Logger {
	attrs := make([]slog.Attr, 0, len(l.labels)+4)
	attrs = append(attrs, l.labels...)
	attrs = append(attrs, appContextAttrs(ctx)...)
	return currentBase().With(attrsToArgs(attrs)...)
}

func currentBase() *slog.Logger {
	return base
}

func attrsToArgs(attrs []slog.Attr) []any {
	args := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		args = append(args, attr)
	}
	return args
}

func appContextAttrs(ctx domain.AppContext) []slog.Attr {
	attrs := []slog.Attr{}
	if flowID := ctx.FlowID(); flowID != "" {
		attrs = append(attrs, slog.String("flow_id", flowID))
	}

	source := ctx.RequestSource()
	if source.Type != "" {
		attrs = append(attrs, slog.String("source_type", string(source.Type)))
	}
	if source.HTTP != nil {
		attrs = append(attrs, slog.String("http_endpoint", source.HTTP.Endpoint))
	}
	if source.Scheduler != nil {
		attrs = append(attrs, slog.String("job", source.Scheduler.Job))
	}

	if user := ctx.User(); user != nil {
		attrs = append(attrs, slog.String("user_id", user.ID.String()))
	}
	return attrs
}

func parseLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("logger: unknown log level %q", raw)
	}
}
