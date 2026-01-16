package initialize

import (
	"fmt"
	"strings"

	"github.com/feimumoke/labequipbms/framework/config"
	"github.com/feimumoke/labequipbms/framework/log"
)

// InitLogger 初始化日志系统
func InitLogger() error {
	logConfig := config.WCConfig.Log
	if logConfig == nil {
		// 如果没有配置，使用默认配置
		logConfig = &config.LogConfig{
			Dir:           "./logs",
			Level:         "INFO",
			EnableConsole: true,
		}
	}

	// 解析日志级别
	level := parseLogLevel(logConfig.Level)

	// 初始化日志系统
	err := log.InitLogger(log.LogConfig{
		LogDir:        logConfig.Dir,
		Level:         level,
		EnableConsole: logConfig.EnableConsole,
	})

	if err != nil {
		return fmt.Errorf("init logger failed: %v", err)
	}

	log.Infof("Logger initialized successfully, log dir: %s, level: %s\n", logConfig.Dir, logConfig.Level)
	return nil
}

// parseLogLevel 解析日志级别
func parseLogLevel(levelStr string) log.LogLevel {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return log.DEBUG
	case "INFO":
		return log.INFO
	case "WARN":
		return log.WARN
	case "ERROR":
		return log.ERROR
	case "FATAL":
		return log.FATAL
	default:
		return log.INFO
	}
}
