package logger

import (
	"github.com/sirupsen/logrus"
)

var ltype logType
var Logger *logrus.Logger

func Panic(message string) {
	Logger = getLogger()
	Logger.Panicf("{{level,%s},{levelNo,%s},{message:%s},{clientIP,%15s}",
		logrus.PanicLevel.MarshalText(), logrus.PanicLevel, message)
}

func Fatal(message string) {
	Logger = getLogger()
	Logger.Fatalf("{{level,%s},{levelNo,%s},{message:%s},{clientIP,%15s}",
		logrus.FatalLevel.MarshalText(), logrus.FatalLevel, message)
}

func Error(message string) {
	Logger = getLogger()
	Logger.Errorf("{{level,%s},{levelNo,%s},{message:%s},{clientIP,%15s}",
		logrus.ErrorLevel.MarshalText(), logrus.ErrorLevel, message)
}

func Warning(message string) {
	Logger = getLogger()
	Logger.Warnf("{{level,%s},{levelNo,%s},{message:%s},{clientIP,%15s}",
		logrus.WarnLevel.MarshalText(), logrus.WarnLevel, message)
}

func Info(message string) {
	Logger = getLogger()
	Logger.Infof("{{level,%s},{levelNo,%s},{message:%s},{clientIP,%15s}",
		logrus.InfoLevel.MarshalText(), logrus.InfoLevel, message)
}

func Debug(message string) {
	Logger = getLogger()
	Logger.Debugf("{{level,%s},{levelNo,%s},{message:%s},{clientIP,%15s}",
		logrus.DebugLevel.MarshalText(), logrus.DebugLevel, message)
}

func Trace(message string) {
	Logger = getLogger()
	Logger.Tracef("{{level,%s},{levelNo,%s},{message:%s},{clientIP,%15s}",
		logrus.TraceLevel.MarshalText(), logrus.TraceLevel, message)
}

func getLogger() *logrus.Logger {
	if ltype == Config.LType && Logger != nil {
		return Logger
	}
	switch Config.LType {
	case CONSOLE:
		Logger = getConsoleLogger()
	case MYSQL:
		Logger = getMysqlLogger()
	case FILE:
		Logger = GetFileLog()
	}

	return Logger
}
