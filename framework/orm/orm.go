package orm

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

var (
	rrGroupIdx   = uint32(0)
	rrReplicaIdx = uint32(0)
)

type Dialector = gorm.Dialector

type SavePointerDialectorInterface = gorm.SavePointerDialectorInterface

type TxBeginner = gorm.TxBeginner

type ConnPoolBeginner = gorm.ConnPoolBeginner

type TxCommitter = gorm.TxCommitter

type Tx = gorm.Tx

type DB = gorm.DB

type DeletedAt = gorm.DeletedAt

type GormDB struct {
	*Config
	inner *gorm.DB
}

func (g *GormDB) InAspectTxMode() bool {
	return g.Config.AspectTxMode
}

type Config struct {
	AspectTxMode bool
	AutoReport   bool
	*gorm.Config
}

func NewConfig() *Config {
	return &Config{
		Config: &gorm.Config{
			SkipDefaultTransaction: true,
		},
	}
}

type Option interface {
	Apply(*Config) error
	//AfterInitialize(*OrmDB) error
}

func Open(dialector Dialector, opts ...Option) (db *GormDB, err error) {
	config := NewConfig()

	// 先应用选项(包括日志配置)
	for _, opt := range opts {
		if err := opt.Apply(config); err != nil {
			return nil, err
		}
	}

	if config.Config == nil {
		config.Config = &gorm.Config{}
	}

	// 如果选项没有设置 Logger，使用默认的
	if config.Logger == nil {
		config.Logger = NewGormLogger()
	}

	db = &GormDB{
		Config: config,
	}

	inner, err := gorm.Open(dialector, config.Config)
	if err != nil {
		return nil, err
	}

	db.Config.Config = inner.Config
	db.inner = inner
	setOrmDB(db)
	if pool, ok := db.ConnPool.(*MultiConnPool); ok {
		db.ConnPool = pool.SetOrmDB(db)
	}

	db.inner = db.inner.Session(&Session{})

	return db, nil
}

func setOrmDB(db *GormDB) {
	db.inner = db.inner.Set("scorm:root_orm_db", db)
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok { //todo:
		db.inner.Statement.ConnPool = pool.SetOrmDB(db)
	}
}

func (db *GormDB) clone() *GormDB {
	return &GormDB{
		Config: &Config{
			AspectTxMode: db.Config.AspectTxMode,
			AutoReport:   db.Config.AutoReport,
			Config:       db.Config.Config,
		},
		inner: db.inner,
	}
}

func (db *GormDB) AddError(err error) (tx GORM) {
	return (&ErrSQLCommon{}).AddError(err)
}

func (db *GormDB) Assign(attrs ...interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Assign(attrs...)
	return c
}

func (db *GormDB) Attrs(attrs ...interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Attrs(attrs...)
	return c
}

func (db *GormDB) Begin(opts ...*sql.TxOptions) GORM {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Begin(opts...)
	if db.AspectTxMode {

	}
	return c
}

func (db *GormDB) Clauses(conds ...clause.Expression) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Clauses(conds...)
	return c
}

func (db *GormDB) Close() *bmserror.BMSError {
	if closer, ok := db.ConnPool.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return bmserror.NewError(constant.ErrDB, err.Error())
		}
	}
	return nil
}

func (db *GormDB) Commit() GORM {
	db.inner.Commit()
	return db
}

func (db *GormDB) Connection(fc func(tx GORM) error) (err error) {
	c := db.clone()
	return c.inner.Connection(func(tx *gorm.DB) error {
		c.inner = tx
		setOrmDB(c)
		return fc(c)
	})
}

func (db *GormDB) Count(count *int64) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Count(count)
	return c
}

func (db *GormDB) Create(value interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Create(value)
	return c
}

func (db *GormDB) CreateInBatches(value interface{}, batchSize int) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.CreateInBatches(value, batchSize)
	return c
}

func (db *GormDB) DB() *sql.DB {
	sDB, err := db.inner.DB()
	if err != nil {
		panic(err)
	}
	return sDB
}

func (db *GormDB) Debug() (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Debug()
	return c
}

func (db *GormDB) Delete(value interface{}, conds ...interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Delete(value, conds...)
	return c
}

func (db *GormDB) Distinct(args ...interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Distinct(args...)
	return c
}

func (db *GormDB) Exec(sql string, values ...interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Exec(sql, values...)
	return c
}

func (db *GormDB) Find(dest interface{}, conds ...interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Find(dest, conds...)
	return c
}

func (db *GormDB) FindInBatches(dest interface{}, batchSize int, fc func(tx GORM, batch int) error) GORM {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
		c.inner = tx
		setOrmDB(c)
		return fc(c, batch)
	})
	return c
}

func (db *GormDB) First(dest interface{}, conds ...interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.First(dest, conds...)
	return c
}

func (db *GormDB) FirstOrCreate(dest interface{}, conds ...interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.FirstOrCreate(dest, conds...)
	return c
}

func (db *GormDB) FirstOrInit(dest interface{}, conds ...interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.FirstOrInit(dest, conds...)
	return c
}

func (db *GormDB) ForShare() GORM {
	c := db.clone()
	c.inner = c.inner.Clauses(clause.Locking{Strength: "SHARE"})
	return c
}

func (db *GormDB) ForUpdate() GORM {
	c := db.clone()
	c.inner = c.inner.Clauses(clause.Locking{Strength: "UPDATE"})
	return c
}

func (db *GormDB) Get(key string) (interface{}, bool) {
	return db.inner.Get(key)
}

func (db *GormDB) GetConfig() *Config {
	return db.Config
}

func (db *GormDB) GetError() *bmserror.BMSError {
	if db.inner.Error != nil {
		return bmserror.NewError(constant.ErrDB, db.inner.Error.Error())
	}
	return nil
}

func (db *GormDB) GetGormDB() *gorm.DB {
	c := db.clone()
	setOrmDB(c) //保证后面hook中拿到的db是最新的
	return c.inner
}

func (db *GormDB) GetStatement() *Statement {
	return db.inner.Statement
}

func (db *GormDB) Group(name string) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Group(name)
	return c
}

func (db *GormDB) Having(query interface{}, args ...interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Having(query, args...)
	return c
}

func (db *GormDB) InstanceGet(key string) (interface{}, bool) {
	return db.inner.InstanceGet(key)
}

func (db *GormDB) InstanceSet(key string, value interface{}) GORM {
	c := db.clone()
	c.inner = c.inner.InstanceSet(key, value)
	return c
}

// 非并发安全，请注意
func (db *GormDB) InstantSet(key string, value interface{}) GORM {
	db.inner.Statement.Settings.Store(key, value)
	return db
}

func (db *GormDB) IsAutoReport() bool {
	return db.Config.AutoReport
}

func (db *GormDB) Last(dest interface{}, conds ...interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Last(dest, conds...)
	return c
}

func (db *GormDB) Limit(limit int) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Limit(limit)
	return c
}

func (db *GormDB) LogMode(enable bool) GORM {
	level := logger.Info
	if !enable {
		level = logger.Silent
	}
	c := db.clone()
	c.inner = c.inner.Session(&Session{Logger: db.Logger.LogMode(level)})
	conf := *(c.Config)
	conf.Config = c.inner.Config
	c.Config = &conf
	return c
}

func (db *GormDB) Model(value interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Model(value)
	return c
}

func (db *GormDB) Not(query interface{}, args ...interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Not(query, args...)
	return c
}

func (db *GormDB) Offset(offset int) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Offset(offset)
	return c
}

func (db *GormDB) Omit(columns ...string) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Omit(columns...)
	return c
}

func (db *GormDB) Or(query interface{}, args ...interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Or(query, args...)
	return c
}

func (db *GormDB) Order(value interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Order(value)
	return c
}

func (db *GormDB) Pluck(column string, dest interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Pluck(column, dest)
	return c
}

func (db *GormDB) RecordNotFound() bool {
	return db.inner.Error == ErrRecordNotFound
}

func (db *GormDB) Raw(sql string, values ...interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Raw(sql, values...)
	return c
}

func (db *GormDB) Rollback() GORM {
	db.inner.Rollback()
	return db
}

func (db *GormDB) RollbackTo(name string) GORM {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.RollbackTo(name)
	return c
}

func (db *GormDB) Row() *sql.Row {
	c := db.clone()
	setOrmDB(c)
	return c.inner.Row()
}

func (db *GormDB) Rows() (*sql.Rows, error) {
	c := db.clone()
	setOrmDB(c)
	return c.inner.Rows()
}

func (db *GormDB) RowsAffected() int64 {
	return db.inner.RowsAffected
}

func (db *GormDB) Save(value interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Save(value)
	return c
}

func (db *GormDB) SavePoint(name string) GORM {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.SavePoint(name)
	return c
}

func (db *GormDB) Scan(dest interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Scan(dest)
	return c
}

func (db *GormDB) ScanRows(rows *sql.Rows, dest interface{}) error {
	c := db.clone()
	setOrmDB(c)
	return c.inner.ScanRows(rows, dest)
}

func (db *GormDB) Scopes(funcs ...func(GORM) GORM) (tx GORM) {
	c := db.clone()
	var fns []func(db *gorm.DB) *gorm.DB
	for i := range funcs {
		fn := funcs[i]
		fns = append(fns, func(db *gorm.DB) *gorm.DB {
			c.inner = db
			setOrmDB(c)
			cc := fn(c)
			return cc.GetGormDB()
		})
	}
	c.inner = c.inner.Scopes(fns...)
	return c
}

func (db *GormDB) Select(query interface{}, args ...interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Select(query, args...)
	return c
}

func (db *GormDB) Session(config *Session) GORM {
	c := db.clone()
	c.inner = c.inner.Session(config)
	return c
}

func (db *GormDB) Set(key string, value interface{}) GORM {
	c := db.clone()
	c.inner = c.inner.Set(key, value)
	return c
}

func (db *GormDB) SetConnMaxLifetime(d time.Duration) {
	if mgr, ok := db.ConnPool.(ConnLifetimeManager); ok {
		mgr.SetConnMaxLifetime(d)
	}
	if pre, ok := db.ConnPool.(*gorm.PreparedStmtDB); ok {
		if mgr, ok := pre.ConnPool.(ConnLifetimeManager); ok {
			mgr.SetConnMaxLifetime(d)
		}
	}
}

func (db *GormDB) SetLogger(log logger.Interface) {
	db.Logger = log
	db.inner.Logger = log
}

func (db *GormDB) SetMaxIdleConns(n int) {
	if mgr, ok := db.ConnPool.(ConnLifetimeManager); ok {
		mgr.SetMaxIdleConns(n)
	}
	if pre, ok := db.ConnPool.(*gorm.PreparedStmtDB); ok {
		if mgr, ok := pre.ConnPool.(ConnLifetimeManager); ok {
			mgr.SetMaxIdleConns(n)
		}
	}
}

func (db *GormDB) SetMaxOpenConns(n int) {
	if mgr, ok := db.ConnPool.(ConnLifetimeManager); ok {
		mgr.SetMaxOpenConns(n)
	}
	if pre, ok := db.ConnPool.(*gorm.PreparedStmtDB); ok {
		if mgr, ok := pre.ConnPool.(ConnLifetimeManager); ok {
			mgr.SetMaxOpenConns(n)
		}
	}
}

func (db *GormDB) SetMultiConnMaxLifetime(ds string, d time.Duration) {
	if mc, ok := db.ConnPool.(MultiConnLifetimeManager); ok {
		mc.SetMultiConnMaxLifetime(ds, d)
		return
	}
	if mgr, ok := db.ConnPool.(ConnLifetimeManager); ok {
		mgr.SetConnMaxLifetime(d)
	}
	if pre, ok := db.ConnPool.(*gorm.PreparedStmtDB); ok {
		if mgr, ok := pre.ConnPool.(ConnLifetimeManager); ok {
			mgr.SetConnMaxLifetime(d)
		}
	}
}
func (db *GormDB) SetMultiMaxIdleConns(ds string, n int) {
	if mc, ok := db.ConnPool.(MultiConnLifetimeManager); ok {
		mc.SetMultiMaxIdleConns(ds, n)
		return
	}
	if mgr, ok := db.ConnPool.(ConnLifetimeManager); ok {
		mgr.SetMaxIdleConns(n)
	}
	if pre, ok := db.ConnPool.(*gorm.PreparedStmtDB); ok {
		if mgr, ok := pre.ConnPool.(ConnLifetimeManager); ok {
			mgr.SetMaxIdleConns(n)
		}
	}
}
func (db *GormDB) SetMultiMaxOpenConns(ds string, n int) {
	if mc, ok := db.ConnPool.(MultiConnLifetimeManager); ok {
		mc.SetMultiMaxOpenConns(ds, n)
		return
	}
	if mgr, ok := db.ConnPool.(ConnLifetimeManager); ok {
		mgr.SetMaxOpenConns(n)
	}
	if pre, ok := db.ConnPool.(*gorm.PreparedStmtDB); ok {
		if mgr, ok := pre.ConnPool.(ConnLifetimeManager); ok {
			mgr.SetMaxOpenConns(n)
		}
	}
}

func (db *GormDB) Table(name string, args ...interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Table(name, args...)
	return c
}

func (db *GormDB) Take(dest interface{}, conds ...interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Take(dest, conds...)
	return c
}

func (db *GormDB) ToSQL(queryFn func(tx GORM) GORM) string {
	return db.inner.ToSQL(func(tx *gorm.DB) *gorm.DB {
		c := db.clone()
		c.inner = tx
		return queryFn(c).GetGormDB()
	})
}

func (db *GormDB) Unscoped() (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Unscoped()
	return c
}

func (db *GormDB) Update(column string, value interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Update(column, value)
	return c
}

func (db *GormDB) UpdateColumn(column string, value interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.UpdateColumn(column, value)
	return c
}

func (db *GormDB) UpdateColumns(values interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.UpdateColumns(values)
	return c
}

func (db *GormDB) Updates(values interface{}) (tx GORM) {
	c := db.clone()
	setOrmDB(c)
	c.inner = c.inner.Updates(values)
	return c
}

func (db *GormDB) Value() interface{} {
	return db.inner.Statement.Dest
}

func (db *GormDB) Where(query interface{}, args ...interface{}) (tx GORM) {
	c := db.clone()
	c.inner = c.inner.Where(query, args...)
	return c
}

func (db *GormDB) WithContext(ctx context.Context) GORM {
	c := db.clone()
	c.inner = c.inner.WithContext(ctx)
	return c
}

func (db *GormDB) Joins(query string, args ...interface{}) GORM {
	c := db.clone()
	c.inner = c.inner.Joins(query, args...)
	return c
}

func (db *GormDB) Preload(query string, args ...interface{}) GORM {
	c := db.clone()
	c.inner = c.inner.Preload(query, args...)
	return c
}

func (db *GormDB) SwitchDataSource(dataSourceName string) GORM {
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok {
		c := db.clone()
		c.inner = c.inner.Session(&Session{Context: c.GetStatement().Context})

		if pool.routed != nil {
			pool = pool.PersistConn(pool.routed)
			c.inner.Statement.ConnPool = pool
		}

		pool = pool.SetDataSrcName(dataSourceName)
		conn := pool.GetCurPersist()
		if conn != nil {
			c.inner.Statement.ConnPool = conn
		} else {
			c.inner.Statement.ConnPool = pool
		}

		return c
	}
	if _, ok := db.inner.Statement.ConnPool.(TxCommitter); ok {
		if pool, ok := db.ConnPool.(*MultiConnPool); ok {
			gormConf := *(db.Config.Config)
			pool = pool.PersistConn(db.inner.Statement.ConnPool)
			gormConf.ConnPool = pool
			conf := *(db.Config)
			conf.Config = &gormConf

			c := db.clone()
			c.inner = c.inner.Session(&Session{})
			c.Config = &conf
			c.inner.Config = conf.Config

			pool = pool.SetDataSrcName(dataSourceName)
			conn := pool.GetCurPersist()
			if conn != nil {
				c.inner.Statement.ConnPool = conn
			} else {
				c.inner.Statement.ConnPool = pool
			}

			return c
		}
	}
	return db
}

func (db *GormDB) GetDataSourceName() string {
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok {
		return pool.GetDataSrcName()
	}
	return DefaultDataSourceName
}

func (db *GormDB) RouteGroup() GORM {
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok {
		c := db.clone()
		c.inner = c.inner.Session(&Session{Context: c.GetStatement().Context})
		pool = pool.SetGroupIdx(getGroupIdx(c))
		c.inner.Statement.ConnPool = pool
		return c
	}
	return db
}

func (db *GormDB) GetGroupKey() string {
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok {
		return getMultiPersistKey(pool.GetDataSrcName(), getGroupIdx(db))
	}
	return DefaultDataSourceName
}

func getMultiPersistKey(ds string, groupIdx int) string {
	return fmt.Sprintf("gorm:%s_#_%v:persisit_key", ds, groupIdx)
}

func (db *GormDB) GroupAt(index int) GORM {
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok {
		c := db.clone()
		c.inner = c.inner.Session(&Session{Context: c.GetStatement().Context})

		c.inner.Statement.ConnPool = pool.SetGroupIdx(index)
		return c
	}
	return db
}

func (db *GormDB) Groups() []GORM {
	var groupDB []GORM
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok {
		groups := pool.routeByDataSourceName(db.inner.Statement.Context)
		for i := range groups {
			c := db.clone()
			c.inner = c.inner.Session(&Session{Context: c.GetStatement().Context})
			c.inner.Statement.ConnPool = pool.SetGroupIdx(i)
			groupDB = append(groupDB, c)
		}
	} else {
		groupDB = append(groupDB, db)
	}
	return groupDB
}

func (db *GormDB) WithRouteKey(key string) GORM {
	c := db.clone()
	c.inner = c.inner.Set(fmt.Sprintf("gorm:%s:route_key", db.GetDataSourceName()), key)
	return c
}

func (db *GormDB) WithRouteStrategy(custom func(key string, groups []GORM) int) GORM {
	c := db.clone()
	c.inner = c.inner.Set(fmt.Sprintf("gorm:%s:route_strategy", db.GetDataSourceName()), custom)
	return c
}

func (db *GormDB) Master() GORM {
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok {
		c := db.clone()
		c.inner = c.inner.Session(&Session{Context: c.GetStatement().Context})

		c.inner.Statement.ConnPool = pool.SetInsIdx(0)
		return c
	}
	return db
}

func (db *GormDB) Replica() GORM {
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok {
		c := db.clone()
		c.inner = c.inner.Session(&Session{Context: c.GetStatement().Context})

		pool = pool.SetInsIdx(getInsIdx(c))

		c.inner.Statement.ConnPool = pool
		return c
	}
	return db
}

func (db *GormDB) ReplicaAt(index int) GORM {
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok {
		c := db.clone()
		c.inner = c.inner.Session(&Session{Context: c.GetStatement().Context})

		c.inner.Statement.ConnPool = pool.SetInsIdx(1 + index)
		return c
	}
	return db
}

func (db *GormDB) Replicas() []GORM {
	var repDB []GORM
	if pool, ok := db.inner.Statement.ConnPool.(*MultiConnPool); ok {
		groups := pool.routeByDataSourceName(db.inner.Statement.Context)
		groupIdx := pool.GetGroupIdx()
		if groupIdx >= len(groups) {
			return repDB
		}
		for i := 1; i < len(groups[groupIdx]); i++ {
			c := db.clone()
			c.inner = c.inner.Session(&Session{Context: c.GetStatement().Context})
			c.inner.Statement.ConnPool = pool.SetInsIdx(i)
			repDB = append(repDB, c)
		}
	} else {
		repDB = append(repDB, db)
	}
	return repDB
}

func (db *GormDB) WithReplicaRouteKey(key string) GORM {
	c := db.clone()
	c.inner = c.inner.Set(fmt.Sprintf("gorm:%s:replica_route_key", db.GetDataSourceName()), key)
	return c
}

func (db *GormDB) WithReplicaRouteStrategy(custom func(key string, reps []GORM) int) GORM {
	c := db.clone()
	c.inner = c.inner.Set(fmt.Sprintf("gorm:%s:replica_route_strategy", db.GetDataSourceName()), custom)
	return c
}

func (db *GormDB) UnsafeGetOrmDB() *gorm.DB {
	setOrmDB(db)
	return db.inner
}

func getGroupIdx(com GORM) int {
	key := ""
	if val, ok := com.Get(fmt.Sprintf("gorm:%s:route_key", com.GetDataSourceName())); ok && val != nil {
		if str, ok := val.(string); ok {
			key = str
		}
	}

	stg := func(key string, groups []GORM) int {
		idx := atomic.AddUint32(&rrGroupIdx, 1)
		return int((idx - 1) % uint32(len(groups)))
	}
	if val, ok := com.Get(fmt.Sprintf("gorm:%s:route_strategy", com.GetDataSourceName())); ok && val != nil {
		if fn, ok := val.(func(key string, _ []GORM) int); ok {
			stg = fn
		}
	}

	return stg(key, com.Groups())
}

func getInsIdx(com GORM) int {
	key := ""
	if val, ok := com.Get(fmt.Sprintf("gorm:%s:replica_route_key", com.GetDataSourceName())); ok && val != nil {
		if str, ok := val.(string); ok {
			key = str
		}
	}

	stg := func(key string, reps []GORM) int {
		idx := atomic.AddUint32(&rrReplicaIdx, 1)
		return int((idx - 1) % uint32(len(reps)))
	}
	if val, ok := com.Get(fmt.Sprintf("gorm:%s:replica_route_strategy", com.GetDataSourceName())); ok && val != nil {
		if fn, ok := val.(func(key string, reps []GORM) int); ok {
			stg = fn
		}
	}

	return stg(key, com.Replicas()) + 1
}
