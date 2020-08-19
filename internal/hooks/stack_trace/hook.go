package stack_trace

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// New returns a new stack_trace instance.
// It automatically extracts stack-trace from errors created with "github.com/pkg/errors"
func New() LogrusStackHook {
	return LogrusStackHook{}
}

// LogrusStackHook is an implementation of logrus.Hook interface.
type LogrusStackHook struct {
	logrus.Hook
}

// Levels provide the levels to be logged.
func (hook LogrusStackHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire log an entry.
func (hook LogrusStackHook) Fire(entry *logrus.Entry) error {
	// extract stack-trace from errors created with "github.com/pkg/errors"
	// packages using Wrap() or WithStack() funcs.
	if err, ok := entry.Data[logrus.ErrorKey]; ok {
		stack := fmt.Sprintf("%+v", err)
		// escape new lines
		//consecutiveNewLines := regexp.MustCompile(`\n`)
		//stack = consecutiveNewLines.ReplaceAllString(stack, "\n")
		entry.Data["stack"] = stack
	}
	return nil
}
