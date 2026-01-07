package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/feimumoke/wechating/framework/constant"
	"github.com/feimumoke/wechating/framework/wcerror"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// 向右补充0 对标python zfill函数
func RightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

// 向左补充0
func LeftPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

// 向左补充0，超出部分不截断
func LeftPadCharWithMinLen(s string, padChar byte, minLen int) string {
	padCountInt := minLen - len(s)
	if padCountInt <= 0 {
		return s
	}
	padStr := string(padChar)
	retStr := strings.Repeat(padStr, padCountInt) + s
	return retStr
}

// 数组的字符串转化为数组字符串
func String2Int64Arry(arryString string) ([]int64, *bmserror.BMSError) {
	result := make([]int64, 0)
	// 传""会报错
	if arryString == "" {
		return result, nil
	}
	err := json.Unmarshal([]byte(arryString), &result)
	if err != nil {
		return nil, bmserror.NewError(constant.ErrParam, err.Error())
	}
	return result, nil
}

// int的数组字符串转化为数组
func String2StringArry(arryString string) ([]string, *bmserror.BMSError) {
	result := make([]string, 0)
	if arryString == "" {
		return result, nil
	}
	err := json.Unmarshal([]byte(arryString), &result)
	if err != nil {
		return nil, bmserror.NewError(constant.ErrParam, err.Error())
	}
	return result, nil
}

// 判断字符串是否为数字字符
func IsNumber(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

// 判断是否为十进制字符
func IsDigit(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func JoinIgnoreEmpty(in []string, sep string) string {
	inIgnoreEmpty := make([]string, 0)
	for _, s := range in {
		if s == "" {
			continue
		}
		inIgnoreEmpty = append(inIgnoreEmpty, s)
	}
	return strings.Join(inIgnoreEmpty, sep)
}

// 是否满足正则表达式
func SatisfyRegex(code string, regexStr string) bool {
	reg := regexp.MustCompile(regexStr)
	return reg.MatchString(code)
}

// 获取最后几位string
func GetLastRuneStr(s string, c int64) string {
	r := []rune(s)
	if c > int64(len(r)) {
		c = int64(len(r))
	}
	return string(r[int64(len(r))-c:])
}

// 获取前几位string
func GetFirstRuneStr(s string, c int64) string {
	r := []rune(s)
	if c >= int64(len(r)) {
		c = int64(len(r))
	}
	return string(r[:c])
}

func StringNumberIsPositiveInteger(number string) bool {
	//转换出错的时候会返回0，可以不用处理err
	numF, _ := strconv.ParseFloat(number, 10)
	ceilNum := math.Ceil(numF)
	//大于0，且向上取整之后和原来的相等
	return numF > 0 && ceilNum == numF
}
func StringNumberIsInteger(number string) bool {
	//转换出错的时候会返回0，可以不用处理err
	numF, _ := strconv.ParseFloat(number, 10)
	ceilNum := math.Ceil(numF)
	//大于0，且向上取整之后和原来的相等
	return numF >= 0 && ceilNum == numF
}

func StructToString(i interface{}) string {
	b, _ := json.Marshal(i)
	return string(b)
}

const CaseTypeKeepOrignal = 0
const CaseTypeKeepUpper = 1
const CaseTypeKeepLower = 2

// case 0 do nothing,case 1 Upper ,case 2 Lower
func Camel2Case(name, split string, caseType int) string {
	buffer := bytes.NewBufferString("")
	beforeUpper := false
	continueUpper := 0
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 && !beforeUpper { //前一个字符为小写
				buffer.WriteString(split)
			}
			if i != 0 && continueUpper > 1 && i < len(name)-1 && unicode.IsLower([]rune(name)[i+1]) { //专属名称结束
				buffer.WriteString(split)
			}
			buffer.WriteRune(r)
			beforeUpper = true
			continueUpper++
		} else {
			switch caseType {
			case CaseTypeKeepOrignal:
				buffer.WriteRune(r)
			case CaseTypeKeepUpper:
				buffer.WriteRune(unicode.ToUpper(r))
			case CaseTypeKeepLower:
				buffer.WriteRune(unicode.ToLower(r))
			}
			beforeUpper = false
			continueUpper = 0
		}
	}
	return buffer.String()
}

// 随机生成字符串, 这个方法是以时间戳作为seed，然后再加上一个随机数，来作为随机因子来进行随机。
// 要求永不重复 建议length设置15位到18位之间
func RandStr(length int) (string, *bmserror.BMSError) {
	// int64长度要求是最大19位
	if length > 18 {
		return "", bmserror.NewError(constant.ErrParam, "The length of the generated string exceeds the upper limit！")
	}
	str := "0123456789"
	bytesBuf := []byte(str)
	result := []byte{}
	rand.Seed(time.Now().UnixNano() + int64(rand.Intn(10000)))
	for i := 0; i < length; i++ {
		result = append(result, bytesBuf[rand.Intn(len(bytesBuf))])
	}
	return string(result), nil
}

func CombineKey(splitKey string, objs ...interface{}) string {
	if len(splitKey) == 0 {
		splitKey = ";"
	}
	var key string
	for _, obj := range objs {
		key = fmt.Sprintf("%s%s%v", key, splitKey, obj)
	}
	key = key[1:]
	return key
}

func JsonToStringSlice(jsonStr string) []string {
	mSlice := make([]string, 0)
	if jsonStr == "" || len(jsonStr) == 0 {
		return mSlice
	}
	_ = json.Unmarshal([]byte(jsonStr), &mSlice)
	return mSlice
}
