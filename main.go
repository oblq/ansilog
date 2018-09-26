package ansilog

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/oblq/ansilog/hooks/pghook"
	"github.com/oblq/ansilog/hooks/stack_trace"
	"github.com/oblq/sprbox"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Level      string
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

func NewLogger(configFilePath string, config *Config) *Logger {
	log := &Logger{Logger: logrus.New()}

	if len(configFilePath) > 0 {
		if compsConfigFile, err := ioutil.ReadFile(configFilePath); err != nil {
			log.Fatalln("wrong config path:", err)
		} else if err = sprbox.Unmarshal(compsConfigFile, &config); err != nil {
			log.Fatalln("can't unmarshal config file:", err)
		}
	}

	log.setup(config)
	return log
}

func (l *Logger) SpareConfig(configFiles []string) (err error) {
	l.Logger = logrus.New()

	var config *Config
	if err = sprbox.LoadConfig(&config, configFiles...); err != nil {
		return err
	}

	//if err = sprbox.LoadConfig(&config.Postgres, configFiles...); err != nil {
	//	return err
	//}
	if err = l.setup(config); err != nil {
		return err
	}
	return
}

func (l *Logger) setup(config *Config) error {
	l.Out = os.Stdout

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.DebugLevel
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

//// prende una funzione come parametro e ritorna il nome, c'è anche su recovery.go
//func (l *Logger) GetFunctionName(function interface{}) string {
//	return runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
//}
//
//func (l *Logger) addStackTraceIfNeeded(skip int, args []interface{}) []interface{} {
//	args = append(args, 0)
//	copy(args[1:], args[0:])
//	args[0] = getStackTrace(skip)
//	return append(args, "\n")
//}
//
//func getStackTrace(skip int) string {
//
//	skip++ // skip itself
//	stackTraceSlice := make([]string, 0)
//
//	for i := skip + 32; i >= skip; i-- {
//		if _, file, line, ok := runtime.Caller(i); ok {
//			//relFilePath, _ := filepath.Rel(Shared.App.LocalNode().Root, file)
//			//stackTraceSlice = append(stackTraceSlice, fmt.Sprintf("%s:%d", relFilePath, line))
//			stackTraceSlice = append(stackTraceSlice, fmt.Sprintf("%s:%d", file, line))
//		}
//	}
//
//	return "\n" + strings.Join(stackTraceSlice, " ->\n") + " ->"
//}
