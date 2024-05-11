package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Logger
// Интерфейс для работы с логгерами
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	SetLevel(level string) error
	SetFormatter(formatter string) error
}

// New
// Конструктор нового логгера
func New() Logger {
	var logger Logger

	logger = &LogrusLogger{
		log: logrus.New(),
	}

	return logger
}

// LogrusLogger
// Логгер использующий библиотеку logrus
type LogrusLogger struct {
	log *logrus.Logger
}

func (l *LogrusLogger) SetLevel(level string) error {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	l.log.SetLevel(logLevel)
	return nil
}

func (l *LogrusLogger) SetFormatter(formatter string) error {
	switch formatter {
	case "plain":
		l.log.SetFormatter(&logrus.TextFormatter{})
	case "json":
		l.log.SetFormatter(&logrus.JSONFormatter{})
	default:
		return fmt.Errorf("unknown log formatter: %v", formatter)
	}
	return nil
}

func (l *LogrusLogger) Debug(args ...interface{}) {
	l.log.Debug(args)
}

func (l *LogrusLogger) Info(args ...interface{}) {
	l.log.Info(args)
}

func (l *LogrusLogger) Warn(args ...interface{}) {
	l.log.Warn(args)
}

func (l *LogrusLogger) Error(args ...interface{}) {
	l.log.Error(args)
}

func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.log.Fatal(args)
}

func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.log.Debugf(format, args)
}

func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.log.Infof(format, args)
}

func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.log.Warnf(format, args)
}

func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.log.Errorf(format, args)
}

func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.log.Fatalf(format, args)
}

func Middleware(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			erw := &ExtendedResponseWriter{
				ResponseWriter: w,
			}
			next.ServeHTTP(erw, r)
			logger.Infof("Request: %v %v, status: %v", r.Method, r.URL.Path, erw.StatusCode)

		})
	}
}

type ExtendedResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *ExtendedResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}
