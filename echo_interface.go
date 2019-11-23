package ansilog

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

func (l Logger) Output() io.Writer {
	return l.Out
}

// SetOutput
func (l Logger) SetOutput(w io.Writer) {
	l.SetOutput(w)
}

func (l Logger) Prefix() string {
	// Is this even valid?  I'm not sure it can be translated since
	// logrus uses a Formatter interface.  Which seems to me to probably be
	// a better way to do it.
	return ""
}

func (l Logger) SetPrefix(p string) {
}

// Level returns the log level
func (l Logger) Level() log.Lvl {
	switch l.Logger.Level {
	case logrus.DebugLevel:
		return log.DEBUG
	case logrus.WarnLevel:
		return log.WARN
	case logrus.ErrorLevel:
		return log.ERROR
	case logrus.InfoLevel:
		return log.INFO
	default:
		l.Panic("Invalid level")
	}

	return log.OFF
}

func (l *Logger) SetHeader(h string) {
	//l.template = l.newTemplate(h)
}

// SetLevel set the log level
func (l Logger) SetLevel(lvl log.Lvl) {
	switch lvl {
	case log.DEBUG:
		l.Logger.SetLevel(logrus.DebugLevel)
	case log.WARN:
		l.Logger.SetLevel(logrus.WarnLevel)
	case log.ERROR:
		l.Logger.SetLevel(logrus.ErrorLevel)
	case log.INFO:
		l.Logger.SetLevel(logrus.InfoLevel)
	default:
		l.Panic("Invalid level")
	}
}

// Printj
func (l Logger) Printj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Print()
}

// Debugj
func (l Logger) Debugj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Debug()
}

// Infoj
func (l Logger) Infoj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Info()
}

// Warnj
func (l Logger) Warnj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Warn()
}

// Errorj
func (l Logger) Errorj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Error()
}

// Fatalj
func (l Logger) Fatalj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Fatal()
}

// Panicj
func (l Logger) Panicj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Panic()
}

// EchoHTTPErrorHandler ------------------------------------------------------------------------------------------------

// NewEchoHTTPErrorHandler return a custom HTTP error handler.
// It sends a JSON response with status code.
func (l *Logger) NewEchoHTTPErrorHandler(debug, log bool) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		he, ok := err.(*echo.HTTPError)

		fields := logrus.Fields{}

		if ok {
			// skip details for common errors
			if he.Code > 399 &&
				he.Code != http.StatusNotFound &&
				he.Code != http.StatusUnauthorized &&
				he.Code != http.StatusUnprocessableEntity {

				switch c.Request().Method {
				case "GET":
					fields["params"] = fmt.Sprintf("%+v", c.Request().URL.Query())
				case "POST", "PUT", "DELETE":
					if b, e := ioutil.ReadAll(c.Request().Body); e == nil {
						fields["params"] = string(b)
					}
				}
				fields["skip"] = 3
				fields["status"] = he.Code
				fields["method"] = c.Request().Method
				fields["host"] = c.Request().Host
				fields["uri"] = c.Request().RequestURI
				fields["user_agent"] = c.Request().UserAgent()

				// extract stack-trace from errors created with "github.com/pkg/errors"
				// packages using Wrap() or WithStack() funcs.
				//fields["stack"] = fmt.Sprintf("%+v", he.Internal)
				_ = c.Request().Body.Close()
			}
		} else {
			he = &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  http.StatusText(http.StatusInternalServerError),
				Internal: err,
			}
		}

		// Send response
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead { // Issue #608
				err = c.NoContent(he.Code)
			} else {
				if debug {
					if log {
						l.WithError(he.Internal).WithFields(fields).Errorln(he.Error())
					}
					he.Message = map[string]interface{}{"message": he.Error()}
				} else if m, ok := he.Message.(string); ok {
					if log {
						l.WithError(he.Internal).WithFields(fields).Errorln(m)
					}
					he.Message = map[string]interface{}{"message": m}
				} else if log {
					l.WithError(he.Internal).WithFields(fields).Errorln(he.Message)
				}

				err = c.JSON(he.Code, he.Message)
			}
		}
	}
}
