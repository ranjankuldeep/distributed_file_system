package logs

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

// Quiet specifies whether to only print machine-readable IDs
var Quiet bool

// logger wraps the logrus Logger together with the exit code
type logger struct {
	*logrus.Logger
	ExitCode int
}

var (
	// Logger is the exposed logger instance
	Logger *logger
	// mutex to ensure thread safety
	mu sync.Mutex
)

// newLogger creates a new instance of logger
func newLogger() *logger {
	l := &logger{
		Logger:   logrus.StandardLogger(), // Use the standard logrus logger
		ExitCode: 1,
	}

	l.ExitFunc = func(_ int) {
		os.Exit(l.ExitCode)
	}

	return l
}

// init initializes the logging system
func init() {
	mu.Lock()
	defer mu.Unlock()

	// Initialize the logger
	Logger = newLogger()

	// Set the output to be stdout in the normal case, but discard all log output in quiet mode
	if Quiet {
		Logger.SetOutput(io.Discard)
	} else {
		Logger.SetOutput(os.Stdout)
	}

	// Enable reporting the caller's file and line number
	Logger.SetReportCaller(true)

	// Customize the log output format
	Logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp:       false,
		ForceColors:            true,
		DisableColors:          false,
		FullTimestamp:          false,
		DisableLevelTruncation: false,
		PadLevelText:           true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := filepath.Base(f.File)
			return "", filename + ":" + fmt.Sprint(f.Line)
		},
	})

	// Disable the stdlib's automatic add of the timestamp in beginning of the log message,
	// as we stream the logs from stdlib log to this logrus instance.
	log.SetFlags(0)
	log.SetOutput(Logger.Writer())
	Logger.SetLevel(logrus.DebugLevel)
}
func IsLoggerInitialized() bool {
	return Logger != nil && Logger.Out != io.Discard
}
