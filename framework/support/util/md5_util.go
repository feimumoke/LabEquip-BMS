package util

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
)

// 返回一个32位md5加密后的字符串
func GetStructMD5Encode(data interface{}) (string, *bmserror.BMSError) {
	bytes, jsonErr := json.Marshal(data)
	if jsonErr != nil {
		return "", bmserror.NewError(constant.ErrJsonEncodeFail, jsonErr.Error())
	}
	str := string(bytes)
	md5Str, err := GetMD5Encode(str)
	if err != nil {
		return "", err.Mark()
	}
	return md5Str, nil
}

// 返回一个32位md5加密后的字符串
func GetMD5Encode(data string) (string, *bmserror.BMSError) {
	h := md5.New()
	_, werr := h.Write([]byte(data))
	if werr != nil {
		return "", bmserror.NewError(constant.ErrParam, "md5 write: %v", werr.Error())
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// 返回一个16位md5加密后的字符串
func Get16MD5Encode(data string) (string, *bmserror.BMSError) {
	md5Str, err := GetMD5Encode(data)
	if err != nil {
		return "", err.Mark()
	}

	return md5Str[8:24], nil
}

// 返回一个128位SHA512加密后的字符串
func GetSha512Encode(data string) (string, *bmserror.BMSError) {
	h := sha512.New()
	_, err := h.Write([]byte(data))
	if err != nil {
		return "", bmserror.NewError(constant.ErrParam, "sha512 write: %v", err.Error())
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// 返回一个128位SHA256加密后的字符串
func GetSha256Encode(data string) (string, *bmserror.BMSError) {
	h := sha256.New()
	_, err := h.Write([]byte(data))
	if err != nil {
		return "", bmserror.NewError(constant.ErrParam, "sha256 write: %v", err.Error())
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
