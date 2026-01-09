package interceptor

import (
	"context"
	"net/http"
	"strings"

	"github.com/feimumoke/labequipbms/apps/basic/manager"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/defines/entity"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/web"
	"github.com/gin-gonic/gin"
)

type LoginStatusInterceptor struct {
	userMng *manager.UserManager
}

func NewLoginStatusInterceptor() *LoginStatusInterceptor {
	return &LoginStatusInterceptor{
		userMng: manager.NewUserManager(),
	}
}

func (l LoginStatusInterceptor) PreHandle(rfCtx *gin.Context, _ interface{}, _ interface{}, header *web.Header, wrapper *web.Handler) *bmserror.BMSError {
	if l.noNeedVerify(wrapper) {
		return nil
	}
	user, err := l.getUserCache(rfCtx)
	log.Infof("getUserCache %v", user)
	if err != nil {
		return err.Mark()
	}

	l.fillHeaderWithUser(header, user)

	return nil
}

func (l LoginStatusInterceptor) getUserCache(rfCtx *gin.Context) (*entity.AccountCacheInfo, *bmserror.BMSError) {
	user, err := l.getUser(rfCtx.Request.Context(), rfCtx.Request)
	if err != nil {
		return nil, err.AddError(constant.ErrNotLogin, "not login")
	}
	if user == nil {
		return nil, bmserror.NewError(constant.ErrNotLogin, "not login")
	}
	return user, nil
}

func (l LoginStatusInterceptor) PostHandle(_ *gin.Context, _ interface{}, _ interface{}, _ *web.Header, _ *web.Handler, handleError *bmserror.BMSError) *bmserror.BMSError {
	return handleError
}

func (l LoginStatusInterceptor) Order() int64 {
	return LoginStatusInterceptorOrder
}

func (l LoginStatusInterceptor) getUser(ctx context.Context, request *http.Request) (*entity.AccountCacheInfo, *bmserror.BMSError) {
	userIDCookie, _ := request.Cookie(constant.CookieAuthUid)
	authSKeyCookie, _ := request.Cookie(constant.CookieAuthSkey)
	if userIDCookie == nil || authSKeyCookie == nil || len(userIDCookie.Value) == 0 || len(authSKeyCookie.Value) == 0 {
		return nil, nil
	}

	user, err := l.userMng.GetUserInfo(ctx, userIDCookie.Value, authSKeyCookie.Value)
	if err != nil {
		return nil, err.Mark()
	}
	return user, nil
}

func (l LoginStatusInterceptor) noNeedVerify(itemWrapper *web.Handler) bool {
	url := itemWrapper.Url

	isWhiteURL := isUrlInWhiteList(url)
	if isWhiteURL {
		return true
	}
	return l.isOpenapi(url)
}

func (l LoginStatusInterceptor) isOpenapi(urlPath string) bool {
	openapiIdentifier := "openapi"
	return strings.Contains(strings.ToLower(urlPath), openapiIdentifier)
}

func (l LoginStatusInterceptor) fillHeaderWithUser(header *web.Header, user *entity.AccountCacheInfo) {
	header.UserID = user.UserID
	header.UserEmail = user.Email
}

func isUrlInWhiteList(url string) bool {
	_, ok := apiWhiteCollection[url]
	return ok
}

var apiWhiteCollection = map[string]interface{}{
	"/api/v2/system/user/account/login/page":                              nil,
	"/api/v2/system/user/account/login/loginSuccessPage":                  nil,
	"/api/v2/system/user/account/login/css/bootstrap.min.css":             nil,
	"/api/v2/system/user/account/login/css/animate.css":                   nil,
	"/api/v2/system/user/account/login/css/style.css":                     nil,
	"/api/v2/system/user/account/login/font-awesome/css/font-awesome.css": nil,
	"/api/v2/system/user/account/login/js/jquery-2.1.1.js":                nil,
	"/api/v2/system/user/account/login/js/bootstrap.min.js":               nil,
	"/api/v2/system/user/account/login/js/plugins/md5/jquery.md5.js":      nil,
	"/api/v2/system/user/account/login/js/js.cookie.js":                   nil,
	"/api/v2/system/user/account/login/js/common/translate.js":            nil,
	//login url
	"/api/v2/apps/system/user/user_login":                nil,
	"/api/v2/apps/system/user/user_pda_login":            nil,
	"/api/v2/apps/system/user/user_google_auth":          nil,
	"/api/v2/apps/system/user/user_google_auth_callback": nil,
	"/api/v2/apps/system/user/get_host_mapping":          nil,
	"/api/v2/apps/system/user/get_captcha":               nil,
	"/api/v2/apps/system/user/simple_password_deadline":  nil,

	// enum
	"/api/v2/apps/basic/system/domain/enums": nil,
}
