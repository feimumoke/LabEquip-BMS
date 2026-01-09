package entity

type ExcelSheetTab struct {
	SheetName     string      `gorm:"column:sheet_name" json:"sheet_name"`
	Data          interface{} `gorm:"column:data" json:"data"`
	ExcludeTitles []string    `gorm:"-" json:"exclude_titles"`
}

type ExcelDataItem struct {
	SheetName  string
	FromRow    int64 // 跟excel中的位置相同
	FromColumn int64 // 跟excel中的位置相同
	ValuesList [][]interface{}
}

const DefaultSheetName = "sheet1"

type ExcelStyleSheet struct {
	SheetName string
	ExcelRows []*ExcelsRow
	Tables    []*ExcelTable
	Charts    []*ExcelChart
}

type ExcelsRow []*ExcelCellItem

type ExcelCellItem struct {
	Data   interface{}
	Width  int64
	Height int64
	axis   string
	foot   string
}

// f.AddTable("Sheet2", "F2", "H6", `{"table_name":"table","table_style":"TableStyleMedium2", "show_first_column":true,"show_last_column":true,"show_row_stripes":false,"show_column_stripes":true}`)

type ExcelCoordinate func() string
type ExcelTable struct {
	HCell  ExcelCoordinate
	VCell  ExcelCoordinate
	Format *TableFormat
}

type TableFormat struct {
	TableName         string `json:"table_name"`
	TableStyle        string `json:"table_style"`
	ShowFirstColumn   bool   `json:"show_first_column"`
	ShowLastColumn    bool   `json:"show_last_column"`
	ShowRowStripes    bool   `json:"show_row_stripes"`
	ShowColumnStripes bool   `json:"show_column_stripes"`
}

type FormatFunc func(sheetName string) string
type ExcelChart struct {
	Cell        ExcelCoordinate
	Format      FormatFunc
	ComboFormat []FormatFunc
}

func (i *ExcelCellItem) SetAxis(axis string) {
	i.axis = axis
}

func (i *ExcelCellItem) GetAxis() string {
	return i.axis
}

func (i *ExcelCellItem) SetFoot(foot string) {
	i.foot = foot
}

func (i *ExcelCellItem) GetFoot() string {
	return i.foot
}

func NewExcelRow() ExcelsRow {
	return make([]*ExcelCellItem, 0)
}

func (row ExcelsRow) Len() int {
	return len(row)
}

func (row ExcelsRow) ValueAt(index int) *ExcelCellItem {
	if row.Len() <= index {
		return nil
	}
	return row[index]
}

func (row *ExcelsRow) AppendItem(data interface{}) *ExcelsRow {
	*row = append(*row, &ExcelCellItem{
		Data:   data,
		Width:  1,
		Height: 1,
	})
	return row
}

func (row *ExcelsRow) AppendItemWithSize(data interface{}, width, height int64) *ExcelsRow {
	*row = append(*row, &ExcelCellItem{
		Data:   data,
		Width:  width,
		Height: height,
	})
	return row
}

func NewDefaultCellItem(data interface{}) *ExcelCellItem {
	return &ExcelCellItem{
		Data:   data,
		Width:  1,
		Height: 1,
	}
}

type GenerateExcelOpt struct {
	IsEncrypt *bool
}
