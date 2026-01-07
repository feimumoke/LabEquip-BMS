package datasource

import (
	"context"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/appcontext"
	"github.com/feimumoke/labequipbms/framework/orm"
	"github.com/feimumoke/labequipbms/framework/transaction"
)

type DataSource interface {
	// 获取 先写库链接
	GetDataSource(ctx context.Context, key *RouterKey) orm.GORM
	// 获取 从库链接
	GetReplica(ctx context.Context, key *RouterKey) orm.GORM
}

type RouterKey struct {
	GroupKey   string
	ReplicaKey string
}

type defaultDataSource struct {
	DS            string
	groupRouter   func(key string, reps []orm.GORM) int
	replicaRouter func(key string, reps []orm.GORM) int
}

func (d defaultDataSource) GetDataSource(ctx context.Context, key *RouterKey) orm.GORM {
	ctxDs := orm.Context(ctx)

	groupKey := ""
	if key != nil {
		groupKey = key.GroupKey
	}
	if ctxDs != nil {
		if ctxDs.GetConfig().AspectTxMode {
			//切面事务
			conn := ctxDs.SwitchDataSource(d.DS)
			conn = conn.WithRouteStrategy(d.groupRouter)
			conn = conn.WithRouteKey(groupKey)
			conn = conn.RouteGroup()
			txCtx := transaction.GetTransactionContext(ctx)
			txKey := conn.GetGroupKey()
			persistConn := txCtx.GetPersistConn(txKey)
			if persistConn != nil {
				return persistConn
			}
			txConn := conn.Begin()
			//txConn.SwitchDataSource(d.DS)
			txCtx.SetPersistConn(txKey, txConn)
			return txConn
		}
		return ctxDs.SwitchDataSource(d.DS).WithRouteStrategy(d.groupRouter).WithRouteKey(groupKey).RouteGroup()
	}
	return appcontext.AppCtx.DBCluster.SwitchDataSource(d.DS).WithRouteStrategy(d.groupRouter).WithRouteKey(groupKey).RouteGroup()
}

func (d defaultDataSource) GetReplica(ctx context.Context, key *RouterKey) orm.GORM {
	ctxDs := orm.Context(ctx)
	groupKey, replicaKey := "", ""
	if key != nil {
		groupKey = key.GroupKey
		replicaKey = key.ReplicaKey
	}

	if ctxDs != nil {
		return ctxDs.SwitchDataSource(d.DS).
			WithRouteStrategy(d.groupRouter).WithReplicaRouteStrategy(d.replicaRouter).
			WithRouteKey(groupKey).WithReplicaRouteKey(replicaKey).RouteGroup()
	}
	return appcontext.AppCtx.DBCluster.SwitchDataSource(d.DS).
		WithRouteStrategy(d.groupRouter).WithReplicaRouteStrategy(d.replicaRouter).
		WithRouteKey(groupKey).WithReplicaRouteKey(replicaKey).RouteGroup()
}

func NewDefaultDataSource(ds string, groupRouter, replicaRouter func(key string, reps []orm.GORM) int) DataSource {
	if ds == "" {
		panic("ds is nil")
	}
	return &defaultDataSource{DS: ds, groupRouter: groupRouter, replicaRouter: replicaRouter}
}

var DefaultBasicSource = NewDefaultDataSource(constant.DataSourceBasic,
	func(key string, reps []orm.GORM) int {
		return 0
	},
	func(key string, reps []orm.GORM) int {
		return 0
	},
)

var DefaultInvSource = NewDefaultDataSource(constant.DataSourceInv,
	func(key string, reps []orm.GORM) int {
		return 0
	},
	func(key string, reps []orm.GORM) int {
		return 0
	},
)

var DefaultBMSSource = NewDefaultDataSource(constant.DataSourceBMS,
	func(key string, reps []orm.GORM) int {
		return 0
	},
	func(key string, reps []orm.GORM) int {
		return 0
	},
)
