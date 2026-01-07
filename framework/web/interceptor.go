package web

import (
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/iters"
	"github.com/gin-gonic/gin"
)

type Interceptor interface {
	//请求之前处理
	PreHandle(ctx *gin.Context, request interface{}, response interface{}, header *Header, wrapper *Handler) *bmserror.BMSError
	//请求处理之后处理
	PostHandle(ctx *gin.Context, request interface{}, response interface{}, header *Header, wrapper *Handler, handleError *bmserror.BMSError) *bmserror.BMSError
	//执行顺序
	Order() int64
}

type interceptorRegistry struct {
	interceptorList []Interceptor
}

var InterceptorCollection = interceptorRegistry{}

func (i *interceptorRegistry) AddInterceptors(interceptor ...Interceptor) {
	i.interceptorList = append(i.interceptorList, interceptor...)
	var orderedList []Interceptor

	iters.From(i.interceptorList).Sort(func(i, j interface{}) bool {
		first := i.(Interceptor)
		second := j.(Interceptor)
		return first.Order() < second.Order()
	}).ToSlice(&orderedList)

	i.interceptorList = orderedList
}

func (i *interceptorRegistry) applyPreHandle(ctx *gin.Context, request interface{}, response interface{}, header *Header, wrapper *Handler) *bmserror.BMSError {
	for _, interceptor := range i.interceptorList {
		err := interceptor.PreHandle(ctx, request, response, header, wrapper)
		if err != nil {
			return err.Mark()
		}
	}
	return nil
}

func (i *interceptorRegistry) applyPostHandle(ctx *gin.Context, request interface{}, response interface{}, header *Header, wrapper *Handler, handleError *bmserror.BMSError) *bmserror.BMSError {
	for _, interceptor := range i.interceptorList {
		err := interceptor.PostHandle(ctx, request, response, header, wrapper, handleError)
		if err != nil {
			return err.Mark()
		}
	}
	return nil
}
