package idutil

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/support/convert"
	"github.com/feimumoke/labequipbms/framework/support/timeutil"
)

func GenEquipNumber(ctx context.Context) (string, *bmserror.BMSError) {
	id, err := DIDCreator.GetID(ctx, constant.EquipID, ALLPointType)
	if err != nil {
		return "", err.Mark()
	}
	return fmt.Sprintf("%s%05d", constant.DistributedIDTypePerfixMap[constant.EquipID], id), nil
}

func GenBorrowTaskId(ctx context.Context, ptId string) (string, *bmserror.BMSError) {
	period, num, err := DailyIDCreator.GetDatePeriodID(ctx, constant.BorrowTaskID, ptId)
	if err != nil {
		return "", err.Mark()
	}
	newPeriod := strconv.FormatInt(period, 10)[2:]
	newNum := ConvertToNewGenerationIDWithBit(num, 4)
	newPtID := GetNewPtID(ptId)
	inboundID := fmt.Sprintf("%s%s%s%s", constant.DistributedIDTypePerfixMap[constant.BorrowTaskID], newPtID, newPeriod, newNum)
	return inboundID, nil
}

func GenerateTaskNumber(ctx context.Context, ptId string, idType constant.DistributedIDType) (string, *bmserror.BMSError) {
	period, idOffset, err := DailyIDCreator.GetBaseDatePeriodID(ctx, timeutil.TodayDateStr(timeutil.DateIntFormatYYMMDD), idType, ptId)
	if err != nil {
		return "", err.Mark()
	}
	return fmt.Sprintf("%s%s%d%s", constant.DistributedIDTypePerfixMap[idType], GetNewPtID(ptId), period, ConvertToNewGenerationIDWithBit(idOffset, 4)), nil
}

func GenUserNo(ctx context.Context) (string, *bmserror.BMSError) {
	id, err := DIDCreator.GetID(ctx, constant.UserNo, ALLPointType)
	if err != nil {
		return "", err.Mark()
	}
	paddingNumStr := LeftPadCharWithMinLen(convert.ToString(id), '0', 10)
	return paddingNumStr, nil
}

func GenerateTransactionID(ctx context.Context, ptId string) (int64, *bmserror.BMSError) {
	period, num, wcErr := DailyIDCreator.GetDatePeriodID(ctx, constant.TransactionID, ptId)
	if wcErr != nil {
		return 0, wcErr.Mark()
	}
	idStr := fmt.Sprintf("%v%v%08d", period, 0, num)
	id, convertErr := convert.StringToInt64(idStr)
	if convertErr != nil {
		return 0, bmserror.NewError(constant.ErrInternalServer, convertErr.Error())
	}
	return id, nil
}

const ALLPointType = "ALL"

func LeftPadCharWithMinLen(s string, padChar byte, minLen int) string {
	padCountInt := minLen - len(s)
	if padCountInt <= 0 {
		return s
	}
	padStr := string(padChar)
	retStr := strings.Repeat(padStr, padCountInt) + s
	return retStr
}
