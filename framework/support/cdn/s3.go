package cdn

import (
	"fmt"
	"github.com/feimumoke/wechating/apps/constants"
	"github.com/feimumoke/wechating/framework/wcerror"
	"time"
)

const (
	defaultTimeout time.Duration = 120 * time.Second
)

// 上传到S3，并且文件名会加上时间戳，形成随机数
func UploadFileToS3(storeFolder constants.CDN_FOLDER, fileName string, fileBytes []byte) (string, *bmserror.BMSError) {
	fileNameWithTimestamp := fmt.Sprint(time.Now().Unix(), fileName)
	return uploadFileToS3WithTTL(storeFolder, fileNameWithTimestamp, fileBytes, constants.DefaultCDNFileTTL, true, nil)
}

// 上传到S3，并且文件名会加上时间戳，形成随机数。并且带过期时间
func UploadFileToS3WithTTL(storeFolder constants.CDN_FOLDER, fileName string, fileBytes []byte, tokenTTL int64) (string, *bmserror.BMSError) {
	fileNameWithTimestamp := fmt.Sprint(time.Now().Unix(), fileName)
	return uploadFileToS3WithTTL(storeFolder,
		fileNameWithTimestamp, fileBytes, tokenTTL, true, nil)
}

// 上传到S3,文件名会无需时间戳,带过期时间,需要外层确保文件名唯一
func UploadFileToS3WithTTLForExport(storeFolder constants.CDN_FOLDER, fileName string,
	fileBytes []byte, tokenTTL int64) (string, *bmserror.BMSError) {
	return uploadFileToS3WithTTL(storeFolder, fileName, fileBytes, tokenTTL, true, NewExportToken(fileName, tokenTTL))
}

// 上传到S3,文件名会无需时间戳,带默认过期时间,需要外层确保文件名唯一
func UploadFileToS3WithDefaultDDL(storeFolder constants.CDN_FOLDER, fileName string, fileBytes []byte) (string, *bmserror.BMSError) {
	return uploadFileToS3WithTTL(storeFolder, fileName, fileBytes, constants.DefaultCDNFileTTL, false, nil)
}

// 上传到S3，理解为将S3当成缓存使用
func UploadFileToS3ByKey(storeFolder constants.CDN_FOLDER, key string, fileBytes []byte) (string, *bmserror.BMSError) {
	return uploadFileToS3WithTTL(storeFolder, key, fileBytes, constants.DefaultCDNFileTTL, false, nil)
}

func uploadFileToS3WithTTL(storeFolder constants.CDN_FOLDER, fileName string,
	fileBytes []byte, tokenTTL int64, storePathWithDate bool, exportToken *ExportResource) (string, *bmserror.BMSError) {
	s3CdnPath := "xx"
	return s3CdnPath, nil
}

type ExportResource struct {
	FileName   string
	ExpireTime int64
}

func NewExportToken(fileName string, tokenTTL int64) *ExportResource {
	return &ExportResource{FileName: fileName,
		ExpireTime: time.Now().Unix() + tokenTTL, // 过期时间等于当前时间加上ttl
	}
}

func (e *ExportResource) GetStr() string {
	return fmt.Sprintf("%s %d", e.FileName, e.ExpireTime)
}
