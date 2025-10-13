package logger

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_build(t *testing.T) {
	var (
		writer = &bytes.Buffer{}
		logger = NewLogger(writer)
	)
	tests := []struct {
		name,
		want string
		level Level
		data  JsonBuilder
	}{
		{
			name:  "info level",
			level: Info,
			data: JsonBuilder{
				{
					"message",
					"simple info message",
				},
				{
					"age",
					64,
				},
			},
			want: fmt.Sprintln(`[INFO] {"message":"simple info message","age":64}`),
		},
		{
			name:  "single raw attribute and raw declaration in json attribute",
			level: Info,
			data: JsonBuilder{
				{
					"node1",
					JsonBuilder{
						{
							"node2",
							JsonBuilder{
								{
									"node3",
									JsonBuilder{
										{
											"node4",
											1234124124,
										},
									},
								},
							},
						},
					},
				},
			},
			want: fmt.Sprintln(`[INFO] {"node1":{"node2":{"node3":{"node4":1234124124}}}}`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer.Reset()

			logger.build(test.level, test.data)
			assert.Equal(t, test.want, writer.String())
		})
	}
}

func Test_jsonEncoder_marshal(t *testing.T) {
	ts := time.Date(2023, time.March, 14, 15, 9, 26, 123456000, time.UTC)

	tests := []struct {
		name    string
		input   JsonBuilder
		wantBuf string
		wantErr error
	}{
		{
			name:    "empty builder",
			input:   nil,
			wantBuf: `{}`,
		},
		{
			name: "primitives and nested",
			input: JsonBuilder{
				{Name: "message", Value: "hello world"},
				{Name: "count", Value: 42},
				{Name: "largeCount", Value: 999288311231235212},
				{Name: "flag", Value: true},
				{Name: "amount", Value: 34.23321225352352},
				{Name: "smallAmount", Value: 12.2},
				{Name: "timestamp", Value: ts},
				{Name: "nested", Value: JsonBuilder{
					{Name: "inner", Value: "value"},
				}},
			},
			wantBuf: `{"message":"hello world","count":42,"largeCount":999288311231235212,"flag":true,"amount":34.23321225352352,"smallAmount":12.2,"timestamp":"2023-03-14T15:09:26.123456Z","nested":{"inner":"value"}}`,
		},
		{
			name: "empty name at top level",
			input: JsonBuilder{
				{Name: "", Value: "bad"},
			},
			wantErr: errJsonName,
		},
		{
			name: "empty name inside nested builder",
			input: JsonBuilder{
				{Name: "outer", Value: JsonBuilder{
					{Name: "", Value: "bad"},
				}},
			},
			wantErr: errJsonName,
		},
		{
			name: "unsupported attribute type",
			input: JsonBuilder{
				{Name: "invalid", Value: struct{}{}},
			},
			wantErr: errJsonUnsupportedType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			enc := &jsonEncoder{}

			err := enc.marshal(tc.input)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			if assert.NoError(t, err) {
				assert.Equal(t, tc.wantBuf, enc.buf.String())
			}
		})
	}
}

func Test_build_level(t *testing.T) {
	var (
		writer  = &bytes.Buffer{}
		message = JsonBuilder{
			{Name: "message", Value: "test"},
		}
	)
	tests := []struct {
		name,
		want string
		levelFn func()
	}{
		{
			name: "info level with debug",
			levelFn: func() {
				NewLogger(writer, WithLevel(Info)).Debug(message)
			},
		},
		{
			name: "info level with info",
			want: fmt.Sprintln(`[INFO] {"message":"test"}`),
			levelFn: func() {
				NewLogger(writer, WithLevel(Info)).Info(message)
			},
		},
		{
			name: "debug level with warn",
			want: fmt.Sprintln(`[WARN] {"message":"test"}`),
			levelFn: func() {
				NewLogger(writer, WithLevel(Debug)).Warn(message)
			},
		},
		{
			name: "without level with fatal",
			want: fmt.Sprintln(`[FATAL] {"message":"test"}`),
			levelFn: func() {
				NewLogger(writer).Fatal(message)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer.Reset()
			test.levelFn()
			assert.Equal(t, test.want, writer.String())
		})
	}
}

func Benchmark_build(b *testing.B) {
	var (
		buf    = bytes.Buffer{}
		logger = NewLogger(&buf)
		data   = JsonBuilder{
			{
				"object1",
				JsonBuilder{
					{
						"message",
						"simple message",
					},
				},
			},
			{
				"required",
				false,
			},
			{
				"object2",
				JsonBuilder{
					{
						"simpleNumber",
						123,
					},
				},
			},
		}
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(data)

		buf.Reset()

		logger.Debug(data)
		buf.Reset()
	}
}
