package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/gob"
	"github.com/feimumoke/wechating/framework/constant"
	"github.com/feimumoke/wechating/framework/wcerror"
	"sync"
)

func init() {
	gob.Register([]byte{})
}

const (
	errTmp = "[encrypt/decrypt err] %v\n"
)

var (
	aesSecretOnce  sync.Once
	piiSecretCache []byte
	piiIvCache     []byte
)

// RawAes128CbcEncrypt aes 128 cbc
func RawAes128CbcEncrypt(raw string, secret, iv []byte) (string, *bmserror.BMSError) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", bmserror.NewError(constant.ErrParam, errTmp, "aes new cipher fail")
	}
	cbc := cipher.NewCBCEncrypter(block, iv)
	padding := pkcs7Padding([]byte(raw), 16)
	enc := make([]byte, len(padding))
	cbc.CryptBlocks(enc, padding)
	// base64 防止返回值乱码
	return base64.StdEncoding.EncodeToString(enc), nil
}

// RawAes128CbcDecrypt aes 128 cbc
func RawAes128CbcDecrypt(encryptData string, secret, iv []byte) (string, *bmserror.BMSError) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", bmserror.NewError(constant.ErrParam, errTmp, "aes new cipher fail")
	}
	// 加密后后的字符串是经过base64编码的 , 因此需要 base64 解码
	enc, decrypt64Err := base64.StdEncoding.DecodeString(encryptData)
	if decrypt64Err != nil {
		return "", bmserror.NewError(constant.ErrParam, errTmp, decrypt64Err.Error())
	}
	cbc := cipher.NewCBCDecrypter(block, iv)
	dec := make([]byte, len(enc))
	cbc.CryptBlocks(dec, enc)
	unPadding, err := pkcs7UnPadding(dec)
	if err != nil {
		return "", bmserror.NewError(constant.ErrParam, errTmp, "aes decrypt un_padding fail")
	}
	return string(unPadding), nil
}

// pkcs7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	//判断缺少几位长度。最少1，最多 blockSize
	padding := blockSize - len(data)%blockSize
	//补足位数。把切片[]byte{byte(padding)}复制padding个
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7UnPadding 填充的反向操作
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		// todo err
		return nil, nil
	}
	//获取填充的个数
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}
