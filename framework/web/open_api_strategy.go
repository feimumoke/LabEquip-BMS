package web

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/config"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/support/convert"
	"github.com/feimumoke/labequipbms/framework/support/monitor"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
	"github.com/feimumoke/labequipbms/framework/support/trace"
	"github.com/feimumoke/labequipbms/framework/support/util"
	"github.com/gin-gonic/gin"
)

const (
	JwtAccountKey               = "account"
	JwtSecuritySecretKey string = "jwt_security_secret"
	JwtSecurityIvKey     string = "jwt_security_iv"
)

type OpenApiHandlerImpl struct {
}

func NewOpenApiHandlerImpl() *OpenApiHandlerImpl {
	return &OpenApiHandlerImpl{}
}

type openApiClaims struct {
	jwt.StandardClaims
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
}

func (g *OpenApiHandlerImpl) assignParamData(ctx *gin.Context, url *Handler, params interface{}) *bmserror.BMSError {
	jwtString, err := g.getJwtString(ctx, url)
	if err != nil {
		return err.Mark()
	}

	jwtToken, parseErr := jwt.ParseWithClaims(jwtString, &openApiClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 校验是不是hs256加密的
		if singingMethod, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, bmserror.NewError(constant.ErrInternalServer, "Signature Algorithm Method Error.")
		} else {
			if singingMethod != jwt.SigningMethodHS256 {
				return nil, bmserror.NewError(constant.ErrInternalServer, "Signature Algorithm Method Error.")
			}
		}

		// 获取account并校验
		account, ok := token.Header[JwtAccountKey]
		if !ok {
			return nil, bmserror.NewError(constant.ErrInternalServer, "Account Is Required.")
		}
		var cfg *config.JwtConfig
		for _, jwtCfg := range config.WCConfig.JwtConfigs {
			if jwtCfg.Account == account {
				cfg = jwtCfg
			}
		}

		if cfg == nil {
			return nil, bmserror.NewError(constant.ErrParam, "Forbidden.")
		}

		// 校验时间戳
		tokenClaims, ok := token.Claims.(*openApiClaims)
		if !ok {
			return nil, bmserror.NewError(constant.ErrParam, "Timestamp invalid.")
		}
		if int64(math.Abs(float64(timeutil.GetCurrentUnix()-tokenClaims.Timestamp))) > 300 {
			return nil, bmserror.NewError(constant.ErrParam, "Timestamp Diff More Than Exceed 300s. Now Timestamp Is %v.", timeutil.GetCurrentUnix())
		}
		secret := cfg.SecretKey
		iv := cfg.AppKey
		if g.isNeedEncrypt(url.Url) {
			if len(secret) == 0 || len(iv) == 0 {
				return nil, bmserror.NewError(constant.ErrParam, "security config error")
			}
			// 设置加密参数，响应时获取加密的secret和iv对响应参数进行加密
			ctx.Set(JwtSecuritySecretKey, secret)
			ctx.Set(JwtSecurityIvKey, iv)
		}

		// 这里要返回秘钥，用于Parse方法内对jwt的校验
		return []byte(secret), nil
	})
	if parseErr != nil {
		// 判断是不是校验或解析问题
		wcErr, ok := parseErr.(*jwt.ValidationError).Inner.(*bmserror.BMSError)
		if ok {
			return wcErr.Mark()
		} else {
			return bmserror.NewError(constant.ErrParam, "Jwt Format Error Or Signature Error.")
		}
	}

	// 解析数据，填充到request中
	tokenClaims, ok := jwtToken.Claims.(*openApiClaims)
	if !ok {
		return bmserror.NewError(constant.ErrParam, "Data Is Required.")
	}

	log.Infof("wcV2API#url:%s,Jwt Raw Data:%s", url.Url, string(tokenClaims.Data))

	jsonErr := json.Unmarshal(tokenClaims.Data, params)
	if jsonErr != nil {
		log.Errorf("High precision unmarshal error [%v]", jsonErr.Error())
		jwtData := make(map[string]interface{})
		jsonErr := json.Unmarshal(tokenClaims.Data, &jwtData)
		if jsonErr != nil {
			return bmserror.NewError(constant.ErrParam, "unmarshal error [%v]", jsonErr.Error())
		}
		jsonValue, jsonErr := json.Marshal(jwtData)
		if jsonErr != nil {
			return bmserror.NewError(constant.ErrJsonEncodeFail, "marshal error [%v]", jsonErr.Error())
		}

		jsonErr = json.Unmarshal(jsonValue, params)
		if jsonErr != nil {
			return bmserror.NewError(constant.ErrInternalServer, "unmarshal error [%v]", jsonErr.Error())
		}
	}

	return nil
}

func (g *OpenApiHandlerImpl) getJwtString(ctx *gin.Context, url *Handler) (string, *bmserror.BMSError) {
	jwtString := ""
	if url.Method == "POST_OPEN_API" {
		type JwtPostStruct struct {
			Jwt *string `json:"jwt"`
		}
		jwtPostString := &JwtPostStruct{}
		bodyByte, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			log.Errorf("read jwt request body err : %v ", err.Error())
			return "", bmserror.NewError(constant.ErrParam, "Jwt Is Required.")
		}
		err = json.Unmarshal(bodyByte, jwtPostString)
		if err != nil {
			log.Errorf("Unmarshal jwt request body err : %v ", err.Error())
			return "", bmserror.NewError(constant.ErrParam, "Jwt Is Required.")
		}
		jwtString = convert.StringValue(jwtPostString.Jwt)
	} else { // GET
		jwtString = ctx.Request.FormValue("jwt")
	}
	if jwtString == "" {
		return "", bmserror.NewError(constant.ErrInternalServer, "Jwt Is Required.")
	}
	return jwtString, nil
}

func (g *OpenApiHandlerImpl) handleApiResponse(ctx *gin.Context, itemWrapper *Handler, response interface{}, handleError *bmserror.BMSError) {
	if g.isNeedEncrypt(itemWrapper.Url) {
		secretVal, _ := ctx.Get(JwtSecuritySecretKey)
		ivVal, _ := ctx.Get(JwtSecurityIvKey)
		secret := convert.ToString(secretVal)
		iv := convert.ToString(ivVal)
		if len(secret) > 0 && len(iv) > 0 {
			genericApiSecurityResponse(ctx, secret, iv, response, handleError)
			return
		}
	}
	setJson(ctx, int64(handleError.Code()), handleError.Message(), response)
}

func (c *OpenApiHandlerImpl) isNeedEncrypt(path string) bool {
	return strings.Contains(path, "order")
}

func handleResponseCookie(c *gin.Context, response interface{}) interface{} {
	if respWithCookie, ok := response.(RespWithCookie); ok {
		//set cookie
		setResponseCookie(c, respWithCookie.CookieList)

		return respWithCookie.Data
	}

	return response
}

func genericApiSecurityResponse(ctx *gin.Context, secret string, iv string, response interface{}, handleError *bmserror.BMSError) {
	resp := handleResponseCookie(ctx, response)
	if handleError == nil {
		jsonData := util.ToJSON(resp)
		//加密处理
		encrypt, err := util.RawAes128CbcEncrypt(jsonData, []byte(secret), []byte(iv))
		if err != nil {
			setJson(ctx, int64(err.Code()), err.Message(), resp)
			return
		}
		setJson(ctx, 0, "success", encrypt)
	} else {
		errCode := int64(handleError.Code())
		//errMsg = handleError.Error()
		setJson(ctx, errCode, handleError.Message(), resp)
	}
}

func setJson(c *gin.Context, code int64, msg string, data interface{}) {
	resp := handleResponseCookie(c, data)
	response := &Response{
		RetCode: code,
		Message: msg,
		Data:    resp,
	}

	c.JSON(http.StatusOK, response)

	ctx := c.Request.Context()
	_, endFunc := monitor.AwesomeStart1(ctx)
	endFunc("result", c.Request.URL.RequestURI(), int(code), "reqID:"+trace.GetOrNewTraceID(ctx)+"| "+toString(response, 1000))
}
