package ansilog

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/urfave/negroni"
)

const defaultLogTemplate = "{{.Time}} | {{.Duration}} | {{.Status}} | {{.Method}} | {{.Host}} | {{.RequestURI}}"

// HttpTracer is a middleware that logs http requests.
type HttpTracer struct {
	*log.Logger

	TimeFormat string

	// Template is the log template used by fmt.
	// Possible fields are:
	// 	Time
	// 	Status
	// 	Duration
	// 	Host
	// 	Method
	// 	ContentLength
	// 	RemoteAddr
	// 	UserAgent
	// 	RequestURI
	// 	Proto
	// Default format is: "{{.Time}} | {{.Duration}} | {{.Status}} | {{.Method}} | {{.Host}} | {{.RequestURI}}"
	Template *template.Template
}

// NewHttpTracer returns a new HttpTracer instance.
func NewHttpTracer(prefix string) *HttpTracer {
	return &HttpTracer{
		Logger:     log.New(os.Stdout, Blue("[")+Yellow(prefix)+Blue("] "), 0),
		TimeFormat: time.RFC822Z, //"2006-01-02 15:04:05"
		Template:   template.Must(template.New("ansilog_parser").Parse(defaultLogTemplate)),
	}
}

func (hl *HttpTracer) trace(rw interface{}, r *http.Request, start time.Time) {
	duration := time.Since(start)

	metricsEntry := struct {
		Time          string
		Proto         string
		RemoteAddr    string
		Status        string
		Method        string
		Duration      string
		ContentLength string
		Host          string
		RequestURI    string
		UserAgent     string
	}{
		Time:       start.Format(hl.TimeFormat),
		Proto:      r.Proto,
		RemoteAddr: fmt.Sprintf("%-14s", r.RemoteAddr),
		Status:     hl.fetchStatusCode(rw),
		Method:     fmt.Sprintf("%-7s", r.Method),
		Duration:   fmt.Sprintf("%20s", duration),
		//ContentLength: fmt.Sprintf("%12s bytes", hl.fetchLength(rw)),
		Host:       fmt.Sprintf("%-22s", r.Host),
		RequestURI: r.RequestURI, // path will exclude '/v1'
		UserAgent:  r.UserAgent(),
	}
	buff := &bytes.Buffer{}
	if err := hl.Template.Execute(buff, metricsEntry); err != nil {
		fmt.Println(Red(err))
	} else {
		hl.Println(buff.String())
	}
}

// fetchStatusCode attempts to see if the passed type implements a Status() method.
// If so, it is called and the value is returned.
func (hl *HttpTracer) fetchStatusCode(rw interface{}) string {
	statusCode := 0

	// Compatible with negroni custom ResponseWriter
	if crw, ok := rw.(interface{ Status() int }); ok {
		statusCode = crw.Status()
	} else if echoResponse, ok := rw.(*echo.Response); ok {
		statusCode = echoResponse.Status
	}

	switch {
	case statusCode < 300:
		return Green(strconv.Itoa(statusCode))
	case statusCode >= 300 && statusCode < 400:
		// redirects
		return Cyan(strconv.Itoa(statusCode))
	case statusCode >= 400 && statusCode < 500:
		// client errors
		return Magenta(strconv.Itoa(statusCode))
	case statusCode >= 500:
		// server error
		return Red(strconv.Itoa(statusCode))
	default:
		return Red("unknown status")
	}
}

func (hl *HttpTracer) fetchLength(rw interface{}) string {
	var length int
	if crw, ok := rw.(responseWriter); ok {
		length = crw.Size()
	} else if echoResponse, ok := rw.(*echo.Response); ok {
		length = int(echoResponse.Size)
	}
	return strconv.Itoa(length)
}

// Middleware ----------------------------------------------------------------------------------------------------------

// responseWriter is a custom http.ResponseWriter that holds statusCode and contentLength
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) Write(bytes []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	size, err := rw.ResponseWriter.Write(bytes)
	rw.size += size
	return size, err
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Size() int {
	return rw.size
}

// HTTPLogHandlerFunc is an http.HandlerFunc middleware.
func (hl *HttpTracer) HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		nrw := &responseWriter{rw, http.StatusOK, 0}
		defer hl.trace(nrw, r, start)
		next.ServeHTTP(nrw, r)
	}
}

// HTTPLogHandler is an http.Handler middleware.
func (hl *HttpTracer) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		nrw := &responseWriter{rw, http.StatusOK, 0}
		defer hl.trace(nrw, r, start)
		next.ServeHTTP(nrw, r)
	})
}

// EchoHTTPHandler is an Echo middleware.
func (hl *HttpTracer) EchoMiddlewareFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		start := time.Now()
		defer hl.trace(c.Response(), r, start)
		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}

// Negroni interface
func (hl *HttpTracer) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	nrw := negroni.NewResponseWriter(rw)
	defer hl.trace(nrw, r, start)
	next(nrw, r)
}