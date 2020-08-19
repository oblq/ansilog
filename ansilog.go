package ansilog

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/oblq/ansilog/internal/hooks/pghook"
	"github.com/oblq/ansilog/internal/hooks/stack_trace"
	"github.com/oblq/swap"
	"github.com/sirupsen/logrus"
)

type Config struct {
	// Out is a writer where logs are written.
	// Optional. Default value is os.Stdout.
	Out io.Writer

	// The logging level the logger should log at.
	// This defaults to `info`, which allows
	// Info(), Warn(), Error() and Fatal() to be logged.
	// Allowed values are:
	// panic, fatal, error, warn, warning, info, debug and trace.
	Level string

	// StackTrace will extract stack-trace from errors created
	// with "github.com/pkg/errors" package
	// using Wrap() or WithStack() funcs.
	StackTrace bool

	// PostgresLevel > 0 will initialize a new postgres instance for the corresponding hook,
	// also, postgres parameters must be provided in that case
	PostgresLevel string

	// Postgres
	Host     string
	Port     int
	DB       string
	User     string
	Password string
}

type Logger struct {
	*logrus.Logger
}

func NewWithConfig(config Config) (logger *Logger, err error) {
	logger = &Logger{Logger: logrus.New()}
	err = logger.setup(config)
	return
}

func NewWithConfigPath(configFilePath string) (logger *Logger, err error) {
	if len(configFilePath) == 0 {
		return nil, errors.New("a valid config file path must be provided")
	}

	logger = &Logger{Logger: logrus.New()}

	var config Config
	if err = swap.Parse(&config, configFilePath); err != nil {
		return
	}

	err = logger.setup(config)
	return
}

func (l *Logger) Configure(configFiles ...string) (err error) {
	l.Logger = logrus.New()

	var config Config
	if err = swap.Parse(&config, configFiles...); err != nil {
		return err
	}

	if err = l.setup(config); err != nil {
		return err
	}
	return
}

func (l *Logger) setup(config Config) error {
	if config.Out != nil {
		l.Out = config.Out
	} else {
		l.Out = os.Stdout
	}

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	l.Logger.Level = level
	// config.Level
	//log.Formatter = &logrus.JSONFormatter{
	//	time.RFC3339,
	//	false,
	//	nil,
	//}

	l.Formatter = &logrus.TextFormatter{
		ForceColors:            true,
		DisableTimestamp:       false,
		FullTimestamp:          true,
		TimestampFormat:        time.RFC3339, //"2006-01-02 15:04:05", // time.RFC3339, // //"2006-01-02 15:04 Z07:00", 2006-01-02T15:04:05-0700
		DisableSorting:         false,
		DisableLevelTruncation: true,
		QuoteEmptyFields:       true,
	}

	//l.Formatter = &logrus.JSONFormatter{
	//	TimestampFormat:   "2006-01-02 15:04:05",
	//	DisableTimestamp:  false,
	//	DisableHTMLEscape: false,
	//	DataKey:           "",
	//	FieldMap:          nil,
	//	CallerPrettyfier:  nil,
	//	PrettyPrint:       true,
	//}

	if config.StackTrace {
		l.AddHook(stack_trace.New())
	}

	if len(config.PostgresLevel) > 0 {

		dbConf := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
			config.Host, config.Port, config.DB, config.User, config.Password)

		db, err := sql.Open("postgres", dbConf)
		if err != nil {
			return fmt.Errorf("[logger] can't connect to postgresql database: %v\nPostgres config: %+v\n", err, config)
		}

		//defer db.Close()

		// NewAsyncHook
		hook := pghook.NewHook(db)
		//defer hook.Flush()

		pgLevel, err := logrus.ParseLevel(config.PostgresLevel)
		if err != nil {
			return fmt.Errorf("[logger] invalid postgres level, no log will be saved to it: %+v", config.PostgresLevel)
		}

		hook.AddFilter(func(entry *logrus.Entry) *logrus.Entry {
			if entry != nil {
				// ignore entries
				if _, ignore := entry.Data["ignore"]; ignore || entry.Level > pgLevel {
					entry = nil
				}
			}

			return entry
		})

		l.AddHook(hook)
	}

	return nil
}
