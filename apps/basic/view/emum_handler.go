package view

import (
	"context"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/web"
)

func (e *UserHandler) GetEnumsView(ctx context.Context, header *web.Header, request interface{}) (interface{}, *bmserror.BMSError) {
	enumsMap := constant.GetEnumValues()
	return enumsMap, nil
}
