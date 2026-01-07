package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func setResponseCookie(ctx *gin.Context, cookieList []*http.Cookie) {
	for _, cookie := range cookieList {
		http.SetCookie(ctx.Writer, cookie)
	}
}
