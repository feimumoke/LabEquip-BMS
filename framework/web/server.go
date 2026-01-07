package web

import (
	"context"

	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/config"
	"github.com/gin-gonic/gin"
)

type Method = string

const (
	HttpGET        Method = "GET"
	HttpPOST       Method = "POST"
	HttpPOSTUpload Method = "POSTUpload"
)

type BasicServer struct {
	handlers []*Handler
	engine   *gin.Engine
	addr     string
}

type RespWithDownloadFile struct {
	fileByteSlice []byte
	fileName      string
}

type RespHeader struct {
	RetCode int
	Message string
	Values  []interface{}
}

func NewRespWithDownloadFile(fileByteSlice []byte, fileName string) *RespWithDownloadFile {
	return &RespWithDownloadFile{
		fileByteSlice: fileByteSlice,
		fileName:      fileName,
	}
}

func NewBasicServer(conf *config.ServerConfig, name string) *BasicServer {
	b := &BasicServer{
		engine: gin.Default(),
		addr:   conf.Addr[name],
	}

	return b
}

func (r *BasicServer) RegisterGET(url string, fun func(context.Context, *Header, interface{}) (interface{}, *bmserror.BMSError), req interface{}) {

	r.handlers = append(r.handlers, &Handler{
		Url:     url,
		Handler: fun,
		Req:     req,
		Method:  HttpGET,
	})
}

func (r *BasicServer) RegisterPOST(url string, fun func(context.Context, *Header, interface{}) (interface{}, *bmserror.BMSError), req interface{}) {

	r.handlers = append(r.handlers, &Handler{
		Url:     url,
		Handler: fun,
		Req:     req,
		Method:  HttpPOST,
	})
}

func (r *BasicServer) RegisterPOSTUpload(url string, fun func(context.Context, *Header, interface{}) (interface{}, *bmserror.BMSError), req interface{}) {

	r.handlers = append(r.handlers, &Handler{
		Url:     url,
		Handler: fun,
		Req:     req,
		Method:  HttpPOSTUpload,
	})
}

func (r *BasicServer) Run() *bmserror.BMSError {
	r.registerURL()
	r.engine.Static("/upload", "./upload")

	err := r.engine.Run(r.addr)
	if err != nil {
		return bmserror.NewError(-1, err.Error())
	}
	return nil
}

func (r *BasicServer) registerURL() {
	apiGroup := r.setAPIMiddleWare()

	for _, handler := range r.handlers {
		tHandler := handler
		switch tHandler.Method {
		case HttpGET:
			apiGroup.GET(tHandler.Url, tHandler.Do)
		case HttpPOST:
			apiGroup.POST(tHandler.Url, tHandler.Do)
		case HttpPOSTUpload:
			apiGroup.POST(tHandler.Url, tHandler.Do)
		}
	}
}

func (r *BasicServer) setAPIMiddleWare() *gin.RouterGroup {
	r.engine.Use(Cors(), APICtxMiddleware(), APIMonitorMiddleware(), AddScormMiddleware())
	apiGroup := r.engine.Group("/api")
	return apiGroup
}
