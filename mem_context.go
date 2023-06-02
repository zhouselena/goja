package goja

import (
	"errors"
)

type visitTracker struct {
	objsVisited    map[objectImpl]bool
	stashesVisited map[*stash]bool
}

func (vt visitTracker) IsObjVisited(obj objectImpl) bool {
	_, ok := vt.objsVisited[obj]
	return ok
}

func (vt visitTracker) VisitObj(obj objectImpl) {
	vt.objsVisited[obj] = true
}

func (vt visitTracker) IsStashVisited(stash *stash) bool {
	_, ok := vt.stashesVisited[stash]
	return ok
}

func (vt visitTracker) VisitStash(stash *stash) {
	vt.stashesVisited[stash] = true
}

type depthTracker struct {
	curDepth int
	maxDepth int
}

func (dt depthTracker) Depth() int {
	return dt.curDepth
}

func (dt *depthTracker) Descend() error {
	if dt.curDepth == dt.maxDepth {
		return ErrMaxDepth
	}
	dt.curDepth++
	return nil
}

func (dt *depthTracker) Ascend() {
	if dt.curDepth == 0 {
		panic("can't ascend with depth 0")
	}
	dt.curDepth--
}

type NativeMemUsageChecker interface {
	NativeMemUsage(goNativeValue interface{}) (uint64, bool)
}

func (self *stash) MemUsage(ctx *MemUsageContext) (uint64, error) {
	if ctx.IsStashVisited(self) {
		return 0, nil
	}
	ctx.VisitStash(self)
	total := uint64(0)
	if self.obj != nil {
		inc, err := self.obj.MemUsage(ctx)
		total += inc
		if err != nil {
			return total, err
		}
	}

	if self.outer != nil {
		inc, err := self.outer.MemUsage(ctx)
		total += inc
		if err != nil {
			return total, err
		}
	}
	if len(self.values) > 0 {
		inc, err := self.values.MemUsage(ctx)
		total += inc
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

type MemUsageContext struct {
	visitTracker
	*depthTracker
	NativeMemUsageChecker
	MemUsageExceedsLimit           func(memUsage uint64) bool
	ArrayLenExceedsThreshold       func(arrayLen int) bool
	ObjectPropsLenExceedsThreshold func(objPropsLen int) bool
}

func NewMemUsageContext(
	vm *Runtime,
	maxDepth int,
	memLimit uint64,
	arrayLenThreshold, objPropsLenThreshold int,
	nativeChecker NativeMemUsageChecker,
) *MemUsageContext {
	return &MemUsageContext{
		visitTracker:          visitTracker{objsVisited: map[objectImpl]bool{}, stashesVisited: map[*stash]bool{}},
		depthTracker:          &depthTracker{curDepth: 0, maxDepth: maxDepth},
		NativeMemUsageChecker: nativeChecker,
		MemUsageExceedsLimit: func(memUsage uint64) bool {
			// memory usage limit above which we should stop mem usage computations
			return memUsage > memLimit
		},
		ArrayLenExceedsThreshold: func(arrayLen int) bool {
			// array length threshold above which we should estimate mem usage
			return arrayLen > arrayLenThreshold
		},
		ObjectPropsLenExceedsThreshold: func(objPropsLen int) bool {
			// number of obj props beyond which we should estimate mem usage
			return objPropsLen > objPropsLenThreshold
		},
	}
}

var (
	ErrMaxDepth = errors.New("reached max depth")
)

type MemUsageReporter interface {
	MemUsage(ctx *MemUsageContext) (uint64, error)
}
