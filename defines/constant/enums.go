package constant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

var EnumValueMap = make(map[string]interface{})
var EnumOrderValueMap = make(map[string][]EnumKeyValue)
var EnumKeyToValueMap = make(map[string]map[string]int64)
var EnumValueToKeyMap = make(map[string]map[int64]string)

func init() {
	// 最常用的枚举，慧君看到有一些YES和NO的枚举，但名字是指定场景的，不能随便使用，否则一旦他修改了，我就受影响了
	// 如果这个枚举不满足你的要求，不要修改这里，你另外写一个
	RegisterEnumValues("CommonEnum", map[string]interface{}{
		"NO":  0,
		"YES": 1,
	})
	RegisterEnumOrderValues("CommonEnumYAndN", []EnumKeyValue{
		{Key: "Y", Value: 1},
		{Key: "N", Value: 0},
	})
}

type EnumKeyValueList []EnumKeyValue

// 自定义json格式化：将数组转化为字典类型
func (enums EnumKeyValueList) MarshalJSON() ([]byte, error) {
	enumList := make([]string, 0)
	for _, orderEnum := range enums {
		valJson, err := json.Marshal(orderEnum.Value)
		if err != nil {
			return nil, err
		}
		enumList = append(enumList, fmt.Sprintf(`"%s":%s`, orderEnum.Key, string(valJson)))
	}
	return []byte("{" + strings.Join(enumList, ",") + "}"), nil
}

func RegisterEnumSortValuesWithoutCamel(key string, values EnumKeyValueList) {
	RegisterEnumSortValuesRoot(key, values, "", false)
}

// 注册有顺序的枚举返回给前端
func RegisterEnumSortValues(key string, values EnumKeyValueList) {
	RegisterEnumSortValuesBySplit(key, values, "_")
}

func RegisterEnumSortValuesBySplit(key string, values EnumKeyValueList, split string) {
	RegisterEnumSortValuesRoot(key, values, split, true)
}

func RegisterEnumSortValuesRoot(key string, values EnumKeyValueList, split string, useCamel bool) {
	if _, ok := EnumValueMap[key]; ok {
		panic(fmt.Sprintf("%s enum key has existed", key))
	}
	if useCamel {
		tValue := make([]EnumKeyValue, 0, len(values))
		for _, one := range values {
			k2 := camel2Case(one.Key, split)
			tValue = append(tValue, EnumKeyValue{
				Key:   k2,
				Value: one.Value,
			})
		}
		EnumValueMap[key] = EnumKeyValueList(tValue)
	} else {
		EnumValueMap[key] = values
	}
}

type EnumKeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func RegisterEnumOrderValuesWithoutCamel(key string, values []EnumKeyValue) {
	RegisterEnumOrderValuesRoot(key, values, "", false)
}

// 注册有顺序的枚举返回给前端
func RegisterEnumOrderValues(key string, values []EnumKeyValue) {
	RegisterEnumOrderValuesBySplit(key, values, "_")
}

func RegisterEnumOrderValuesBySplit(key string, values []EnumKeyValue, split string) {
	RegisterEnumOrderValuesRoot(key, values, split, true)
}

func RegisterEnumOrderValuesRoot(key string, values []EnumKeyValue, split string, useCamel bool) {
	if _, ok := EnumValueMap[key]; ok {
		panic(fmt.Sprintf("%s enum key has existed", key))
	}

	if useCamel {
		tValue := make([]EnumKeyValue, 0, len(values))
		for _, one := range values {
			k2 := camel2Case(one.Key, split)
			tValue = append(tValue, EnumKeyValue{
				Key:   k2,
				Value: one.Value,
			})
		}
		EnumValueMap[key] = tValue
		EnumOrderValueMap[key] = tValue
	} else {
		EnumValueMap[key] = values
		EnumOrderValueMap[key] = values
	}
}

func RegisterEnumValues(key string, values map[string]interface{}) {
	RegisterEnumValuesBySplit(key, values, "_")
}

func RegisterEnumValuesBySplit(key string, values map[string]interface{}, split string) {
	RegisterEnumValuesRoot(key, values, split, true)
}

// key保持原样，不做任何修改
func RegisterEnumValuesWithoutCamel(key string, values map[string]interface{}) {
	RegisterEnumValuesRoot(key, values, "", false)
}

func RegisterEnumValuesRoot(key string, values map[string]interface{}, split string, useCamel bool) {
	if _, ok := EnumValueMap[key]; ok {
		panic(fmt.Sprintf("%s enum key has existed", key))
	}

	tValue := make(map[string]interface{})
	if useCamel {
		for k, v := range values {
			k2 := camel2Case(k, split)
			tValue[k2] = v
		}
		EnumValueMap[key] = tValue
	} else {
		EnumValueMap[key] = values
	}
}

func RegisterEnumValuesWithDigit(key string, values map[string]interface{}, split string, useCamel bool) {
	if _, ok := EnumValueMap[key]; ok {
		panic(fmt.Sprintf("%s enum key has existed", key))
	}

	tValue := make(map[string]interface{})
	if useCamel {
		for k, v := range values {
			k2 := camel2CaseWithDigit(k, split)
			tValue[k2] = v
		}
		EnumValueMap[key] = tValue
	} else {
		EnumValueMap[key] = values
	}
}

func GetEnumValues() map[string]interface{} {
	return EnumValueMap
}

func GetOrderEnumValues() map[string][]EnumKeyValue {
	return EnumOrderValueMap
}

// camel2Case 私有方法驼峰转大写+指定分隔符
func camel2Case(name, split string) string {
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
			buffer.WriteRune(unicode.ToUpper(r))
			beforeUpper = false
			continueUpper = 0
		}
	}
	return buffer.String()
}

// camel2CaseWithDigit 私有方法驼峰转大写+包含数字
func camel2CaseWithDigit(name, split string) string {
	buffer := bytes.NewBufferString("")
	beforeUpper := false
	continueUpper := 0
	for i, r := range name {
		if unicode.IsUpper(r) || unicode.IsDigit(r) {
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
			buffer.WriteRune(unicode.ToUpper(r))
			beforeUpper = false
			continueUpper = 0
		}
	}
	return buffer.String()
}

func RegisterEnumValuesV2(key string, values map[string]int64) {
	if _, ok := EnumValueMap[key]; ok {
		panic(fmt.Sprintf("%s enum key has existed", key))
	}

	tValue := make(map[string]int64)
	valueToKeyMap := make(map[int64]string)
	for k, v := range values {
		k2 := camel2Case(k, "_")
		tValue[k2] = v
		valueToKeyMap[v] = k
	}
	EnumValueMap[key] = tValue
	EnumKeyToValueMap[key] = values
	EnumValueToKeyMap[key] = valueToKeyMap
}

func IsEnumExisted(key string, value int64) bool {
	valueToKeyMap, existed := EnumValueToKeyMap[key]
	if !existed {
		return false
	}
	_, existed = valueToKeyMap[value]
	return existed
}

func GetEnumName(key string, value int64) string {
	return EnumValueToKeyMap[key][value]
}

func GetEnumValuesByEnumKey(key string) interface{} {
	return EnumValueMap[key]
}
