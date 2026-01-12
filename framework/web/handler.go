package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/support/convert"
	"github.com/feimumoke/labequipbms/framework/support/monitor"
	"github.com/feimumoke/labequipbms/framework/support/trace"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Header struct {
	PointID    string `header:"point_id"`
	Region     string `header:"region"`
	UserID     string `header:"user_id"`
	UserEmail  string
	SupplierID int64  `header:"supplier_id"`
	ClientIP   string `header:"client_ip"`
}

type Response struct {
	RetCode int64       `json:"retcode"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type RespWithCookie struct {
	CookieList []*http.Cookie
	Data       interface{}
}

type Handler struct {
	Url     string
	Handler func(ctx context.Context, header *Header, request interface{}) (interface{}, *bmserror.BMSError)
	Req     interface{}
	Method  Method
}

const (
	Success     = 0
	DefaultFail = -1
)

func (h *Handler) Do(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			ctx := c.Request.Context()

			errMsg := string(debug.Stack())

			_, endFunc := monitor.AwesomeStart1(ctx)
			endFunc("panic", c.Request.URL.RequestURI(), -1, "reqID:"+trace.GetOrNewTraceID(ctx)+"\n"+errMsg)
			log.Errorf("panic info %v", errMsg)
			h.setJson(c, DefaultFail, "inner error", nil)
		}
	}()

	// 解析 header
	header, err := parseRequestHeader(c)
	if err != nil {
		h.setJson(c, DefaultFail, err.Error(), nil)
		return
	}

	newReq := reflect.New(reflect.TypeOf(h.Req).Elem()).Interface() // h.Req 是全局的，只能读
	err = h.getRequest(c, newReq)
	if err != nil {
		h.setJson(c, DefaultFail, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	if h.Method != HttpPOSTUpload {
		log.CtxInfof(ctx, "basic_api_request, URL: %v, head: %v, request: %v", h.Url, header, newReq)
	}
	response, wcErr := h.Handler(ctx, header, newReq)
	log.CtxInfof(ctx, "basic_api_response, response: %v", response)
	if wcErr != nil {
		log.CtxErrorf(ctx, "basic_api_response_err, error: %v", wcErr)
		h.setJson(c, DefaultFail, wcErr.Error(), response)
		return
	}

	h.setJson(c, Success, "", response)
}

func (h *Handler) getRequest(c *gin.Context, request interface{}) error {
	if h.Method == HttpGET {
		return fillGetParam(c, request)
	} else if h.Method == HttpPOST {
		return c.ShouldBindBodyWith(request, binding.JSON)
	} else if h.Method == HttpPOSTUpload {
		return assignSingleForm(c, request)
	}
	return errors.New("method not exist")
}

func assignSingleForm(c *gin.Context, request interface{}) error {
	const defaultMaxMemory = 32 << 20 // 32 MB
	err := c.Request.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		return err
	}
	multipartForm := c.Request.MultipartForm
	t := reflect.TypeOf(request).Elem()
	v := reflect.ValueOf(request).Elem()
	filename := "default.png"
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if isPbGenerateField(field.Name) {
			continue
		}
		fieldType := field.Type
		fieldTag := field.Tag
		fieldValue := v.Field(i)
		fileJsonName := getJsonName(fieldTag)
		if isByteSlice(fieldType) {
			fileSlice := multipartForm.File[fileJsonName]

			if len(fileSlice) == 0 {
				return bmserror.NewError(constant.ErrInternalServer, "no upload file")
			}
			uploadFile := fileSlice[0]
			filename = uploadFile.Filename
			file, err := uploadFile.Open()
			if err != nil {
				return bmserror.NewError(constant.ErrInternalServer, "no upload file")
			}
			defer file.Close()
			var buf bytes.Buffer
			_, err = io.Copy(&buf, file)
			if err != nil {
				return bmserror.NewError(constant.ErrInternalServer, err.Error())
			}

			bys := buf.Bytes()
			fieldValue.SetBytes(bys)
		} else {
			_, isFieldExist := multipartForm.Value[fileJsonName]
			if !isFieldExist {
				if fileJsonName == "file_name" {
					fmt.Println(filename)
				}
				continue
			}
			postFieldValueList := multipartForm.Value[fileJsonName]
			if isString(fieldType) && len(postFieldValueList) > 0 {
				strValPointer := reflect.ValueOf(&postFieldValueList[0])
				fieldValue.Set(strValPointer)
			}

			if isNumber(fieldType) {
				intVal, err := convert.StringToInt64(postFieldValueList[0])
				if err != nil {
					log.Infof("str to int err: %v", err.Error())
				}
				intValPointer := reflect.ValueOf(&intVal)
				fieldValue.Set(intValPointer)
			}
		}
	}
	return nil
}

func (h *Handler) setJson(c *gin.Context, code int64, msg string, data interface{}) {
	response := &Response{
		RetCode: code,
		Message: msg,
		Data:    data,
	}

	c.JSON(http.StatusOK, response)

	ctx := c.Request.Context()
	_, endFunc := monitor.AwesomeStart1(ctx)
	endFunc("result", c.Request.URL.RequestURI(), int(code), "reqID:"+trace.GetOrNewTraceID(ctx)+"| "+toString(response, 1000))
}

func toString(o interface{}, limit int) string {
	buffer, _ := json.Marshal(o)
	if len(buffer) > limit {
		buffer = buffer[:limit]
	}
	return string(buffer)
}

func parseRequestHeader(c *gin.Context) (*Header, error) {
	header := &Header{}
	err := c.ShouldBindHeader(header)
	if err != nil {
		return nil, err
	}

	// 获取客户端 IP
	header.ClientIP = c.ClientIP()

	// 获取请求路径
	requestPath := c.Request.URL.Path

	// 白名单路径不需要验证登录态
	if IsWhiteListPath(requestPath) {
		return header, nil
	}

	//:1:admin_001:c300a91e-87e
	// 从 HTTP Header 中获取 token
	// 支持两种方式：
	// 1. Authorization: Bearer <token>
	// 2. X-User-Email 和直接的 token
	token := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// 兼容直接传 token 的方式
		token = c.GetHeader("token")
		if token == "" {
			token = c.GetHeader("Token")
		}
	}

	userEmail := c.GetHeader("X-User-Email")

	// 如果没有 token，返回未授权错误
	if token == "" {
		return nil, errors.New("unauthorized: token is required")
	}

	// 验证 token 并获取用户信息
	ctx := c.Request.Context()
	authService := GetAuthService()
	accountInfo, authErr := authService.ValidateToken(ctx, token)
	if authErr != nil {
		return nil, errors.New("unauthorized: " + authErr.Error())
	}

	// 填充用户信息到 header
	header.UserID = accountInfo.UserID
	header.UserEmail = accountInfo.Email
	if userEmail != "" {
		header.UserEmail = userEmail
	}

	// 将用户信息存储到 gin.Context 中，供后续使用
	c.Set("user_id", accountInfo.UserID)
	c.Set("user_email", accountInfo.Email)
	c.Set("user_name", accountInfo.UserName)
	c.Set("account_info", accountInfo)

	return header, nil
}

func fillGetParam(c *gin.Context, req interface{}) error {
	requestParams := make(map[string]interface{})
	for k, v := range c.Request.URL.Query() {
		if len(v) < 1 {
			continue
		}

		l := reflect.TypeOf(req).Elem().NumField()
		for i := 0; i < l; i++ {
			t := reflect.TypeOf(req).Elem().Field(i)
			s := strings.Split(t.Tag.Get("json"), ",")
			if len(s) == 0 || k != s[0] {
				continue
			}
			kind := t.Type.Kind()
			if kind == reflect.Ptr {
				kind = t.Type.Elem().Kind()
			}

			if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 ||
				kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
				v2, err := strconv.Atoi(v[0])
				if err != nil {
					return err
				}
				requestParams[k] = v2
			} else if kind == reflect.Float32 {
				v3, err := strconv.ParseFloat(v[0], 32)
				if err != nil {
					return err
				}
				requestParams[k] = v3
			} else if kind == reflect.Float64 {
				v4, err := strconv.ParseFloat(v[0], 64)
				if err != nil {
					return err
				}
				requestParams[k] = v4
			} else {
				requestParams[k] = v[0]
			}
			break
		}
	}

	buf, err := json.Marshal(requestParams)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buf, req)
	if err != nil {
		return err
	}
	return nil
}
