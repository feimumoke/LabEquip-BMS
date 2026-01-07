package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/support/convert"
	"github.com/feimumoke/labequipbms/framework/support/monitor"
	"github.com/feimumoke/labequipbms/framework/support/trace"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gin-gonic/gin/render"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func APICatMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		start := time.Now()
		url := c.Request.URL.RequestURI()

		//trace id
		ctx, unSetTraceFunc := trace.Init(context.TODO())
		defer unSetTraceFunc(ctx)
		traceID := trace.GetOrNewTraceID(ctx)

		req := map[string]interface{}{}
		err := c.ShouldBindBodyWith(&req, binding.JSON)
		if err != nil {
			log.Errorf("parse req body %v", err.Error())
		}
		reqJson, err := json.Marshal(req)
		if err != nil {
			log.Errorf("parse req body %v", err.Error())
		}
		reqParam := truncate(string(reqJson), 500)

		log.Infof("WCAPI url:%v,req:%v", url, reqParam)

		// 请求前
		var catMsg string
		var catStatus = convert.Int(0)
		_, endFunc := monitor.AwesomeStart1(context.TODO())

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}

		defer func() {
			var resp = blw.body.String()
			if x := recover(); x != nil {
				errMsg := string(debug.Stack())
				*catStatus = -1
				catMsg = errMsg

				log.Errorf("panic info %v", errMsg)
				response := &Response{
					Message: "inner error",
					RetCode: -1,
				}
				resp = errMsg

				c.Render(http.StatusOK, render.JSON{Data: response})
			}

			latency := time.Since(start)

			resp = truncate(resp, 500)

			log.Infof("WCAPI url:%v,consumed time:%vms:,req:%v, resp:%v", url, latency.Milliseconds(), reqParam, resp)
			endFunc("API", url, *catStatus, "reqID:"+traceID+","+catMsg)
		}()

		//保存 resp
		c.Writer = blw

		c.Next()

		// 请求后

		resp := &Response{}
		err = json.Unmarshal(blw.body.Bytes(), resp)
		if err != nil {
			log.Errorf("resp json unmarshal err:%v", err.Error())
		}

		if resp.RetCode != 0 {
			*catStatus = -1
		}
		catMsg = blw.body.String()

	}
}

func truncate(str string, limit int) string {
	if len(str) > limit {
		str = str[:limit]
	}
	return str
}
