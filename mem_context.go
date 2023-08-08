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

type MemUsageContext struct {
	visitTracker
	*depthTracker
	NativeMemUsageChecker
	ArrayLenExceedsThreshold       func(arrayLen int) bool
	ObjectPropsLenExceedsThreshold func(objPropsLen int) bool
	memoryLimit                    uint64
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
		memoryLimit:           memLimit,
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

// MemUsageLimitExceeded ensures a limit function is defined and checks against the limit. If limit is breached
// it will return true
func (m *MemUsageContext) MemUsageLimitExceeded(memUsage uint64) bool {
	return memUsage > m.memoryLimit
}

var (
	ErrMaxDepth = errors.New("reached max depth")
)

type MemUsageReporter interface {
	// The newMemUsage value is used to allow tracking significant differences in how
	// we track memory usage vs a more accurate tracking. This will help
	// us evaluate the impact of any memory usage value change.
	MemUsage(ctx *MemUsageContext) (memUsage uint64, newMemUsage uint64, err error)
}
