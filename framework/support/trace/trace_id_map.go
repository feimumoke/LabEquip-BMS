package trace

import (
	"github.com/sillyousu/goid"
	"sync"
)

var (
	gLogIDMap     = make(map[int64]string, 1000000)
	gLogIDMapLock sync.RWMutex
)

func setLogTraceID(logid string) {
	if gLogIDMap != nil {
		gLogIDMapLock.Lock()
		goroutineId := goid.Goid()
		gLogIDMap[goroutineId] = logid
		gLogIDMapLock.Unlock()
	}
}

func unsetTraceID() {
	if gLogIDMap != nil {
		gLogIDMapLock.Lock()
		delete(gLogIDMap, goid.Goid())
		gLogIDMapLock.Unlock()
	}
}

func getTraceIDFromLocalMap() (string, bool) {
	var traceID string
	var ok bool

	goroutineId := goid.Goid()
	if gLogIDMap != nil {
		gLogIDMapLock.RLock()
		traceID, ok = gLogIDMap[goroutineId]
		gLogIDMapLock.RUnlock()
	}

	return traceID, ok
}
