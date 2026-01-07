package util

import (
	"archive/zip"
	"encoding/json"
	"github.com/feimumoke/wechating/framework/constant"
	"github.com/feimumoke/wechating/framework/log"
	"github.com/feimumoke/wechating/framework/wcerror"
	"io"
	"os"
)

func GetPageOffset(pageNo int64, count int64) int64 {
	var pageNoInt int64 = 1
	var countInt int64 = 10

	if pageNo > 0 {
		pageNoInt = pageNo
	}

	if count > 0 {
		countInt = count
	}

	return (pageNoInt - 1) * countInt
}

const UploadFileTokenTTL = 24 * 3600 * 30

// 压缩工具
// 输入zip文件名, 待压缩文件列表
// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func ArchiveZipFiles(filename string, files []*FileInfo) *bmserror.BMSError {
	newZipFile, cErr := os.Create(filename)
	if cErr != nil {
		return bmserror.NewError(constant.ErrInternalServer, cErr.Error())
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if aErr := AddFileToZip(zipWriter, file); aErr != nil {
			return aErr.Mark()
		}
	}
	return nil
}

type FileInfo struct {
	FilePath string
	FileName string
}

func AddFileToZip(zipWriter *zip.Writer, fileInfo *FileInfo) *bmserror.BMSError {
	fileToZip, err := os.Open(fileInfo.FilePath)
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = fileInfo.FileName

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	_, err = io.Copy(writer, fileToZip)
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, err.Error())
	}
	return nil
}

func ToJSON(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		log.Errorf("json marshal err:%v", err.Error())
	}
	return string(b)
}

func GetProvinceByPtID(ptID string) string {
	return ptID[0:2]
}
