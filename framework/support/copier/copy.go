package copier

import (
	"encoding/json"
)

const ErrParam = -200001 //Parameter Invalid

/*
*
1. dest 必须为指针引用，否则 Unmarshal 会失败
2. 错误由 json 包 Marshal 和 Unmarshal 返回即可
*/
func Copy(source interface{}, dest interface{}) error {

	sourceBytes, err := json.Marshal(source)
	if err != nil {
		return err
	}
	err = json.Unmarshal(sourceBytes, dest)
	if err != nil {
		return err
	}
	return nil
}

func CopyWithIgnore(source interface{}, dest interface{}, ignoreField []string) error {
	var srcMap map[string]interface{}
	sourceBytes, err := json.Marshal(source)
	if err != nil {
		return err
	}
	err = json.Unmarshal(sourceBytes, &srcMap)
	if err != nil {
		return err
	}

	for _, field := range ignoreField {
		delete(srcMap, field)
	}

	err = Copy(srcMap, dest)
	if err != nil {
		return err
	}
	return nil
}

// source不变，修改dest
// source中的属性优先，覆盖dest中的属性
func MergeWithIgnore(source interface{}, dest interface{}, ignoreField []string) error {
	sourceBytes, err := json.Marshal(source)
	if err != nil {
		return err
	}
	var srcMap map[string]interface{}
	err = json.Unmarshal(sourceBytes, &srcMap)
	if err != nil {
		return err
	}

	for _, field := range ignoreField {
		delete(srcMap, field)
	}

	destBytes, err := json.Marshal(dest)
	if err != nil {
		return err
	}
	var destMap map[string]interface{}
	err = json.Unmarshal(destBytes, &destMap)
	if err != nil {
		return err
	}

	for key, field := range destMap {
		if _, ok := srcMap[key]; !ok {
			srcMap[key] = field
		}
	}

	err = Copy(srcMap, dest)
	if err != nil {
		return err
	}
	return nil
}

func Merge(source interface{}, dest interface{}) error {
	sourceBytes, err := json.Marshal(source)
	if err != nil {
		return err
	}
	var srcMap map[string]interface{}
	err = json.Unmarshal(sourceBytes, &srcMap)
	if err != nil {
		return err
	}

	destBytes, err := json.Marshal(dest)
	if err != nil {
		return err
	}
	var destMap map[string]interface{}
	err = json.Unmarshal(destBytes, &destMap)
	if err != nil {
		return err
	}

	for key, field := range destMap {
		if _, ok := srcMap[key]; !ok {
			srcMap[key] = field
		}
	}

	err = Copy(srcMap, dest)
	if err != nil {
		return err
	}
	return nil
}

// 根据json标签，生成map
// json序列化的方式
func CopyToJsonMap(source interface{}) (map[string]interface{}, error) {
	var srcMap map[string]interface{}
	sourceBytes, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(sourceBytes, &srcMap)
	if err != nil {
		return nil, err
	}
	return srcMap, nil
}
