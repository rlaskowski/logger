package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

const (
	Info  Level = "INFO"
	Debug Level = "DEBUG"
	Warn  Level = "WARN"
	Err   Level = "ERROR"
)

type (
	Logger struct {
		mu     sync.Mutex
		option *Option
		writer io.Writer
	}

	Option struct {
		level Level
	}

	OptionFn func(optio *Option)

	Level string
)

func NewLogger(w io.Writer, options ...OptionFn) *Logger {
	option := &Option{}
	for _, opt := range options {
		opt(option)
	}
	return &Logger{
		mu:     sync.Mutex{},
		writer: w,
		option: option,
	}
}

func WithLevel(level Level) OptionFn {
	return func(option *Option) {
		option.level = level
	}
}

func (l *Logger) Info(msg ...any) {
	l.build(Info, msg)
}

func (l *Logger) Debug(msg ...any) {
	l.build(Debug, msg)
}

func (l *Logger) Warn(msg ...any) {
	l.build(Warn, msg)
}

func (l *Logger) Err(msg ...any) {
	l.build(Err, msg)
}

func (l *Logger) format(level Level, msg string) ([]byte, error) {
	buf := &bytes.Buffer{}
	if _, err := fmt.Fprintf(buf, "[%s] %s", string(level), msg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (l *Logger) build(level Level, msg ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.option.level != "" && l.option.level != level {
		return
	}
	builder := strings.Builder{}
	for _, d := range msg {
		b, err := json.Marshal(d)
		if err != nil {
			return
		}
		builder.WriteString(string(b))
	}
	f, err := l.format(level, builder.String())
	if err != nil {
		return
	}
	//nolint:errcheck
	fmt.Fprint(l.writer, string(f))
}
