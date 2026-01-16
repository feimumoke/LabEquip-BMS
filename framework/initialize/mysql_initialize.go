package initialize

import (
	"time"

	"github.com/feimumoke/labequipbms/defines/constant"
	"github.com/feimumoke/labequipbms/framework/bmserror"
	"github.com/feimumoke/labequipbms/framework/config"
	"github.com/feimumoke/labequipbms/framework/orm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func buildDBClusters() (*orm.GormDB, *bmserror.BMSError) {
	dbConfig := config.WCConfig.DataSource
	dialector := orm.NewMultiDialector()
	idleMap := make(map[string]int)
	openMap := make(map[string]int)
	maxLifeMap := make(map[string]int)
	for ds, groups := range dbConfig {
		idleMap[ds] = groups.MaxIdleConns
		openMap[ds] = groups.MaxOpenConns
		maxLifeMap[ds] = groups.ConnMaxLifetime
		for _, group := range groups.Groups {
			if ds == orm.DefaultDataSourceName {
				salves := make([]gorm.Dialector, 0)
				for _, s := range group.ReplicasDsn {
					salves = append(salves, mysql.Open(s))
				}
				dialector.AddDefaultDialector(mysql.Open(group.MasterDsn), salves...)
			} else {
				salves := make([]gorm.Dialector, 0)
				for _, s := range group.ReplicasDsn {
					salves = append(salves, mysql.Open(s))
				}
				dialector.AddDialectorFor(ds, mysql.Open(group.MasterDsn), salves...)
			}
		}

	}

	// 创建数据库连接，应用 GORM 日志配置
	cluster, err := orm.Open(dialector, &orm.GormLogLevelOption{})
	if err != nil {
		return nil, bmserror.NewError(constant.ErrDB, err.Error())
	}
	for ds, groups := range dbConfig {
		cluster.SetMultiMaxIdleConns(ds, groups.MaxIdleConns)
		cluster.SetMultiMaxOpenConns(ds, groups.MaxOpenConns)
		cluster.SetMultiConnMaxLifetime(ds, time.Duration(groups.ConnMaxLifetime)*time.Second)
	}
	return cluster, nil
}
