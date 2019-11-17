package stack_trace

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// NewHook is the initializer for LogrusStackHook (implementing logrus.Hook).
// extract stack-trace from errors created with "github.com/pkg/errors"
func NewHook(levelsToLog []logrus.Level) LogrusStackHook {
	return LogrusStackHook{LevelsToLog: levelsToLog}
}

// StandardHook is a convenience initializer for LogrusStackHook.
func DefaultHook() LogrusStackHook {
	return LogrusStackHook{
		LevelsToLog: []logrus.Level{
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
	}
}

// LogrusStackHook is an implementation of logrus.Hook interface.
type LogrusStackHook struct {
	// Levels set the levels for which the stack must be logged.
	LevelsToLog []logrus.Level
}

// Levels provides the levels to filter.
func (hook LogrusStackHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is called by logrus when something is logged.
func (hook LogrusStackHook) Fire(entry *logrus.Entry) error {
	// extract stack-trace from errors created with "github.com/pkg/errors"
	// packages using Wrap() or WithStack() funcs.
	if err, ok := entry.Data[logrus.ErrorKey]; ok {
		entry.Data["stack"] = fmt.Sprintf("%+v", err)
		fmt.Println(entry.Data["stack"])
	}
	return nil
}
