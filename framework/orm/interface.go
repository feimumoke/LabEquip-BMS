package orm

import (
	"gorm.io/gorm"
	"time"
)

type ConnPool = gorm.ConnPool

type ConnLifetimeManager interface {
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
}

type MultiConnLifetimeManager interface {
	ConnLifetimeManager
	SetMultiConnMaxLifetime(ds string, d time.Duration)
	SetMultiMaxIdleConns(ds string, n int)
	SetMultiMaxOpenConns(ds string, n int)
}
