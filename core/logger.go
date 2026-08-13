package dge

import (
	"fmt"
	"io"
	"time"
)

type LogLevel string

const (
	LevelInfo    LogLevel = "INFO"
	LevelError   LogLevel = "ERROR"
	LevelWarning LogLevel = "WARNING"
)

type EvenType string

const (
	EventSuccess EvenType = "SUCCESS"
	EventStart   EvenType = "START"
	EventFailed  EvenType = "FAILED"
	EventSkipped EvenType = "SKIPPED"
)

type GraphLogger interface {
	FlushLog(et EvenType, logLv LogLevel, msg, vId, gId string)
}

type Log struct {
	fd io.Writer
}

func NewLogger(fd io.Writer) GraphLogger {
	return &Log{
		fd: fd,
	}
}

func (l *Log) FlushLog(et EvenType, logLv LogLevel, msg, vId, gId string) {
	log := fmt.Sprintf("[%s] [%s] [%s] [%s] [%s] %s\n", time.Now().Format("15:04:05"), logLv, gId, vId, et, msg)
	_, err := l.fd.Write([]byte(log))
	if err != nil {
		errWriteLog := fmt.Sprintf("[%s] [%s] [%s] [%s] [%s] %s\n", time.Now().Format("15:04:05"), LevelWarning, gId, vId, "", err.Error())
		fmt.Printf(errWriteLog)
	}
}
