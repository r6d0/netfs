package logger

import (
	"log"
	"os"
)

type LogLevel uint8

const (
	Info LogLevel = iota
	Error
)

type LoggerConfig struct {
	Level LogLevel
}

type Logger struct {
	level  LogLevel
	logger *log.Logger
}

func (lg *Logger) Info(message string, args ...any) {
	if lg.level == Info {
		lg.logger.Printf("[INFO] "+message, args...)
	}
}

func (lg *Logger) Error(message string, args ...any) {
	if lg.level == Error {
		lg.logger.Printf("[ERROR] "+message, args...)
	}
}

func NewLogger(config LoggerConfig) *Logger {
	lg := Logger{level: config.Level, logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)}
	return &lg
}
