package logging

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/sirupsen/logrus"
)

type Fields = logrus.Fields

type FieldLogger interface {
	logrus.FieldLogger
}

var (
	log  logrus.FieldLogger
	once sync.Once
)

// Init initialize global instance of logging library.
func Init(fields Fields) FieldLogger {
	once.Do(func() {
		logger := logrus.New()

		logger.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})

		logger.SetLevel(logrus.InfoLevel)

		log = logger.WithFields(fields)
	})
	return log
}

type ctxKey int

const (
	loggerKey ctxKey = iota
)

func AddToContext(ctx context.Context, log FieldLogger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

func FromContext(ctx context.Context) FieldLogger {
	log, ok := ctx.Value(loggerKey).(FieldLogger)
	if ok {
		return log
	}
	return log
}

func WithFields(fields Fields) FieldLogger {
	return log.WithFields(fields)
}

func WithField(key string, value any) FieldLogger {
	return log.WithField(key, value)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Warning(args ...interface{}) {
	log.Warning(args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func LogCommand(command interface{}) {
	context, _ := json.Marshal(command)

	commandName := reflect.TypeOf(command).Elem().Name()

	var commandAsFields logrus.Fields
	json.Unmarshal(context, &commandAsFields)

	log.
		WithFields(logrus.Fields{
			"command_name": commandName,
			"command":      commandAsFields,
		}).Info("Start command " + commandName)
}

func LogCommandFailed(command interface{}, err error) {
	commandName := reflect.TypeOf(command).Elem().Name()
	log.
		WithFields(logrus.Fields{
			"reason": err.Error(),
		}).Info("Command failed" + commandName)
}
