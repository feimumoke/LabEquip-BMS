package interceptor

import (
	"context"
	"os"

	"github.com/feimumoke/labequipbms/apps/basic/manager"
	"github.com/feimumoke/labequipbms/framework/appcontext"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/web"
	"github.com/gin-gonic/gin"
)

type InterceptorOrderType = int64

const (
	AddTraceInterceptorOrder InterceptorOrderType = iota
	LoginStatusInterceptorOrder
	ApiPermissionInterceptorOrder
	LanguageInterceptorOrder
	GoogleCallbackRedirectInterceptorOrder
	LoginURLRedirectInterceptorOrder
	OpHistoryInterceptorOrder
	MTIConfigCheckInterceptorOrder
	POIBogorConfigCheckInterceptorOrder
	AdminPortalPermissionCheckerOrder
	PiiRateLimitInterceptorOrder
)

const (
	ApiPermissionControlSwitchKey = "apiPermissionControlSwitch"
	SwitchOpen                    = "1"
	SwitchWarn                    = "2"
)

type ApiPermissionInterceptor struct {
	userMng *manager.UserManager
}

func NewApiPermissionInterceptor() *ApiPermissionInterceptor {
	return &ApiPermissionInterceptor{userMng: manager.NewUserManager()}
}

func (a *ApiPermissionInterceptor) PreHandle(rfCtx *gin.Context, _ interface{}, _ interface{}, header *web.Header, wrapper *web.Handler) *bmserror.BMSError {
	// 上报 Prometheus

	if header.UserEmail == "" {
		//用户未登录，无需登录接口直接放行
		return nil
	}
	if !isApi() {
		// 非api，直接返回
		return nil
	}
	//检查API权限
	return a.CheckApiPermission(rfCtx.Request.Context(), header, wrapper.Url)
}

func (a *ApiPermissionInterceptor) PostHandle(_ *gin.Context, _ interface{}, _ interface{}, _ *web.Header, _ *web.Handler, handleError *bmserror.BMSError) *bmserror.BMSError {
	return handleError
}

func (a *ApiPermissionInterceptor) Order() int64 {
	return ApiPermissionInterceptorOrder
}

func isApi() bool {
	moduleName := os.Getenv("MODULE_NAME")
	return moduleName == "api"
}

// CheckApiPermission 抽出公共方法,以便rms 调用grpc使用
func (a *ApiPermissionInterceptor) CheckApiPermission(ctx context.Context, header *web.Header, path string) *bmserror.BMSError {
	// 开关控制
	if !a.isOpenApiPermissionControl(ctx, header.PointID) {
		// 开关关闭，直接返回
		log.Infof("permission control closed, ptId: %v, userEmail: %v, url: %v", header.PointID, header.UserEmail, path)
		return nil
	}

	//1.获取用户角色列表
	hasPermission := a.userMng.IsUserHasPermission(ctx, header.UserID, path)
	if hasPermission {
		return nil
	}
	return bmserror.NewError(ErrUnAuthorized, "permission denied")
}

const ErrUnAuthorized = -100401 //UnAuth
func (a *ApiPermissionInterceptor) isOpenApiPermissionControl(ctx context.Context, ptId string) bool {
	val := appcontext.AppCtx.ConfigFunc(ctx, ApiPermissionControlSwitchKey, ptId)

	// 开启开关或者开启告警，认为是开启
	return val == SwitchOpen || val == SwitchWarn
}

func (a *ApiPermissionInterceptor) isWarnApiPermissionControl(ctx context.Context, ptId string) bool {
	val := appcontext.AppCtx.ConfigFunc(ctx, ApiPermissionControlSwitchKey, ptId)
	return val == SwitchWarn
}
