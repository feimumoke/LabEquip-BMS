package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/feimumoke/labequipbms/apps/register"
	"github.com/feimumoke/labequipbms/framework/asynctask"
	"github.com/feimumoke/labequipbms/framework/config"
	"github.com/feimumoke/labequipbms/framework/crontask"
	"github.com/feimumoke/labequipbms/framework/initialize"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/web"
)

// main 程序入口
func main() {
	path, _ := os.Getwd()
	fmt.Println("Working directory:", path)

	configPath := filepath.Join(path, "server", "_config", "conf.yaml")
	jsonPath := filepath.Join(path, "server", "_config") + string(filepath.Separator)
	fmt.Println("Config path:", configPath)

	// 加载配置文件
	if err := config.DoInitWcConfigWithPath(configPath); err != nil {
		panic(fmt.Sprintf("Load config failed: %v", err))
	}

	// 初始化日志系统
	if err := initialize.InitLogger(); err != nil {
		panic(fmt.Sprintf("Init logger failed: %v", err))
	}

	// 使用日志系统记录启动信息
	log.Infof("=================================================\n")
	log.Infof("LabEquip-BMS Backend Service Starting...\n")
	log.Infof("=================================================\n")
	log.Infof("Working Directory: %s\n", path)
	log.Infof("Config Path: %s\n", configPath)

	// 初始化其他组件
	if wcError := initialize.Initialize(jsonPath); wcError != nil {
		log.Fatalf("Initialize failed: %v\n", wcError)
	}

	// 初始化任务和服务
	cornRunner := crontask.NewCornRunner()
	r := asynctask.InitAsyncRunnerInProcess(cornRunner)
	s := web.NewBasicServer(config.WCConfig.Server, "api")
	register.RegisterApiAndTask(s, r)
	r.ManualProcessing(s)

	log.Infof("=================================================\n")
	log.Infof("Initialize Success!\n")
	log.Infof("API Server listening on: %s\n", config.WCConfig.Server.Addr["api"])
	log.Infof("=================================================\n")

	// 启动服务
	go func() {
		if err := s.Run(); err != nil {
			log.Fatalf("Server run failed: %v\n", err)
		}
	}()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Infof("=================================================\n")
	log.Infof("Shutting down server...\n")
	log.Infof("=================================================\n")

	// 关闭日志系统
	log.Close()

	time.Sleep(time.Second * 2)
	log.Infof("Server exited\n")
}
