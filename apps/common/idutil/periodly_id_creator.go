package idutil

import (
	"context"
	"fmt"

	"math"
	"strconv"
	"time"

	cmanager "github.com/feimumoke/labequipbms/apps/common/manager"
	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
)

type IPeriodlyIDCreator interface {
	GetDataFormate() string
	GetDeleteTime() int64
	GetDatePeriodID(ctx context.Context, idType constant.DistributedIDType, ptNo int64) (int64, int64, *bmserror.BMSError)
}

var DIDCreator *DistributedIDCreator

type DistributedIDCreator struct {
	Generator *cmanager.Generator
}

func InitDIDCreator() {
	idmap := make(map[string]*cmanager.IDSegment)
	gen := cmanager.NewIdGenerator(idmap)
	DIDCreator = &DistributedIDCreator{Generator: gen}
}

func (d *DistributedIDCreator) GetID(ctx context.Context, idType constant.DistributedIDType, ptId string) (int64,
	*bmserror.BMSError) {
	step := constant.IDTypeToStep[idType]
	id, err := d.Generator.GetID(ctx, idType, 0, 0, step)
	if err != nil {
		return 0, err.Mark()
	}
	return id, nil
}

var DailyIDCreator *DistributedDailyIDCreator

type DistributedDailyIDCreator struct {
	Generator *cmanager.Generator
}

func InitDailyIDCreator() {
	idmap := make(map[string]*cmanager.IDSegment)
	gen := cmanager.NewIdGenerator(idmap)
	DailyIDCreator = &DistributedDailyIDCreator{Generator: gen}
}

func (d *DistributedDailyIDCreator) GetDataFormate() string {
	return fmt.Sprintf("%v", time.Now().Format("20060102"))
}
func (d *DistributedDailyIDCreator) GetDeleteTime() int64 {
	return 180 * 24 * 60 * 60
}

func (d *DistributedDailyIDCreator) GetDatePeriodID(ctx context.Context, idType constant.DistributedIDType, ptId string) (int64, int64,
	*bmserror.BMSError) {
	dataFormate := timeutil.TodayDateStr(timeutil.DateIntFormat)
	period, ctErr := strconv.ParseInt(dataFormate, 10, 64)
	if ctErr != nil {
		return 0, 0, bmserror.NewError(constant.ErrInternalServer, "dateformate invalid ", dataFormate)
	}
	step := constant.IDTypeToStep[idType]
	id, err := d.Generator.GetID(ctx, idType, period, 0, step)
	if err != nil {
		return 0, 0, err.Mark()
	}
	return period, id, nil
}

func (d *DistributedDailyIDCreator) GetBaseDatePeriodID(ctx context.Context, dataFormate string, idType constant.DistributedIDType, ptId string) (int64, int64,
	*bmserror.BMSError) {
	period, ctErr := strconv.ParseInt(dataFormate, 10, 64)
	if ctErr != nil {
		return 0, 0, bmserror.NewError(constant.ErrInternalServer, "dateformate invalid ", dataFormate)
	}
	step := constant.IDTypeToStep[idType]
	id, err := d.Generator.GetID(ctx, idType, period, 0, step)
	if err != nil {
		return 0, 0, err.Mark()
	}
	return period, id, nil
}

func (d *DistributedDailyIDCreator) GenerateCommonOrderId(ctx context.Context, ptId string, idType int64, prefix string) (string, *bmserror.BMSError) {
	datePeriod, ret, idErr := d.GetDatePeriodID(ctx, idType, ptId)
	if idErr != nil {
		return "", bmserror.NewError(constant.ErrDB, idErr.Error())
	}
	newDatePeriod := strconv.FormatInt(datePeriod, 10)[2:]
	newRet := ConvertToNewGenerationIDWithBit(ret, 4)
	newptId := GetNewPtID(ptId)
	taskId := fmt.Sprintf("%s%s%s%s", prefix, newptId, newDatePeriod, newRet)
	return taskId, nil
}

func GetNewPtID(ptId string) string {
	if len(ptId) == 3 {
		return ptId + "000"
	}
	return ptId
}

/*
// 3位流水 纯数字/最末位为字母/次末位为字母 时的范围
1-999
1000-3573
3574-9657
// 4位流水
1-9999
10000-35973
35974-102897
// 5位流水
1-99999
100000-359973
359974-1035297
// 6位流水
1-999999 999999
1000000-3599973
3599974-10359297
*/

const CharNumbers int64 = 26

func ConvertToNewGenerationIDWithBit(id int64, bit int64) string {
	format := "%0" + strconv.FormatInt(bit, 10) + "d"
	strID := fmt.Sprintf(format, id)

	// 最末位为字母时 输出格式/模/最小值/最大值
	lastCharFormat := "%0" + strconv.FormatInt(bit-1, 10) + "d%c"
	lastCharMod := int64(math.Pow10(int(bit-1))) - 1
	lastCharSmallest := int64(math.Pow10(int(bit)))
	lastCharBiggest := lastCharSmallest + lastCharMod*CharNumbers - 1

	// 次末位为字母时 输出格式/模/最小值/最大值
	secondLastCharFormat := "%0" + strconv.FormatInt(bit-2, 10) + "d%c%c"
	secondLastCharMod := int64(math.Pow10(int(bit-2))) - 1
	secondLastCharSmallest := lastCharBiggest + 1
	secondLastCharBiggest := secondLastCharSmallest + secondLastCharMod*CharNumbers*CharNumbers - 1

	if id >= 1 && id < lastCharSmallest {
	} else if id >= lastCharSmallest && id <= lastCharBiggest {
		// tmp是这个范围的相对位置, 最末位不同字符的范围是 lastCharMod, 所以对 tmp / lastCharMod 就得到最末位字符
		tmp := id - lastCharSmallest + 1
		q := (tmp - 1) / lastCharMod
		r := tmp % lastCharMod
		if r == 0 {
			r = lastCharMod
		}
		strID = fmt.Sprintf(lastCharFormat, r, 65+q)
	} else if id >= secondLastCharSmallest && id <= secondLastCharBiggest {
		// tmp是这个范围的相对位置, 次末位不同字符的范围是 secondLastCharMod * 26, 所以对 tmp / (secondLastCharMod * 26) 就得到次末位字符
		// 在次末位相同的这一段上, 最末位不同字符的范围是 secondLastCharMod, 所以对 tmp在次末位相同的这一段的相对位置 / secondLastCharMod 就得到最末位字符
		tmp := id - secondLastCharSmallest + 1
		qq := (tmp - 1) / (secondLastCharMod * CharNumbers)
		q := (tmp - secondLastCharMod*CharNumbers*qq - 1) / secondLastCharMod
		r := tmp % secondLastCharMod
		if r == 0 {
			r = secondLastCharMod
		}
		strID = fmt.Sprintf(secondLastCharFormat, r, 65+qq, 65+q)
	}
	return strID
}
