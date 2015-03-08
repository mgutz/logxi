
![demo](https://github.com/mgutz/logxi/raw/master/images/demo.gif)

# logxi

log XI (eleven) is a structured [12 factor app](http://12factor.net/logs)
logger built for performance and happy development.

TL;DR

*   Structured with key-value pairs
*   Faster and less allocations than logrus and log15
*   Loggers and their level are configurable via environment variables
*   JSON formatter or very efficient key-value text formatter for production
*   Happy, colorful, developer friendly logger in terminal. Warnings
    and errors are emphasized with their call stack.
*   Defines a simple interface for logging
*   Colored logging works on Windows
*   Loggers are named
*   Has level guards to avoid cost of passing arguments

### Installation

    go get -u github.com/mgutz/logxi/v1

### Getting Started

```go
import "github.com/mgutz/logxi/v1"

// create package variable for Logger interface
var logger log.Logger

func main() {
    // use default logger
    if log.IsInfo() {
        log.Info("Hello", "name", "mario")
    }

    // create a logger for your package, assigning a unique
    // name which can be enabled from environment variables
    logger = log.New("pkg")

    // specify a writer
    modelLogger = log.NewLogger(os.Stdout, "models")

    db, err := sql.Open("postgres", "dbname=testdb")
    if err != nil {
        modelLogger.Error("Could not open database", "err", err)
    }

    a := "apple"
    o := "orange"
    if log.IsDebug() {
        // use key-value pairs after message
        logger.Debug("OK", "a", a, "o", o)
    }
}
```

logxi defaults to showing warnings and above. To view all logs

    LOGXI=* go run main.go

## Higlights

This logger package

*   Is fast in production environment

    A logger should be efficient and minimize performance tax.
    logxi encodes JSON 2X faster than logrus and log15 with primitive types.
    When diagnosing a problem in production, troubleshooting often means
    enabling small trace data in `Debug` and `Info` statements for some
    period of time.

    ```
# primitive types
BenchmarkLogxi         100000     20795 ns/op    3458 B/op     72 allocs/op
BenchmarkLogrus         20000     64447 ns/op    8691 B/op    180 allocs/op
BenchmarkLog15          20000     98889 ns/op    8756 B/op    220 allocs/op

# nested object
BenchmarkLogxiComplex   20000     75537 ns/op    8900 B/op    200 allocs/op
BenchmarkLogrusComplex  10000    103580 ns/op   11872 B/op    262 allocs/op
BenchmarkLog15Complex   10000    157817 ns/op   12259 B/op    294 allocs/op

* BenchmarkLog15Complex doesn't marshal nested object as JSON
```

*   Is developer friendly in the terminal. The HappyDevFormatter
    is colorful, prints file and line numbers for traces, warnings
    and errors. Errors print the call stack.

    `HappyDevFormatter` is not too concerned with performance
    and delegates to JSONFormatter internally.

*   Logs machine parsable output in production environments.
    The default formatter for non terminals is `JSONFormatter`.

    `TextFormatter` may also be used which is MUCH faster than
    JSON but there is no guarantee it can be easily parsed.

*   Has level guards to avoid the cost of building arguments. Get in the
    habit of using guards.

        if log.IsDebug() {
            log.Debug("some ", "key1", expensive())
        }

        if log.IsInfo() {
            log.Info("some ", "key1", expensive())
        }

        if log.IsWarn() {
            log.Warn("some ", "key1", expensive())
        }

    Error and Fatal do not have guards, they always log.

*   Conforms to a logging interface so it can be replaced.

        type Logger interface {
            Trace(msg string, args ...interface{})
            Debug(msg string, args ...interface{})
            Info(msg string, args ...interface{})
            Warn(msg string, args ...interface{})
            Error(msg string, args ...interface{})
            Fatal(msg string, args ...interface{})
            Log(level int, msg string, args []interface{})

            SetLevel(int)
            IsTrace() bool
            IsDebug() bool
            IsInfo() bool
            IsWarn() bool
            // Error, Fatal not needed, those SHOULD always be logged
        }

*   Standardizes on key-value pair argument sequence for machine parsing

    ```go
log.Debug("inside Fn()", "key1", value1, "key2", value2)
```
    logxi logs `FIX_IMBALANCED_PAIRS =>` if key-value pairs are imbalanced

*   Supports Color Schemes (256 colors)

    `log.New` creates a logger that supports color schemes

        logger := log.New("mylog")

    To customize scheme

        # emphasize errors with white text on red background
        LOGXI_COLORS="ERR=white:red" yourapp

        # emphasize errors with pink = 200 on 256 colors table
        LOGXI_COLORS="ERR=200" yourapp

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
variable. Valid values are `"happy", "text", "JSON"`

    # Use JSON in production with custom time
    LOGXI_FORMAT=JSON,t=2006-01-02T15:04:05.000000-0700 yourapp

The "happy" formatter has more options

*   fit - will try to fit a key-value pair onto the same line. If
    result characters are longer than `maxcol` then the pair will be
    put on the next line and indented

    "happy" defaults to pretty mode where it puts each key-value pair
    on its own line and indented

*   maxcol - maximum number of columns before forcing a key to be on its
    own line. If you want everything on a single line, set this to high
    value like 1000. Default is 80.

*   context - the number of context lines to print on source. Set to -1
    to see only file:lineno. Default is 2.

### Color Schemes

The color scheme may be set with `LOGXI_COLORS` environment variable. For
example, the default dark scheme is emulated like this

    # on non-Windows
    export LOGXI_COLORS=key=cyan+h,value,misc=blue+h,source=magenta,TRC,DBG,WRN=yellow,INF=green,ERR=red+h
    yourapp

    # on windows, stick to basic colors, styles liked bold, etc will not work
    SET LOGXI_COLORS=key=cyan,value,misc=blue,source=magenta,TRC,DBG,WRN=yellow,INF=green,ERR=red
    yourapp

    # color only errors
    LOGXI_COLORS=ERR=red yourapp

See [ansi](http://github.com/mgutz/ansi) package for styling. An empty
value, like "value" and "DBG" above means use default foreground and
background on terminal.

Keys

*   \*  - default color
*   TRC - trace color
*   DBG - debug color
*   WRN - warn color
*   INF - info color
*   ERR - error color
*   key - key color
*   value - value color unless WRN or ERR
*   misc - time and log name color
*   source - source context color (excluding error line)

## Extending

What about hooks? There are least two ways to do this

*   Implement your own `io.Writer` to write to external services. Be sure to set
    the formatter to JSON to faciliate decoding with Go's built-in streaming
    decoder.
*   Create an external filter. See `v1/cmd/filter` as an example.

What about log rotation? 12 factor apps only concern themselves with
STDOUT. Use shell redirection operators to write to a file.

There are many utilities to rotate logs which accept STDIN as input. They can
do many things like send alerts, etc. The two obvious choices are Apache's `rotatelogs`
utility and `lograte`.

```sh
yourapp | rotatelogs yourapp 86400
```

## Testing

```
# install godo task runner
go get -u gopkg.in/godo.v1/cmd/godo

# install dependencies
godo install -v

# run test
godo test

# run bench with allocs (requires manual cleanup)
godo allocs
```

## License

MIT License
