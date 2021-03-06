// Package log provides utilities for providing output to the user.
package log

import (
	"fmt"
	"os"

	logging "github.com/op/go-logging"
)

// Logger is a re-exported logger from `go-logging`. Originally, we provided
// wrapper functions around the specific logging methods that we use, but this
// causes the package, file, and line numbers to be useless (they all point to
// the wrapper instead of the caller). Instead, we only use the documented log
// levels below:
//
// _Debug_ messages are used for tracing execution and diagnosing unintended
// error cases.
//
// _Notice_ messages are non-error events that the user should be informed of,
// such as notifications that an action has occurred.
//
// _Warning_ messages are non-fatal error events that the user should be
// informed of. Generally, the user can do something to fix these.
//
// _Fatal_ messages are non-recoverable errors, and cause an `os.Exit(1)`. They
// should be used for foreseen error conditions that we cannot continue from.
//
// _Panic_ messages are errors that are unforeseen, should never happen, and
// indicate that something has gone terribly wrong. They are akin to assertion
// failures, and are generally only used as sanity checks for invariants.
var Logger = logging.MustGetLogger("fossa-cli")

// Initialize configures logging. If `interactive` is true, then logging will
// include colors and ANSI codes (e.g. spinners). If `debug` is true, then
// logging will include debugging output.
func Initialize(interactive, debug bool) {
	// If `interactive`, then use ANSI codes (spinner + colors)
	useSpinner = interactive
	var colorOn, colorOff string
	if interactive {
		colorOn = "%{color}"
		colorOff = "%{color:reset}"
	}

	// If `debug`, then log in debug format and at debug level.
	formatter := logging.MustStringFormatter(colorOn + "%{level}" + colorOff + " %{message}")
	if debug {
		formatter = logging.MustStringFormatter(colorOn + "%{time} %{level} %{shortpkg}/%{shortfile}/%{shortfunc}" + colorOff + " %{message}")
	}
	stderrBackend := logging.AddModuleLevel(logging.NewBackendFormatter(logging.NewLogBackend(os.Stderr, "", 0), formatter))
	if debug {
		stderrBackend.SetLevel(logging.DEBUG, "")
	} else {
		stderrBackend.SetLevel(logging.WARNING, "")
	}
	logging.SetBackend(stderrBackend)
}

// Print outputs a literal message to STDOUT.
func Print(args ...interface{}) {
	fmt.Print(args...)
}

// Printf outputs a formatted message to STDOUT.
func Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

type Fields map[string]interface{}

type Entry struct {
	Message string
	Error   error
	Fields  Fields
}

func (e Entry) String() string {
	var fields string
	for key, val := range e.Fields {
		switch v := val.(type) {
		case fmt.Stringer:
			fields += fmt.Sprintf(" %s=%#v", key, v.String())
		default:
			fields += fmt.Sprintf(" %s=%#v", key, v)
		}
	}
	if e.Error != nil {
		return fmt.Sprintf("%s error=%#v %s", e.Message, e.Error, fields)
	}
	return fmt.Sprintf("%s%s", e.Message, fields)
}
