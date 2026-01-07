package paginator

import (
	"reflect"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/orm"
	"github.com/feimumoke/labequipbms/framework/support/expression"
)

type PageIn struct {
	Pageno     int64  // 页码
	Count      int64  // 数量
	OrderBy    string // 为空字符串 代表不需要排序
	IsGetTotal bool   // 为False代表不需要获取总数
}

var LimitOnePage = &PageIn{Pageno: 1, Count: 1}

func Paginator(qs orm.GORM, pageIn *PageIn, out interface{}) (int64, *bmserror.BMSError) {
	/*
		如果pageIn 为nil，则不需要分页，当pageIn 不为nil时，pageno和count都大于0时才会用到offset和limit
	*/
	if pageIn == nil {
		err := qs.Find(out).GetError()
		if err != nil {
			return 0, bmserror.NewError(constant.ErrDB, err.Error())
		}
		return 0, nil
	}
	total := int64(0)
	if pageIn.IsGetTotal {
		err := qs.Count(&total).GetError()
		if err != nil {
			return 0, bmserror.NewError(constant.ErrDB, err.Error())
		}
	}
	if pageIn.OrderBy != "" {
		qs = qs.Order(pageIn.OrderBy)
	}
	if pageIn.Pageno > 0 && pageIn.Count > 0 {
		qs = qs.Offset(int((pageIn.Pageno - 1) * pageIn.Count)).Limit(int(pageIn.Count))
	}
	err := qs.Find(out).GetError()
	if err != nil {
		return 0, bmserror.NewError(constant.ErrDB, err.Error())
	}
	return total, nil
}

// 内存分页
func PageList(l interface{}, page, count int64) {
	if l == nil || page <= 0 {
		return
	}
	kind := reflect.ValueOf(l).Type().Kind()
	if kind != reflect.Ptr {
		panic("PageList kind not reflect.Ptr")
	}
	kind = reflect.ValueOf(l).Elem().Type().Kind()
	if kind != reflect.Slice {
		panic("PageList kind not reflect.Slice")
	}

	total := reflect.ValueOf(l).Elem().Len()

	if total == 0 {
		return
	}

	begin := int((page - 1) * count)
	end := int(page * count)

	if total > end {
		l2 := reflect.ValueOf(l).Elem().Slice(begin, end)
		reflect.ValueOf(l).Elem().Set(l2)
	} else if total > begin {
		l2 := reflect.ValueOf(l).Elem().Slice(begin, total)
		reflect.ValueOf(l).Elem().Set(l2)
	} else {
		l2 := reflect.ValueOf(l).Elem().Slice(0, 0)
		reflect.ValueOf(l).Elem().Set(l2)
	}
}

// 面向基础类型分页
func PagingBaseType(l interface{}, page, count int64) interface{} {

	switch t := l.(type) {
	case []int64:
		totalCount := int64(len(l.([]int64)))
		return t[(page-1)*count : expression.IfInt64(page*count < totalCount, page*count, totalCount)]
	case []string:
		totalCount := int64(len(l.([]string)))
		return t[(page-1)*count : expression.IfInt64(page*count < totalCount, page*count, totalCount)]
	case []float64:
		totalCount := int64(len(l.([]float64)))
		return t[(page-1)*count : expression.IfInt64(page*count < totalCount, page*count, totalCount)]
	default:
		return pageSlice(l, page, count)
	}
	return nil
}

// support []Struct{}, []int64{}...
func pageSlice(l interface{}, page int64, count int64) interface{} {
	kind := reflect.ValueOf(l).Type().Kind()
	if kind != reflect.Slice {
		panic("PageList kind not reflect.Slice")
	}
	total := reflect.ValueOf(l).Len()
	begin := int((page - 1) * count)
	end := int(page * count)
	if total > end {
		return reflect.ValueOf(l).Slice(begin, end).Interface()
	} else if total > begin {
		return reflect.ValueOf(l).Slice(begin, total).Interface()
	} else {
		return reflect.ValueOf(l).Slice(0, 0).Interface()
	}
	return nil
}
