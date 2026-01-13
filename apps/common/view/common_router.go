package view

import (
	pbbasic "github.com/feimumoke/labequipbms/api_idl/apps/basic"
	"github.com/feimumoke/labequipbms/framework/asynctask"
	"github.com/feimumoke/labequipbms/framework/web"
)

func InitCommonView(s *web.BasicServer, r *asynctask.AsyncRunner) {
	cv := &CommonView{}
	s.RegisterPOSTUpload("/apps/common/upload_file", cv.UploadFileView, &pbbasic.UploadFileRequest{})
}
