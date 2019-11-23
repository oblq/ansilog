package ansilog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/urfave/negroni"
)

const defaultLogTemplate = "{{.Time}} | {{.Latency}} | {{.Status}} | {{.Method}} | {{.Host}} | {{.RequestURI}}"

// HttpTracer is a middleware that logs http requests.
type HttpTracer struct {
	*log.Logger

	TimeFormat string

	// Template is the log template used by fmt.
	// Possible fields are:
	// 	Time
	// 	Status
	// 	Latency
	// 	Host
	// 	Method
	// 	ContentLength
	// 	RemoteAddr
	// 	UserAgent
	// 	RequestURI
	// 	Proto
	// Default format is: "{{.Time}} | {{.Latency}} | {{.Status}} | {{.Method}} | {{.Host}} | {{.RequestURI}}"
	Template *template.Template

	// Colors ------------------------------------------------

	ConsoleColorsMode ConsoleColorsModeEnum
	outIsTerm         bool

	black   painter
	white   painter
	red     painter
	green   painter
	yellow  painter
	blue    painter
	magenta painter
	cyan    painter
}

// NewHttpTracer returns a new HttpTracer instance.
func NewHttpTracer(prefix string) *HttpTracer {
	tracer := &HttpTracer{
		TimeFormat:        time.RFC822Z, //"2006-01-02 15:04:05"
		Template:          template.Must(template.New("ansilog_parser").Parse(defaultLogTemplate)),
		ConsoleColorsMode: ConsoleColorsModeAuto,
		outIsTerm:         true,
	}

	var out io.Writer
	out = os.Stdout
	tracer.outIsTerm = IsTerm(out)

	var mustPaint = func() bool {
		return tracer.ConsoleColorsMode == ConsoleColorsModeEnabled ||
			(tracer.ConsoleColorsMode == ConsoleColorsModeAuto && tracer.outIsTerm)
	}

	tracer.black = NewDynamicPainter(black, mustPaint)
	tracer.white = NewDynamicPainter(white, mustPaint)
	tracer.red = NewDynamicPainter(red, mustPaint)
	tracer.green = NewDynamicPainter(green, mustPaint)
	tracer.yellow = NewDynamicPainter(yellow, mustPaint)
	tracer.blue = NewDynamicPainter(blue, mustPaint)
	tracer.magenta = NewDynamicPainter(magenta, mustPaint)
	tracer.cyan = NewDynamicPainter(cyan, mustPaint)

	tracer.Logger = log.New(out, tracer.blue("[")+tracer.yellow(prefix)+tracer.blue("] "), 0)

	return tracer
}

func (hl *HttpTracer) trace(rw interface{}, r *http.Request, start time.Time) {
	latency := time.Since(start)
	if latency > time.Minute {
		latency = latency - latency%time.Second
	}

	metricsEntry := struct {
		Time          string
		Proto         string
		RemoteAddr    string
		Status        string
		Method        string
		Latency       string
		ContentLength string
		Host          string
		RequestURI    string
		UserAgent     string
	}{
		Time:       start.Format(hl.TimeFormat),
		Proto:      r.Proto,
		RemoteAddr: fmt.Sprintf("%-14s", r.RemoteAddr),
		Status:     fmt.Sprintf("%3s", hl.fetchStatusCode(rw)),
		Method:     fmt.Sprintf("%-7s", hl.coloredMethod(r.Method)),
		Latency:    fmt.Sprintf("%13s", latency),
		//ContentLength: fmt.Sprintf("%12s bytes", hl.fetchLength(rw)),
		Host:       fmt.Sprintf("%-22s", r.Host),
		RequestURI: r.RequestURI, // path will exclude '/v1'
		UserAgent:  r.UserAgent(),
	}
	buff := &bytes.Buffer{}
	if err := hl.Template.Execute(buff, metricsEntry); err != nil {
		fmt.Println(hl.red(err))
	} else {
		hl.Println(buff.String())
	}
}

// MethodColor is the ANSI color for appropriately logging http method to a terminal.
func (hl *HttpTracer) coloredMethod(method string) string {
	switch method {
	case "GET":
		return hl.green(method)
		//return Blue(method)
	case "POST":
		return hl.blue(method)
		//return Cyan(method)
	case "PUT":
		return hl.cyan(method)
		//return Yellow(method)
	case "DELETE":
		return hl.red(method)
	case "PATCH":
		return hl.yellow(method)
		//return hl.green(method)
	case "HEAD":
		return hl.magenta(method)
	case "OPTIONS":
		return hl.white(method)
	default:
		return method
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
	case statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices: // 200 300
		return hl.green(strconv.Itoa(statusCode))
	case statusCode >= http.StatusMultipleChoices && statusCode < http.StatusBadRequest: // redirects... 300 400
		return hl.cyan(strconv.Itoa(statusCode))
	case statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError: // client errors... 400 500
		return hl.magenta(strconv.Itoa(statusCode))
	case statusCode >= http.StatusInternalServerError: // server error
		return hl.red(strconv.Itoa(statusCode))
	default:
		return hl.red("unknown status")
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
