package excel

import (
	"bytes"
	"fmt"
	"github.com/feimumoke/wechating/apps/entity"
	"github.com/feimumoke/wechating/framework/constant"
	"github.com/feimumoke/wechating/framework/log"
	"github.com/feimumoke/wechating/framework/support/collection"
	"github.com/feimumoke/wechating/framework/support/convert"
	"github.com/feimumoke/wechating/framework/support/env"
	"github.com/feimumoke/wechating/framework/support/expression"
	"github.com/feimumoke/wechating/framework/support/util"
	"github.com/feimumoke/wechating/framework/wcerror"
	"github.com/feimumoke/wechating/framework/web"
	"github.com/xuri/excelize/v2"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var arr = [...]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
	"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF",
	"AG", "AH", "AI", "AJ", "AK", "AL", "AM", "AN", "AO", "AP", "AQ", "AR", "AS", "AT", "AU", "AV", "AW", "AX", "AY", "AZ"}

// 自定义时间格式，存在于 styles.go 文件
var timeStyleDict = map[string]int{
	"mm-dd-yy":      14,
	"d-mmm-yy":      15,
	"d-mmm":         16,
	"mmm-yy":        17,
	"h:mm am/pm":    18,
	"h:mm:ss am/pm": 19,
	"h:mm":          20,
	"h:mm:ss":       21,
}

func ParseResultItemType(result interface{}) (reflect.Type, *bmserror.BMSError) {
	t := reflect.TypeOf(result)
	if t.Kind() != reflect.Ptr {
		return nil, bmserror.NewError(constant.ErrParam, "parse excel, result param is not pointer")
	}

	typ := t.Elem()
	if typ.Kind() != reflect.Slice {
		return nil, bmserror.NewError(constant.ErrParam, "parse excel, result param is not sliceType pointer")
	}

	if typ.Elem().Kind() != reflect.Ptr {
		return nil, bmserror.NewError(constant.ErrParam, "parse excel, result param is not sliceType[pointer] pointer")
	}

	sliceElemPointerType := typ.Elem()
	if sliceElemPointerType.Elem().Kind() != reflect.Struct {
		return nil, bmserror.NewError(constant.ErrParam, "parse excel, result param is not slice[*struct] pointer")
	}

	return sliceElemPointerType.Elem(), nil
}

func getSheetName(f *excelize.File) (string, *bmserror.BMSError) {
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return "", bmserror.NewError(constant.ErrParam, "excel file no sheet")
	}
	sheetName := f.GetSheetName(0)
	return sheetName, nil
}

func fillStructByReflect(row int, columns []string, assignStruct interface{}) *bmserror.BMSError {
	structType := reflect.TypeOf(assignStruct).Elem()
	structValue := reflect.ValueOf(assignStruct).Elem()

	// 遍历数据行，防止panic
	for i := 0; i < structType.NumField() && i < len(columns); i++ {
		errorMessage := getExcelLocation(row+1, i+1) + " column can only contain number"
		typeErrorMessage := getExcelLocation(row+1, i+1) + " the uploaded data type is illegal"

		field := structType.Field(i)

		result := structValue.Field(i)

		switch field.Type.Kind() {
		case reflect.String:
			value := strings.TrimSpace(columns[i])
			result.SetString(value)
		case reflect.Int64:
			if columns[i] == "" || len(columns[i]) == 0 {
				columns[i] = "0"
			}
			convertedValue, err := convert.StringToInt64(columns[i])
			if err != nil {
				return bmserror.NewError(constant.ErrParam, errorMessage)
			}
			// 注意这里的参数类型默认是int64
			result.SetInt(convertedValue)
		case reflect.Float64:
			if columns[i] == "" || len(columns[i]) == 0 {
				columns[i] = "0.00"
			}
			f, err := strconv.ParseFloat(columns[i], 64)
			if err != nil {
				return bmserror.NewError(constant.ErrParam, errorMessage)
			}
			result.SetFloat(f)
		case reflect.Slice, reflect.Array:
			splitSlice := strings.Split(columns[i], ",")
			kind := field.Type.Elem().Kind()
			switch kind {
			case reflect.String:
				result.Set(reflect.ValueOf(splitSlice))
			case reflect.Int64:
				var array []int64
				for _, v := range splitSlice {
					if v == "" || len(v) == 0 {
						continue
					}

					convertedValue, err := convert.StringToInt64(v)
					if err != nil {
						return bmserror.NewError(constant.ErrParam, errorMessage)
					}
					array = append(array, convertedValue)
				}
				result.Set(reflect.ValueOf(array))
			case reflect.Float64:
				var array []float64
				for _, v := range splitSlice {
					if v == "" || len(v) == 0 {
						continue
					}

					convertedValue, err := convert.StringToFloat(v)
					if err != nil {
						return bmserror.NewError(constant.ErrParam, errorMessage)
					}
					array = append(array, convertedValue)
				}
				result.Set(reflect.ValueOf(array))
			default:
				return bmserror.NewError(constant.ErrParam, typeErrorMessage)
			}
		default:
			return bmserror.NewError(constant.ErrParam, typeErrorMessage)
		}
	}
	return nil
}

func UnmarshalFromExcel(buf []byte, resultList interface{}, isSkipFirstRow ...bool) *bmserror.BMSError {
	// 校验是否符合类型
	itemType, err := ParseResultItemType(resultList)
	if err != nil {
		return err
	}
	// 得到一个新的数据流
	bytesReader := bytes.NewReader(buf)
	// 读取数据流并返回填充的电子表格文件内容
	f, err1 := excelize.OpenReader(bytesReader)
	if err1 != nil {
		return bmserror.NewError(constant.ErrParam, err1.Error())
	}
	// get sheet name
	sheetName, err2 := getSheetName(f)
	if err2 != nil {
		return bmserror.NewError(constant.ErrParam, err2.Error())
	}

	// 获取行数去封装数据
	rows, err3 := f.GetRows(sheetName)
	if err3 != nil {
		return bmserror.NewError(constant.ErrParam, err3.Error())
	}

	resultSliceValue := reflect.ValueOf(resultList).Elem()
	emptySign := false

	// 是否跳过第一行
	isSkip := false
	firstRow := 0
	if len(isSkipFirstRow) > 0 && isSkipFirstRow[0] {
		isSkip = true
		firstRow = 1
	}
	for row, columns := range rows {
		// 跳过第一行
		if isSkip && row == 0 {
			continue
		}
		assignStructItem := reflect.New(itemType).Interface()
		if row == firstRow {
			err := verifyExcelHead(columns, assignStructItem)
			if err != nil {
				return err.Mark()
			}
			continue
		}

		verifyResult := verifyRowWhetherEmpty(columns)
		if verifyResult {
			emptySign = true
			continue
		}

		if emptySign {
			return bmserror.NewError(constant.ErrParam, "excel does not allow blank rows in the data!")
		}

		// 填充数据
		err := fillStructByReflect(row, columns, assignStructItem)
		if err != nil {
			return err.AddError(constant.ErrParam, err.Error())
		}

		importValue := reflect.ValueOf(assignStructItem)

		resultSliceValue = reflect.Append(resultSliceValue, importValue)
	}

	reflect.ValueOf(resultList).Elem().Set(resultSliceValue)
	return nil
}

func getExcelLocation(row, column int) string {
	axis := getCellAxis(getExcelColumnChar(int64(column)), int64(row))
	return axis
}

// 防止出现panic
func verifyRowWhetherEmpty(columns []string) bool {

	for i := 0; i < len(columns); i++ {
		if columns[i] != "" {
			return false
		}
	}
	return true
}
func verifyExcelHead(columns []string, assignStruct interface{}) *bmserror.BMSError {
	structType := reflect.TypeOf(assignStruct).Elem()
	fieldNum := structType.NumField()

	// 防止出现panic情况
	if len(columns) < fieldNum {
		return bmserror.NewError(constant.ErrParam, "Template is different from the system")
	}

	titleErrorSign := false
	for i := 0; i < fieldNum; i++ {
		header := structType.Field(i).Tag.Get("title")
		// 忽略大小写和首尾空格
		if strings.EqualFold(header, strings.TrimSpace(columns[i])) {
			continue
		}
		titleErrorSign = true
	}
	if titleErrorSign {
		return bmserror.NewError(constant.ErrParam, "Template is different from the system")
	}

	titleRedundantColumn := false
	if len(columns) > fieldNum {
		for i := fieldNum; i < len(columns); i++ {
			if columns[i] == "" {
				continue
			}
			titleRedundantColumn = true
		}
		if titleRedundantColumn {
			return bmserror.NewError(constant.ErrParam, "Template is different from the system")
		}
	}
	return nil
}

func GenerateExcel(itemList interface{}, excelName string) (*web.RespWithDownloadFile, *bmserror.BMSError) {
	fileBytes, errRead := GenerateFileBytes(itemList, excelName)
	if errRead != nil {
		return nil, bmserror.NewError(constant.ErrParam, "download excel fail")
	}

	resp := web.NewRespWithDownloadFile(fileBytes, excelName)

	return resp, nil
}

func GenerateExcelFileAndFilePath(itemList interface{}, excelName string) ([]byte, string, *bmserror.BMSError) {
	if !strings.HasSuffix(excelName, ".xlsx") {
		return nil, "", bmserror.NewError(constant.ErrParam, "The Excel name should end in .xlsx")
	}

	// 生成Excel的文件
	excel := excelize.NewFile()
	sheetName := "Sheet1"
	excel, err := generateExcelSheet(itemList, sheetName, excel)
	if err != nil {
		return nil, "", err.Mark()
	}

	fileBytes, filePath, err := generateFileBytesAndPath(excel, excelName)
	if err != nil {
		return nil, "", err.Mark()
	}

	return fileBytes, filePath, nil
}

func generateFileBytesAndPath(excel *excelize.File, excelName string) ([]byte, string, *bmserror.BMSError) {

	if !strings.HasSuffix(excelName, ".xlsx") {
		return nil, "", bmserror.NewError(constant.ErrParam, "The Excel name should end in .xlsx")
	}

	//store excel 存储Excel
	exportDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, "", bmserror.NewError(constant.ErrParam, err.Error())
	}

	// 保存目录
	filePath := exportDir + excelName
	err = excel.SaveAs(filePath)
	if err != nil {
		return nil, "", bmserror.NewError(constant.ErrParam, err.Error())
	}

	fileBytes, errRead := ioutil.ReadFile(filePath)
	if errRead != nil {
		return nil, "", bmserror.NewError(constant.ErrParam, "download excel fail")
	}
	return fileBytes, filePath, nil
}

func GenerateExcelWithMultipleSheet(sheetList []*entity.ExcelSheetTab, excelName string) ([]byte, string, *bmserror.BMSError) {
	excel, err := generateExcelSheetList(sheetList)
	if err != nil {
		return nil, "", err.Mark()
	}
	fileBytes, filePath, err := generateFileBytesAndPath(excel, excelName)
	if err != nil {
		return nil, "", err.Mark()
	}

	return fileBytes, filePath, nil
}

func GenerateExcelWithTemplate(templateFileName, excelName string, excelDataItems ...*entity.ExcelDataItem) ([]byte, string, *bmserror.BMSError) {
	// 所有的模板都必须放到 /web/download_template/ 目录下
	fileTemplate := env.GetProjectDir() + "/web/download_template/" + templateFileName
	file, openErr := excelize.OpenFile(fileTemplate)
	if openErr != nil {
		return nil, "", bmserror.NewError(constant.ErrParam, "open template file fail|file:%v,err:%v", fileTemplate, openErr)
	}
	if err := SetExcelData(file, excelDataItems...); err != nil {
		return nil, "", bmserror.NewError(constant.ErrParam, "set excel data fail|file:%v,err:%v", fileTemplate, err)
	}
	fileBytes, filePath, err := generateFileBytesAndPath(file, excelName)
	if err != nil {
		return nil, "", err.Mark()
	}
	return fileBytes, filePath, nil
}

func generateExcelSheetList(sheetList []*entity.ExcelSheetTab) (*excelize.File, *bmserror.BMSError) {

	// 生成Excel的文件
	excel := excelize.NewFile()

	for i := 0; i < len(sheetList); i++ {
		var err *bmserror.BMSError
		excel, err = generateExcelSheet(sheetList[i].Data, sheetList[i].SheetName, excel)
		if err != nil {
			return nil, err.Mark()
		}
	}
	if len(sheetList) != 0 {
		excel.DeleteSheet("Sheet1")
	}

	return excel, nil

}

func generateExcelSheet(itemList interface{}, sheetName string, excel *excelize.File) (*excelize.File, *bmserror.BMSError) {
	excel.NewSheet(sheetName)
	errCheck := CheckItemListType(itemList)
	if errCheck != nil {
		return nil, errCheck.Mark()
	}

	fieldValue := reflect.ValueOf(itemList)
	fieldType := reflect.TypeOf(itemList).Elem().Elem()
	sliceLength := fieldValue.Len()
	fieldNum := fieldType.NumField()

	// 引入rowIndex 变量避免空列
	rowIndex := int64(0)
	for i := 0; i < fieldNum; i++ {
		name := fieldType.Field(i).Tag.Get("title")
		if name == "" {
			continue
		}

		// 设置头部
		rowIndex++
		axis := getCellAxis(getExcelColumnChar(rowIndex), 1)
		err := excel.SetCellValue(sheetName, axis, name)
		if err != nil {
			return nil, bmserror.NewError(constant.ErrParam, "this column name should be %v", name)
		}
	}

	//set rows
	for i := 0; i < sliceLength; i++ {
		// 得到某一个具体的结构体的
		structValue := fieldValue.Index(i).Elem()
		// 初始化新行的列索引
		rowIndex = int64(0)
		for j := 0; j < fieldNum; j++ {
			name := fieldType.Field(j).Tag.Get("title")
			if name == "" {
				continue
			}

			elem := structValue.Field(j)

			result, err := getCellValue(elem, fieldType.Field(j))
			if err != nil {
				return nil, err.Mark()
			}
			//var result string
			//
			//switch elem.Kind() {
			//case reflect.Int64:
			//	s := strconv.FormatInt(structValue.Field(j).Int(), 10)
			//	result = s
			//case reflect.Float64:
			//	value := strconv.FormatFloat(structValue.Field(j).Float(), 'f', -1, 64)
			//	result = value
			//case reflect.String:
			//	result = structValue.Field(j).String()
			//case reflect.Slice:
			//	var buf bytes.Buffer
			//	elemKind := fieldType.Field(j).Type.Elem().Kind()
			//	switch elemKind {
			//	case reflect.Int64:
			//		for k := 0; k < elem.Len(); k++ {
			//			value := strconv.FormatInt(elem.Index(k).Int(), 10)
			//			buf.WriteString(value)
			//			if k != elem.Len()-1 {
			//				buf.WriteString(",")
			//			}
			//		}
			//	case reflect.Float64:
			//		for k := 0; k < elem.Len(); k++ {
			//			value := strconv.FormatFloat(elem.Index(k).Float(), 'f', -1, 64)
			//			buf.WriteString(value)
			//			if k != elem.Len()-1 {
			//				buf.WriteString(",")
			//			}
			//		}
			//	case reflect.String:
			//		for k := 0; k < elem.Len(); k++ {
			//			buf.WriteString(elem.Index(k).String())
			//			if k != elem.Len()-1 {
			//				buf.WriteString(",")
			//			}
			//		}
			//	default:
			//		return nil, bmserror.NewError(constant.ErrParam, "type does not meet the requirements")
			//	}
			//	if buf.String() != "" {
			//		result = buf.String()
			//	}
			//}

			currentRowHeight := int64(i + 2)
			rowIndex++
			axis := getCellAxis(getExcelColumnChar(rowIndex), currentRowHeight)
			setCellErr := excel.SetCellValue(sheetName, axis, result)
			if setCellErr != nil {
				return nil, bmserror.NewError(constant.ErrParam, "In the %v row, In the %v column failed to generate Excel", i+1, j)
			}
		}
	}
	return excel, nil
}

// GenerateExcelFileAndFilePathWithTitle 带表头 适用于动态表头Excel
func GenerateExcelFileAndFilePathWithTitle(titleList []string, itemList [][]string, excelName string) ([]byte, string, *bmserror.BMSError) {
	if !strings.HasSuffix(excelName, ".xlsx") {
		return nil, "", bmserror.NewError(constant.ErrParam, "The Excel name should end in .xlsx")
	}

	// 生成Excel的文件
	excel := excelize.NewFile()
	excel, err := generateExcelSheetWithTitle(titleList, itemList, excel)
	if err != nil {
		return nil, "", err.Mark()
	}

	fileBytes, filePath, err := generateFileBytesAndPath(excel, excelName)
	if err != nil {
		return nil, "", err.Mark()
	}

	return fileBytes, filePath, nil
}

// generateExcelSheetWithTitle 带表头 适用于动态表头Excel
func generateExcelSheetWithTitle(titleList []string, itemList [][]string, excel *excelize.File) (*excelize.File, *bmserror.BMSError) {
	// 生成Excel的文件
	sheetName := "Sheet1"
	excel.NewSheet(sheetName)
	// 引入rowIndex 变量避免空列
	rowIndex := int64(0)
	for _, name := range titleList {
		// 设置头部 列序号
		rowIndex++
		axis := getCellAxis(getExcelColumnChar(rowIndex), 1)
		err := excel.SetCellValue(sheetName, axis, name)
		if err != nil {
			return nil, bmserror.NewError(constant.ErrParam, "this column name should be %v", name)
		}
	}

	//set rows
	for i := 0; i < len(itemList); i++ {
		// 得到某一个具体的结构体的
		line := itemList[i]
		// 初始化新行的列索引
		rowIndex = int64(0)
		for j := 0; j < len(line); j++ {

			result := line[j]

			currentRowHeight := int64(i + 2)
			rowIndex++
			axis := getCellAxis(getExcelColumnChar(rowIndex), currentRowHeight)
			setCellErr := excel.SetCellValue(sheetName, axis, result)
			if setCellErr != nil {
				return nil, bmserror.NewError(constant.ErrParam, "In the %v row, In the %v column failed to generate Excel", i+1, j)
			}
		}
	}
	return excel, nil
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

func CheckItemListTypeDynamic(itemList interface{}) *bmserror.BMSError {
	t := reflect.TypeOf(itemList)
	if t.Kind() != reflect.Slice {
		return bmserror.NewError(constant.ErrInternalServer, "param is not sliceType pointer")
	}

	if t.Elem().Kind() != reflect.Ptr {
		return bmserror.NewError(constant.ErrInternalServer, "param is not pointer")
	}

	return nil
}

func GenerateFileBytes(itemList interface{}, excelName string) ([]byte, *bmserror.BMSError) {
	fileBytes, _, err := GenerateExcelFileAndFilePath(itemList, excelName)
	if err != nil {
		return nil, err.Mark()
	}
	return fileBytes, nil
}

func getIsEncrypt(fieldIsEncrypt string, opt *entity.GenerateExcelOpt) bool {
	return opt != nil && convert.BoolValue(opt.IsEncrypt) && fieldIsEncrypt == constant.CommonEnumYES
}

func generateStreamExcelSheet(i int, sheet *entity.ExcelSheetTab, excel *excelize.File, opt *entity.GenerateExcelOpt) (*excelize.File, *bmserror.BMSError) {
	itemList := sheet.Data
	sheetName := sheet.SheetName
	if i == 0 {
		if sheetName != "Sheet1" {
			excel.SetSheetName("Sheet1", sheetName)
		}
	} else {
		excel.NewSheet(sheetName)
	}
	errCheck := CheckItemListType(itemList)
	if errCheck != nil {
		return nil, errCheck.Mark()
	}

	fieldValue := reflect.ValueOf(itemList)
	fieldType := reflect.TypeOf(itemList).Elem().Elem()
	sliceLength := fieldValue.Len()
	fieldNum := fieldType.NumField()

	excludeTitles := collection.NewStringSet(sheet.ExcludeTitles...)
	// 获取流式写入器
	streamWriter, err := excel.NewStreamWriter(sheetName)
	if err != nil {
		return nil, bmserror.NewError(constant.ErrParam, err.Error())
	}

	titleList := make([]interface{}, 0)

	// 设置头部
	for i := 0; i < fieldNum; i++ {
		name := fieldType.Field(i).Tag.Get("title")
		if name == "" {
			continue
		}
		if excludeTitles.Contains(name) {
			continue
		}
		titleList = append(titleList, excelize.Cell{Value: name})
	}
	if err := streamWriter.SetRow("A1", titleList); err != nil {
		return nil, bmserror.NewError(constant.ErrParam, err.Error())
	}
	//默认时间格式
	defaultStyle, _ := excel.NewStyle(&excelize.Style{NumFmt: 22, Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"}})

	//set rows
	for i := 0; i < sliceLength; i++ {
		// 得到某一个具体的结构体的
		structValue := fieldValue.Index(i).Elem()

		rowList := make([]interface{}, 0)

		//跳过的字段
		unUsedColumns := 0
		for j := 0; j < fieldNum; j++ {
			name := fieldType.Field(j).Tag.Get("title")
			if name == "" {
				unUsedColumns++
				continue
			}
			if excludeTitles.Contains(name) {
				unUsedColumns++
				continue
			}
			structValue := structValue.Field(j)

			// 直接写入对于列表类型中间分隔符为空格
			//rowList = append(rowList, structValue)
			structType := fieldType.Field(j)
			var cellStyle *int
			if structValue.Kind() == reflect.Struct {
				switch structValue.Interface().(type) {
				case time.Time:
					timeFormat := fieldType.Field(j).Tag.Get("format")
					if timeFormat != "" { //自定义时间格式
						if _, ok := timeStyleDict[timeFormat]; ok {
							cellStyle = convert.Int(timeStyleDict[timeFormat])
							break
						}
					}
					cellStyle = convert.Int(defaultStyle)
				default:
					break
				}
			}
			fieldIsEncrypt := fieldType.Field(j).Tag.Get("encrypt")
			isEncrypt := getIsEncrypt(fieldIsEncrypt, opt)
			elem, err := getCell(structValue, structType, cellStyle, isEncrypt)
			if err != nil {
				return nil, err.Mark()
			}
			rowList = append(rowList, elem)
		}
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err := streamWriter.SetRow(cell, rowList); err != nil {
			return nil, bmserror.NewError(constant.ErrParam, err.Error())
		}
	}
	if fErr := streamWriter.Flush(); fErr != nil {
		return nil, bmserror.NewError(constant.ErrParam, fErr.Error())
	}

	return excel, nil
}

func getCellValue(cellValue reflect.Value, cellType reflect.StructField) (interface{}, *bmserror.BMSError) {
	var result interface{}

	switch cellValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		result = cellValue.Int()
	case reflect.Int64:
		if cellValue.Int() > 19700101000 { //比较长的整数转字符串不然会变成科学计数法
			result = strconv.FormatInt(cellValue.Int(), 10)
		} else {
			result = cellValue.Int()
		}
	case reflect.Float64, reflect.Float32:
		result = cellValue.Float()
	case reflect.String:
		result = cellValue.String()
	case reflect.Slice:
		var buf bytes.Buffer
		elemKind := cellType.Type.Elem().Kind()
		switch elemKind {
		case reflect.Int64:
			for k := 0; k < cellValue.Len(); k++ {
				value := strconv.FormatInt(cellValue.Index(k).Int(), 10)
				buf.WriteString(value)
				if k != cellValue.Len()-1 {
					buf.WriteString(",")
				}
			}
		case reflect.Float64:
			for k := 0; k < cellValue.Len(); k++ {
				value := strconv.FormatFloat(cellValue.Index(k).Float(), 'f', -1, 64)
				buf.WriteString(value)
				if k != cellValue.Len()-1 {
					buf.WriteString(",")
				}
			}
		case reflect.String:
			for k := 0; k < cellValue.Len(); k++ {
				buf.WriteString(cellValue.Index(k).String())
				if k != cellValue.Len()-1 {
					buf.WriteString(",")
				}
			}
		default:
			return "", bmserror.NewError(constant.ErrParam, "type does not meet the requirements")
		}
		if buf.String() != "" {
			result = buf.String()
		}
	case reflect.Struct:
		switch cellValue.Interface().(type) {
		case time.Time:
			//这里应该是 time.Time,后续需要处理excel数据设置为时间格式 其他结构体需要自行处理或者在这补充
			orignalTime := cellValue.Interface().(time.Time)
			if orignalTime.Year() == 1970 || orignalTime.Year() == 1 { //1970的时间不处理
				timStr := "-"
				return timStr, nil
			}
			//excel处理只能传utc时间，所以将时间转为utc时间，时分秒没变，相当于如果是东八区时间，调小八小时后转为utc时间
			result = time.Date(orignalTime.Year(), orignalTime.Month(), orignalTime.Day(), orignalTime.Hour(), orignalTime.Minute(), orignalTime.Second(), orignalTime.Nanosecond(), time.FixedZone("CST", 0)).UTC()
		default:
			return "", bmserror.NewError(constant.ErrParam, "type does not meet the requirements")
		}

	}
	return result, nil
}

func generateStreamExcelSheetList(sheetList []*entity.ExcelSheetTab, opt *entity.GenerateExcelOpt) (*excelize.File, *bmserror.BMSError) {
	// 生成Excel的文件
	excel := excelize.NewFile()
	for i := 0; i < len(sheetList); i++ {
		var err *bmserror.BMSError
		excel, err = generateStreamExcelSheet(i, sheetList[i], excel, opt)
		if err != nil {
			return nil, err.Mark()
		}
	}
	return excel, nil
}

// excel流式写入器
func GenerateStreamExcelWithMultipleSheet(sheetList []*entity.ExcelSheetTab, excelName string, opt *entity.GenerateExcelOpt) ([]byte, string, *bmserror.BMSError) {
	excel, err := generateStreamExcelSheetList(sheetList, opt)
	if err != nil {
		return nil, "", err.Mark()
	}
	fileBytes, filePath, err := generateFileBytesAndPath(excel, excelName)
	if err != nil {
		return nil, "", err.Mark()
	}

	return fileBytes, filePath, nil
}

func getCell(cellValue reflect.Value, cellType reflect.StructField, cellStyle *int, isEncrypt bool) (*excelize.Cell, *bmserror.BMSError) {
	result := &excelize.Cell{}
	var val interface{}
	var err *bmserror.BMSError
	if isEncrypt {
		val = "***"
	} else {
		val, err = getCellValue(cellValue, cellType)
		if err != nil {
			return nil, err.Mark()
		}
	}
	result.Value = val
	if cellStyle != nil {
		result.StyleID = *cellStyle
	}
	return result, nil
}

func GenerateExeclWithStyle(sheetList []*entity.ExcelStyleSheet, excelName string) ([]byte, string, *bmserror.BMSError) {
	if !strings.HasSuffix(excelName, ".xlsx") {
		return nil, "", bmserror.NewError(constant.ErrParam, "The Excel name should end in .xlsx")
	}
	excel := excelize.NewFile()
	sheetOneExist := false
	for i, sheet := range sheetList {
		sheetName := expression.IfString(sheet.SheetName == "", "Sheet"+strconv.Itoa(i+1), sheet.SheetName)
		if sheetName == "Sheet1" {
			sheetOneExist = true
		}
		excel.NewSheet(sheetName)
		colIndex := int64(1)
		colHeightMap := make(map[string]int64)
		for _, rowInfo := range sheet.ExcelRows {
			for col := 0; col < rowInfo.Len(); col++ {
				cellVal := rowInfo.ValueAt(col)
				data := cellVal.Data
				cellWidth := cellVal.Width
				cellHeight := cellVal.Height
				if cellWidth == 0 {
					continue
				}
				if cellHeight == 0 {
					colIndex += cellWidth
					continue
				}
				curHeight := colHeightMap[getExcelColumnChar(colIndex)]
				axis := getCellAxis(getExcelColumnChar(colIndex), curHeight+1)
				cellVal.SetAxis(axis)
				cellVal.SetFoot(axis)
				for c := colIndex; c < colIndex+cellWidth; c++ {
					colHeightMap[getExcelColumnChar(c)] += cellHeight
				}
				if err := excel.SetCellValue(sheetName, axis, data); err != nil {
					return nil, "", bmserror.NewError(constant.ErrParam, "SetCellValue %v - %v err %v", axis, data, err)
				}
				if cellWidth > 1 || cellHeight > 1 {
					end := getCellAxis(getExcelColumnChar(colIndex+cellWidth-1), curHeight+cellHeight)
					cellVal.SetFoot(end)
					log.Infof("MergeCell for %v from %v to %v", data, axis, end)
					if err := excel.MergeCell(sheetName, axis, end); err != nil {
						return nil, "", bmserror.NewError(constant.ErrParam, "MergeCell %v-%v err %v", axis, end, err)
					}
				}
				colIndex += cellWidth
			}
			colIndex = 1
		}
		for _, table := range sheet.Tables {
			hcell := table.HCell()
			vcell := table.VCell()
			format := util.ToJSON(table.Format)
			log.Infof("AddTable hcell-%v vcell- %v format %v", hcell, vcell, format)
			if hcell != "" && vcell != "" {
				tab := &excelize.Table{
					Range:             fmt.Sprintf("%v:%v", hcell, vcell),
					Name:              table.Format.TableName,
					StyleName:         table.Format.TableStyle,
					ShowColumnStripes: table.Format.ShowColumnStripes,
					ShowFirstColumn:   table.Format.ShowFirstColumn,
					ShowHeaderRow:     nil,
					ShowLastColumn:    table.Format.ShowLastColumn,
					ShowRowStripes:    convert.Bool(table.Format.ShowRowStripes),
				}
				if err := excel.AddTable(sheetName, tab); err != nil {
					return nil, "", bmserror.NewError(constant.ErrParam, "AddTable %v-%v err %v", hcell, vcell, err)
				}
			}
		}
		for _, chart := range sheet.Charts {
			cell := chart.Cell()
			format := chart.Format(sheetName)
			var comb []string
			for _, combFormat := range chart.ComboFormat {
				comb = append(comb, combFormat(sheetName))
			}
			log.Infof("AddChart cell- %v format %v comb %v", cell, format, comb)
			if cell != "" && format != "" {
				if err := excel.AddChart(sheetName, cell, nil); err != nil {
					return nil, "", bmserror.NewError(constant.ErrParam, "AddChart %v err %v", cell, err)
				}
			}
		}

	}
	if !sheetOneExist {
		excel.DeleteSheet("Sheet1")
	}
	return generateFileBytesAndPath(excel, excelName)
}

// SplitCellName("AK74") // return "AK", 74, nil
func SplitCellName(cell string) (string, int, *bmserror.BMSError) {
	alpha := func(r rune) bool {
		return ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z')
	}
	if strings.IndexFunc(cell, alpha) == 0 {
		i := strings.LastIndexFunc(cell, alpha)
		if i >= 0 && i < len(cell)-1 {
			col, rowstr := cell[:i+1], cell[i+1:]
			if row, err := strconv.Atoi(rowstr); err == nil && row > 0 {
				return col, row, nil
			}
		}
	}
	return "", -1, bmserror.NewError(constant.ErrParam, "cell is wrong %v", cell)
}
