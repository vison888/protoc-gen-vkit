package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DebugLevel  = Level(1)
	InfoLevel   = Level(2)
	WarnLevel   = Level(3)
	ErrorLevel  = Level(4)
	FileMaxLine = 65536
)

// Level 日志输出级别。
type Level int32

func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "[debug]"
	case InfoLevel:
		return "[info]"
	case WarnLevel:
		return "[warn]"
	case ErrorLevel:
		return "[error]"
	default:
		return "[none]"
	}
}

var (
	logDir string

	mutex     sync.Mutex
	logFile   *os.File
	outputBuf []byte
	lineCount int32
)

func init() {
	logDir = "./logs"
	tryNewFile(true)
}

//logs/appname/podname.time.log
func tryNewFile(force bool) {
	if lineCount > FileMaxLine || force {
		// builf file path
		timeStr := time.Now().Format("2006-01-02-15:04:05")
		fileDir := logDir
		filePath := fmt.Sprintf("%s/%s.log", fileDir, timeStr)
		//try create dir
		_, err := os.Stat(fileDir)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(fileDir, os.ModePerm)
				if err != nil {
					fmt.Println("create forder fail fileDir=:" + fileDir)
					return
				}
			}
		}
		// new file
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println("open log file failed, err:", err)
			return
		}
		lineCount = 0
		if logFile != nil {
			logFile.Close()
		}
		logFile = file
	}
}

func formatAndWrite(l Level, format string, v ...interface{}) {
	now := time.Now()
	mutex.Lock()
	defer mutex.Unlock()

	outputBuf = outputBuf[:0]
	formatHeader(&outputBuf, l, now)
	s := fmt.Sprintf(format, v...)
	outputBuf = append(outputBuf, s...)
	outputBuf = append(outputBuf, '\n')
	logFile.Write(outputBuf)
	lineCount++
	tryNewFile(false)
}

//[level][time][NODE_NAME][POD_NAME][APP_NAME] msg
func formatHeader(buf *[]byte, l Level, t time.Time) {
	*buf = append(*buf, l.String()...)
	timeStr := t.Format("[2006-01-02 15:04:05.000000]")
	*buf = append(*buf, timeStr...)
}

func Infof(format string, v ...interface{}) {
	formatAndWrite(InfoLevel, format, v...)
}

func Warnf(format string, v ...interface{}) {
	formatAndWrite(WarnLevel, format, v...)
}

func Errorf(format string, v ...interface{}) {
	formatAndWrite(ErrorLevel, format, v...)
}

func Debugf(format string, v ...interface{}) {
	formatAndWrite(DebugLevel, format, v...)
}

func Info(format string, v ...interface{}) {
	Infof(format, v...)
}

func Warn(format string, v ...interface{}) {
	Warnf(format, v...)
}

func Error(format string, v ...interface{}) {
	Errorf(format, v...)
}

func Debug(format string, v ...interface{}) {
	Debugf(format, v...)
}

func CanServerLog(xct string) bool {
	return !strings.Contains(xct, "multipart/form-data")
}
