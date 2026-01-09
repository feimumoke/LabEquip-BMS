package transaction

import (
	"context"

	"runtime/debug"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/support/monitor"
)

type TxAdvice interface {
	InitAdvice(ctx context.Context) interface{}
	BeforeCommit(ctx context.Context, info interface{}, err *bmserror.BMSError) *bmserror.BMSError
	DoCommit(ctx context.Context, info interface{}, err *bmserror.BMSError) *bmserror.BMSError
	AfterCommit(ctx context.Context, info interface{}, err *bmserror.BMSError) *bmserror.BMSError
}

const TransactionAspect = "tx_aspect"

func init() {
	RegisterAdvice(TransactionAspect, &TransactionContext{})
}

type TxAdviceContext struct {
	Info   interface{}
	Advice TxAdvice
}

type TxAspectInfo struct {
	CurrentDbIdentifier string
	CrossCuttings       map[string]*TxAdviceContext
}

// 当前ctx是否存在事务
func GetTxAspect(ctx context.Context) *TxAspectInfo {
	aspects := ctx.Value(ExtraTxAspectKey)
	if aspects == nil {
		return nil
	}
	return aspects.(*TxAspectInfo)
}

func GetTransactionContext(ctx context.Context) *TransactionContext {
	aspects := GetTxAspect(ctx)
	if aspects == nil {
		return nil
	}
	adviceContext := aspects.CrossCuttings[TransactionAspect]
	if adviceContext == nil {
		return nil
	}
	return adviceContext.Info.(*TransactionContext)
}

var CommonAdvice = make(map[string]TxAdvice)

func RegisterAdvice(name string, advice TxAdvice) {
	CommonAdvice[name] = advice
}

func InitAspect(ctx context.Context, id string) *TxAspectInfo {
	aspectInfo := &TxAspectInfo{CurrentDbIdentifier: id, CrossCuttings: make(map[string]*TxAdviceContext)}
	log.Infof("InitAspect: %v", id)
	defer func() {
		if p := recover(); p != nil {
			//打印调用栈信息
			errStack := string(debug.Stack())
			log.CtxErrorf(ctx, "Transaction InitAspect panic: %v", errStack)
			_ = monitor.AwesomeReportEventWithoutTrans(ctx, TransactionModule, "transaction", "-1", string(debug.Stack()))
		}
	}()
	for name, advice := range CommonAdvice {
		aspectInfo.CrossCuttings[name] = &TxAdviceContext{
			Info:   advice.InitAdvice(ctx),
			Advice: advice,
		}
	}
	return aspectInfo
}

func BeforeTxCommit(ctx context.Context, err *bmserror.BMSError) (rerr *bmserror.BMSError) {
	var wcerr *bmserror.BMSError
	if err != nil {
		log.Infof("exec BeforeTxCommit error %v", err)
	}
	defer func() {
		if wcerr != nil {
			log.CtxErrorf(ctx, "Transaction advice BeforeCommit err: %v", wcerr.DebugError())
		}
		if p := recover(); p != nil {
			//打印调用栈信息
			errStack := string(debug.Stack())
			log.CtxErrorf(ctx, "Transaction advice BeforeCommit panic: %v", errStack)
			_ = monitor.AwesomeReportEventWithoutTrans(ctx, "transaction", "BeforeTxCommit", "-1", string(debug.Stack()))
			rerr = bmserror.NewError(constant.ErrInternalServer, "%v", p)
		}
	}()
	aspects := GetTxAspect(ctx)
	if aspects == nil {
		return nil
	}
	var errList []*bmserror.BMSError
	for name, adviceCtx := range aspects.CrossCuttings {
		if bfErr := adviceCtx.Advice.BeforeCommit(ctx, adviceCtx.Info, err); bfErr != nil {
			log.Errorf("Transaction %v BeforeCommit %v Advice err %v", aspects.CurrentDbIdentifier, name, bfErr)
			errList = append(errList, bfErr)
			// 先提交 比较重要 采用错误传播
			if err == nil {
				err = bfErr
			}
		}
	}
	if len(errList) > 0 {
		wcerr = DealErrorList(errList)
	}
	return wcerr
}

func DoTxCommit(ctx context.Context, err *bmserror.BMSError) (rerr *bmserror.BMSError) {
	var wcerr *bmserror.BMSError
	if err != nil {
		log.Infof("exec DoCommit error %v", err)
	}
	defer func() {
		if wcerr != nil {
			log.CtxErrorf(ctx, "Transaction advice DoCommit err: %v", wcerr.DebugError())
		}
		if p := recover(); p != nil {
			//打印调用栈信息
			errStack := string(debug.Stack())
			log.CtxErrorf(ctx, "Transaction advice DoCommit panic: %v", errStack)
			_ = monitor.AwesomeReportEventWithoutTrans(ctx, "transaction", "DoCommit", "-1", string(debug.Stack()))
			rerr = bmserror.NewError(constant.ErrInternalServer, "%v", p)
		}
	}()
	aspects := GetTxAspect(ctx)
	if aspects == nil {
		return nil
	}
	var errList []*bmserror.BMSError
	for name, adviceCtx := range aspects.CrossCuttings {
		if bfErr := adviceCtx.Advice.DoCommit(ctx, adviceCtx.Info, err); bfErr != nil {
			log.Errorf("Transaction %v DoCommit %v Advice err %v", aspects.CurrentDbIdentifier, name, bfErr)
			errList = append(errList, bfErr)
			// 先提交 比较重要 采用错误传播
			if err == nil {
				err = bfErr
			}
		}
	}
	if len(errList) > 0 {
		wcerr = DealErrorList(errList)
	}
	return wcerr
}

const TransactionModule = "TransactionException"

func AfterTxCommit(ctx context.Context, err *bmserror.BMSError) (rerr *bmserror.BMSError) {
	var wcerr *bmserror.BMSError
	if err != nil {
		log.Infof("exec AfterTxCommit err %v", err)
	}
	defer func() {
		if wcerr != nil {
			log.CtxErrorf(ctx, "Transaction advice AfterCommit err: %v", wcerr.DebugError())
		}
		if p := recover(); p != nil {
			//打印调用栈信息
			errStack := string(debug.Stack())
			log.CtxErrorf(ctx, "Transaction advice AfterCommit panic: %v", errStack)
			_ = monitor.AwesomeReportEventWithoutTrans(ctx, TransactionModule, "transaction", "-1", string(debug.Stack()))
			rerr = bmserror.NewError(constant.ErrInternalServer, "%v", p)
		}
	}()
	aspects := GetTxAspect(ctx)
	if aspects == nil {
		return nil
	}
	var errList []*bmserror.BMSError
	for name, adviceCtx := range aspects.CrossCuttings {
		if afErr := adviceCtx.Advice.AfterCommit(ctx, adviceCtx.Info, err); afErr != nil {
			log.Errorf("Transaction %v AfterCommit %v Advice err %v", aspects.CurrentDbIdentifier, name, afErr)
			errList = append(errList, afErr)
		}
	}
	if len(errList) > 0 {
		wcerr = DealErrorList(errList)
	}
	return wcerr
}

func DealErrorList(errList []*bmserror.BMSError) *bmserror.BMSError {
	if len(errList) == 0 {
		return nil
	}
	var returnErr *bmserror.BMSError
	for _, err := range errList {
		if err != nil {
			if returnErr == nil {
				returnErr = err
				continue
			}
			returnErr = returnErr.AddError(err.Code(), err.Message())
		}
	}
	if returnErr != nil {
		return returnErr.Mark()
	}
	return nil
}

func ExtraAdviceInfoInTx(ctx context.Context, adviceKey string) interface{} {
	aspects := GetTxAspect(ctx)
	if aspects == nil {
		return nil
	}
	adviceContext := aspects.CrossCuttings[adviceKey]
	if adviceContext == nil {
		return nil
	}
	return adviceContext.Info
}
