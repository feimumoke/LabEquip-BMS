package task

import (
	"context"
	"github.com/feimumoke/wechating/framework/constant"
	"github.com/feimumoke/wechating/framework/log"
	"github.com/feimumoke/wechating/framework/support/trace"
	"github.com/feimumoke/wechating/framework/wcerror"
	"github.com/panjf2000/ants/v2"
	"runtime/debug"
	"time"
)

// 线程池
type RoutinePool struct {
	antsPool *ants.Pool
}
type PoolLogger struct{}

func NewPoolLogger() *PoolLogger {
	return &PoolLogger{}
}

func (l *PoolLogger) Printf(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// 只能通过New方法创建实例，确保属性完成初始化
func NewRoutinePool(size int) (*RoutinePool, error) {
	pool, err := ants.NewPool(size, ants.WithLogger(NewPoolLogger()))
	if err != nil {
		return nil, err
	}
	return &RoutinePool{
		antsPool: pool,
	}, nil
}

type Task func(ctx context.Context) *bmserror.BMSError

// 修改线程池大小
func (pool *RoutinePool) Tune(size int) {
	if pool.antsPool == nil {
		return
	}
	//pool.antsPool.Tune(size)
}

func WrapperGoroutinePoolTask(task func(ctx context.Context) *bmserror.BMSError) func(ctx context.Context) *bmserror.BMSError {
	return func(ctx context.Context) *bmserror.BMSError {
		//set log id
		defer trace.UnsetCtxTraceID(context.TODO())
		traceID := trace.GetOrNewTraceID(ctx)
		ctx = trace.SetCtxTraceID(ctx, traceID)

		start := time.Now().Unix()
		defer func() {
			log.Infof("Goroutine Pool Task consumed:[%v]", time.Now().Unix()-start)
		}()

		return task(ctx)
	}
}

// 同步执行单个任务
func (pool *RoutinePool) SubmitTask(ctx context.Context, task Task) *bmserror.BMSError {
	if pool.antsPool == nil {
		return bmserror.NewError(constant.ErrInternalServer, "pls init ants pool")
	}
	wcErr := make(chan *bmserror.BMSError)
	err := pool.antsPool.Submit(func() {
		defer func() {
			if p := recover(); p != nil {
				wcErr <- bmserror.NewError(constant.ErrInternalServer, "ants pool submit task error because single task panic:%v", p)
				log.Infof("ants pool submit task error because single task err:%v, panic: %v", p, string(debug.Stack()))
			}
		}()
		t := WrapperGoroutinePoolTask(task)
		select {
		case <-ctx.Done():
			wcErr <- bmserror.NewError(constant.ErrInternalServer, ctx.Err().Error())
			return
		default:
		}
		wcErr <- t(ctx)
	})
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, "ants pool submit task error:(%v)", err.Error())
	}
	return <-wcErr
}

// 多个任务并发执行，全部任务执行成功或某个任务执行失败时返回
func (pool *RoutinePool) SubmitTaskList(ctx context.Context, taskList []Task) *bmserror.BMSError {
	if pool.antsPool == nil {
		return bmserror.NewError(constant.ErrInternalServer, "pls init ants pool")
	}
	group, cancelCtx := NewErrGroupWithContext(ctx)
	for i := 0; i < len(taskList); i++ {
		task := taskList[i]
		group.Add()
		// 协程池执行
		err := pool.antsPool.Submit(func() {
			group.Go(func() (gErr *bmserror.BMSError) {
				defer func() {
					if p := recover(); p != nil {
						gErr = bmserror.NewError(constant.ErrInternalServer, "ants pool submit task error because multi task panic:%v", p)
						log.Infof("ants pool submit task error because single task err:%v panic: %v", p, string(debug.Stack()))
					}
				}()
				t := WrapperGoroutinePoolTask(task)
				select {
				case <-cancelCtx.Done():
					return bmserror.NewError(constant.ErrInternalServer, cancelCtx.Err().Error())
				default:
				}
				gErr = t(cancelCtx)
				return gErr
			})
		})
		if err != nil {
			if group.cancel != nil {
				group.cancel()
			}
			return bmserror.NewError(constant.ErrInternalServer, "ants pool submit task error:(%v)", err.Error())
		}
	}
	return group.Wait()
}

// SubmitTaskAndExecuteAsync 提交任务并异步执行，当pool无空闲协程时此方法阻塞至空闲出协程（根据NewPool参数控制）
func (pool *RoutinePool) SubmitTaskAndExecuteAsync(ctx context.Context, task Task) *bmserror.BMSError {
	if pool.antsPool == nil {
		return bmserror.NewError(constant.ErrInternalServer, "pls init ants pool")
	}
	err := pool.antsPool.Submit(func() {
		defer func() {
			if p := recover(); p != nil {
				log.Infof("ants pool submit task error because single task panic: %v", string(debug.Stack()))
			}
		}()
		task(ctx)
	})
	if err != nil {
		return bmserror.NewError(constant.ErrInternalServer, "ants pool submit task error:(%v)", err.Error())
	}
	return nil
}
