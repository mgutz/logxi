# logxi

A [12 factor app](http://12factor.net/logs) logger built for performance
and happy development.

### Installation

    go get -a github.com/mgutz/logxi/v1

## Higlights

This logger package

1.  Is fast in production environment.

    logxi encodes JSON 2X faster than logrus and log15 with primitive types.
    When diagnosing a problem in production, troubleshooting usually means
    enabling small trace data in `Debug` and `Info` statements for an
    extended period of time.

    ```
# primitive types
BenchmarkLogxi          100000     14962 ns/op    2784 B/op     44 allocs/op
BenchmarkLogrus          50000     36554 ns/op    7656 B/op    174 allocs/op
BenchmarkLog15           30000     47822 ns/op    8024 B/op    192 allocs/op

# nested object
BenchmarkLogxiComplex    30000     43831 ns/op    8739 B/op    184 allocs/op
BenchmarkLogrusComplex   30000     51598 ns/op   10832 B/op    256 allocs/op
BenchmarkLog15Complex    20000     74030 ns/op   12072 B/op    278 allocs/op
```

2.  Logs machine parsable output in production environments.
    The recommended formatter for production is `JSONFormatter`.

    `TextFormatter` may also be used if you don't care about
    JSON and want the fastest logs with key, value pairs.

3.  Is developer friendly in development environments. The default
    formatter in terminals is colorful, prints file and line numbers
    when warnings and errors occur.

    The default formatter in TTY mode is `HappyDevFormatter`. It  is
    not concerned with performance and should never be used
    in prouction environments.

4.  Has level guards to avoid the cost of parameters. These should
    always be used in production mode if tracing may be enabled.

    ```go
if log.IsDebug() {
    log.Debug("some ")
}

if log.IsInfo() {
    log.Info("some ")
}

if log.IsWarn() {
    log.Warn("some ")
}

// Error and Fatal do not have guards, they MUST always log.
```

5.  Conforms to a logging interface so it can be replaced.

    ```go
type Logger interface {
    Debug(string, ...interface{})
    Info(string, ...interface{})
    Warn(string, ...interface{})
    Error(string, ...interface{})
    Fatal(string, ...interface{})

    SetLevel(int)
    SetFormatter(Formatter)

    IsDebug() bool
    IsInfo() bool
    IsWarn() bool
    // Error, Fatal not needed, those SHOULD always be logged
}
```

6.  Standardizes on key, value pairs for machine parsing instead
    of `map` and `fmt.Printf`.

    ```go
log.Debug("inside Fn()", "key1", value1, "key2", value2)
```

    logxi logs `IMBALANCED_PAIRS=>` if key/value pairs are imbalanced

7.  Loggers can be enabled/disabled via environment variables.

    `LOGXI` acccepts a list space separated names with colons to indicate the
    log level. See `LevelAtoi` in `logger.go` for values.

    The following defaults all loggers to `LevelError`. The log named
    `"models"` is set to `LevelDebug`. Any log having suffix `"controller"` is
    set to `LevelError`. The log named `"trace"` is disabled and will use
    `NullLog`.

    ```sh
LOGXI="*:ERR models:DBG *controller:ERR trace:OFF" yourapp
```

    The format may also be selected via `LOGXI_FORMAT` environment
    variable. Valid values are `"text"` and `"JSON"`.

    ```sh
LOGXI_FORMAT=JSON yourapp
```

### Extending

What about hooks? Implement your own `io.Writer` to write to external
services and use `JSONFormatter` for your writer to parse the
stream.

What about other writers? 12 factor apps only concern themselves with
STDOUT. Use shell redirection operators to write to file or use an
`io.Writer`.

What about formatting? Key/value pairs only.

### License

MIT License
