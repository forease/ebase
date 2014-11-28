/**
 * Ebase frame for daemon program
 * Author Jonsen Yang
 * Date 2013-07-05
 * Copyright (c) 2013 ForEase Times Technology Co., Ltd. All rights reserved.
 */
package ebase

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
)

// Log levels to control the logging output.
const (
	LevelCritical = iota
	LevelError
	LevelWarning
	LevelInfo
	LevelDebug
	LevelTrace
)

type BaseLog struct {
	Loger    *log.Logger
	LogFile  string
	LogLevel int
	LogType  string
	//lChan chan
}

type LogOptions struct {
	Type   string
	File   string
	Level  int
	Flag   int
	Enable bool
}

// 初始化日志
// 创建一个goroutine来专门记录日志
// 需要写日志时，通过chan传递
func NewLog(opt *LogOptions) (l *BaseLog) {
	var loger *log.Logger

	if opt.Level == -1 {
		opt.Level = 3
	}

	switch opt.Type {
	case "consloe":
		loger = log.New(os.Stdout, "", opt.Flag)
	case "file":
		out, err := os.OpenFile(opt.File, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			fmt.Println("Open log file error", err)
			os.Exit(1)
		}
		loger = log.New(out, "", opt.Flag)
	default:
		var err error
		loger, err = syslog.NewLogger(syslog.Priority(opt.Level), opt.Flag)
		if err != nil {
			fmt.Println("init syslog error ", err)
			os.Exit(1)
		}
	}

	l = &BaseLog{Loger: loger, LogFile: opt.File,
		LogType: opt.Type, LogLevel: opt.Level}

	return
}

func (l *BaseLog) Critical(v ...interface{}) {
	if l.LogLevel >= LevelCritical {
		l.Loger.Output(2, "[Crt]"+fmt.Sprintln(v...))
	}
}

func (l *BaseLog) Error(v ...interface{}) {
	if l.LogLevel >= LevelError {
		l.Loger.Output(2, "[Err] "+fmt.Sprintln(v...))
	}
}

func (l *BaseLog) Warn(v ...interface{}) {
	if l.LogLevel >= LevelWarning {
		l.Loger.Output(2, "[War] "+fmt.Sprintln(v...))
	}
}

func (l *BaseLog) Info(v ...interface{}) {
	if l.LogLevel >= LevelInfo {
		l.Loger.Output(2, "[Inf] "+fmt.Sprintln(v...))
	}
}

func (l *BaseLog) Debug(v ...interface{}) {
	if l.LogLevel >= LevelDebug {
		l.Loger.Output(2, "[Dbg] "+fmt.Sprintln(v...))
	}
}

func (l *BaseLog) Trace(v ...interface{}) {
	if l.LogLevel >= LevelTrace {
		l.Loger.Output(2, "[Trc] "+fmt.Sprintln(v...))
	}
}

func (l *BaseLog) Println(v ...interface{}) {
	l.Loger.Output(2, fmt.Sprintln(v...))
}

func (l *BaseLog) Panic(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.Loger.Output(2, s)
	panic(s)
}

func (l *BaseLog) Criticalf(format string, v ...interface{}) {
	if l.LogLevel >= LevelCritical {
		l.Loger.Output(2, "[Crt] "+fmt.Sprintf(format, v...))
	}
}

func (l *BaseLog) Errorf(format string, v ...interface{}) {
	if l.LogLevel >= LevelError {
		l.Loger.Output(2, "[Err] "+fmt.Sprintf(format, v...))
	}
}

func (l *BaseLog) Warnf(format string, v ...interface{}) {
	if l.LogLevel >= LevelWarning {
		l.Loger.Output(2, "[War] "+fmt.Sprintf(format, v...))
	}
}

func (l *BaseLog) Infof(format string, v ...interface{}) {
	if l.LogLevel >= LevelInfo {
		l.Loger.Output(2, "[Inf] "+fmt.Sprintf(format, v...))
	}
}

func (l *BaseLog) Debugf(format string, v ...interface{}) {
	if l.LogLevel >= LevelDebug {
		l.Loger.Output(2, "[Dbg] "+fmt.Sprintf(format, v...))
	}
}

func (l *BaseLog) Tracef(format string, v ...interface{}) {
	if l.LogLevel >= LevelTrace {
		l.Loger.Output(2, "[Trc] "+fmt.Sprintf(format, v...))
	}
}

func (l *BaseLog) Printf(format string, v ...interface{}) {
	l.Loger.Output(2, fmt.Sprintf(format, v...))
}

func (l *BaseLog) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Loger.Output(2, s)
	panic(s)
}
