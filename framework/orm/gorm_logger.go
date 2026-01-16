package orm

import (
	"context"
	"errors"
	"time"

	"github.com/feimumoke/labequipbms/framework/config"
	"github.com/feimumoke/labequipbms/framework/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// GormLogger GORM 自定义日志器，集成到我们的日志系统
type GormLogger struct {
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	Colorful                  bool
}

// NewGormLogger 创建新的 GORM 日志器
func NewGormLogger() *GormLogger {
	return &GormLogger{
		LogLevel:                  logger.Info,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: false,
		Colorful:                  false, // 文件日志不需要颜色
	}
}

// LogMode 设置日志级别
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info 输出 Info 级别日志
func (l GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		log.CtxInfof(ctx, msg, data...)
	}
}

// Warn 输出 Warn 级别日志
func (l GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		log.CtxWarnf(ctx, msg, data...)
	}
}

// Error 输出 Error 级别日志
func (l GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		log.CtxErrorf(ctx, msg, data...)
	}
}

// Trace 输出 SQL 跟踪日志
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 获取调用位置
	fileWithLine := utils.FileWithLineNum()

	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		// 错误日志
		log.CtxErrorf(ctx, "[GORM] %s | %.3fms | rows:%d | %s | error:%v\n",
			fileWithLine,
			float64(elapsed.Nanoseconds())/1e6,
			rows,
			sql,
			err,
		)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		// 慢查询警告
		log.CtxWarnf(ctx, "[GORM] SLOW SQL >= %v | %s | %.3fms | rows:%d | %s\n",
			l.SlowThreshold,
			fileWithLine,
			float64(elapsed.Nanoseconds())/1e6,
			rows,
			sql,
		)
	case l.LogLevel == logger.Info:
		// 普通 SQL 日志
		log.CtxInfof(ctx, "[GORM] %s | %.3fms | rows:%d | %s\n",
			fileWithLine,
			float64(elapsed.Nanoseconds())/1e6,
			rows,
			sql,
		)
	}
}

// LogModeFromString 从字符串解析日志级别
func LogModeFromString(mode string) logger.LogLevel {
	switch mode {
	case "silent", "SILENT":
		return logger.Silent
	case "error", "ERROR":
		return logger.Error
	case "warn", "WARN":
		return logger.Warn
	case "info", "INFO":
		return logger.Info
	default:
		return logger.Info
	}
}

// GormLogLevelOption GORM 日志级别配置选项
type GormLogLevelOption struct{}

// Apply 应用日志级别配置
func (o *GormLogLevelOption) Apply(cfg *Config) error {
	if cfg.Logger == nil {
		gormLogger := NewGormLogger()

		// 从配置中读取 GORM 日志级别
		if config.WCConfig != nil && config.WCConfig.Log != nil {
			level := config.WCConfig.Log.GormLogLevel
			if level == "" {
				level = "INFO" // 默认级别
			}
			gormLogger.LogLevel = LogModeFromString(level)
			log.Infof("GORM log level set to: %s\n", level)
		}

		cfg.Logger = gormLogger
	}
	return nil
}
