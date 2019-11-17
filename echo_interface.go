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

// EchoHTTPErrorHandler return a custom HTTP error handler. It sends a JSON response
// with status code.
func (l *Logger) EchoHTTPErrorHandler(debug, log bool) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		he, ok := err.(*echo.HTTPError)
		if ok {
			// skip details for common errors
			if he.Code > 399 &&
				he.Code != http.StatusNotFound &&
				he.Code != http.StatusUnauthorized &&
				he.Code != http.StatusUnprocessableEntity {

				var params string
				switch c.Request().Method {
				case "GET":
					params = fmt.Sprintf("%+v", c.Request().URL.Query())
				case "POST", "PUT", "DELETE":
					if b, e := ioutil.ReadAll(c.Request().Body); e == nil {
						params = string(b)
					}
				}

				fields := logrus.Fields{
					"skip":           3,
					"status":         he.Code,
					"method":         c.Request().Method,
					"host":           c.Request().Host,
					"uri":            c.Request().RequestURI,
					"user_agent":     c.Request().UserAgent(),
					"params":         params,
					"internal_error": he.Internal.Error(),
				}

				_ = c.Request().Body.Close()

				if log {
					l.WithError(err).WithFields(fields).Errorln(he.Message)
				}
			} else if log {
				l.WithError(he).Errorln(he.Message)
			}
		} else {
			he = &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  http.StatusText(http.StatusInternalServerError),
				Internal: err,
			}
			if log {
				l.WithError(he).Errorln(he.Message)
			}
		}

		// Send response
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead { // Issue #608
				err = c.NoContent(he.Code)
			} else {
				if debug {
					// err is *echo.HTTPError
					//he.Message = err.Error() // produce a string
					he.Message = map[string]interface{}{
						"code":     he.Code,
						"message":  he.Message,
						"internal": he.Internal.Error(),
					}
				} else if m, ok := he.Message.(string); ok {
					he.Message = map[string]interface{}{"message": m}
				}

				err = c.JSON(he.Code, he.Message)
			}
		}
	}
}

//// returned from EchoMiddleware() below.
//func (l *Logger) echoMiddlewareHandler(c echo.Context, next echo.HandlerFunc) error {
//	req := c.Request()
//	res := c.Response()
//	start := time.Now()
//	if err := next(c); err != nil {
//		c.Error(err)
//	}
//	stop := time.Now()
//
//	p := req.URL.Path
//	if p == "" {
//		p = "/"
//	}
//
//	bytesIn := req.Header.Get(echo.HeaderContentLength)
//	if bytesIn == "" {
//		bytesIn = "0"
//	}
//
//	l.WithFields(map[string]interface{}{
//		"time_rfc3339":  time.Now().Format(time.RFC3339),
//		"remote_ip":     c.RealIP(),
//		"host":          req.Host,
//		"uri":           req.RequestURI,
//		"method":        req.Method,
//		"path":          p,
//		"referer":       req.Referer(),
//		"user_agent":    req.UserAgent(),
//		"status":        res.Status,
//		"latency":       strconv.FormatInt(stop.Sub(start).Nanoseconds()/1000, 10),
//		"latency_human": stop.Sub(start).String(),
//		"bytes_in":      bytesIn,
//		"bytes_out":     strconv.FormatInt(res.Size, 10),
//	}).Info("Handled request")
//
//	return nil
//}
//
//// EchoMiddleware is an echo middleware that log http requests metrics
//func (l *Logger) EchoMiddleware() echo.MiddlewareFunc {
//	return func(next echo.HandlerFunc) echo.HandlerFunc {
//		return func(c echo.Context) error {
//			return l.echoMiddlewareHandler(c, next)
//		}
//	}
//}
