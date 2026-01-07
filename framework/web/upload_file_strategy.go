package web

import (
	"bytes"
	"io"
	"reflect"
	"strings"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/support/convert"
	"github.com/feimumoke/labequipbms/framework/support/iters"
	"github.com/gin-gonic/gin"
)

type UploadFileHandlerImpl struct {
}

func (g *UploadFileHandlerImpl) assignParamData(ctx *gin.Context, itemWrapper *Handler, request interface{}) *bmserror.BMSError {
	//todo check content type
	if !isMultipartForm(ctx) {
		return bmserror.NewError(constant.ErrBadRequest, "content-type invalid")
	}

	const defaultMaxMemory = 32 << 20 // 32 MB
	err := ctx.Request.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		return bmserror.NewError(constant.ErrBadRequest, err.Error())
	}
	multipartForm := ctx.Request.MultipartForm
	t := reflect.TypeOf(request).Elem()
	v := reflect.ValueOf(request).Elem()
	for i := 0; i < t.NumField(); i++ {

		currentStructFieldType := t.Field(i)
		if isPbGenerateField(currentStructFieldType.Name) {
			continue
		}

		currentFieldReflectType := currentStructFieldType.Type
		currentStructFieldTag := currentStructFieldType.Tag
		currentFieldReflectValue := v.Field(i)

		fileJsonName := getJsonName(currentStructFieldTag)
		if isByteSlice(currentFieldReflectType) {
			fileSlice := multipartForm.File[fileJsonName]

			if len(fileSlice) == 0 {
				return bmserror.NewError(constant.ErrInternalServer, "no upload file")
			}
			uploadFile := fileSlice[0]
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
			currentFieldReflectValue.SetBytes(bys)
		} else {
			jsonName := getJsonName(currentStructFieldTag)
			_, isFieldExist := multipartForm.Value[jsonName]
			if !isFieldExist {
				continue
			}

			postFieldValueList := multipartForm.Value[jsonName]
			if isString(currentFieldReflectType) && len(postFieldValueList) > 0 {
				strValPointer := reflect.ValueOf(&postFieldValueList[0])
				currentFieldReflectValue.Set(strValPointer)
			}

			if isNumber(currentFieldReflectType) {
				intVal, err := convert.StringToInt64(postFieldValueList[0])
				if err != nil {
					log.Infof("str to int err: %v", err.Error())
				}
				intValPointer := reflect.ValueOf(&intVal)
				currentFieldReflectValue.Set(intValPointer)
			}

		}

	}
	return nil

}

func isMultipartForm(ctx *gin.Context) bool {
	contentType := ctx.Request.Header.Get("Content-Type")
	return strings.Contains(contentType, "multipart/form-data")
}

func isPbGenerateField(field string) bool {
	pbGenerateField := []string{"state", "sizeCache", "unknownFields"}
	return iters.From(pbGenerateField).Contains(field)
}

func (g *UploadFileHandlerImpl) handleApiResponse(ctx *gin.Context, itemWrapper *Handler, response interface{}, handleError *bmserror.BMSError) {
	setJson(ctx, int64(handleError.Code()), handleError.Message(), response)
}

func isByteSlice(t reflect.Type) bool {
	if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return true

	}

	return false

}
func isString(t reflect.Type) bool {

	if t.Kind() == reflect.String {
		return true
	}

	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.String {
		return true
	}
	return false
}

func isNumber(refType reflect.Type) bool {
	if refType.Kind() != reflect.Ptr {
		return false
	}
	refType = refType.Elem()
	kind := refType.Kind()
	if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 ||
		kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
		return true
	}
	return false
}

func getJsonName(tag reflect.StructTag) string {
	jsonName := strings.Split(tag.Get("json"), ",")[0]
	return jsonName
}
