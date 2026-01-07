package appcontext

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/feimumoke/labequipbms/framework/log"
	"github.com/feimumoke/labequipbms/framework/orm"
)

type AppContext struct {
	Logger     log.Logger
	DBCluster  *orm.GormDB
	AppData    map[string]interface{}
	ConfigFunc func(ctx context.Context, key, level string) string
}

var AppCtx = &AppContext{AppData: make(map[string]interface{})}

func RegisterDBCluster(cluster *orm.GormDB) {
	AppCtx.DBCluster = cluster
}

func BindContext(ctx context.Context) context.Context {
	ctx = orm.BindContext(ctx, AppCtx.DBCluster)
	return ctx
}

func InitAppData(path string) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	if umErr := json.Unmarshal(buf, &AppCtx.AppData); umErr != nil {
		panic(umErr)
	}
}
