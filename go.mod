module github.com/oblq/ansilog

go 1.14

require (
	github.com/labstack/echo/v4 v4.1.11
	github.com/labstack/gommon v0.3.0
	github.com/lib/pq v1.2.0
	github.com/mattn/go-isatty v0.0.9
	github.com/oblq/sprbox v1.5.0
	github.com/oblq/swap v0.0.1
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/negroni v1.0.0
	golang.org/x/crypto v0.0.0-20191108234033-bd318be0434a // indirect
)

replace github.com/oblq/sprbox => ../sprbox
