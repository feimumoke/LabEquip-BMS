package convert

import (
	"fmt"
	"reflect"
)

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

// StringValue returns the value of the string pointer passed in or
// "" if the pointer is nil.
func StringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// BoolValue returns the value of the bool pointer passed in or
// false if the pointer is nil.
func BoolValue(v *bool) bool {
	if v != nil {
		return *v
	}
	return false
}

// Int returns a pointer to the int value passed in.
func Int(v int) *int {
	return &v
}

// Uint returns a pointer to the uint value passed in.
func Uint(v uint) *uint {
	return &v
}

// UintValue returns the value of the uint pointer passed in or
// 0 if the pointer is nil.
func UintValue(v *uint) uint {
	if v != nil {
		return *v
	}
	return 0
}

// Int8 returns a pointer to the int8 value passed in.
func Int8(v int8) *int8 {
	return &v
}

// Int8Value returns the value of the int8 pointer passed in or
// 0 if the pointer is nil.
func Int8Value(v *int8) int8 {
	if v != nil {
		return *v
	}
	return 0
}

// Int8Slice converts a slice of int8 values into a slice of
// int8 pointers
func Int8Slice(src []int8) []*int8 {
	dst := make([]*int8, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// Int8ValueSlice converts a slice of int8 pointers into a slice of
// int8 values
func Int8ValueSlice(src []*int8) []int8 {
	dst := make([]int8, len(src))
	for i := 0; i < len(src); i++ {
		if src[i] != nil {
			dst[i] = *(src[i])
		}
	}
	return dst
}

// Int16 returns a pointer to the int16 value passed in.
func Int16(v int16) *int16 {
	return &v
}

// Int16Value returns the value of the int16 pointer passed in or
// 0 if the pointer is nil.
func Int16Value(v *int16) int16 {
	if v != nil {
		return *v
	}
	return 0
}

// Int32 returns a pointer to the int32 value passed in.
func Int32(v int32) *int32 {
	return &v
}

// Int32 returns a pointer to the int32 value passed in.
func Int64ToInt32(v int64) *int32 {
	i := int32(v)
	return &i
}

// Int32Value returns the value of the int32 pointer passed in or
// 0 if the pointer is nil.
func Int32Value(v *int32) int32 {
	if v != nil {
		return *v
	}
	return 0
}

// Int64 returns a pointer to the int64 value passed in.
func Int64(v int64) *int64 {
	return &v
}

// Int64Value returns the value of the int64 pointer passed in or
// 0 if the pointer is nil.
func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}
	return 0
}

// uint64 returns a pointer to the uint64 value passed in.
func UInt64(v uint64) *uint64 {
	return &v
}

// uint64 returns the value of the uint64 pointer passed in or
// 0 if the pointer is nil.
func UInt64Value(v *uint64) uint64 {
	if v != nil {
		return *v
	}
	return 0
}

// uint64 returns a pointer to the uint64 value passed in.
func UInt32(v uint32) *uint32 {
	return &v
}

// uint64 returns the value of the uint64 pointer passed in or
// 0 if the pointer is nil.
func UInt32Value(v *uint32) uint32 {
	if v != nil {
		return *v
	}
	return 0
}

// Float64 returns a pointer to the float64 value passed in.
func Float64(v float64) *float64 {
	return &v
}

// Float64Value returns the value of the float64 pointer passed in or
// 0 if the pointer is nil.
func Float64Value(v *float64) float64 {
	if v != nil {
		return *v
	}
	return 0
}

// Float64 returns a pointer to the float64 value passed in.
func Float32(v float32) *float32 {
	return &v
}

// Float64Value returns the value of the float64 pointer passed in or
// 0 if the pointer is nil.
func Float32Value(v *float32) float32 {
	if v != nil {
		return *v
	}
	return 0
}

// .
func Interface(v interface{}) interface{} {
	return v
}

func StructToMapWithContainFields(in interface{}, tagName string, containFields []string) (map[string]interface{}, error) {
	out, err := StructToMap(in, tagName)
	if err != nil {
		return nil, err
	}
	newOut := map[string]interface{}{}
	for _, field := range containFields {
		newOut[field] = out[field]
	}
	return newOut, nil
}

func StructToMapWithIgnoreFields(in interface{}, tagName string, ignoreFields []string) (map[string]interface{}, error) {
	out, err := StructToMap(in, tagName)
	if err != nil {
		return nil, err
	}
	for _, field := range ignoreFields {
		delete(out, field)
	}
	return out, nil
}

// ToMap 结构体转为Map[string]interface{}
func StructToMap(in interface{}, tagName string) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct { // 非结构体返回错误提示
		return nil, fmt.Errorf("ToMap only accepts struct or struct pointer; got %T", v)
	}

	t := v.Type()
	// 遍历结构体字段
	// 指定tagName值为map中key;字段值为map中value
	for i := 0; i < v.NumField(); i++ {
		fi := t.Field(i)
		if tagValue := fi.Tag.Get(tagName); tagValue != "" {
			out[tagValue] = v.Field(i).Interface()
		}
	}
	return out, nil
}

//func ByteToPString(in []byte) *string {
//	// str和bytes共用一片内存，这样做的好处是减少内存拷贝
//	return (*string)(unsafe.Pointer(&in))
//}
