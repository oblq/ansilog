package ansilog

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/oblq/ansilog/internal/hooks/pghook"
	"github.com/oblq/ansilog/internal/hooks/stack_trace"
	"github.com/oblq/sprbox"
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

	// Postgres
	PostgresLevel string
	Host          string
	Port          int
	DB            string
	User          string
	Password      string
}

type Logger struct {
	*logrus.Logger
}

func NewWithConfig(config Config) (logger *Logger, err error) {
	logger = &Logger{Logger: logrus.New()}
	err = logger.setup(config)
	return
}

func NewWithConfigFile(configFilePath string) (logger *Logger, err error) {
	if len(configFilePath) == 0 {
		return nil, errors.New("a valid config file path must be provided")
	}

	logger = &Logger{Logger: logrus.New()}

	var config Config
	var compsConfigFile []byte
	if compsConfigFile, err = ioutil.ReadFile(configFilePath); err != nil {
		return
	} else if err = sprbox.Unmarshal(compsConfigFile, &config); err != nil {
		return
	}

	err = logger.setup(config)
	return
}

func (l *Logger) SpareConfig(configFiles []string) (err error) {
	l.Logger = logrus.New()

	var config Config
	if err = sprbox.LoadConfig(&config, configFiles...); err != nil {
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
		TimestampFormat:        "2006-01-02 15:04:05", // time.RFC3339, // //"2006-01-02 15:04 Z07:00",
		DisableSorting:         false,
		DisableLevelTruncation: true,
		QuoteEmptyFields:       true,
	}

	if config.StackTrace {
		l.AddHook(stack_trace.DefaultHook())
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
		hook := pghook.NewHook(db, nil) //, map[string]interface{}{"this": "is logged every time"})
		//defer hook.Flush()

		pgLevel, err := logrus.ParseLevel(config.PostgresLevel)
		if err != nil {
			return fmt.Errorf("[logger] invalid postgres level, no log will be saved to it: %+v", config.PostgresLevel)
		}

		//hook.InsertFunc = func(db *sql.DB, entry *logrus.Entry) error {
		//	jsonData, err := json.Marshal(entry.Data)
		//	if err != nil {
		//		return err
		//	}
		//
		//	var errID int
		//	err = db.QueryRow("INSERT INTO logs(level, message, message_data, created_at) VALUES ($1,$2,$3,$4) returning id", entry.Level, entry.Message, jsonData, entry.Time).Scan(&errID)
		//
		//	//entry.WithField("pg_err_id", errID)
		//	entry.Data["pg_err_id"] = errID
		//
		//	return err
		//}

		hook.AddFilter(func(entry *logrus.Entry) *logrus.Entry {
			// ignore entries
			if _, ok := entry.Data["ignore"]; ok {
				entry = nil
			}
			if entry.Level > pgLevel {
				entry = nil
			}
			return entry
		})

		l.AddHook(hook)
	}

	return nil
}
