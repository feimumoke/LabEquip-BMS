package orm

import (
	"context"
	"database/sql"

	"github.com/feimumoke/labequipbms/framework/bmserror"
	"gorm.io/gorm"
)

type keyType string

const commonKey keyType = "gorm-root-key"

// get sqlCommon from context
func getSQLCommon(ctx context.Context) GORM {
	if sqlCommon, ok := ctx.Value(commonKey).(GORM); ok {
		return sqlCommon
	}
	return nil
}

// BindContext add sqlCommon into ctx
func BindContext(ctx context.Context, db GORM) context.Context {
	return context.WithValue(ctx, commonKey, db)
}

// open api for for repository
func Context(ctx context.Context) GORM {
	if sqlCommon := getSQLCommon(ctx); sqlCommon != nil {
		return sqlCommon.WithContext(ctx)
	}
	panic("can't get sqlCommon")
}

type Statement = gorm.Statement
type Session = gorm.Session

type GORM interface {
	dataSourceOperation
	SingleDB
	GroupDB
	ClusterDB
	WithContext(ctx context.Context) GORM
}

type SingleDB interface {
	Exec(query string, args ...interface{}) GORM
	Close() *bmserror.BMSError
	GetError() *bmserror.BMSError
	Where(query interface{}, args ...interface{}) GORM
	Or(query interface{}, args ...interface{}) GORM
	Not(query interface{}, args ...interface{}) GORM
	Limit(limit int) GORM
	Offset(offset int) GORM
	Order(value interface{}) GORM
	First(out interface{}, where ...interface{}) GORM
	Take(out interface{}, where ...interface{}) GORM
	Last(out interface{}, where ...interface{}) GORM
	Find(out interface{}, where ...interface{}) GORM
	Select(query interface{}, args ...interface{}) (tx GORM)
	Scan(dest interface{}) GORM
	Rows() (*sql.Rows, error)
	ScanRows(rows *sql.Rows, result interface{}) error
	Pluck(column string, value interface{}) GORM
	Count(value *int64) GORM
	FirstOrCreate(out interface{}, where ...interface{}) GORM
	FirstOrInit(dest interface{}, conds ...interface{}) GORM
	Updates(values interface{}) GORM
	BatchUpdate(tableName string, itemList interface{}) (int64, *bmserror.BMSError)
	Save(value interface{}) GORM
	Create(value interface{}) GORM
	CreateInBatches(value interface{}, batchSize int) GORM
	Delete(value interface{}, where ...interface{}) GORM
	Model(value interface{}) GORM
	Table(name string, args ...interface{}) (tx GORM)
	Debug() GORM
	Begin(opts ...*sql.TxOptions) GORM
	Commit() GORM
	Rollback() GORM
	RecordNotFound() bool
	RowsAffected() int64
	Set(name string, value interface{}) GORM
	Get(name string) (value interface{}, ok bool)
	InstantSet(name string, value interface{}) GORM
	ForUpdate() GORM
	ForShare() GORM
	GetConfig() *Config
	GetStatement() *Statement
	SavePoint(name string) GORM
	RollbackTo(name string) GORM
	GetGormDB() *gorm.DB
}

type dataSourceOperation interface {
	SwitchDataSource(dataSourceName string) GORM
	GetDataSourceName() string
}

type GroupDB interface {
	Master() GORM
	Replica() GORM
	ReplicaAt(index int) GORM
	Replicas() []GORM
	WithReplicaRouteKey(key string) GORM
	WithReplicaRouteStrategy(custom func(key string, reps []GORM) int) GORM
}

type ClusterDB interface {
	GetGroupKey() string
	RouteGroup() GORM
	GroupAt(index int) GORM
	Groups() []GORM
	WithRouteKey(key string) GORM
	WithRouteStrategy(custom func(key string, reps []GORM) int) GORM
}
