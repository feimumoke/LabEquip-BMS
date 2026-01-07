package web

import (
	"fmt"
	"net/http"

	"github.com/feimumoke/labequipbms/framework/appcontext"
	"github.com/feimumoke/labequipbms/framework/support/monitor"
	"github.com/feimumoke/labequipbms/framework/support/trace"
	"github.com/gin-gonic/gin"
)

func APICtxMiddleware() func(*gin.Context) {
	return func(c *gin.Context) {
		//trace id
		ctx, unSetTraceFunc := trace.Init(c.Request.Context())
		defer unSetTraceFunc(ctx)
		c.Next()
	}
}

//func APIRecoverMiddleware() func(*gin.Context) {
//	return func(c *gin.Context) {
//		defer func() {
//			if err := recover(); err != nil {
//				ctx := c.Keys["ctx"].(context.Context)
//
//				errMsg := string(debug.Stack())
//
//				_, endFunc := monitor.AwesomeStart1(ctx)
//				endFunc("panic", c.Request.URL.RequestURI(), -1, "reqID:"+trace.GetOrNewTraceID(ctx)+"\n"+errMsg)
//				log.Errorf("panic info %v", errMsg)
//
//				c.JSON(http.StatusOK, &Response{
//					Message: "inner error",
//					RetCode: DefaultFail,
//				})
//			}
//		}()
//		fmt.Println("--2")
//		c.Next()
//		fmt.Println("==2")
//	}
//}

func APIMonitorMiddleware() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx1, endFunc := monitor.AwesomeStart1(ctx)

		c.Request = c.Request.WithContext(ctx1)
		c.Next()

		endFunc("API", c.Request.URL.RequestURI(), 0, "")
	}
}

func AddScormMiddleware() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = appcontext.BindContext(ctx)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func Cors() gin.HandlerFunc {

	return func(c *gin.Context) {
		method := c.Request.Method
		fmt.Println("------- " + method)
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}

}
