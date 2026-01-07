package transaction

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync/atomic"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/appcontext"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/orm"
	"github.com/feimumoke/labequipbms/framework/support/monitor"
	"github.com/feimumoke/labequipbms/framework/support/trace"
	"gorm.io/gorm"
)

type businessFunc func(ctx context.Context) *bmserror.BMSError

type propagationOption struct {
	newContextInTx bool
}

type ExtraDataKeyType string

const (
	ExtraDataKey     ExtraDataKeyType = "extra_data_key"
	ExtraTxAspectKey ExtraDataKeyType = "extra_tx_aspect_key"
)

type ExtraData struct {
	start              bool  // InTransactionCount 并发时有问题 虽然事务中不允许并发
	InTransactionCount int64 //进入事务次数
	NeedSendCount      int64 //发送了多少次事件
	CompleteCh         chan struct{}
	FailCh             chan struct{}
}

func InitExtraData() *ExtraData {
	e := &ExtraData{
		InTransactionCount: 0,
		NeedSendCount:      0,
		CompleteCh:         make(chan struct{}, 512),
		FailCh:             make(chan struct{}, 512),
	}
	return e
}

func IsCtxInTransaction(ctx context.Context) bool {
	return true
}

type PropagationOption func(opt *propagationOption)

func PropagationNever(ctx context.Context, f businessFunc, opts ...PropagationOption) *bmserror.BMSError {
	com := orm.Context(ctx)
	if isInTx(com) {
		err := bmserror.NewError(constant.ErrDB, fmt.Sprintf("run propagation never in transaction"))
		com.GetConfig().Logger.Error(ctx, "%v", err)
		return err
	}
	return f(ctx)
}

func PropagationSupports(ctx context.Context, f businessFunc, opts ...PropagationOption) *bmserror.BMSError {
	return f(ctx)
}

func PropagationRequiresNew(ctx context.Context, f businessFunc, opts ...PropagationOption) (err *bmserror.BMSError) {
	var opt propagationOption
	for _, o := range opts {
		o(&opt)
	}
	com := orm.Context(ctx)
	com.GetStatement().ConnPool = com.GetConfig().ConnPool
	if opt.newContextInTx {
		com.GetStatement().Context = context.Background()
	}
	tx := com.Begin()
	err = tx.GetError()
	if err != nil {
		return err
	}
	defer func() {
		if panik := recover(); panik != nil {
			err = bmserror.NewError(constant.ErrDB, fmt.Sprintf("panic: %v", panik))
		}
		if err != nil {
			com.GetConfig().Logger.Error(ctx, "run propagation requires new got err: %v", err)
			suberr := tx.Rollback().GetError()
			if suberr != nil {
				com.GetConfig().Logger.Error(ctx, "run propagation requires new and rollback got err: %v", suberr)
				err = bmserror.NewError(constant.ErrDB, fmt.Sprintf("got err:%v and rollback failed: %v", err, suberr))
			}
		} else {
			err = tx.Commit().GetError()
			if err != nil {
				com.GetConfig().Logger.Error(ctx, "run propagation requires new and commit got err: %v", err)
			}
		}
	}()

	err = f(orm.BindContext(ctx, tx))

	return err
}

func PropagationRequiredV1(ctx context.Context, f businessFunc, opts ...PropagationOption) (err *bmserror.BMSError) {
	var opt propagationOption
	for _, o := range opts {
		o(&opt)
	}
	com := orm.Context(ctx)
	if opt.newContextInTx {
		com.GetStatement().Context = context.Background()
	}

	if !isInTx(com) {
		tx := com.Begin()
		err = tx.GetError()
		if err != nil {
			return err
		}
		defer func() {
			if panik := recover(); panik != nil {
				err = bmserror.NewError(constant.ErrDB, fmt.Sprintf("panic: %v", panik))
			}
			if err != nil {
				com.GetConfig().Logger.Error(ctx, "run propagation required got err: %v", err)
				suberr := tx.Rollback().GetError()
				if suberr != nil {
					com.GetConfig().Logger.Error(ctx, "run propagation required and rollback got err: %v", suberr)
					err = bmserror.NewError(constant.ErrDB, fmt.Sprintf("got err:%v and rollback failed: %v", err, suberr))
				}
			} else {
				err = tx.Commit().GetError()
				if err != nil {
					com.GetConfig().Logger.Error(ctx, "run propagation required and commit got err: %v", err)
				}
			}
		}()
		ctx = orm.BindContext(ctx, tx)
	}

	err = f(ctx)

	return err
}
func PropagationNotSupported(ctx context.Context, f businessFunc, opts ...PropagationOption) *bmserror.BMSError {
	ctx = appcontext.BindContext(ctx)
	return f(ctx)
}

func PropagationRequired(ctx context.Context, f businessFunc, opts ...PropagationOption) (wcErr *bmserror.BMSError) {
	var opt propagationOption
	for _, o := range opts {
		o(&opt)
	}
	db := orm.Context(ctx)
	if opt.newContextInTx {
		db.GetStatement().Context = context.Background()
	}
	db.GetConfig().AspectTxMode = true

	extraDataT := ctx.Value(ExtraDataKey)
	if extraDataT == nil {
		extraDataT = InitExtraData()
		ctx = context.WithValue(ctx, ExtraDataKey, extraDataT)
	}

	extraData := extraDataT.(*ExtraData)

	atomic.AddInt64(&extraData.InTransactionCount, 1)

	defer func() {
		atomic.AddInt64(&extraData.InTransactionCount, -1)
		if extraData.InTransactionCount == 0 {
			log.Infof("PropagationRequired send message")
			if wcErr == nil {
				extraData.CompleteCh <- struct{}{}
			} else {
				extraData.FailCh <- struct{}{}
			}
		}
	}()

	needCommit := false
	if !extraData.start {
		ctx = context.WithValue(ctx, ExtraTxAspectKey, InitAspect(ctx, trace.GetOrNewTraceID(ctx)))
		extraData.start = true
		needCommit = true
		defer func() {
			if panik := recover(); panik != nil {
				wcErr = bmserror.NewError(constant.ErrDB, fmt.Sprintf("panic: %v", panik))
			}
			AfErr := AfterTxCommit(ctx, wcErr)
			if AfErr != nil {
				log.Errorf("AfterCommit err %v", AfErr)
			}
		}()
	}
	innerFunc := func(ctx context.Context) (busErr *bmserror.BMSError) {
		defer func() {
			if busErr != nil {
				log.CtxErrorf(ctx, "PropagationRequired err: %v", busErr.DebugError())
			}
			if p := recover(); p != nil {
				//打印调用栈信息
				errStack := string(debug.Stack())
				log.CtxErrorf(ctx, "PropagationRequired inner panic: %v", errStack)

				_ = monitor.AwesomeReportEventWithoutTrans(ctx, TransactionModule, "transaction", "-1", string(debug.Stack()))
				busErr = bmserror.NewError(constant.ErrInternalServer, "%v", p)
			}
		}()
		busErr = f(ctx)

		return busErr
	}
	ctx = orm.BindContext(ctx, db)
	wcErr = innerFunc(ctx)
	if needCommit {
		if bfErr := DoTxCommit(ctx, wcErr); bfErr != nil {
			wcErr = bfErr
		}
		if bfErr := BeforeTxCommit(ctx, wcErr); bfErr != nil {
			wcErr = bfErr
		}
	}
	return wcErr
}

func PropagationMandatory(ctx context.Context, f businessFunc, opts ...PropagationOption) error {
	com := orm.Context(ctx)
	if !isInTx(com) {
		err := fmt.Errorf("run propagation mandatory in NO transaction")
		com.GetConfig().Logger.Error(ctx, "%v", err)
		return err
	}
	return f(ctx)
}

func PropagationNested(ctx context.Context, f businessFunc, opts ...PropagationOption) (err error) {
	var opt propagationOption
	for _, o := range opts {
		o(&opt)
	}
	com := orm.Context(ctx)
	if opt.newContextInTx {
		com.GetStatement().Context = context.Background()
	}

	if !isInTx(com) {
		tx := com.Begin()
		err = tx.GetError()
		if err != nil {
			return err
		}
		defer func() {
			if panik := recover(); panik != nil {
				err = fmt.Errorf("panic: %v", panik)
			}
			if err != nil {
				com.GetConfig().Logger.Error(ctx, "run propagation nested got err: %v", err)
				suberr := tx.Rollback().GetError()
				if suberr != nil {
					com.GetConfig().Logger.Error(ctx, "run propagation nested and rollback got err: %v", suberr)
					err = fmt.Errorf("got err:%v and rollback failed: %v", err, suberr)
				}
			} else {
				err = tx.Commit().GetError()
				if err != nil {
					com.GetConfig().Logger.Error(ctx, "run propagation nested and commit got err: %v", err)
				}
			}
		}()
		ctx = orm.BindContext(ctx, tx)
	} else {
		name := fmt.Sprintf("sp_%p", com)
		tx := com.SavePoint(name)
		err = tx.GetError()
		if err != nil {
			return err
		}
		defer func() {
			if panik := recover(); panik != nil {
				err = fmt.Errorf("panic: %v", panik)
			}
			if err != nil {
				com.GetConfig().Logger.Error(ctx, "run propagation nested got err: %v", err)
				suberr := tx.RollbackTo(name).GetError()
				if suberr != nil {
					com.GetConfig().Logger.Error(ctx, "run propagation nested and rollback to %s got err: %v", name, suberr)
					err = fmt.Errorf("got err:%v and rollback to %s failed: %v", err, name, suberr)
				}
			}
		}()
		ctx = orm.BindContext(ctx, tx)
	}

	err = f(ctx)

	return err
}

func isInTx(com orm.GORM) bool {
	pool := com.GetStatement().ConnPool
	if pool == nil {
		return false
	}

	_, ok := pool.(gorm.TxCommitter)

	return ok
}
