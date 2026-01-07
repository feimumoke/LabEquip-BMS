package web

import (
	"context"
	"net/http"

	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/support/util"
)

const RealIpHeaderConfigName = "real_ip_header"
const XForwordedForHeader = "X-FORWARDED-FOR"

func logReq(ctx context.Context, data interface{}, header *Header) {
	reqJson := util.ToJSON(data)
	headerJson := util.ToJSON(header)

	log.CtxInfof(ctx, "WCAPI#req:#[#param:%s,header:%s#]#", reqJson, headerJson)
}

func logResp(ctx context.Context, data interface{}, request interface{}, header *Header, handleError *bmserror.BMSError, elapsedMs string) {

	reqJson := util.ToJSON(request)
	headerJson := util.ToJSON(header)
	respJson := util.ToJSON(data)

	if len(respJson) > 500 {
		respJson = respJson[:500]
	}

	if handleError == nil {
		log.CtxInfof(ctx, "WcV2API#resp:[%s]#[#param:%s,header:%s,resp:%s#]#", elapsedMs, reqJson, headerJson, respJson)
	} else {
		log.CtxErrorf(ctx, "WcV2API#resp:[%s]#[#param:%s,header:%s,resp:%s,err:%s#]#", elapsedMs, reqJson, headerJson, respJson, handleError.DebugError())
	}
}

func parseSysCodeFromHeader(header http.Header) string {
	return header.Get("account")
}

func parseRefererFromHeader(request *http.Request) string {
	return request.Header.Get("referer")
}

func parseWcsWhsIDFromHeader(header http.Header) string {
	return header.Get("x-whs")
}

func parseWcsSupplierFromHeader(header http.Header) string {
	return header.Get("x-supplier")
}

func parseWcsDeviceFromHeader(header http.Header) string {
	return header.Get("x-device")
}

func parseWcsTraceIDFromHeader(header http.Header) string {
	return header.Get("x-traceid")
}

func parseXCountryFromHeader(header http.Header) string {
	return header.Get("X-Country")
}
