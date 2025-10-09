package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	var (
		buf    = bytes.Buffer{}
		logger = NewLogger(&buf)
	)
	tests := []struct {
		name,
		want string
		level Level
		data  any
	}{
		{
			name:  "info level",
			level: Info,
			data: struct {
				Message string
				Value   int
			}{Message: "simple info message", Value: 11341},
			want: `[INFO] {"Message":"simple info message","Value":11341}`,
		},
		{
			name:  "warn level with short message",
			level: Warn,
			data:  "short message",
			want:  `[WARN] "short message"`,
		},
		{
			name:  "debug level",
			level: Debug,
			data: struct {
				Name   string  `json:"firstName"`
				Age    int     `json:"age"`
				Amount float64 `json:"amount"`
			}{
				Name:   "John Carby",
				Age:    45,
				Amount: 120.32,
			},
			want: `[DEBUG] {"firstName":"John Carby","age":45,"amount":120.32}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger.build(test.level, test.data)
			assert.Equal(t, test.want, buf.String())
			buf.Reset()
		})
	}
}

func Test_withLevel(t *testing.T) {
	var (
		buf    = bytes.Buffer{}
		logger = NewLogger(&buf, WithLevel(Level("DEBUG")))
	)
	msg := struct {
		Message  string `json:"message"`
		MetaData string `json:"metaData"`
	}{Message: "simple test message", MetaData: "debug"}

	logger.Info(msg)
	assert.Empty(t, buf.String())
}

func Benchmark_build(b *testing.B) {
	var (
		buf     = bytes.Buffer{}
		logger  = NewLogger(&buf)
		message = struct{ Message string }{Message: "simple message"}
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(message)
		logger.Debug(message)
		logger.Warn(message)
	}
}
