package logger

import "log"

type Logger struct {
	Level int
}

const (
	LevelDebug = iota
	LevelInfo
	LevelWarning
	LevelError
)

func New(l int) *Logger {
	if l < LevelDebug {
		l = LevelDebug
	}
	return &Logger{Level: l}
}

func (l *Logger) Debug(s string, args ...any) {
	if l.Level <= LevelDebug {
		log.Printf("DEBUG - "+s, args...)
	}
}

func (l *Logger) Info(s string, args ...any) {
	if l.Level <= LevelInfo {
		log.Printf("INFO - "+s, args...)
	}
}

func (l *Logger) Warning(s string, args ...any) {
	if l.Level <= LevelWarning {
		log.Printf("WARNING - "+s, args...)
	}
}

func (l *Logger) Error(s string, args ...any) {
	if l.Level <= LevelError {
		log.Printf("ERROR - "+s, args...)
	}
}
