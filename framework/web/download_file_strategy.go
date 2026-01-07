package web

import (
	"net/http"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/gin-gonic/gin"
)

type DownloadFileHandlerImpl struct {
}

func (g *DownloadFileHandlerImpl) assignParamData(ctx *gin.Context, url *Handler, params interface{}) *bmserror.BMSError {
	fE := fillGetParam(ctx, params)
	if fE != nil {
		return bmserror.NewError(constant.ErrParam, fE.Error())
	}
	return nil
}

func (g *DownloadFileHandlerImpl) handleApiResponse(ctx *gin.Context, itemWrapper *Handler, response interface{}, handleError *bmserror.BMSError) {
	if handleError != nil {
		setJson(ctx, int64(handleError.Code()), handleError.Message(), response)
		return
	}

	resp, ok := response.(*RespWithDownloadFile)
	if !ok {
		setJson(ctx, constant.ErrInternalServer, "response is not type [RespWithDownloadFile]", response)
		return
	}
	mime := http.DetectContentType(resp.fileByteSlice)
	ctx.Header("Content-Type", mime)
	ctx.Header("Content-Disposition", "attachment; filename="+resp.fileName)
	ctx.Writer.Write(resp.fileByteSlice)
}
