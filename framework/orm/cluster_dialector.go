package orm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/feimumoke/labequipbms/framework/log"
	mysql2 "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	DefaultDataSourceName = "default"
)

type Plugin = gorm.Plugin

type Migrator = gorm.Migrator

type RouteStrategy func(db GORM, pools [][]ConnPool) (groupIndex int, instanceIndex int)

type MultiDialector struct {
	Default [][]Dialector
	Others  map[string][][]Dialector

	stg RouteStrategy
}

// 构造一个不包含任何数据源的MultiDialector
func NewMultiDialector() *MultiDialector {
	return &MultiDialector{
		Others: map[string][][]Dialector{},
	}
}

// 为Default数据源添加一组Dialector
func (m *MultiDialector) AddDefaultDialector(master Dialector, replicas ...Dialector) *MultiDialector {
	m.Default = append(m.Default, append([]Dialector{master}, replicas...))
	return m
}

// 为指定的数据源添加一组Dialector
func (m *MultiDialector) AddDialectorFor(name string, master Dialector, replicas ...Dialector) *MultiDialector {
	if m.Others == nil {
		m.Others = map[string][][]Dialector{}
	}

	val, _ := m.Others[name]
	val = append(val, append([]Dialector{master}, replicas...))
	m.Others[name] = val

	return m
}

// 为MultiDialector指定路由策略
// 传入的策略为一个处理函数：
//
//	入参db：表示本次操作的Orm句柄，包含本次SQL操作的信息
//	入参pools：表示目前可选的连接池，ConnPool的二维数组，维度含义与上文的MultiDialector的维度含义一致。ConnPool里包含数据源的信息
//	出参groupIndex：水平分库的下标，如果没有水平分库，则返回0
//	出参instanceIndex：主从分库的下标，下标0表示主库，其他表示从库
//
// 当某一个数据源处于事务中，路由策略将失效，事务过程将一直采用开启事务的连接
func (m *MultiDialector) WithRouteStrategy(stg RouteStrategy) *MultiDialector {
	m.stg = stg
	return m
}

func (m *MultiDialector) Name() string {
	return "multi"
}

func (m *MultiDialector) Initialize(db *gorm.DB) error {
	if m.Default == nil {
		ds, ok := m.Others[DefaultDataSourceName]
		if !ok || ds == nil {
			return fmt.Errorf("default data source is required")
		}
		m.Default = ds
	}
	delete(m.Others, DefaultDataSourceName)

	if len(m.Default) == 0 || len(m.Default[0]) == 0 {
		return fmt.Errorf("default data source expects ONE dsn at least")
	}

	err := m.Default[0][0].Initialize(db)
	if err != nil {
		return fmt.Errorf("default [0,0]th dialector initialize failed: %v", err)
	}

	dbConfig := db.Config

	multiPool := NewMultiConnPool()
	multiPool.Default = [][]ConnPool{{db.ConnPool}}
	multiPool.stg = m.stg
	db.ConnPool = multiPool

	log := db.Logger
	db.Logger = log.LogMode(logger.Silent)

	defer func() {
		db.Logger = log
		if err != nil {
			multiPool.Close()
			db.ConnPool = nil
		}
	}()

	var pools [][]ConnPool

	pools, err = m.multiInitialize(DefaultDataSourceName, m.Default, db, multiPool.Default[0][0])
	if err != nil {
		return fmt.Errorf("default dialectors initialize failed: %v", err)
	}
	multiPool.Default = pools

	multiPool.Others = map[string][][]ConnPool{}
	for k, v := range m.Others {
		pools, err = m.multiInitialize(k, v, db, nil)
		if err != nil {
			return fmt.Errorf("%s dialectors initialize failed: %v", k, err)
		}
		multiPool.Others[k] = pools
	}

	db.Config = dbConfig //把默认数据源的首次初始化结果作为db的配置

	return nil
}

func (m *MultiDialector) Migrator(db *gorm.DB) Migrator {
	return m.Default[0][0].Migrator(db)
}

func (m *MultiDialector) DataTypeOf(field *schema.Field) string {
	return m.Default[0][0].DataTypeOf(field)
}

func (m *MultiDialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return m.Default[0][0].DefaultValueOf(field)
}

func (m *MultiDialector) BindVarTo(writer clause.Writer, stmt *Statement, v interface{}) {
	m.Default[0][0].BindVarTo(writer, stmt, v)
}

func (m *MultiDialector) QuoteTo(writer clause.Writer, str string) {
	m.Default[0][0].QuoteTo(writer, str)
}

func (m *MultiDialector) Explain(sql string, vars ...interface{}) string {
	return m.Default[0][0].Explain(sql, vars...)
}

func (m *MultiDialector) SavePoint(tx *gorm.DB, name string) error {
	return tx.Exec("SAVEPOINT " + name).Error
}

func (m *MultiDialector) RollbackTo(tx *gorm.DB, name string) error {
	return tx.Exec("ROLLBACK TO SAVEPOINT " + name).Error
}

// 遍历多个Dialector，依次使用Dialector初始化得到ConnPool。首个ConnPool可以由外部传入，不自动创建
func (m *MultiDialector) multiInitialize(name string, dials [][]Dialector, db *gorm.DB, firstConnPool ConnPool) ([][]ConnPool, error) {
	pools := make([][]ConnPool, 0, len(dials))
	var err error
	defer func() {
		if err != nil {
			closeConnPools(pools)
		}
	}()

	for i := range dials {
		pool := make([]ConnPool, 0, len(dials[i]))
		for j := range dials[i] {
			if firstConnPool != nil {
				pool = append(pool, &NormalConnPool{
					pool:          firstConnPool,
					dsName:        name,
					groupIdx:      i,
					insIdx:        j,
					defaultDBName: m.getDefaultDBName(dials[i][j]),
				})
				firstConnPool = nil
				continue
			}
			config := m.cloneConfig(db.Config)
			db.Config = config
			err = dials[i][j].Initialize(db)
			if err != nil {
				return nil, fmt.Errorf("[%d,%d]th dialector initialize failed: %v", i, j, err)
			}
			pool = append(pool, &NormalConnPool{
				pool:          db.ConnPool,
				dsName:        name,
				groupIdx:      i,
				insIdx:        j,
				defaultDBName: m.getDefaultDBName(dials[i][j]),
			})
		}
		pools = append(pools, pool)
	}

	return pools, nil
}

// 目前只针对mysql的Dialector有效
func (m *MultiDialector) getDefaultDBName(dial Dialector) []byte {
	if d, ok := dial.(*mysql.Dialector); ok {
		dsn := d.DSN
		cfg, err := mysql2.ParseDSN(dsn)
		if err != nil {
			return nil
		}
		if cfg.DBName == "" {
			return nil
		}

		data, _ := json.Marshal(cfg.DBName)
		return data
	}

	return nil
}

func (m *MultiDialector) cloneConfig(config *gorm.Config) *gorm.Config {
	conf := *config
	conf.ClauseBuilders = map[string]clause.ClauseBuilder{}
	conf.Plugins = map[string]gorm.Plugin{}
	return &conf
}

type MultiConnPool struct {
	Default [][]ConnPool            //只读
	Others  map[string][][]ConnPool //只读

	db          *GormDB
	stg         RouteStrategy
	routed      ConnPool
	dataSrcName string

	groupIdx int
	insIdx   int

	persistConn map[string]ConnPool //当开启事务，或者使用Connection接口时，连接应该被固化
}

type rooter interface {
	SetRoot(*MultiConnPool) ConnPool
}

func NewMultiConnPool() *MultiConnPool {
	return &MultiConnPool{
		dataSrcName: DefaultDataSourceName,
		persistConn: map[string]ConnPool{},
	}
}

func (cp *MultiConnPool) Close() error {
	var errs []error
	err := closeConnPools(cp.Default)
	if err != nil {
		errs = append(errs, fmt.Errorf("close default conn pool failed: %v", err))
	}
	for k, v := range cp.Others {
		err = closeConnPools(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("close %s conn pool failed: %v", k, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%+v", errs)
	}

	return nil
}

func closeConnPools(pools [][]ConnPool) error {
	var errs []error
	for i := range pools {
		for j := range pools[i] {
			if closer, ok := pools[i][j].(interface{ Close() error }); ok && closer != nil {
				err := closer.Close()
				if err != nil {
					errs = append(errs, fmt.Errorf("[%d,%d]th conn pool close failed: %v", i, j, err))
				}
			} else {
				errs = append(errs, fmt.Errorf("[%d,%d]th conn pool could not close", i, j))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%+v", errs)
	}

	return nil
}

func (cp *MultiConnPool) SetMaxIdleConns(n int) {
	setMaxIdleConnsForPools(n, cp.Default)

	for _, v := range cp.Others {
		setMaxIdleConnsForPools(n, v)
	}
}
func (cp *MultiConnPool) SetMaxOpenConns(n int) {
	setMaxOpenConnsForPools(n, cp.Default)
	for _, v := range cp.Others {
		setMaxOpenConnsForPools(n, v)
	}
}
func (cp *MultiConnPool) SetConnMaxLifetime(d time.Duration) {
	setConnMaxLifetimeForPools(d, cp.Default)
	for _, v := range cp.Others {
		setConnMaxLifetimeForPools(d, v)
	}
}

func (cp *MultiConnPool) SetMultiConnMaxLifetime(ds string, d time.Duration) {
	if ds == DefaultDataSourceName {
		setConnMaxLifetimeForPools(d, cp.Default)
	} else {
		if groups, ok := cp.Others[ds]; ok {
			setConnMaxLifetimeForPools(d, groups)
		}
	}
}
func (cp *MultiConnPool) SetMultiMaxIdleConns(ds string, n int) {
	if ds == DefaultDataSourceName {
		setMaxIdleConnsForPools(n, cp.Default)
	} else {
		if groups, ok := cp.Others[ds]; ok {
			setMaxIdleConnsForPools(n, groups)
		}
	}
}
func (cp *MultiConnPool) SetMultiMaxOpenConns(ds string, n int) {
	if ds == DefaultDataSourceName {
		setMaxOpenConnsForPools(n, cp.Default)
	} else {
		if groups, ok := cp.Others[ds]; ok {
			setMaxOpenConnsForPools(n, groups)
		}
	}
}

func setMaxIdleConnsForPools(n int, pools [][]ConnPool) {
	for i := range pools {
		for j := range pools[i] {
			if mgr, ok := pools[i][j].(ConnLifetimeManager); ok {
				mgr.SetMaxIdleConns(n)
			}
		}
	}
}

func setMaxOpenConnsForPools(n int, pools [][]ConnPool) {
	for i := range pools {
		for j := range pools[i] {
			if mgr, ok := pools[i][j].(ConnLifetimeManager); ok {
				mgr.SetMaxOpenConns(n)
			}
		}
	}
}

func setConnMaxLifetimeForPools(d time.Duration, pools [][]ConnPool) {
	for i := range pools {
		for j := range pools[i] {
			if mgr, ok := pools[i][j].(ConnLifetimeManager); ok {
				mgr.SetConnMaxLifetime(d)
			}
		}
	}
}

func (cp *MultiConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	conn := cp.routed
	if conn == nil {
		cl := cp.clone()
		cl.route(ctx, cl.db)
		conn = cl.routed
	}

	if r, ok := conn.(rooter); ok {
		conn = r.SetRoot(cp)
	}

	return conn.PrepareContext(ctx, query)
}

func (cp *MultiConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	conn := cp.routed
	if conn == nil {
		cl := cp.clone()
		cl.route(ctx, cl.db)
		conn = cl.routed
	}

	if r, ok := conn.(rooter); ok {
		conn = r.SetRoot(cp)
	}

	return conn.ExecContext(ctx, query, args...)
}

func (cp *MultiConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	conn := cp.routed
	if conn == nil {
		cl := cp.clone()
		cl.route(ctx, cl.db)
		conn = cl.routed
	}

	if r, ok := conn.(rooter); ok {
		conn = r.SetRoot(cp)
	}

	return conn.QueryContext(ctx, query, args...)
}

func (cp *MultiConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	conn := cp.routed
	if conn == nil {
		cl := cp.clone()
		cl.route(ctx, cl.db)
		conn = cl.routed
	}

	if r, ok := conn.(rooter); ok {
		conn = r.SetRoot(cp)
	}

	return conn.QueryRowContext(ctx, query, args...)
}

func (cp *MultiConnPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (t ConnPool, err error) {
	conn := cp.routed
	if conn == nil {
		cl := cp.clone()
		cl.route(ctx, cl.db)
		conn = cl.routed
	}

	if r, ok := conn.(rooter); ok {
		conn = r.SetRoot(cp)
	}

	if tx, ok := conn.(TxBeginner); ok {
		return tx.BeginTx(ctx, opts)
	}

	if tx, ok := conn.(ConnPoolBeginner); ok {
		return tx.BeginTx(ctx, opts)
	}

	return nil, ErrInvalidTransaction
}

func (cp *MultiConnPool) GetDBConn() (*sql.DB, error) {
	pool := cp.routed
	if pool == nil {
		cl := cp.clone()
		cl.route(cl.db.GetStatement().Context, cl.db)
		pool = cl.routed
	}

	if r, ok := pool.(rooter); ok {
		pool = r.SetRoot(cp)
	}

	if getter, ok := pool.(gorm.GetDBConnector); ok {
		return getter.GetDBConn()
	}
	if db, ok := pool.(*sql.DB); ok {
		return db, nil
	}

	return nil, ErrInvalidDB
}

func (cp *MultiConnPool) clone() *MultiConnPool {
	persistConn := map[string]ConnPool{}
	for k, v := range cp.persistConn {
		persistConn[k] = v
	}

	return &MultiConnPool{
		Default:     cp.Default,
		Others:      cp.Others,
		db:          cp.db,
		stg:         cp.stg,
		routed:      cp.routed,
		dataSrcName: cp.dataSrcName,
		groupIdx:    cp.groupIdx,
		insIdx:      cp.insIdx,
		persistConn: persistConn,
	}
}

func (cp *MultiConnPool) SetRouteStrategy(stg RouteStrategy) *MultiConnPool {
	cl := cp.clone()
	cl.clearRoute()
	cl.stg = stg
	return cl
}

func (cp *MultiConnPool) SetOrmDB(db *GormDB) *MultiConnPool {
	cl := cp.clone()
	cl.db = db
	return cl
}

func (cp *MultiConnPool) SetDataSrcName(name string) *MultiConnPool {
	cl := cp.clone()
	if cl.dataSrcName != name {
		cl.clearRoute()
	}
	cl.dataSrcName = name
	return cl
}

func (cp *MultiConnPool) GetDataSrcName() string {
	return cp.dataSrcName
}

func (cp *MultiConnPool) SetGroupIdx(idx int) *MultiConnPool {
	cl := cp.clone()
	if cl.groupIdx != idx {
		cl.clearRoute()
	}
	cl.groupIdx = idx
	return cl
}

func (cp *MultiConnPool) GetGroupIdx() int {
	return cp.groupIdx
}

func (cp *MultiConnPool) SetInsIdx(idx int) *MultiConnPool {
	cl := cp.clone()
	if cl.insIdx != idx {
		cl.clearRoute()
	}
	cl.insIdx = idx
	return cl
}

func (cp *MultiConnPool) GetInsIdx() int {
	return cp.insIdx
}

func (cp *MultiConnPool) SetGroupIdxAndInsIdx(groupIdx int, insIdx int) *MultiConnPool {
	cl := cp.clone()
	if cl.groupIdx != groupIdx || cl.insIdx != insIdx {
		cl.clearRoute()
	}
	cl.groupIdx = groupIdx
	cl.insIdx = insIdx
	return cl
}

func (cp *MultiConnPool) PersistConn(conn ConnPool) *MultiConnPool {
	cl := cp.clone()
	cl.routed = conn
	cl.persistConn[cl.dataSrcName] = conn
	if rooter, ok := conn.(interface{ SetRoot(*MultiConnPool) }); ok {
		rooter.SetRoot(cl)
	}
	return cl
}

func (cp *MultiConnPool) GetCurPersist() ConnPool {
	if pool, ok := cp.persistConn[cp.dataSrcName]; ok {
		return pool
	}

	return nil
}

func (cp *MultiConnPool) routeByDataSourceName(ctx context.Context) [][]ConnPool {
	var groups [][]ConnPool

	if cp.dataSrcName == DefaultDataSourceName {
		groups = cp.Default
	} else {
		groups, _ = cp.Others[cp.dataSrcName]
	}
	return groups
}

func (cp *MultiConnPool) route(ctx context.Context, com GORM) (ConnPool, error) {
	if cp.routed != nil {
		return cp.routed, nil
	}

	if conn, ok := cp.persistConn[cp.dataSrcName]; ok && conn != nil {
		cp.routed = conn
		return cp.routed, nil
	}

	groups := cp.routeByDataSourceName(ctx)
	if len(groups) == 0 || len(groups[0]) == 0 {
		return nil, fmt.Errorf("data source [%s] has no conn pool", cp.dataSrcName)
	}

	if cp.stg != nil {
		cp.groupIdx, cp.insIdx = cp.stg(com, groups)
	}

	if cp.groupIdx >= len(groups) {
		return nil,
			fmt.Errorf("data source [%s] group index [%d] out of range [%d]",
				cp.dataSrcName, cp.groupIdx, len(groups))
	}

	if cp.insIdx >= len(groups[cp.groupIdx]) {
		return nil,
			fmt.Errorf("data source [%s] instance index [%d] of group index [%d] out of range [%d]",
				cp.dataSrcName, cp.insIdx, cp.groupIdx, len(groups[cp.groupIdx]))
	}

	cp.routed = groups[cp.groupIdx][cp.insIdx]
	return cp.routed, nil
}

func (cp *MultiConnPool) clearRoute() {
	cp.routed = nil
}

type NormalConnPool struct {
	root          *MultiConnPool
	pool          ConnPool
	dsName        string
	groupIdx      int
	insIdx        int
	defaultDBName []byte
}

func (np *NormalConnPool) clone() *NormalConnPool {
	return &NormalConnPool{
		root:          np.root,
		pool:          np.pool,
		dsName:        np.dsName,
		groupIdx:      np.groupIdx,
		insIdx:        np.insIdx,
		defaultDBName: np.defaultDBName,
	}
}

func (np *NormalConnPool) GetDataSourceName() string {
	return np.dsName
}

func (np *NormalConnPool) GetGroupName() string {
	return fmt.Sprintf("group%d", np.groupIdx)
}

func (np *NormalConnPool) GetRoleName() string {
	if np.insIdx == 0 {
		return "master"
	}
	return fmt.Sprintf("replica%d", np.insIdx-1)
}

func (np *NormalConnPool) GetDefaultDBName() string {
	if len(np.defaultDBName) == 0 {
		return ""
	}

	var name string
	json.Unmarshal(np.defaultDBName, &name)
	return name
}

func (np *NormalConnPool) SetRoot(m *MultiConnPool) ConnPool {
	c := np.clone()
	c.root = m
	return c
}

func (np *NormalConnPool) Close() error {
	return closeConnPool(np.pool)
}

func (np *NormalConnPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx ConnPool, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if db, ok := np.pool.(ConnPoolBeginner); ok {

		start := time.Now()
		defer func() {
			np.root.db.Logger.Trace(ctx, start, func() (sql string, rowsAffected int64) {
				return "BEGIN", -1
			}, err)
		}()
		if np.root.db.AutoReport {
			defer func() {
				log.Infof(fmt.Sprintf("MySQL.%s.%s.%s-%s", np.GetDataSourceName(), np.GetGroupName(), np.GetRoleName(), "Begin"))
			}()
		}

		tx, err = db.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &NormalTx{
			conn:   tx,
			parent: np,
			ctx:    ctx,
		}, nil
	}
	if db, ok := np.pool.(TxBeginner); ok {
		start := time.Now()
		defer func() {
			np.root.db.Logger.Trace(ctx, start, func() (sql string, rowsAffected int64) {
				return "BEGIN", -1
			}, err)
		}()
		if np.root.db.AutoReport {
			defer func() {
				log.Infof(fmt.Sprintf("MySQL.%s.%s.%s-%s", np.GetDataSourceName(), np.GetGroupName(), np.GetRoleName(), "Begin"))
			}()
		}

		tx, err = db.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &NormalTx{
			conn:   tx,
			parent: np,
			ctx:    ctx,
		}, nil
	}

	return nil, ErrInvalidTransaction
}
func (np *NormalConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return np.pool.PrepareContext(ctx, query)
}
func (np *NormalConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return np.pool.ExecContext(ctx, query, args...)
}
func (np *NormalConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return np.pool.QueryContext(ctx, query, args...)
}
func (np *NormalConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return np.pool.QueryRowContext(ctx, query, args...)
}

func (np *NormalConnPool) GetDBConn() (*sql.DB, error) {
	if getter, ok := np.pool.(gorm.GetDBConnector); ok {
		return getter.GetDBConn()
	}
	if db, ok := np.pool.(*sql.DB); ok {
		return db, nil
	}
	return nil, ErrInvalidDB
}

func (np *NormalConnPool) SetMaxIdleConns(n int) {
	if mgr, ok := np.pool.(ConnLifetimeManager); ok {
		mgr.SetMaxIdleConns(n)
	}
}
func (np *NormalConnPool) SetMaxOpenConns(n int) {
	if mgr, ok := np.pool.(ConnLifetimeManager); ok {
		mgr.SetMaxOpenConns(n)
	}
}
func (np *NormalConnPool) SetConnMaxLifetime(d time.Duration) {
	if mgr, ok := np.pool.(ConnLifetimeManager); ok {
		mgr.SetConnMaxLifetime(d)
	}
}

type NormalTx struct {
	conn ConnPool

	ctx    context.Context
	parent *NormalConnPool
}

func (nt *NormalTx) GetDataSourceName() string {
	return nt.parent.GetDataSourceName()
}

func (nt *NormalTx) GetGroupName() string {
	return nt.parent.GetGroupName()
}

func (nt *NormalTx) GetRoleName() string {
	return nt.parent.GetRoleName()
}

func (nt *NormalTx) GetDefaultDBName() string {
	return nt.parent.GetDefaultDBName()
}

func (nt *NormalTx) Commit() (err error) {
	if tx, ok := nt.conn.(TxCommitter); ok {
		start := time.Now()
		defer func() {
			nt.parent.root.db.Logger.Trace(nt.ctx, start, func() (sql string, rowsAffected int64) {
				return "COMMIT", -1
			}, err)
		}()
		if nt.parent.root.db.AutoReport {

			defer func() {
				log.Infof(fmt.Sprintf("MySQL.%s.%s.%s-%s", nt.GetDataSourceName(), nt.GetGroupName(), nt.GetRoleName(), "Commit"))
			}()
		}

		return tx.Commit()
	}
	return ErrInvalidTransaction
}

func (nt *NormalTx) Rollback() (err error) {
	if tx, ok := nt.conn.(TxCommitter); ok {
		start := time.Now()
		defer func() {
			nt.parent.root.db.Logger.Trace(nt.ctx, start, func() (sql string, rowsAffected int64) {
				return "ROLLBACK", -1
			}, err)
		}()
		if nt.parent.root.db.AutoReport {
			defer func() {
				log.Infof(fmt.Sprintf("MySQL.%s.%s.%s-%s", nt.GetDataSourceName(), nt.GetGroupName(), nt.GetRoleName(), "Rollback"))
			}()
		}

		return tx.Rollback()
	}
	return ErrInvalidTransaction
}

func (nt *NormalTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nt.conn.PrepareContext(ctx, query)
}
func (nt *NormalTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nt.conn.ExecContext(ctx, query, args...)
}
func (nt *NormalTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nt.conn.QueryContext(ctx, query, args...)
}
func (nt *NormalTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nt.conn.QueryRowContext(ctx, query, args...)
}

func closeConnPool(pool ConnPool) error {
	if closer, ok := pool.(interface{ Close() error }); ok && closer != nil {
		return closer.Close()
	}

	return fmt.Errorf("could not close")
}
