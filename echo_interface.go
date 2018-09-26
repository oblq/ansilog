package ansilog

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
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

// EchoHTTPErrorHandler is a custom echo http error handler which add some useful info to the log
func (l *Logger) EchoHTTPErrorHandler(err error, c echo.Context) {

	if he, ok := err.(*echo.HTTPError); ok {

		if he.Code > 399 &&
			he.Code != http.StatusNotFound &&
			he.Code != http.StatusUnauthorized &&
			he.Code != http.StatusUnprocessableEntity {

			var params string

			switch c.Request().Method {
			case "GET":
				params = fmt.Sprintf("%+v", c.Request().URL.Query())
			case "POST", "PUT", "DELETE":
				defer c.Request().Body.Close()
				if b, e := ioutil.ReadAll(c.Request().Body); e == nil {
					params = string(b)
				}
			}

			fields := logrus.Fields{
				"skip":       3,
				"http":       true,
				"status":     he.Code,
				"host":       c.Request().Host,
				"method":     c.Request().Method,
				"path":       c.Request().URL.Path,
				"user_agent": c.Request().UserAgent(),
				"params":     params,
			}

			message, ok := he.Message.(string)
			if !ok {
				message = err.Error()
			}

			l.WithError(err).WithFields(fields).Errorln(message)
		}
	} else {
		code := http.StatusInternalServerError
		errorPage := fmt.Sprintf("%d.html", code)
		if err := c.File(errorPage); err != nil {
			l.WithError(err).Errorln(err.Error())
		}
		l.WithError(err).Errorln(err.Error())
	}
}

// returned from EchoMiddleware() below.
func (l *Logger) echoMiddlewareHandler(c echo.Context, next echo.HandlerFunc) error {
	req := c.Request()
	res := c.Response()
	start := time.Now()
	if err := next(c); err != nil {
		c.Error(err)
	}
	stop := time.Now()

	p := req.URL.Path
	if p == "" {
		p = "/"
	}

	bytesIn := req.Header.Get(echo.HeaderContentLength)
	if bytesIn == "" {
		bytesIn = "0"
	}

	l.WithFields(map[string]interface{}{
		"time_rfc3339":  time.Now().Format(time.RFC3339),
		"remote_ip":     c.RealIP(),
		"host":          req.Host,
		"uri":           req.RequestURI,
		"method":        req.Method,
		"path":          p,
		"referer":       req.Referer(),
		"user_agent":    req.UserAgent(),
		"status":        res.Status,
		"latency":       strconv.FormatInt(stop.Sub(start).Nanoseconds()/1000, 10),
		"latency_human": stop.Sub(start).String(),
		"bytes_in":      bytesIn,
		"bytes_out":     strconv.FormatInt(res.Size, 10),
	}).Info("Handled request")

	return nil
}

// EchoMiddleware is an echo middleware that log http requests metrics
func (l *Logger) EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return l.echoMiddlewareHandler(c, next)
		}
	}
}
