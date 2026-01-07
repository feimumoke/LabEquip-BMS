package orm

import (
	"context"
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"time"
)

var (
	// ErrRecordNotFound record not found error
	ErrRecordNotFound = gorm.ErrRecordNotFound
	// ErrInvalidTransaction invalid transaction when you are trying to `Commit` or `Rollback`
	ErrInvalidTransaction = gorm.ErrInvalidTransaction
	// ErrNotImplemented not implemented
	ErrNotImplemented = gorm.ErrNotImplemented
	// ErrMissingWhereClause missing where clause
	ErrMissingWhereClause = gorm.ErrMissingWhereClause
	// ErrUnsupportedRelation unsupported relations
	ErrUnsupportedRelation = gorm.ErrUnsupportedRelation
	// ErrPrimaryKeyRequired primary keys required
	ErrPrimaryKeyRequired = gorm.ErrPrimaryKeyRequired
	// ErrModelValueRequired model value required
	ErrModelValueRequired = gorm.ErrModelValueRequired
	// ErrInvalidData unsupported data
	ErrInvalidData = gorm.ErrInvalidData
	// ErrUnsupportedDriver unsupported driver
	ErrUnsupportedDriver = gorm.ErrUnsupportedDriver
	// ErrRegistered registered
	ErrRegistered = gorm.ErrRegistered
	// ErrInvalidField invalid field
	ErrInvalidField = gorm.ErrInvalidField
	// ErrEmptySlice empty slice found
	ErrEmptySlice = gorm.ErrEmptySlice
	// ErrDryRunModeUnsupported dry run mode unsupported
	ErrDryRunModeUnsupported = gorm.ErrDryRunModeUnsupported
	// ErrInvalidDB invalid db
	ErrInvalidDB = gorm.ErrInvalidDB
	// ErrInvalidValue invalid value
	ErrInvalidValue = gorm.ErrInvalidValue
	// ErrInvalidValueOfLength invalid values do not match length
	ErrInvalidValueOfLength = gorm.ErrInvalidValueOfLength
	// ErrPreloadNotAllowed preload is not allowed when count is used
	ErrPreloadNotAllowed = gorm.ErrPreloadNotAllowed
)

func TryGetErrByString(errStr string) error {
	var err error
	switch errStr {
	case ErrRecordNotFound.Error():
		err = ErrRecordNotFound
	case ErrInvalidTransaction.Error():
		err = ErrInvalidTransaction
	case ErrInvalidTransaction.Error():
		err = ErrInvalidTransaction
	case ErrNotImplemented.Error():
		err = ErrNotImplemented
	case ErrMissingWhereClause.Error():
		err = ErrMissingWhereClause
	case ErrUnsupportedRelation.Error():
		err = ErrUnsupportedRelation
	case ErrPrimaryKeyRequired.Error():
		err = ErrPrimaryKeyRequired
	case ErrModelValueRequired.Error():
		err = ErrModelValueRequired
	case ErrInvalidData.Error():
		err = ErrInvalidData
	case ErrUnsupportedDriver.Error():
		err = ErrUnsupportedDriver
	case ErrRegistered.Error():
		err = ErrRegistered
	case ErrInvalidField.Error():
		err = ErrInvalidField
	case ErrEmptySlice.Error():
		err = ErrEmptySlice
	case ErrDryRunModeUnsupported.Error():
		err = ErrDryRunModeUnsupported
	case ErrInvalidDB.Error():
		err = ErrInvalidDB
	case ErrInvalidValue.Error():
		err = ErrInvalidValue
	case ErrInvalidValueOfLength.Error():
		err = ErrInvalidValueOfLength
	case ErrPreloadNotAllowed.Error():
		err = ErrPreloadNotAllowed
	}
	return err
}

type ErrSQLCommon struct {
	err error
}

func (e *ErrSQLCommon) Joins(query string, args ...interface{}) GORM {
	return e
}

func (e *ErrSQLCommon) Preload(query string, args ...interface{}) GORM {
	return e
}

func (e *ErrSQLCommon) AddError(err error) (tx GORM) {
	if e.err != nil {
		e.err = fmt.Errorf("%v,%v", e.err, err)
	} else {
		e.err = err
	}
	return e
}

func (e *ErrSQLCommon) Assign(attrs ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Attrs(attrs ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Begin(opts ...*sql.TxOptions) GORM {
	return e
}

func (e *ErrSQLCommon) Clauses(conds ...clause.Expression) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Close() *bmserror.BMSError {
	if e.err != nil {
		return bmserror.NewError(constant.ErrDB, e.err.Error())
	}
	return nil
}

func (e *ErrSQLCommon) Commit() GORM {
	return e
}

func (e *ErrSQLCommon) Connection(fc func(tx GORM) error) (err error) {
	return e.err
}

func (e *ErrSQLCommon) Count(count *int64) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Create(value interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) CreateInBatches(value interface{}, batchSize int) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) DB() *sql.DB {
	return nil
}

func (e *ErrSQLCommon) Debug() (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Delete(value interface{}, conds ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Distinct(args ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Exec(sql string, values ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Find(dest interface{}, conds ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) FindInBatches(dest interface{}, batchSize int, fc func(tx GORM, batch int) error) GORM {
	return e
}

func (e *ErrSQLCommon) First(dest interface{}, conds ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) FirstOrCreate(dest interface{}, conds ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) FirstOrInit(dest interface{}, conds ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) ForShare() GORM {
	return e
}

func (e *ErrSQLCommon) ForUpdate() GORM {
	return e
}

func (e *ErrSQLCommon) Get(key string) (interface{}, bool) {
	return nil, false
}

func (e *ErrSQLCommon) GetConfig() *Config {
	return nil
}

func (e *ErrSQLCommon) GetError() *bmserror.BMSError {
	if e.err != nil {
		return bmserror.NewError(constant.ErrDB, e.err.Error())
	}
	return nil
}

func (e *ErrSQLCommon) GetGormDB() *gorm.DB {
	return nil
}

func (e *ErrSQLCommon) GetStatement() *Statement {
	return nil
}

func (e *ErrSQLCommon) Group(name string) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Having(query interface{}, args ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) InstanceGet(key string) (interface{}, bool) {
	return nil, false
}

func (e *ErrSQLCommon) InstanceSet(key string, value interface{}) GORM {
	return e
}

func (e *ErrSQLCommon) InstantSet(key string, value interface{}) GORM {
	return e
}

func (e *ErrSQLCommon) IsAutoReport() bool {
	return false
}

func (e *ErrSQLCommon) Last(dest interface{}, conds ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Limit(limit int) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) LogMode(enable bool) GORM {
	return e
}

func (e *ErrSQLCommon) Model(value interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Not(query interface{}, args ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Offset(offset int) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Omit(columns ...string) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Or(query interface{}, args ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Order(value interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Pluck(column string, dest interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) RecordNotFound() bool {
	return e.err == ErrRecordNotFound
}

func (e *ErrSQLCommon) Raw(sql string, values ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Rollback() GORM {
	return e
}

func (e *ErrSQLCommon) RollbackTo(name string) GORM {
	return e
}

func (e *ErrSQLCommon) Row() *sql.Row {
	return nil
}

func (e *ErrSQLCommon) Rows() (*sql.Rows, error) {
	return nil, e.err
}

func (e *ErrSQLCommon) RowsAffected() int64 {
	return 0
}

func (e *ErrSQLCommon) Save(value interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) SavePoint(name string) GORM {
	return e
}

func (e *ErrSQLCommon) Scan(dest interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) ScanRows(rows *sql.Rows, dest interface{}) error {
	return e.err
}

func (e *ErrSQLCommon) Scopes(funcs ...func(GORM) GORM) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Select(query interface{}, args ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Session(config *Session) GORM {
	return e
}

func (e *ErrSQLCommon) Set(key string, value interface{}) GORM {
	return e
}

func (e *ErrSQLCommon) SetConnMaxLifetime(d time.Duration) {
}

func (e *ErrSQLCommon) SetLogger(log logger.Interface) {
}

func (e *ErrSQLCommon) SetMaxIdleConns(n int) {
}

func (e *ErrSQLCommon) SetMaxOpenConns(n int) {
}

func (e *ErrSQLCommon) Table(name string, args ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Take(dest interface{}, conds ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) ToSQL(queryFn func(tx GORM) GORM) string {
	return ""
}

func (e *ErrSQLCommon) Unscoped() (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Update(column string, value interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) UpdateColumn(column string, value interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) UpdateColumns(values interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Updates(values interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) Value() interface{} {
	return e
}

func (e *ErrSQLCommon) Where(query interface{}, args ...interface{}) (tx GORM) {
	return e
}

func (e *ErrSQLCommon) WithContext(ctx context.Context) GORM {
	return e
}

func (e *ErrSQLCommon) SwitchDataSource(dataSourceName string) GORM {
	return e
}

func (e *ErrSQLCommon) GetDataSourceName() string {
	return ""
}

func (e *ErrSQLCommon) Master() GORM {
	return e
}

func (e *ErrSQLCommon) Replica() GORM {
	return e
}

func (e *ErrSQLCommon) ReplicaAt(index int) GORM {
	return e
}

func (e *ErrSQLCommon) Replicas() []GORM {
	return []GORM{e}
}

func (e *ErrSQLCommon) WithReplicaRouteKey(key string) GORM {
	return e
}

func (e *ErrSQLCommon) WithReplicaRouteStrategy(custom func(key string, reps []GORM) int) GORM {
	return e
}

func (e *ErrSQLCommon) RouteGroup() GORM {
	return e
}

func (e *ErrSQLCommon) GetGroupKey() string {
	return ""
}

func (e *ErrSQLCommon) GroupAt(index int) GORM {
	return e
}

func (e *ErrSQLCommon) Groups() []GORM {
	return []GORM{e}
}

func (e *ErrSQLCommon) WithRouteKey(key string) GORM {
	return e
}

func (e *ErrSQLCommon) WithRouteStrategy(custom func(key string, reps []GORM) int) GORM {
	return e
}

func (e *ErrSQLCommon) NextCallback() {
}

func (e *ErrSQLCommon) AbortCallback(err error) {
	e.AddError(err)
}

func (e *ErrSQLCommon) BatchUpdate(tableName string, itemList interface{}) (int64, *bmserror.BMSError) {
	return 0, nil
}
