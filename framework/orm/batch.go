package orm

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/collection"
	"github.com/feimumoke/labequipbms/framework/support/iters"
)

const batchCreateMaxNum = 100

func (o *GormDB) BatchUpdate(tableName string, itemList interface{}) (int64, *bmserror.BMSError) {

	affectRow := int64(0)
	sqlArray, err := GenerateBatchUpdateSQL(tableName, itemList)
	if err != nil {
		return affectRow, err.Mark()
	}
	for _, sqlStr := range sqlArray {
		db := o.Exec(sqlStr)
		if err := db.GetError(); err != nil {
			return affectRow, bmserror.NewError(constant.ErrDB, err.Error())
		}
		affectRow += db.RowsAffected()
	}
	return affectRow, nil
}

// 单条更新 传入一个结构体的指针 返回一个map(要更新的字段=》要更新的值)
func GenerateUpdateMap(item interface{}) (map[string]interface{}, *bmserror.BMSError) {
	updateMap := map[string]interface{}{}
	t := reflect.TypeOf(item)
	if t.Kind() != reflect.Ptr {
		return nil, bmserror.NewError(constant.ErrInternalServer, "param is not pointer")
	}

	if t.Elem().Kind() != reflect.Struct {
		return nil, bmserror.NewError(constant.ErrInternalServer, "param is not *struct pointer")
	}
	fieldType := t.Elem()
	fieldNum := fieldType.NumField()

	fieldValue := reflect.ValueOf(item).Elem()
	for i := 0; i < fieldNum; i++ {
		elem := fieldValue.Field(i)
		gormTag := fieldType.Field(i).Tag.Get("gorm")
		fieldName := GetFieldName(gormTag)
		if strings.HasPrefix(fieldName, "id;") { //id是主键不需要放到里面
			continue
		}

		if len(strings.TrimSpace(fieldName)) == 0 {
			return nil, bmserror.NewError(constant.ErrParam, "the structure attribute should have tag")
		}
		if elem.Kind() == reflect.Ptr && elem.IsNil() {
			continue
		} else {
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			switch elem.Kind() {
			case reflect.Int64:
				updateMap[fieldName] = elem.Int()
			case reflect.String:
				updateMap[fieldName] = elem.String()
				//if strings.Contains(elem.String(), "'") {
				//	temp = fmt.Sprintf("'%v'", strings.ReplaceAll(elem.String(), "'", "\\'"))
				//} else {
				//	temp = fmt.Sprintf("'%v'", elem.String())
				//}
			case reflect.Float64:
				updateMap[fieldName] = elem.Float()
			case reflect.Bool:
				updateMap[fieldName] = elem.Bool()
			default:
				return nil, bmserror.NewError(constant.ErrParam, "type conversion error, param is %v", fieldType.Field(i).Tag.Get("json"))
			}
		}
	}
	if len(updateMap) == 0 {
		return nil, bmserror.NewError(constant.ErrParam, "type conversion error, param is all null")
	}
	return updateMap, nil
}
func GenerateBatchUpdateSQL(tableName string, itemList interface{}) ([]string, *bmserror.BMSError) {

	errCheck := CheckItemListType(itemList)
	if errCheck != nil {
		return nil, errCheck.Mark()
	}

	fieldValue := reflect.ValueOf(itemList)
	fieldType := reflect.TypeOf(itemList).Elem().Elem()
	sliceLength := fieldValue.Len()
	fieldNum := fieldType.NumField()

	// 检验结构体标签是否为空和重复
	verifyTagDuplicate := make(map[string]string)
	count := 0
	for i := 0; i < fieldNum; i++ {
		fieldTag := fieldType.Field(i).Tag.Get("gorm")

		fieldName := GetFieldName(fieldTag)
		if len(strings.TrimSpace(fieldName)) == 0 {
			return nil, bmserror.NewError(constant.ErrParam, "the structure attribute should have tag")
		}

		if strings.HasPrefix(fieldName, "id;") {
			count++
		}

		_, ok := verifyTagDuplicate[fieldName]
		if !ok {
			verifyTagDuplicate[fieldName] = fieldName
		} else {
			return nil, bmserror.NewError(constant.ErrParam, "the structure attribute %v tag is not allow duplication", fieldName)
		}

	}

	if count != 1 {
		return nil, bmserror.NewError(constant.ErrParam, "the structure attribute should have a primary key")
	}

	IDSet := collection.NewStringSet()
	var IDList []string
	updateMap := make(map[string][]*string)
	for i := 0; i < sliceLength; i++ {
		// 得到某一个具体的结构体的
		structValue := fieldValue.Index(i).Elem()
		for j := 0; j < fieldNum; j++ {
			elem := structValue.Field(j)
			gormTag := fieldType.Field(j).Tag.Get("gorm")
			fieldTag := GetFieldName(gormTag)

			if elem.Kind() == reflect.Ptr && elem.IsNil() {
				updateMap[fieldTag] = append(updateMap[fieldTag], nil) // 如果为nil的指针，则填入nil，保持每个field中数组数量的一致性
			} else {
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				var temp string
				switch elem.Kind() {
				case reflect.Int64:
					temp = strconv.FormatInt(elem.Int(), 10)
				case reflect.String:
					if strings.Contains(elem.String(), "'") {
						temp = fmt.Sprintf("'%v'", strings.ReplaceAll(elem.String(), "'", "\\'"))
					} else {
						temp = fmt.Sprintf("'%v'", elem.String())
					}
				case reflect.Float64:
					temp = strconv.FormatFloat(elem.Float(), 'f', -1, 64)
				case reflect.Bool:
					temp = strconv.FormatBool(elem.Bool())
				default:
					return nil, bmserror.NewError(constant.ErrParam, "type conversion error, param is %v", fieldType.Field(j).Tag.Get("json"))
				}

				if strings.HasPrefix(fieldTag, "id;") {
					id, err := strconv.ParseInt(temp, 10, 64)
					if err != nil {
						return nil, bmserror.NewError(constant.ErrParam, err.Error())
					}
					// id 的合法性校验
					if id < 1 {
						return nil, bmserror.NewError(constant.ErrParam, "this structure should have a primary key and gt 0")
					}
					if IDSet.Contains(temp) {
						return nil, bmserror.NewError(constant.ErrParam, "this structure data id can not repeat: %v", temp)
					}
					IDSet.Add(temp)
					IDList = append(IDList, temp)
					continue
				}
				updateMap[fieldTag] = append(updateMap[fieldTag], &temp)
			}

		}
	}
	// 过滤掉 updateMap 中都是 nil 的字段，不用更新
	for fieldTag, valList := range updateMap {
		isFieldAllNil := true
		for _, val := range valList {
			if val != nil {
				isFieldAllNil = false
				break
			}
		}
		if isFieldAllNil { // 如果全为nil的，废弃掉
			delete(updateMap, fieldTag)
		}
	}

	var newIDList []string
	iters.From(IDList).Select(func(i interface{}) interface{} {
		return i.(string)
	}).Distinct().ToSlice(&newIDList)

	if len(IDList) != len(newIDList) {
		var repeatedIDList []string
		iters.From(IDList).Except(iters.From(newIDList)).ToSlice(&repeatedIDList)
		return nil, bmserror.NewError(constant.ErrParam, "this structure data id %v can not repeat", strings.Join(repeatedIDList, ","))
	}

	length := len(IDList)
	size := batchCreateMaxNum
	SQLQuantity := getSQLQuantity(length, size)
	var SQLArray []string
	k := 0

	updateFieldCount := len(updateMap)
	for i := 0; i < SQLQuantity; i++ {
		count := 0

		var record bytes.Buffer
		record.WriteString("UPDATE " + tableName + " SET ")

		for fieldName, fieldValueList := range updateMap {
			record.WriteString(fieldName)
			record.WriteString(" = CASE " + "id")

			for j := k; j < len(IDList) && j < len(fieldValueList) && j < size+k; j++ {
				if fieldValueList[j] == nil { // 如果要更新的值为nil，说明不用更新，设置为它自己（这里不能conine，避出现）
					record.WriteString(" WHEN " + IDList[j] + " THEN " + fieldName)
				} else {
					record.WriteString(" WHEN " + IDList[j] + " THEN " + *fieldValueList[j])
				}
			}
			record.WriteString(" ELSE " + fieldName) // 如果没变更则设置成db原来字段的值
			count++
			if count != updateFieldCount {
				record.WriteString(" END, ")
			}
		}
		record.WriteString(" END WHERE ")
		record.WriteString("id" + " IN (")
		min := size + k
		if len(IDList) < min {
			min = len(IDList)
		}
		record.WriteString(strings.Join(IDList[k:min], ","))
		record.WriteString(");")

		k += size
		SQLArray = append(SQLArray, record.String())
	}

	return SQLArray, nil
}

func CheckItemListType(itemList interface{}) *bmserror.BMSError {
	t := reflect.TypeOf(itemList)
	if t.Kind() != reflect.Slice {
		return bmserror.NewError(constant.ErrInternalServer, "param is not sliceType pointer")
	}

	if t.Elem().Kind() != reflect.Ptr {
		return bmserror.NewError(constant.ErrInternalServer, "param is not pointer")
	}

	if t.Elem().Elem().Kind() != reflect.Struct {
		return bmserror.NewError(constant.ErrInternalServer, "param is not slice[*struct] pointer")
	}

	return nil
}

func getSQLQuantity(length, size int) int {
	SQLQuantity := int(math.Ceil(float64(length) / float64(size)))
	return SQLQuantity
}

func GetFieldName(fieldTag string) string {
	fieldTagArr := strings.Split(fieldTag, ":")
	if len(fieldTagArr) == 0 {
		return ""
	}

	fieldName := fieldTagArr[len(fieldTagArr)-1]

	return fieldName
}
