package expression

import (
	"reflect"
)

const (
	MaxInt64 = 1<<63 - 1
	MinInt64 = -1 << 63
	MaxInt   = int(^uint(0) >> 1) //     最大值，根据二进制补码，第一位为0，其余为1
	MINInt   = ^MaxInt            // 最小值，第一位为1，其余为0，最大值取反即可
)

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func IfInt64(condition bool, trueVal, falseVal int64) int64 {
	if condition {
		return trueVal
	}
	return falseVal
}

func IfString(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func Max(nums ...int64) int64 {
	var maxNum int64 = MinInt64
	for _, num := range nums {
		if num > maxNum {
			maxNum = num
		}
	}
	return maxNum
}

func Min(nums ...int64) int64 {
	var minNum int64 = MaxInt64
	for _, num := range nums {
		if num < minNum {
			minNum = num
		}
	}
	return minNum
}

func MinInt(nums ...int) int {
	var minNum = MaxInt
	for _, num := range nums {
		if num < minNum {
			minNum = num
		}
	}
	return minNum
}

func Default(src, def interface{}) interface{} {
	// 确保src 和def是相同的指针类型
	if src == nil || reflect.ValueOf(src).IsNil() {
		return def
	}
	return src
}

func Abs(num int64) int64 {
	if num < 0 {
		return -num
	}
	return num
}

func IfNilUseDefaultInt64(pointer *int64, defaultVal int64) int64 {
	if pointer == nil {
		return defaultVal
	}
	return *pointer
}

// 判断nums是否每个元素都在assertFunc中返回true
func AllMatch(assertFunc func(int64) bool, nums ...int64) bool {
	for _, n := range nums {
		if !assertFunc(n) {
			return false
		}
	}
	return true
}
