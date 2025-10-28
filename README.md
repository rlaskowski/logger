# logger

Small structured logger that writes JSON lines to any `io.Writer`. Each entry
comes with a level prefix and a JSON payload built via a lightweight
`JsonBuilder`.

- Supports levels (`Debug`, `Info`, `Warn`, `Error`, `Fatal`) and optional level filtering.
- Produces compact JSON without reflection-based overhead.
- Allows nested attributes by nesting `JsonBuilder` values.
- Handles common primitive types (`string`, `int`, `float64`, `bool`, `time.Time`).

## Usage

```go
package main

import (
	"os"
	"time"

	"github.com/rafal/logger"
)

func main() {
	log := logger.NewLogger(os.Stdout, logger.WithLevel(logger.Debug))

	log.Info(logger.JsonBuilder{
		{Name: "message", Value: "service started"},
		{Name: "timestamp", Value: time.Now()},
		{Name: "config", Value: logger.JsonBuilder{
			{Name: "port", Value: 8080},
			{Name: "region", Value: "eu-west-1"},
		}},
	})
}
```

Running the example prints:

```
[INFO] {"message":"service started","timestamp":"2023-03-14T15:09:26.123456Z","config":{"port":8080,"region":"eu-west-1"}}
```

## Development

Run the tests and benchmarks with:

```
go test ./... -bench=. -benchmem
```

Example benchmark output:

```
Benchmark_build-8                 1319697        873.1 ns/op      0 B/op   0 allocs/op
Benchmark_jsonEncoderMarshal-8    2581470        466.3 ns/op      0 B/op   0 allocs/op
Benchmark_buildFiltered-8       578627546          2.1 ns/op      0 B/op   0 allocs/op
```
