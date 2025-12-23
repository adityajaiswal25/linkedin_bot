package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Logger represents a structured logger
type Logger struct {
	level  Level
	format Format
	writer io.Writer
	logger *log.Logger
}

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Format int

const (
	TextFormat Format = iota
	JSONFormat
)

var globalLogger *Logger

// NewLogger creates a new logger instance and sets it as global
func NewLogger(levelStr, formatStr, outputStr string) (*Logger, error) {
	level, err := parseLevel(levelStr)
	if err != nil {
		return nil, err
	}

	format, err := parseFormat(formatStr)
	if err != nil {
		return nil, err
	}

	writer, err := getWriter(outputStr)
	if err != nil {
		return nil, err
	}

	l := &Logger{
		level:  level,
		format: format,
		writer: writer,
		logger: log.New(writer, "", 0),
	}

	globalLogger = l
	return l, nil
}

// GetLogger returns the global logger
func GetLogger() *Logger {
	return globalLogger
}

// Close closes the logger if it has a file writer
func (l *Logger) Close() error {
	if closer, ok := l.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.log(DebugLevel, message, fields)
}

// Info logs an info message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.log(InfoLevel, message, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.log(WarnLevel, message, fields)
}

// Error logs an error message
func (l *Logger) Error(message string, fields map[string]interface{}) {
	l.log(ErrorLevel, message, fields)
}

// Package-level helpers for convenience
func Debug(message string, fields map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(message, fields)
	}
}

func Info(message string, fields map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Info(message, fields)
	}
}

func Warn(message string, fields map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(message, fields)
	}
}

func Error(message string, fields map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Error(message, fields)
	}
}

func (l *Logger) log(level Level, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Level:   level.String(),
		Message: message,
		Fields:  fields,
	}

	var output string
	switch l.format {
	case JSONFormat:
		data, _ := json.Marshal(entry)
		output = string(data)
	case TextFormat:
		output = fmt.Sprintf("[%s] %s: %s", entry.Time, entry.Level, entry.Message)
		if len(fields) > 0 {
			fieldStrs := make([]string, 0, len(fields))
			for k, v := range fields {
				fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", k, v))
			}
			output += " " + strings.Join(fieldStrs, " ")
		}
	}

	l.logger.Println(output)
}

type LogEntry struct {
	Time    string                 `json:"time"`
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func parseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	default:
		return InfoLevel, fmt.Errorf("invalid log level: %s", s)
	}
}

func parseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "json":
		return JSONFormat, nil
	case "text":
		return TextFormat, nil
	default:
		return TextFormat, fmt.Errorf("invalid log format: %s", s)
	}
}

func getWriter(output string) (io.Writer, error) {
	if output == "stdout" {
		return os.Stdout, nil
	}

	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}
