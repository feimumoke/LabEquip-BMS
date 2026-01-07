package collection

import (
	"reflect"
	"strings"
)

func Contain(targetObj interface{}, collection interface{}) bool {
	collectionValues := reflect.ValueOf(collection)
	switch reflect.TypeOf(collection).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < collectionValues.Len(); i++ {
			if collectionValues.Index(i).Interface() == targetObj {
				return true
			}
		}
	case reflect.Map:
		if collectionValues.MapIndex(reflect.ValueOf(targetObj)).IsValid() {
			return true
		}
	}
	return false
}

func ContainNot(targetObj interface{}, collection interface{}) bool {
	return !Contain(targetObj, collection)
}

func MapGetOrDefault(targetMap interface{}, key interface{}, defaultValue interface{}) interface{} {
	collectionValues := reflect.ValueOf(targetMap)
	if reflect.TypeOf(targetMap).Kind() != reflect.Map {
		return defaultValue
	}

	if collectionValues.MapIndex(reflect.ValueOf(key)).IsValid() {
		return collectionValues.MapIndex(reflect.ValueOf(key)).Interface()
	}
	return defaultValue
}

func IsEmptyStringSlice(collection []string) bool {
	if collection == nil {
		return true
	}
	if len(collection) < 1 {
		return true
	}
	if len(collection) == 1 && strings.Trim(collection[0], " ") == "" {
		return true
	}
	return false
}

func IsNotEmptyStringSlice(collection []string) bool {
	return !IsEmptyStringSlice(collection)
}

/*
特别注意：当输入的其中一个数组是nil时，这个数组会被忽略。如果输入的其中一个数组是空的slice，则会做交集操作，最终一定会返回空数组
*/
func GetInterStringList(items ...[]string) []string {
	// 获得多个items的交集，可能返回的nil
	var dst []string
	for idx, currItem := range items {
		if idx == 0 {
			dst = currItem
			continue
		}
		dst = getInterStringList(dst, currItem)
	}
	return dst
}

func getInterStringList(a1 []string, a2 []string) []string {
	// 两个slice的交集，可能返回nil
	if a1 == nil {
		return a2
	}
	if a2 == nil {
		return a1
	}
	s := NewStringSet(a1...)
	return s.InterSet(NewStringSet(a2...)).ToSlice()
}

func GetUnionStringList(items ...[]string) []string {
	// 获得多个items的并集，可能返回的nil，需要保持数组内字符串的顺序
	var dst []string
	for idx, currItem := range items {
		if idx == 0 {
			dst = currItem
			continue
		}
		dst = getUnionStringList(dst, currItem)
	}
	return dst
}

func getUnionStringList(a1 []string, a2 []string) []string {
	if len(a1) == 0 {
		return a2
	}
	if len(a2) == 0 {
		return a1
	}
	s := NewStringSet(a1...)
	for _, item := range a2 {
		if s.Contains(item) {
			continue
		}
		a1 = append(a1, item)
	}
	return a1
}

// 求差集
func GetDifferenceStringList(list1, list2 []string) []string {
	if len(list1) == 0 {
		return list1
	}
	if len(list2) == 0 {
		return list1
	}
	set := NewStringSet(list2...)
	ans := make([]string, 0)
	for _, str := range list1 {
		if !set.Contains(str) {
			ans = append(ans, str)
		}
	}
	return ans
}
