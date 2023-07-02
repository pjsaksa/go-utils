package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	debugColor   = "\x1B[90m"
	errorColor   = "\x1B[91m"
	eventColor   = "\x1B[94m"
	fatalColor   = "\x1B[41m"
	infoColor    = "\x1B[97m"
	warningColor = "\x1B[93m"
	//
	resetColor = "\x1B[m"
)

// ------------------------------------------------------------

var mainOutput io.Writer = os.Stderr
var forcedTime time.Time

func SetOutput(w io.Writer)      { mainOutput = w }
func ResetOutput()               { mainOutput = os.Stderr }
func setForcedTime(ft time.Time) { forcedTime = ft }
func resetForcedTime()           { forcedTime = time.Time{} }

// ------------------------------------------------------------

func DebugMsg(format string, v ...any) Message {
	message := fmt.Sprintf(format, v...)
	return Message{
		panic: false,
		Line:  debugColor + message + resetColor,
	}
}

func ErrorMsg(format string, v ...any) Message {
	message := fmt.Sprintf(format, v...)
	return Message{
		panic: false,
		Line:  errorColor + "ERROR: " + message + resetColor,
	}
}

func EventMsg(format string, v ...any) Message {
	message := fmt.Sprintf(format, v...)
	return Message{
		panic: false,
		Line:  eventColor + "EVENT: " + message + resetColor,
	}
}

func InfoMsg(format string, v ...any) Message {
	message := fmt.Sprintf(format, v...)
	return Message{
		panic: false,
		Line:  infoColor + "INFO: " + message + resetColor,
	}
}

func WarningMsg(format string, v ...any) Message {
	message := fmt.Sprintf(format, v...)
	return Message{
		panic: false,
		Line:  warningColor + "WARNING: " + message + resetColor,
	}
}

func FatalMsg(format string, v ...any) Message {
	message := fmt.Sprintf(format, v...)
	return Message{
		panic: true,
		Line:  fatalColor + "FATAL: " + message + resetColor,
	}
}

// ------------------------------------------------------------

func DEBUG(format string, v ...any) {
	DebugMsg(format, v...).write()
}

func ERROR(format string, v ...any) {
	ErrorMsg(format, v...).write()
}

func EVENT(format string, v ...any) {
	EventMsg(format, v...).write()
}

func INFO(format string, v ...any) {
	InfoMsg(format, v...).write()
}

func WARNING(format string, v ...any) {
	WarningMsg(format, v...).write()
}

func FATAL(format string, v ...any) {
	FatalMsg(format, v...).write()
}
