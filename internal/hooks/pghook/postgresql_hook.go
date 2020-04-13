package pghook

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type filter func(*logrus.Entry) *logrus.Entry

// Sync Hook -----------------------------------------------------------------------------------------------------------

// Hook to send logs to a PostgreSQL database
type Hook struct {
	Extra      map[string]interface{}
	db         *sql.DB
	mu         sync.RWMutex
	InsertFunc func(*sql.DB, *logrus.Entry) error `json:"-"`
	filters    []filter
}

// Levels return the available logging levels.
func (hook *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *Hook) Fire(entry *logrus.Entry) error {
	newEntry := hook.newEntry(entry)
	if newEntry == nil {
		// entry is ignored.
		return nil
	}
	return hook.InsertFunc(hook.db, newEntry)
}

//func (hook *Hook) Close() error {
//	return hook.db.Close()
//}

//AddFilter adds filter that can modify or ignore entry.
func (hook *Hook) AddFilter(fn filter) {
	hook.filters = append(hook.filters, fn)
}

// NewHook creates a PGHook to be added to an instance of logger.
func NewHook(db *sql.DB) *Hook {
	return &Hook{
		db: db,
		InsertFunc: func(db *sql.DB, entry *logrus.Entry) error {
			jsonData, err := json.Marshal(entry.Data)
			if err != nil {
				return err
			}

			_, err = db.Exec("INSERT INTO logs(level, message, message_data, created_at) VALUES ($1,$2,$3,$4);",
				entry.Level, entry.Message, jsonData, entry.Time)
			if err != nil {
				return fmt.Errorf("postgres: %s", err.Error())
			}
			return nil
		},
	}
}

// Async Hook -----------------------------------------------------------------------------------------------------------

type AsyncHook struct {
	*Hook
	buf        chan *logrus.Entry
	flush      chan bool
	wg         sync.WaitGroup
	Ticker     *time.Ticker
	InsertFunc func(*sql.Tx, *logrus.Entry) error
}

// Levels return the available logging levels.
func (hook *AsyncHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire log an entry.
func (hook *AsyncHook) Fire(entry *logrus.Entry) error {
	if entry == nil {
		// entry is ignored.
		return nil
	}
	newEntry := hook.newEntry(entry)
	if newEntry == nil {
		// entry is ignored.
		return nil
	}
	hook.wg.Add(1)
	hook.buf <- newEntry
	return nil
}

//AddFilter adds filter that can modify or ignore entry.
func (hook *AsyncHook) AddFilter(fn filter) {
	hook.filters = append(hook.filters, fn)
}

// NewAsyncHook creates a hook to be added to an instance of logger.
// The hook created will be asynchronous, and it's the responsibility of the user to call the Flush method
// before exiting to empty the log queue.
// Once the buffer (bufferSize) is full, logging will start blocking,
// waiting for slots to be available in the queue.
// bufferSize st to 0 will use the default value (8192).
func NewAsyncHook(db *sql.DB, bufferSize uint) *AsyncHook {
	if bufferSize == 0 {
		bufferSize = 8192
	}

	hook := &AsyncHook{
		Hook:   NewHook(db),
		buf:    make(chan *logrus.Entry, bufferSize),
		flush:  make(chan bool),
		Ticker: time.NewTicker(time.Second),
		InsertFunc: func(tx *sql.Tx, entry *logrus.Entry) error {
			jsonData, err := json.Marshal(entry.Data)
			if err != nil {
				return err
			}

			_, err = tx.Exec("INSERT INTO logs(level, message, error, message_data, created_at) VALUES ($1,$2,$3,$4);",
				entry.Level, entry.Message, jsonData, entry.Time)
			if err != nil {
				return fmt.Errorf("postgres: %s", err.Error())
			}
			return nil
		},
	}

	go func() {
		for {
			var err error
			tx, err := hook.db.Begin()
			if err != nil {
				fmt.Println("[pglogrus] Can't create db transaction:", err)
				// Don't create new transactions too fast, it will flood stderr
				select {
				case <-hook.Ticker.C:
					continue
				}
			}

			var numEntries int
		Loop:
			for {
				select {
				case entry := <-hook.buf:
					err = hook.InsertFunc(tx, entry)
					if err != nil {
						fmt.Printf("[pglogrus] Can't insert entry (%v): %v\n", entry, err)
					}
					numEntries++

				case <-hook.Ticker.C:
					if numEntries > 0 {
						break Loop
					}

				case <-hook.flush:
					err = tx.Commit()
					if err != nil {
						fmt.Println("[pglogrus] Can't commit transaction:", err)
					}
					hook.flush <- true
					return
				}

			}

			err = tx.Commit()
			if err != nil {
				fmt.Println("[pglogrus] Can't commit transaction:", err)
			}

			for i := 0; i < numEntries; i++ {
				hook.wg.Done()
			}
		}
	}()

	return hook
}

// Flush waits for the log queue to be empty.
// This func is meant to be used when the hook was created with NewAsyncHook.
func (hook *AsyncHook) Flush() {
	hook.Ticker = time.NewTicker(100 * time.Millisecond)
	hook.wg.Wait()

	hook.flush <- true
	<-hook.flush
}

// newEntry will prepare a new logrus entry to be logged in the DB
// the extra fields are added to entry Data
func (hook *Hook) newEntry(entry *logrus.Entry) *logrus.Entry {
	hook.mu.RLock() // Claim the mutex as a RLock - allowing multiple go routines to log simultaneously
	defer hook.mu.RUnlock()

	// Don't modify entry.Data directly, as the entry will used after this hook was fired
	data := map[string]interface{}{}

	// Merge extra fields
	for k, v := range hook.Extra {
		data[k] = v
	}
	for k, v := range entry.Data {
		data[k] = v
		if k == logrus.ErrorKey {
			asError, isError := v.(error)
			_, isMarshaler := v.(json.Marshaler)
			if isError && !isMarshaler {
				data[k] = newMarshalableError(asError)
			}
		}
	}

	newEntry := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    data,
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
	}

	// Apply filters
	for _, fn := range hook.filters {
		newEntry = fn(newEntry)
		if newEntry == nil {
			break
		}
	}
	return newEntry
}

// MISC ----------------------------------------------------------------------------------------------------------------

//func fieldsFilter(fields []string) filter {
//	return func(entry *logrus.Entry) *logrus.Entry {
//		for _, field := range fields {
//			delete(entry.Data, field)
//		}
//		return entry
//	}
//}
