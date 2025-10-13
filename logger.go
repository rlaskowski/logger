package logger

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"sync"
	"time"
)

const (
	Debug = iota
	Info
	Warn
	Error
	Fatal
)

var (
	errJsonName            = errors.New("name cannot be empty")
	errJsonUnsupportedType = errors.New("unsupported json type")
)

type (
	Logger struct {
		option          *Option
		jsonEncoderPool sync.Pool
		writer          io.Writer
	}

	Option struct {
		level Level
	}

	OptionFn func(option *Option)

	JsonBuilder []JsonAttr

	JsonAttr struct {
		Name  string
		Value any
	}

	Level int

	jsonEncoder struct {
		buf bytes.Buffer
	}
)

// Returns the printable representation of a log Level.
func getLevelName(level Level) string {
	var levelName string
	switch level {
	case Debug:
		levelName = "DEBUG"
	case Info:
		levelName = "INFO"
	case Warn:
		levelName = "WARN"
	case Error:
		levelName = "ERROR"
	case Fatal:
		levelName = "FATAL"
	default:
		levelName = "UNKNOWN"
	}
	return levelName
}

// Constructs a Logger configured with the supplied options and writer.
func NewLogger(w io.Writer, options ...OptionFn) *Logger {
	option := &Option{}
	for _, opt := range options {
		opt(option)
	}
	return &Logger{
		jsonEncoderPool: sync.Pool{
			New: func() any {
				return &jsonEncoder{}
			},
		},
		writer: w,
		option: option,
	}
}

func WithLevel(level Level) OptionFn {
	return func(option *Option) {
		option.level = level
	}
}

func (l *Logger) Info(jb JsonBuilder) {
	l.build(Info, jb)
}

func (l *Logger) Debug(jb JsonBuilder) {
	l.build(Debug, jb)
}

func (l *Logger) Warn(jb JsonBuilder) {
	l.build(Warn, jb)
}

func (l *Logger) Err(jb JsonBuilder) {
	l.build(Error, jb)
}

func (l *Logger) Fatal(jb JsonBuilder) {
	l.build(Fatal, jb)
}

// Build renders the JsonBuilder into the writer when the level passes the filter.
// It prefixes the payload with the level name, skips work when marshalling fails,
// and writes a trailing newline so consecutive entries appear on separate lines.
func (l *Logger) build(level Level, jb JsonBuilder) {
	if l.option.level > level {
		return
	}
	enc := l.newJsonEncoder()
	defer l.disposeJsonEncoder(enc)

	enc.buf.WriteString("[" + getLevelName(level) + "] ")

	if err := enc.marshal(jb); err != nil {
		return
	}
	enc.buf.WriteByte('\n')

	//nolint:errcheck
	l.writer.Write(enc.buf.Bytes())
}

// Retrieves a pooled encoder to avoid per-call allocations.
func (l *Logger) newJsonEncoder() *jsonEncoder {
	return l.jsonEncoderPool.Get().(*jsonEncoder)
}

// Resets the encoder state and returns it to the pool.
func (l *Logger) disposeJsonEncoder(enc *jsonEncoder) {
	enc.buf.Reset()
	l.jsonEncoderPool.Put(enc)
}

// Serializes a JsonBuilder into JSON respecting nested builders.
func (e *jsonEncoder) marshal(jb JsonBuilder) error {
	e.buf.WriteByte('{')
	for i, attr := range jb {
		if i > 0 {
			e.buf.WriteByte(',')
		}
		if attr.Name == "" {
			return errJsonName
		}
		e.buf.Write(strconv.AppendQuote(e.buf.AvailableBuffer(), attr.Name))
		e.buf.WriteByte(':')
		switch v := attr.Value.(type) {
		case int:
			e.buf.Write(strconv.AppendInt(e.buf.AvailableBuffer(), int64(v), 10))
		case float64:
			e.buf.Write(strconv.AppendFloat(e.buf.AvailableBuffer(), v, 'f', -1, 64))
		case time.Time:
			e.buf.Write(strconv.AppendQuote(e.buf.AvailableBuffer(), v.Format(time.RFC3339Nano)))
		case bool:
			e.buf.Write(strconv.AppendBool(e.buf.AvailableBuffer(), v))
		case string:
			e.buf.Write(strconv.AppendQuote(e.buf.AvailableBuffer(), v))
		case JsonBuilder:
			if err := e.marshal(v); err != nil {
				return err
			}
		default:
			return errJsonUnsupportedType
		}
	}
	e.buf.WriteByte('}')
	return nil
}
