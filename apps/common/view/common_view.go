package view

import (
	"context"
	"os"
	"strings"

	pbbasic "github.com/feimumoke/labequipbms/api_idl/apps/basic"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/convert"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
	"github.com/feimumoke/labequipbms/framework/web"
	"github.com/google/uuid"
)

type CommonView struct {
}

func (e *CommonView) UploadFileView(ctx context.Context, header *web.Header, request interface{}) (interface{}, *bmserror.BMSError) {
	req := request.(*pbbasic.UploadFileRequest)

	// 获取原始文件名和扩展名
	originalFilename := req.GetFileName()
	if originalFilename == "" {
		originalFilename = "file.jpg" // 默认文件名
	}

	// 获取文件扩展名
	ext := ""
	if idx := strings.LastIndex(originalFilename, "."); idx != -1 {
		ext = originalFilename[idx:] // 包含点号，如 ".jpg"
	} else {
		ext = ".jpg" // 默认扩展名
	}

	// 创建上传目录
	dateStr := timeutil.TodayDateStr(timeutil.DateIntFormat)
	uploadDir := "./uploads/" + dateStr + "/"
	os.MkdirAll(uploadDir, os.ModePerm)
	os.Chmod(uploadDir, os.ModePerm)

	// 生成唯一文件名（UUID + 扩展名）
	generateUUID, _ := uuid.NewUUID()
	filename := generateUUID.String() + ext
	fullPath := uploadDir + filename

	// 创建文件
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, bmserror.NewError(constant.ErrFile, "create file err %v", err.Error())
	}
	defer f.Close()

	// 写入文件内容
	_, err2 := f.Write(req.File)
	if err2 != nil {
		return nil, bmserror.NewError(constant.ErrFile, "write file err %v", err2.Error())
	}

	// 返回访问 URL（格式：/uploads/20240113/uuid.jpg）
	fileUrl := "/uploads/" + dateStr + "/" + filename

	return &pbbasic.UploadFileResponse{
		Url: convert.String(fileUrl),
	}, nil
}
