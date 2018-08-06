package logrus

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"runtime"
)

const (
	FMT_DIR_NAME  = "20060102"            // 表示目录的名称格式
	FMT_FILE_NAME = "2006010215"          // 表示文件的名称格式
	FMT_LOG_DATE  = "2006-01-02 15:04:05" // 每条日志的名称格式
)

const Record logrus.Level = 100

type LOG_DEST int // 日志打印位置

const (
	LOG_DEST_CONSOLE LOG_DEST = iota + 1
	LOG_DEST_FILE
)

type LOG_CHANGE_MODE int // 整点变换日志文件的方式

const (
	LOG_CHANGE_MODE_SELF  LOG_CHANGE_MODE = iota + 1 // 通过自身检测时间变化来改变日志文件
	LOG_CHANGE_MODE_OTHER                            // 通过外部程序
)

type Logger struct {
	logrus         *logrus.Entry
	logLevel       logrus.Level
	logDir         string
	logDest        LOG_DEST
	curFile        map[logrus.Level]*os.File // 表示当前打开的各个级别的日志文件
	curPath        map[logrus.Level]string   // 表示当前打开的各个级别的日志文件名
	recordsFile    map[string]*os.File
	recordsPath    map[string]string // 表示当前打开的各个级别的日志文件名
	logMutex       *sync.Mutex
	changeFileMode LOG_CHANGE_MODE
}

func NewLogger(pLogLevel, pLogDest, pLogDir string) *Logger {

	logLevel := logrus.DebugLevel

	switch pLogLevel {
	case "debug":
		logLevel = logrus.DebugLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	case "fatal":
		logLevel = logrus.FatalLevel
	case "panic":
		logLevel = logrus.PanicLevel
	default:
		panic("unrecongnised log level: " + pLogLevel)
	}

	logrus.SetLevel(logLevel)

	logDest := LOG_DEST_CONSOLE

	switch pLogDest {
	case "console":
		logDest = LOG_DEST_CONSOLE
	case "file":
		logDest = LOG_DEST_FILE
	default:
		panic("unrecongnised log dest: " + pLogDest)
	}

	logger := &Logger{
		logrus:         logrus.WithFields(logrus.Fields{}),
		logDir:         pLogDir,
		logLevel:       logLevel,
		logDest:        logDest,
		curFile:        make(map[logrus.Level]*os.File),
		curPath:        make(map[logrus.Level]string),
		recordsFile:    make(map[string]*os.File),
		recordsPath:    make(map[string]string),
		logMutex:       new(sync.Mutex),
		changeFileMode: LOG_CHANGE_MODE_SELF,
	}

	logrus.SetFormatter(&MultiLineFormatter{})

	return logger
}
func (this *Logger) Reload(pLogLevel, pLogDest, pLogDir string) {

	this.logMutex.Lock()
	defer this.logMutex.Unlock()

	logLevel := logrus.DebugLevel

	switch pLogLevel {
	case "debug":
		logLevel = logrus.DebugLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	case "fatal":
		logLevel = logrus.FatalLevel
	case "panic":
		logLevel = logrus.PanicLevel
	}

	logrus.SetLevel(logLevel)

	logDest := LOG_DEST_CONSOLE

	switch pLogDest {
	case "console":
		logDest = LOG_DEST_CONSOLE
	case "file":
		logDest = LOG_DEST_FILE
	}

	this.logDir = pLogDir
	this.logLevel = logLevel
	this.logDest = logDest

	this.curFile = make(map[logrus.Level]*os.File)
	this.curPath = make(map[logrus.Level]string)

}
func (this *Logger) Stop() {

	this.logMutex.Lock()
	defer this.logMutex.Unlock()

	for _, file := range this.curFile {
		if file != nil {
			file.Close()
		}
	}
}
func (this *Logger) ChangeLogFileJustHour() { // 用于整点时变换日志文件

	if this.changeFileMode == LOG_CHANGE_MODE_SELF { // 通过自身来改变日志文件目录
		return
	}

	this.logMutex.Lock()
	defer this.logMutex.Unlock()

	for level, file := range this.curFile { // 只变更已有的文件

		if file != nil {
			file.Close()
		}

		newPath := this.GetLogFilePath(this.logDir, level, "")

		newFile := this.GetLogFile(newPath)

		this.curFile[level] = newFile
		this.curPath[level] = newPath
	}

}
func (this *Logger) GetLogFile(path string) *os.File { // 获取路径对应的文件

	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {

		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {

			fmt.Println("GetLogFile: mkdirall error --> " + err.Error())

		}

	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)

	if err != nil {

		fmt.Println("GetLogFile: open file error --> " + err.Error())

	}

	return file
}
func (this *Logger) GetLogFilePath(rootDir string, logrusLevel logrus.Level, filename string) string { // 获取当前时间和对应日志类型应该使用的日志文件名

	t := time.Now()

	dirName := t.Format(FMT_DIR_NAME) // 目录名

	if logrusLevel == Record {

		fileName := fmt.Sprintf("%s.log", filename) // 按日志类型分类日志文件

		return filepath.Join(rootDir, filename, dirName, fileName)

	}

	logType := "info"
	switch logrusLevel {
	case logrus.DebugLevel, logrus.InfoLevel:
		logType = "info"
	case logrus.WarnLevel:
		logType = "warn"
	case logrus.ErrorLevel:
		logType = "error"
	case logrus.FatalLevel:
		logType = "fatal"
	case logrus.PanicLevel:
		logType = "panic"
	}

	fileName := fmt.Sprintf("%s_%s.log", logType, t.Format(FMT_FILE_NAME)) // 按日志类型分类日志文件

	return filepath.Join(rootDir, logType, dirName, fileName) // 根路径、日志类型、日期目录、日志文件
}

//*********************************************************************************
func (this *Logger) Debugf(format string, err ...interface{}) {
	this.WriteToLog(logrus.DebugLevel, true, "", format, err)
}
func (this *Logger) Infof(format string, err ...interface{}) {
	this.WriteToLog(logrus.InfoLevel, true, "", format, err)
}
func (this *Logger) Warnf(format string, err ...interface{}) {
	this.WriteToLog(logrus.WarnLevel, true, "", format, err)
}
func (this *Logger) Errorf(format string, err ...interface{}) {
	this.WriteToLog(logrus.ErrorLevel, true, "", format, err)
}
func (this *Logger) Fatalf(format string, err ...interface{}) {
	this.WriteToLog(logrus.FatalLevel, true, "", format, err)
}
func (this *Logger) Panicf(format string, err ...interface{}) {
	this.WriteToLog(logrus.PanicLevel, true, "", format, err)
}
func (this *Logger) Recordf(filename, format string, err ...interface{}) {
	this.WriteToLog(Record, true, filename, format, err)
}

func (this *Logger) Debug(err ...interface{}) {

	this.WriteToLog(logrus.DebugLevel, false, "", "", err)
}
func (this *Logger) Info(err ...interface{}) {
	this.WriteToLog(logrus.InfoLevel, false, "", "", err)
}
func (this *Logger) Warn(err ...interface{}) {
	this.WriteToLog(logrus.WarnLevel, false, "", "", err)
}
func (this *Logger) Error(err ...interface{}) {
	this.WriteToLog(logrus.ErrorLevel, false, "", "", err)
}
func (this *Logger) Fatal(err ...interface{}) {
	this.WriteToLog(logrus.FatalLevel, false, "", "", err)
}
func (this *Logger) Panic(err ...interface{}) {
	this.WriteToLog(logrus.PanicLevel, false, "", "", err)
}
func (this *Logger) Record(filename string, err ...interface{}) {
	this.WriteToLog(Record, false, filename, "", err)
}

//*********************************************************************************
func (this *Logger) WriteToLog(logLevel logrus.Level, append bool, filename, format string, contents ...interface{}) {

	this.logMutex.Lock()
	defer this.logMutex.Unlock()

	if logLevel == Record {

		filepath := this.GetLogFilePath(this.logDir, Record, filename)

		if path, exist := this.recordsPath[filename]; !exist || path != filepath {

			if this.recordsFile[filename] != nil {
				this.recordsFile[filename].Close()
			}

			this.recordsFile[filename] = this.GetLogFile(filepath)
			this.recordsPath[filename] = filepath

		}

		logrus.SetOutput(this.recordsFile[filename])

		if append {
			logrus.Infof(format, contents)
		} else {
			logrus.Info(contents)
		}

		return
	}

	if this.logLevel < logLevel {
		return
	}

	if this.logDest == LOG_DEST_CONSOLE { // 将日志打到控制台上

		logrus.SetOutput(os.Stdout)

	} else if this.logDest == LOG_DEST_FILE { // 将日志打到文件中

		filepath := this.GetLogFilePath(this.logDir, logLevel, "") // 获取当前应该使用的日志文件路径

		if this.changeFileMode == LOG_CHANGE_MODE_SELF {

			if path, exist := this.curPath[logLevel]; !exist || filepath != path {

				this.ChangeLogFileAndPath(filepath, logLevel)

			}

		} else {

			if _, exist := this.curFile[logLevel]; !exist {

				this.ChangeLogFileAndPath(filepath, logLevel)

			}
		}

		logrus.SetOutput(this.curFile[logLevel])
	}

	line := this.GetCallLineNumber()

	prefix := fmt.Sprintf("%04d -- ", line)

	switch logLevel {
	case logrus.DebugLevel:
		if append {
			this.logrus.Debugf(prefix+format, contents)
		} else {
			this.logrus.Debug(prefix, contents)
		}
	case logrus.InfoLevel:
		if append {
			this.logrus.Infof(prefix+format, contents)
		} else {
			this.logrus.Info(prefix, contents)
		}
	case logrus.WarnLevel:
		if append {
			this.logrus.Warnf(prefix+format, contents)
		} else {
			this.logrus.Warn(prefix, contents)
		}
	case logrus.ErrorLevel:
		if append {
			this.logrus.Errorf(prefix+format, contents)
		} else {
			this.logrus.Error(prefix, contents)
		}
	case logrus.FatalLevel:
		if append {
			this.logrus.Fatalf(prefix+format, contents)
		} else {
			this.logrus.Fatal(prefix, contents)
		}
	case logrus.PanicLevel:
		if append {
			this.logrus.Panicf(prefix+format, contents)
		} else {
			this.logrus.Panic(prefix, contents)
		}
	}
}
func (this *Logger) ChangeLogFileAndPath(newPath string, logLevel logrus.Level) {

	oldFile := this.curFile[logLevel]

	if oldFile != nil {
		oldFile.Close()
	}

	newFile := this.GetLogFile(newPath)

	if logLevel == logrus.DebugLevel || logLevel == logrus.InfoLevel { // 把这两个级别的日志打到一个文件中以便于观察

		this.curFile[logrus.DebugLevel] = newFile
		this.curPath[logrus.DebugLevel] = newPath

		this.curFile[logrus.InfoLevel] = newFile
		this.curPath[logrus.InfoLevel] = newPath

	} else {

		this.curFile[logLevel] = newFile
		this.curPath[logLevel] = newPath

	}

}
func (this *Logger) GetCallLineNumber() int {

	_, _, line, _ := runtime.Caller(3)

	// f := runtime.FuncForPC(pc)

	return line // , f.Name()
}
