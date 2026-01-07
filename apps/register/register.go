package register

import (
	"github.com/feimumoke/labequipbms/apps/basic/view"
	"github.com/feimumoke/labequipbms/framework/asynctask"
	"github.com/feimumoke/labequipbms/framework/interceptor"
	"github.com/feimumoke/labequipbms/framework/web"
)

func RegisterApiAndTask(s *web.BasicServer, a *asynctask.AsyncRunner) {
	//common
	initBasic(s, a)
	initBusiness(s, a)
	initInterceptor()
}

func initBasic(s *web.BasicServer, a *asynctask.AsyncRunner) {
	view.InitBasicView(s, a)
}

func initBusiness(s *web.BasicServer, a *asynctask.AsyncRunner) {
}

func initInterceptor() {
	interceptorList := []web.Interceptor{
		interceptor.NewLoginStatusInterceptor(),
		interceptor.NewApiPermissionInterceptor(),
		interceptor.NewGoogleCallbackRedirectInterceptor(),
	}
	web.InterceptorCollection.AddInterceptors(interceptorList...)
}
