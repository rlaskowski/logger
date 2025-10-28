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
		buf     bytes.Buffer
		scratch []byte
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
	enc.scratch = enc.scratch[:0]
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
		e.writeQuotedString(attr.Name)
		e.buf.WriteByte(':')
		switch v := attr.Value.(type) {
		case int:
			e.writeInt(int64(v))
		case float64:
			e.writeFloat(v)
		case time.Time:
			e.writeTime(v)
		case bool:
			e.writeBool(v)
		case string:
			e.writeQuotedString(v)
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

func (e *jsonEncoder) writeQuotedString(s string) {
	e.scratch = e.scratch[:0]
	e.scratch = strconv.AppendQuote(e.scratch, s)
	e.buf.Write(e.scratch)
}

func (e *jsonEncoder) writeInt(i int64) {
	e.scratch = e.scratch[:0]
	e.scratch = strconv.AppendInt(e.scratch, i, 10)
	e.buf.Write(e.scratch)
}

func (e *jsonEncoder) writeFloat(f float64) {
	e.scratch = e.scratch[:0]
	e.scratch = strconv.AppendFloat(e.scratch, f, 'f', -1, 64)
	e.buf.Write(e.scratch)
}

func (e *jsonEncoder) writeBool(bv bool) {
	e.scratch = e.scratch[:0]
	e.scratch = strconv.AppendBool(e.scratch, bv)
	e.buf.Write(e.scratch)
}

func (e *jsonEncoder) writeTime(t time.Time) {
	e.scratch = e.scratch[:0]
	e.scratch = append(e.scratch, '"')
	e.scratch = t.AppendFormat(e.scratch, time.RFC3339Nano)
	e.scratch = append(e.scratch, '"')
	e.buf.Write(e.scratch)
}
