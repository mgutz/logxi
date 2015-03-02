# logxi

A [12 factor app](http://12factor.net/logs) logger built for performance
and happy development.

### Installation

    go get -u github.com/mgutz/logxi/v1

### Getting Started

```go
import "github.com/mgutz/logxi/v1"

// create package variable for logger interface
var logger log.Logger

func main() {
    // use default logger
    if log.IsInfo() {
        log.Info("Hello", "name", "mario")
    }

    // create a logger for your package, assigning a unique
    // name which can be enabled from environment variables
    logger = log.New(os.Stdout, "yourpkg")

    db, err := sql.Open("postgres", "dbname=testdb")
    if err != nil {
        // use key-value pairs after message
        logger.Error("Could not open database", "err", err)
    }
    
    if log.IsDebug() {
        logger.Debug("OK")
    }
}
```

Run your application with Debug enabled while developing 
(otherwise only errors are logged)

    LOGXI=* go run main.go

## Higlights

This logger package

*   Is fast in production environment

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

*   Logs machine parsable output in production environments.
    The recommended formatter for production is `JSONFormatter`.

    `TextFormatter` may also be used if you don't care about
    JSON and want the fastest logs with key, value pairs.

*   Is developer friendly in development environments. The default
    formatter in terminals is colorful, prints file and line numbers
    when warnings and errors occur.

    The default formatter in TTY mode is `HappyDevFormatter`. It  is
    not concerned with performance and should never be used
    in production environments.

*   Has level guards to avoid the cost of arguments. These _SHOULD_
    always be used.

        if log.IsDebug() {
            log.Debug("some ")
        }

        if log.IsInfo() {
            log.Info("some ")
        }

        if log.IsWarn() {
            log.Warn("some ")
        }

    Error and Fatal do not have guards, they MUST always log.

*   Conforms to a logging interface so it can be replaced.

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
            // Error, Fatal have no guards, they SHOULD always log
        }

*   Standardizes on key, value pairs for machine parsing instead
    of `map` and `fmt.Printf`.

    ```go
log.Debug("inside Fn()", "key1", value1, "key2", value2)
```

    logxi logs `IMBALANCED_PAIRS=>` if key/value pairs are imbalanced
    
## Configuration

### Enabling/Disabling Loggers

By default logxi logs entries whose level is `LevelWarn` or above when 
using a terminal. For non-terminals, entries with level `LevelError` and 
above are logged.

To quickly see all entries use short form

    # enable all, disable log named foo
    LOGXI=*,-foo yourapp

To better control logs in production, use long form which allows
for granular control of levels

    # the above statement is equivalent to this
    LOGXI=*=DBG,foo=OFF yourapp
    
`DBG` should obviously not be used in production unless for 
troubleshooting. See `LevelAtoi` in `logger.go` for values.
For example, there is a problem in the data access layer
in production.

    # Set all to Error and set data related packages to Debug
    LOGXI=*=ERR,models=DBG,dat*=DBG,api=DBG yourapp

### Format

The format may be set via `LOGXI_FORMAT` environment
variable. Valid values are `"text"` and `"JSON"`.

    # Use JSON in production
    LOGXI_FORMAT=JSON yourapp

## Extending

What about hooks? Implement your own `io.Writer` to write to external
services and use `JSONFormatter`.

What about other writers? 12 factor apps only concern themselves with
STDOUT. Use shell redirection operators to write to file or create
a custom `io.Writer`.

What about formatting? Key-value pairs only.

## License

MIT License
