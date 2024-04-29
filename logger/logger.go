package logger

import "github.com/sirupsen/logrus"

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
}

func New(logLevelString string) (Logger, error) {
	var logger Logger

	logLevel, err := logrus.ParseLevel(logLevelString)
	if err != nil {
		return nil, err
	}

	ll := logrus.New()
	ll.Level = logLevel

	logger = ll

	return logger, nil
}

type LogrusLogger struct {
	log *logrus.Logger
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
