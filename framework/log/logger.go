package log

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/feimumoke/labequipbms/framework/support/trace"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger 日志接口
type Logger interface {
	Print(v ...interface{})
	CtxLogInfof(ctx context.Context, format string, v ...interface{})
	CtxLogDebugf(ctx context.Context, format string, v ...interface{})
	CtxLogErrorf(ctx context.Context, format string, v ...interface{})
	CtxLogFatalf(ctx context.Context, format string, v ...interface{})
}

// BMSLogger 日志实现
type BMSLogger struct {
	logDir         string
	currentLogFile *os.File
	logger         *log.Logger
	level          LogLevel
	enableConsole  bool
	mu             sync.RWMutex
	currentDate    string
}

var (
	globalLogger *BMSLogger
	once         sync.Once
)

// LogConfig 日志配置
type LogConfig struct {
	LogDir        string   // 日志目录
	Level         LogLevel // 日志级别
	EnableConsole bool     // 是否输出到控制台
}

// InitLogger 初始化日志系统
func InitLogger(config LogConfig) error {
	var err error
	once.Do(func() {
		globalLogger = &BMSLogger{
			logDir:        config.LogDir,
			level:         config.Level,
			enableConsole: config.EnableConsole,
		}
		err = globalLogger.init()
	})
	return err
}

// init 初始化日志文件
func (l *BMSLogger) init() error {
	// 创建日志目录
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return fmt.Errorf("create log dir failed: %v", err)
	}

	// 打开当前日志文件
	if err := l.rotateLogFile(); err != nil {
		return err
	}

	// 启动日志切割协程
	go l.watchLogRotate()

	return nil
}

// rotateLogFile 切换日志文件
func (l *BMSLogger) rotateLogFile() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// 如果是同一天，不需要切换
	if l.currentDate == dateStr && l.currentLogFile != nil {
		return nil
	}

	// 关闭旧的日志文件
	if l.currentLogFile != nil {
		l.currentLogFile.Close()
	}

	// 创建新的日志文件
	logFileName := filepath.Join(l.logDir, fmt.Sprintf("bms-%s.log", dateStr))
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file failed: %v", err)
	}

	l.currentLogFile = file
	l.currentDate = dateStr

	// 配置输出
	var writers []io.Writer
	writers = append(writers, file)
	if l.enableConsole {
		writers = append(writers, os.Stdout)
	}

	multiWriter := io.MultiWriter(writers...)
	l.logger = log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)

	return nil
}

// watchLogRotate 监控日志切割
func (l *BMSLogger) watchLogRotate() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		dateStr := now.Format("2006-01-02")

		// 如果日期变化，切换日志文件
		if l.currentDate != dateStr {
			if err := l.rotateLogFile(); err != nil {
				fmt.Printf("rotate log file failed: %v\n", err)
			}
		}

		// 压缩旧日志（保留最近7天的日志）
		l.cleanOldLogs()
	}
}

// cleanOldLogs 清理旧日志
func (l *BMSLogger) cleanOldLogs() {
	files, err := os.ReadDir(l.logDir)
	if err != nil {
		return
	}

	cutoffDate := time.Now().AddDate(0, 0, -7)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 检查是否是日志文件
		if filepath.Ext(file.Name()) != ".log" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// 删除超过7天的日志
		if info.ModTime().Before(cutoffDate) {
			logPath := filepath.Join(l.logDir, file.Name())
			os.Remove(logPath)
			Infof("Cleaned old log file: %s\n", logPath)
		}
	}
}

// logf 记录日志
func (l *BMSLogger) logf(level LogLevel, format string, v ...interface{}) {
	if l == nil || l.logger == nil {
		// 如果日志系统未初始化，使用标准输出
		fmt.Printf(format, v...)
		return
	}

	if level < l.level {
		return
	}

	l.mu.RLock()
	logger := l.logger
	l.mu.RUnlock()

	levelName := levelNames[level]
	msg := fmt.Sprintf(format, v...)

	traceID, _ := trace.GetTraceIDFromLocalMap()

	// 使用 Output 来跳过调用栈，显示正确的文件和行号
	logger.Output(3, fmt.Sprintf("%v [%s] %s", traceID, levelName, msg))

	// FATAL 级别直接退出
	if level == FATAL {
		os.Exit(1)
	}
}

// logfWithContext 带 context 的日志记录（自动提取 traceID）
func (l *BMSLogger) logfWithContext(ctx context.Context, level LogLevel, format string, v ...interface{}) {
	if l == nil || l.logger == nil {
		// 如果日志系统未初始化，使用标准输出
		fmt.Printf(format, v...)
		return
	}

	if level < l.level {
		return
	}

	l.mu.RLock()
	logger := l.logger
	l.mu.RUnlock()

	levelName := levelNames[level]
	msg := fmt.Sprintf(format, v...)

	// 从 context 中提取 traceID
	traceID := extractTraceID(ctx)
	if traceID != "" {
		msg = fmt.Sprintf("%s %s", traceID, msg)
	}

	// 使用 Output 来跳过调用栈，显示正确的文件和行号
	logger.Output(3, fmt.Sprintf("[%s] %s", levelName, msg))

	// FATAL 级别直接退出
	if level == FATAL {
		os.Exit(1)
	}
}

// extractTraceID 从 context 中提取 traceID
func extractTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// 尝试从 context 中获取 trace_id
	if traceID, ok := ctx.Value("logid").(string); ok {
		return traceID
	}

	return ""
}

// 全局日志函数

// Infof 记录 INFO 级别日志
func Infof(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logf(INFO, format, v...)
	} else {
		fmt.Printf(format, v...)
	}
}

// Debugf 记录 DEBUG 级别日志
func Debugf(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logf(DEBUG, format, v...)
	} else {
		fmt.Printf(format, v...)
	}
}

// Warnf 记录 WARN 级别日志
func Warnf(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logf(WARN, format, v...)
	} else {
		fmt.Printf(format, v...)
	}
}

// Errorf 记录 ERROR 级别日志
func Errorf(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logf(ERROR, format, v...)
	} else {
		fmt.Printf(format, v...)
	}
}

// Fatalf 记录 FATAL 级别日志并退出程序
func Fatalf(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logf(FATAL, format, v...)
	} else {
		fmt.Printf(format, v...)
		os.Exit(1)
	}
}

// 带 Context 的日志函数

// CtxInfof 带上下文的 INFO 日志（自动提取 traceID）
func CtxInfof(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logfWithContext(ctx, INFO, format, v...)
	} else {
		fmt.Printf(format, v...)
	}
}

// CtxDebugf 带上下文的 DEBUG 日志（自动提取 traceID）
func CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logfWithContext(ctx, DEBUG, format, v...)
	} else {
		fmt.Printf(format, v...)
	}
}

// CtxWarnf 带上下文的 WARN 日志（自动提取 traceID）
func CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logfWithContext(ctx, WARN, format, v...)
	} else {
		fmt.Printf(format, v...)
	}
}

// CtxErrorf 带上下文的 ERROR 日志（自动提取 traceID）
func CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logfWithContext(ctx, ERROR, format, v...)
	} else {
		fmt.Printf(format, v...)
	}
}

// CtxFatalf 带上下文的 FATAL 日志（自动提取 traceID）
func CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logfWithContext(ctx, FATAL, format, v...)
	} else {
		fmt.Printf(format, v...)
		os.Exit(1)
	}
}

// Close 关闭日志系统
func Close() {
	if globalLogger != nil && globalLogger.currentLogFile != nil {
		globalLogger.currentLogFile.Close()
	}
}
