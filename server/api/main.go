package main

import (
	"fmt"
	"os"
	"time"

	"github.com/feimumoke/labequipbms/apps/register"
	"github.com/feimumoke/labequipbms/framework/asynctask"
	"github.com/feimumoke/labequipbms/framework/config"
	"github.com/feimumoke/labequipbms/framework/crontask"
	"github.com/feimumoke/labequipbms/framework/initialize"
	"github.com/feimumoke/labequipbms/framework/libinit"
	"github.com/feimumoke/labequipbms/framework/web"
)

// toc
func main() {

	path, _ := os.Getwd()
	fmt.Println(path)

	configPath := fmt.Sprintf("%s/server/_config/conf.yaml", path)
	jsonPath := fmt.Sprintf("%s/server/_config/", path)
	fmt.Println("configPath: " + configPath)

	if err := config.DoInitWcConfigWithPath(configPath); err != nil {
		panic(err)
	}
	if wcError := initialize.Initialize(jsonPath); wcError != nil {
		panic(wcError)
	}

	libinit.Init()
	cornRunner := crontask.NewCornRunner()
	r := asynctask.InitAsyncRunnerInProcess(cornRunner)
	s := web.NewBasicServer(config.WCConfig.Server, "api")
	register.RegisterApiAndTask(s, r)
	r.ManualProcessing(s)
	fmt.Println("Initialize success")
	s.Run()

	time.Sleep(time.Second * 2000)
}
