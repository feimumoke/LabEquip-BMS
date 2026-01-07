package excel

import (
	"github.com/feimumoke/wechating/apps/entity"
	"github.com/feimumoke/wechating/framework/constant"
	"github.com/feimumoke/wechating/framework/support/convert"
	"github.com/feimumoke/wechating/framework/wcerror"
	"github.com/xuri/excelize/v2"
	"io/ioutil"
	"time"
)

type ExcelHeaderField struct {
	ColumnIdx         int64
	ColumnDisplayName string
}

func GenerateExcelFile(header map[string]ExcelHeaderField, rows []map[string]interface{}) (string, *bmserror.BMSError) {
	excel := excelize.NewFile()
	sheetName := "Sheet1"
	excel.NewSheet(sheetName)

	err := setExcelHeader(excel, sheetName, header)
	if err != nil {
		return "", bmserror.NewError(constant.ErrParam, err.Error())
	}

	//set rows
	setRowErr := setExcelRowsValue(excel, sheetName, header, rows)
	if setRowErr != nil {
		return "", bmserror.NewError(constant.ErrParam, setRowErr.Error())
	}

	//store excel
	excelStoreFileName := convert.ToString(time.Now().Unix()) + ".xlsx"
	exportDir, err := ioutil.TempDir("", "exportDir")
	if err != nil {
		return "", bmserror.NewError(constant.ErrParam, err.Error())
	}

	filePath := exportDir + excelStoreFileName
	err = excel.SaveAs(filePath)
	if err != nil {
		return "", bmserror.NewError(constant.ErrParam, err.Error())
	}

	return filePath, nil
}

func setExcelRowsValue(excel *excelize.File, sheetName string, header map[string]ExcelHeaderField, rows []map[string]interface{}) *bmserror.BMSError {
	for rowIdx, row := range rows {
		currentRowHeight := rowIdx + 2
		for columnKey, columnValue := range row {
			columnChar, err := getColumnChar(header, columnKey)
			if err != nil {
				return err
			}
			axis := getCellAxis(columnChar, int64(currentRowHeight))
			setCellErr := excel.SetCellValue(sheetName, axis, columnValue)
			if setCellErr != nil {
				return bmserror.NewError(constant.ErrParam, setCellErr.Error())
			}
		}
	}
	return nil
}

func getColumnChar(header map[string]ExcelHeaderField, key string) (string, *bmserror.BMSError) {
	if field, ok := header[key]; ok {
		return getExcelColumnChar(field.ColumnIdx), nil
	} else {
		return "", bmserror.NewError(constant.ErrInternalServer, "excel header not contain [%s]", key)
	}
}

func getExcelColumnChar(idx int64) string {
	if idx/26 == 0 {
		return toChar(idx)
	}

	if idx == 26 {
		return "Z"
	}

	return toChar(idx/26) + toChar(idx%26)
}

func toChar(idx int64) string {
	if idx == 0 {
		return "A"
	}
	if idx > 26 {
		return "Z"
	}

	var arr = [...]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	return arr[idx-1]
}

func getCellAxis(column string, rowHeight int64) string {
	return column + convert.ToString(rowHeight)
}

func setExcelHeader(excel *excelize.File, sheetName string, header map[string]ExcelHeaderField) error {
	for _, v := range header {
		axis := getCellAxis(getExcelColumnChar(v.ColumnIdx), 1)
		err := excel.SetCellValue(sheetName, axis, v.ColumnDisplayName)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetExcelData(excel *excelize.File, excelDataItems ...*entity.ExcelDataItem) error {
	for _, excelItem := range excelDataItems {
		for rowIdx, values := range excelItem.ValuesList {
			row := int64(rowIdx) + excelItem.FromRow
			for columnIdx, v := range values {
				column := int64(columnIdx) + excelItem.FromColumn
				axis := getCellAxis(getExcelColumnChar(column), row)
				err := excel.SetCellValue(excelItem.SheetName, axis, v)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
