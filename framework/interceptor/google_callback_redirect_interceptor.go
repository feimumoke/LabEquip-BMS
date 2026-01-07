package interceptor

import (
	"fmt"
	"github.com/feimumoke/wechating/framework/constant"
	"github.com/feimumoke/wechating/framework/wcerror"
	"github.com/feimumoke/wechating/framework/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GoogleCallbackRedirectInterceptor struct {
}

func NewGoogleCallbackRedirectInterceptor() *GoogleCallbackRedirectInterceptor {
	return &GoogleCallbackRedirectInterceptor{}
}

func (l *GoogleCallbackRedirectInterceptor) PreHandle(rfCtx *gin.Context, _ interface{}, _ interface{}, header *web.Header, wrapper *web.Handler) *bmserror.BMSError {
	return nil
}

func (l *GoogleCallbackRedirectInterceptor) PostHandle(ginCtx *gin.Context, request interface{}, response interface{}, header *web.Header, handler *web.Handler, handleError *bmserror.BMSError) *bmserror.BMSError {
	isGoogleCallback := l.isGoogleCallback(handler)
	if !isGoogleCallback {
		return nil
	}
	//处理cookie
	respWithCookie, ok := response.(*web.RespWithCookie)
	if ok {
		for _, cookie := range respWithCookie.CookieList {
			http.SetCookie(ginCtx.Writer, cookie)
		}
	}

	//google登录成功
	if handleError == nil {
		redirectURL := l.getRedirectURL(header)
		http.Redirect(ginCtx.Writer, ginCtx.Request, redirectURL, 302)
		return nil
	}

	//登录失败，并且用户状态为禁止登录
	if handleError.Code() == constant.ErrUserLoginForbidden {
		redirectURL := l.getForbiddenRedirectURL(header)
		http.Redirect(ginCtx.Writer, ginCtx.Request, redirectURL, 302)
		return nil
	}

	return handleError
}

func (l *GoogleCallbackRedirectInterceptor) Order() int64 {
	return GoogleCallbackRedirectInterceptorOrder
}

func (l *GoogleCallbackRedirectInterceptor) isGoogleCallback(itemWrapper *web.Handler) bool {
	return itemWrapper.Url == "/auth/callback"
}

const host = "127.0.0.1:8080"

func (l *GoogleCallbackRedirectInterceptor) getRedirectURL(header *web.Header) string {
	redirectURLFormat := "https://%s%s"
	result := fmt.Sprintf(redirectURLFormat, host, "/wechat")
	return result
}
func (l *GoogleCallbackRedirectInterceptor) getForbiddenRedirectURL(header *web.Header) string {
	redirectURLFormat := "http://%s%s"
	return fmt.Sprintf(redirectURLFormat, host, "/wechat/403")
}
