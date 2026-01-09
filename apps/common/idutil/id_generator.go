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

func GenBorrowTaskId(ctx context.Context) (string, *bmserror.BMSError) {
	period, num, err := DailyIDCreator.GetDatePeriodID(ctx, constant.BorrowTaskID, ALLPointType)
	if err != nil {
		return "", err.Mark()
	}
	newPeriod := strconv.FormatInt(period, 10)[2:]
	newNum := ConvertToNewGenerationIDWithBit(num, 4)
	inboundID := fmt.Sprintf("%s%s%s", constant.DistributedIDTypePerfixMap[constant.BorrowTaskID], newPeriod, newNum)
	return inboundID, nil
}

func GenerateTaskNumber(ctx context.Context, idType constant.DistributedIDType) (string, *bmserror.BMSError) {
	period, idOffset, err := DailyIDCreator.GetBaseDatePeriodID(ctx, timeutil.TodayDateStr(timeutil.DateIntFormatYYMMDD), idType, ALLPointType)
	if err != nil {
		return "", err.Mark()
	}
	return fmt.Sprintf("%s%d%s", constant.DistributedIDTypePerfixMap[idType], period, ConvertToNewGenerationIDWithBit(idOffset, 4)), nil
}

func GenUserNo(ctx context.Context) (string, *bmserror.BMSError) {
	id, err := DIDCreator.GetID(ctx, constant.UserNo, ALLPointType)
	if err != nil {
		return "", err.Mark()
	}
	paddingNumStr := LeftPadCharWithMinLen(convert.ToString(id), '0', 10)
	return paddingNumStr, nil
}

func GenerateTransactionID(ctx context.Context) (string, *bmserror.BMSError) {
	period, num, wcErr := DailyIDCreator.GetDatePeriodID(ctx, constant.TransactionID, ALLPointType)
	if wcErr != nil {
		return "", wcErr.Mark()
	}
	idStr := fmt.Sprintf("%v%v%08d", period, 0, num)
	id, convertErr := convert.StringToInt64(idStr)
	if convertErr != nil {
		return "", bmserror.NewError(constant.ErrInternalServer, convertErr.Error())
	}
	return convert.Int64ToString(id), nil
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
