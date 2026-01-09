package convert

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
)

func floatToString(f float64) string {
	return strconv.FormatFloat(f, 'E', -1, 64)
}
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}
func boolToString(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

func ToString(o interface{}) string {
	return fmt.Sprintf("%+v", o)
}

func StringToInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

func StringToInt64ByDefault(str string, def int64) int64 {
	if str == "" {
		return def
	}
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return def
	}
	return val
}

func StringToFloat(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func StringListToIntList(str []string) ([]int64, error) {

	result := make([]int64, 0)
	for _, s := range str {
		intValue, convertErr := StringToInt64(s)
		if convertErr != nil {
			return nil, convertErr
		}
		result = append(result, intValue)
	}

	return result, nil
}

func NumberRound(number float64, precision int) float64 {
	style := fmt.Sprintf(".%vf", precision)
	value, _ := strconv.ParseFloat(fmt.Sprintf("%"+style, number), 64)
	return value
}

func NumberPercentRound(number float64, precision int) string {
	style := fmt.Sprintf(".%vf", precision-2)
	return fmt.Sprintf("%"+style, number*100) + "%"
}

// StringToInt64Array [0,1,2]或者0,1,2
func StringToInt64Array(str string) ([]int64, *bmserror.BMSError) {
	res := make([]int64, 0)
	if len(str) == 0 {
		return res, nil
	}
	if str[0:1] == "[" && len(str) < 2 {
		return nil, bmserror.NewError(constant.ErrParam, "covert string to slice error")
	}
	if str[0:1] == "[" {
		str = str[1 : len(str)-1]
	}
	data := strings.Split(str, ",")
	for _, item := range data {
		i, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			return nil, bmserror.NewError(constant.ErrParam, err.Error())
		}
		res = append(res, i)
	}
	return res, nil
}
