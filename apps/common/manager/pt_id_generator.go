package cmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/datasource"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/orm"
	"github.com/feimumoke/labequipbms/framework/support/expression"
	"github.com/feimumoke/labequipbms/framework/transaction"
)

type Generator struct {
	ds        datasource.DataSource
	mu        sync.Mutex
	IDTypeMap map[string]*IDSegment
}

func NewIdGenerator(idTypeMap map[string]*IDSegment) *Generator {
	return &Generator{ds: datasource.NewDefaultDataSource(constant.DataSourceBasic,
		func(key string, reps []orm.GORM) int {
			return 0
		},
		func(key string, reps []orm.GORM) int {
			return 0
		},
	), IDTypeMap: idTypeMap}
}

type IDSegment struct {
	IDType constant.DistributedIDType
	Offset int64
	Max    int64
}

type PtNoType = int64

const DefaultInitID int64 = 1
const Unlimited = int64(^uint(0) >> 1)
const DefaultStep int64 = 100

func (i *Generator) GetID(ctx context.Context, idType constant.DistributedIDType, datePeriod int64, ptNo PtNoType, step int64) (int64, *bmserror.BMSError) {
	i.mu.Lock()
	defer i.mu.Unlock()
	key := getKey(idType, datePeriod, ptNo)
	segment := i.getSegment(key)
	var err *bmserror.BMSError
	if segment == nil || segment.Offset >= segment.Max {
		var idStep = expression.If(step == 0, DefaultStep, step).(int64)
		segment, err = i.createNewSegment(ctx, idType, datePeriod, ptNo, idStep)
		if err != nil {
			return 0, err.Mark()
		}
	}
	newID := segment.Offset
	segment.Offset++
	return newID, nil
}

func (i *Generator) getSegment(key string) *IDSegment {
	segment := i.IDTypeMap[key]
	return segment
}

func (i *Generator) createNewSegment(ctx context.Context, idType constant.DistributedIDType, datePeriod int64, ptNo PtNoType, step int64) (*IDSegment, *bmserror.BMSError) {
	var idOffset int64
	var limitNum int64
	var err *bmserror.BMSError
	for {
		var idOffsetDB int64
		//不管外层用不用事务，这里肯定不能用事务；否则会出现异常
		err = transaction.PropagationNotSupported(ctx, func(ctx context.Context) *bmserror.BMSError {
			idOffsetDBInner, err := i.getIDOffset(ctx, idType, datePeriod, ptNo, step, Unlimited)
			if err != nil {
				return err.Mark()
			}
			idOffsetDB = idOffsetDBInner
			return nil
		})
		if err != nil {
			log.Errorf("get id offset fail,err=%v", err)
			return nil, err.Mark()
		}

		if limitNum > 10 {
			panic("get id offset over maximum limit")
		}

		if idOffsetDB >= 1 {
			idOffset = idOffsetDB
			break
		}
		limitNum++
	}
	idSegment := &IDSegment{
		IDType: idType,
		Offset: idOffset,
		Max:    idOffset + step,
	}
	key := getKey(idType, datePeriod, ptNo)
	i.setSegment(key, idSegment)
	return idSegment, nil
}

func (i *Generator) setSegment(key string, idSegment *IDSegment) {
	i.IDTypeMap[key] = idSegment
}

func getKey(idType constant.DistributedIDType, datePeriod int64, ptId PtNoType) string {
	return fmt.Sprintf("%v%v%v", idType, datePeriod, ptId)
}

func (i *Generator) getIDOffset(ctx context.Context, idType constant.DistributedIDType, datePeriod int64, ptNo PtNoType, step int64, maxValue int64) (int64, *bmserror.BMSError) {
	idList, err := i.search(ctx, idType, datePeriod, ptNo)
	if err != nil {
		log.Errorf("get id type fail, err=%v", err)
		return 0, bmserror.NewError(constant.ErrInternalServer, "get id type fail")
	}
	if len(idList) == 0 {
		//初始化
		err := i.create(ctx, &DistributedIDcreatorTab{
			IDType:      idType,
			DatePeriod:  datePeriod,
			PtNo:        ptNo,
			IDValue:     DefaultInitID,
			Description: "",
			Mtime:       time.Now().Unix(),
		})
		if err != nil {
			log.Infof("create id err,err=%v", err)
		}
		return 0, nil
	}
	idOffset := idList[0].IDValue
	if maxValue > 0 && idOffset >= maxValue {
		return 0, bmserror.NewError(constant.ErrDB, "id value over maximum limit")
	}
	rows, err := i.update(ctx, idType, datePeriod, ptNo, idOffset, idOffset+step)
	if err != nil {
		return 0, err.Mark()
	}
	if rows > 0 {
		return idOffset, nil
	}
	return 0, nil
}
