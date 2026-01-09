package monitor

import (
	"context"

	"github.com/feimumoke/labequipbms/framework/log"
)

type AwesomeEnd1 func(moduleName, interfaceName string, status int, data string) bool
type AwesomeEnd2 func(moduleName, interfaceName, status, data string) bool

type Monitor interface {
	AwesomeStart1(context.Context) (context.Context, AwesomeEnd1)
	AwesomeStart2(context.Context) (context.Context, AwesomeEnd2)
}

var DoMonitor Monitor = &EmptyMonitor{}

func SetDoMonitor(m Monitor) {
	DoMonitor = m
}

func AwesomeStart1(ctx context.Context) (context.Context, AwesomeEnd1) {
	return DoMonitor.AwesomeStart1(ctx)
}

func AwesomeStart2(ctx context.Context) (context.Context, AwesomeEnd2) {
	return DoMonitor.AwesomeStart2(ctx)
}

func AwesomeReportEventWithoutTrans(ctx context.Context, moduleName, interfaceName, status, data string) error {

	return nil
}

type EmptyMonitor struct{}

func (e *EmptyMonitor) AwesomeStart1(ctx context.Context) (context.Context, AwesomeEnd1) {
	return ctx, func(moduleName, interfaceName string, status int, data string) bool {
		log.Infof("AwesomeStart1 Report for moduleName [%v] interfaceName [%v] data [%v] -- status %v", moduleName, interfaceName, data, status)
		return true
	}
}

func (e *EmptyMonitor) AwesomeStart2(ctx context.Context) (context.Context, AwesomeEnd2) {
	return ctx, func(moduleName, interfaceName, status, data string) bool { return true }
}
