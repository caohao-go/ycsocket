package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"server_golang/common/types"
)

const (
	LevelError  = "ERROR"
	LevelWarn   = "WARN"
	LevelNotice = "NOTICE"
	LevelInfo   = "INFO"
	LevelDebug  = "DEBUG"
)

var (
	logPath  string
	logLevel int
	zoneID   int
	mu       sync.Mutex
)

func Init(path string, level int, zone int) {
	logPath = path
	logLevel = level
	zoneID = zone
	os.MkdirAll(logPath, 0777)
}

func Error(msg string, fileNames ...string) {
	fileName := "trace"
	if len(fileNames) > 0 {
		fileName = fileNames[0]
	}
	writeLog(LevelError, msg, fileName)
}

func Warn(msg string, fileNames ...string) {
	fileName := "trace"
	if len(fileNames) > 0 {
		fileName = fileNames[0]
	}
	writeLog(LevelWarn, msg, fileName)
}

func Info(msg string, fileNames ...string) {
	fileName := "trace"
	if len(fileNames) > 0 {
		fileName = fileNames[0]
	}
	writeLog(LevelInfo, msg, fileName)
}

func Debug(msg string, fileNames ...string) {
	fileName := "trace"
	if len(fileNames) > 0 {
		fileName = fileNames[0]
	}
	writeLog(LevelDebug, msg, fileName)
}

func Notice(msg string, fileNames ...string) {
	fileName := "trace"
	if len(fileNames) > 0 {
		fileName = fileNames[0]
	}
	writeLog(LevelNotice, msg, fileName)
}

func writeLog(level, msg, fileName string) {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	dateStr := now.Format("2006-01-02 15:04:05")
	logFileName := filepath.Join(logPath, fmt.Sprintf("%s.%s.log", fileName, now.Format("20060102")))

	callerInfo := getCallerInfo()

	var logLine string
	switch level {
	case LevelInfo:
		logLine = fmt.Sprintf("[%s] [%s] [%d] - \"%s\"\n", level, dateStr, zoneID, msg)
		appendToFile(logFileName, logLine)
	case LevelDebug:
		logLine = fmt.Sprintf("[%s] [%s] [%d] [%s] \"%s\"\n", level, dateStr, zoneID, callerInfo, msg)
		appendToFile(logFileName+".debug", logLine)
	default:
		logLine = fmt.Sprintf("[%s] [%s] [%d] [%s] \"%s\"\n", level, dateStr, zoneID, callerInfo, msg)
		appendToFile(logFileName+".wf", logLine)
	}
}

func getCallerInfo() string {
	var parts []string
	for i := 3; i < 10; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		parts = append(parts, fmt.Sprintf("%s|%d", filepath.Base(file), line))
	}
	return strings.Join(parts, " => ")
}

func appendToFile(path, content string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", path, err)
		return
	}
	defer f.Close()
	f.WriteString(content)
}

// InstanceLogger 带上下文的请求级日志器
type InstanceLogger struct {
	FileName string
	IP       string
	Params   types.Map
}

func NewInstanceLogger(fileName string) *InstanceLogger {
	return &InstanceLogger{FileName: fileName}
}

func (l *InstanceLogger) SetClientIP(ip string) {
	l.IP = ip
}

func (l *InstanceLogger) SetParams(params types.Map) {
	l.Params = params
}

func (l *InstanceLogger) LogInfo(msg string) {
	Info(msg, l.FileName)
}

func (l *InstanceLogger) LogError(msg string) {
	Error(msg, l.FileName)
}

func (l *InstanceLogger) LogDebug(msg string) {
	Debug(msg, l.FileName)
}
