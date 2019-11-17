package stack_trace

import (
	"strings"

	stack2 "github.com/oblq/ansilog/internal/hooks/stack_trace/stack"

	"github.com/sirupsen/logrus"
)

// NewHook is the initializer for LogrusStackHook{} (implementing logrus.Hook).
// Set levels to callerLevels for which "caller" value may be set, providing a
// single frame of stack. Set levels to stackLevels for which "stack" value may
// be set, providing the full stack (minus logrus).
func NewHook(callerLevels []logrus.Level, stackLevels []logrus.Level) LogrusStackHook {
	return LogrusStackHook{
		CallerLevels: callerLevels,
		StackLevels:  stackLevels,
	}
}

// StandardHook is a convenience initializer for LogrusStackHook{} with
// default args.
func DefaultHook() LogrusStackHook {
	return LogrusStackHook{
		CallerLevels: []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel}, //logrus.AllLevels, logrus.InfoLevel,
		StackLevels:  []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel},
	}
}

// LogrusStackHook is an implementation of logrus.Hook interface.
// use logger.WithField("skip", 3).Errorln("pippo") to skip frames from other libraries

type LogrusStackHook struct {
	// Set levels to CallerLevels for which "caller" value may be set,
	// providing a single frame of stack.
	CallerLevels []logrus.Level

	// Set levels to StackLevels for which "stack" value may be set,
	// providing the full stack (minus logrus).
	StackLevels []logrus.Level
}

// Levels provides the levels to filter.
func (hook LogrusStackHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is called by logrus when something is logged.
func (hook LogrusStackHook) Fire(entry *logrus.Entry) error {
	var skipFrames int
	if len(entry.Data) == 0 {
		// When WithField(s) is not used, we have 8 logrus frames to skip.
		skipFrames = 8
	} else {
		// When WithField(s) is used, we have 6 logrus frames to skip.
		skipFrames = 6
	}

	if skip, ok := entry.Data["skip"].(int); ok {
		skipFrames += skip + 1 // add 1 to remove also the call to add skip field
	}

	// Get the complete stack track past skipFrames count.
	_stack := stack2.Callers(skipFrames)

	var stack stack2.Stack

	// Remove logrus's own frames that seem to appear after the code is through
	// certain hooks. e.g. http handler in a separate package.
	// This is a workaround.
	for _, frame := range _stack {
		if !strings.Contains(frame.File, "github.com/sirupsen/logrus") {
			stack = append(stack, frame)
		}
	}

	if len(stack) > 0 {
		// If we have a frame, we set it to "caller" field for assigned levels.
		for _, level := range hook.CallerLevels {
			if entry.Level == level {
				entry.Message += " <- " + stack[0].String() + "\n"
				entry.Data["caller"] = stack[0].String()
				break
			}
		}

		// Set the available frames to "stack" field.
		for _, level := range hook.StackLevels {
			if entry.Level == level {
				entry.Message += " <- " + stack.String() + "\n"
				entry.Data["stack"] = stack.String()
				entry.Data["caller"] = stack[0].String()
				break
			}
		}
	}

	return nil
}
