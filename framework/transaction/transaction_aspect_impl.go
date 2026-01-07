package transaction

import (
	"context"
	"sync"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/orm"
)

type TransactionContext struct {
	Mutex            sync.Mutex
	multiPersistConn map[string]orm.GORM
}

func (s *TransactionContext) GetPersistConn(key string) orm.GORM {
	if conn, ok := s.multiPersistConn[key]; ok {
		return conn
	}
	return nil
}

func (s *TransactionContext) SetPersistConn(key string, conn orm.GORM) {
	s.multiPersistConn[key] = conn
}

func (s *TransactionContext) InitAdvice(_ context.Context) interface{} {
	return &TransactionContext{multiPersistConn: make(map[string]orm.GORM)}
}

func (s *TransactionContext) BeforeCommit(ctx context.Context, info interface{}, err *bmserror.BMSError) *bmserror.BMSError {
	return nil
}

func (s *TransactionContext) DoCommit(ctx context.Context, info interface{}, err *bmserror.BMSError) *bmserror.BMSError {
	txCtx := info.(*TransactionContext)
	if txCtx == nil {
		log.Errorf("TransactionContext get null dsn in DoCommit")
		return nil
	}
	var commitErr *bmserror.BMSError
	if err == nil {
		for name, conn := range txCtx.multiPersistConn {
			if innerErr := conn.Commit().GetError(); innerErr != nil {
				log.CtxErrorf(ctx, "%v commit err: %+v", name, innerErr.Error())
				if commitErr == nil {
					commitErr = bmserror.NewError(constant.ErrDB, innerErr.Error())
				} else {
					commitErr.AddError(constant.ErrDB, innerErr.Error())
				}
			}
		}

	} else {
		for name, conn := range txCtx.multiPersistConn {
			if innerErr := conn.Rollback().GetError(); innerErr != nil {
				log.CtxErrorf(ctx, "%v rollback err: %+v", name, innerErr.Error())
				if commitErr == nil {
					commitErr = bmserror.NewError(constant.ErrDB, innerErr.Error())
				} else {
					commitErr.AddError(constant.ErrDB, innerErr.Error())
				}
			}
		}
	}
	return nil
}

func (s TransactionContext) AfterCommit(ctx context.Context, info interface{}, err *bmserror.BMSError) *bmserror.BMSError {

	return nil
}
