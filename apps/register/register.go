package register

import (
	"github.com/feimumoke/labequipbms/apps/basic/view"
	bmsview "github.com/feimumoke/labequipbms/apps/bms/view"
	"github.com/feimumoke/labequipbms/apps/common/idutil"
	"github.com/feimumoke/labequipbms/framework/asynctask"
	"github.com/feimumoke/labequipbms/framework/interceptor"
	"github.com/feimumoke/labequipbms/framework/web"
)

func Init() {
	idutil.InitDIDCreator()
	idutil.InitDailyIDCreator()
}

func RegisterApiAndTask(s *web.BasicServer, a *asynctask.AsyncRunner) {
	//common
	Init()
	initBasic(s, a)
	initBusiness(s, a)
	initInterceptor()
}

func initBasic(s *web.BasicServer, a *asynctask.AsyncRunner) {
	view.InitBasicView(s, a)
}

func initBusiness(s *web.BasicServer, a *asynctask.AsyncRunner) {
	bmsview.InitBMSView(s, a)
}

func initInterceptor() {
	interceptorList := []web.Interceptor{
		interceptor.NewLoginStatusInterceptor(),
		interceptor.NewApiPermissionInterceptor(),
		interceptor.NewGoogleCallbackRedirectInterceptor(),
	}
	web.InterceptorCollection.AddInterceptors(interceptorList...)
}
