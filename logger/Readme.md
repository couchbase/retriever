PACKAGE DOCUMENTATION

package logger
    import "github.com/couchbaselabs/retriever/logger"



CONSTANTS

const (
    LevelError = logLevel(iota)
    LevelWarn
    LevelInfo
    LevelDebug
)

const (
    Default logDevice = iota << 1 // log to stdout
    File                          // user specified file name
    Remote                        // remote host
)

const DEFAULT_PATH = "/tmp"

const MAX_CLEANUP_COUNTER = 1000


TYPES

type LogWriter struct {
    // contains filtered or unexported fields
}


func NewLogger(module string, level logLevel) (*LogWriter, error)
    Create a new instance of a logWriter


func (lw *LogWriter) DisableKeys(keys []string) error
    disable component keys

func (lw *LogWriter) DisableTransactionLogging()
    Disable logging to a transaction file

func (lw *LogWriter) EnableKeys(keys []string) error
    enable component keys

func (lw *LogWriter) EnableTransactionLogging()
    Set the logging to the log to a transaction file

func (lw *LogWriter) LogDebug(transactionId string, key string, format string, args ...interface{})
    log debug. transaction id, component id, log message

func (lw *LogWriter) LogError(transactionId string, key string, format string, args ...interface{})
    log error transaction id, component id, log message

func (lw *LogWriter) LogInfo(transactionId string, key string, format string, args ...interface{})
    log info. transaction id, component id, log message

func (lw *LogWriter) LogWarn(transactionId string, key string, format string, args ...interface{})
    log warning transaction id, component id, log message

func (lw *LogWriter) SetFile(path string) error
    Set the output device

func (lw *LogWriter) SetLogHost(string) error
    Set the remote host

func (lw *LogWriter) SetLogLevel(level logLevel) error
    Set the log level


type TransactionLogger struct {
    // contains filtered or unexported fields
}




