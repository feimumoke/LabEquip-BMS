package initialize

import (
	"github.com/feimumoke/labequipbms/framework/appcontext"
	"github.com/feimumoke/labequipbms/framework/bmserror"
)

func Initialize(jsonpath string) *bmserror.BMSError {
	clusters, wcError := buildDBClusters()
	if wcError != nil {
		return wcError.Mark()
	}
	appcontext.RegisterDBCluster(clusters)
	appcontext.InitAppData(jsonpath + "static.json")
	return nil
}
