package libinit

import "github.com/feimumoke/wechating/apps/common/cservice"

func Init() {
	cservice.InitDIDCreator()
	cservice.InitDailyIDCreator()
}
