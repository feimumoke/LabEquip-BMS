package trace

import (
	"sync"

	"github.com/petermattis/goid"
)

var (
	gLogIDMap     = make(map[int64]string, 1000000)
	gLogIDMapLock sync.RWMutex
)

func setLogTraceID(logid string) {
	if gLogIDMap != nil {
		gLogIDMapLock.Lock()
		goroutineId := goid.Get()
		gLogIDMap[goroutineId] = logid
		gLogIDMapLock.Unlock()
	}
}

func unsetTraceID() {
	if gLogIDMap != nil {
		gLogIDMapLock.Lock()
		delete(gLogIDMap, goid.Get())
		gLogIDMapLock.Unlock()
	}
}

func getTraceIDFromLocalMap() (string, bool) {
	var traceID string
	var ok bool

	goroutineId := goid.Get()
	if gLogIDMap != nil {
		gLogIDMapLock.RLock()
		traceID, ok = gLogIDMap[goroutineId]
		gLogIDMapLock.RUnlock()
	}

	return traceID, ok
}
