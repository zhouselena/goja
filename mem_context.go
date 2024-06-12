package goja

import (
	"errors"
	"math"
)

type visitTracker struct {
	objsVisited    map[objectImpl]struct{}
	stashesVisited map[*stash]struct{}
}

func (vt visitTracker) IsObjVisited(obj objectImpl) bool {
	_, ok := vt.objsVisited[obj]
	return ok
}

func (vt visitTracker) VisitObj(obj objectImpl) {
	vt.objsVisited[obj] = struct{}{}
}

func (vt visitTracker) IsStashVisited(stash *stash) bool {
	_, ok := vt.stashesVisited[stash]
	return ok
}

func (vt visitTracker) VisitStash(stash *stash) {
	vt.stashesVisited[stash] = struct{}{}
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
	ComputeSampleStep              func(totalItems int) int
	memoryLimit                    uint64
}

func NewMemUsageContext(
	vm *Runtime,
	maxDepth int,
	memLimit uint64,
	arrayLenThreshold, objPropsLenThreshold int,
	sampleRate float64,
	nativeChecker NativeMemUsageChecker,
) *MemUsageContext {
	return &MemUsageContext{
		visitTracker:          visitTracker{objsVisited: make(map[objectImpl]struct{}), stashesVisited: make(map[*stash]struct{})},
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
		ComputeSampleStep: func(totalItems int) int {
			return computeSampleStep(totalItems, sampleRate)
		},
	}
}

// MemUsageLimitExceeded ensures a limit function is defined and checks against the limit. If limit is breached
// it will return true
func (m *MemUsageContext) MemUsageLimitExceeded(memUsage uint64) bool {
	return memUsage > m.memoryLimit
}

func computeMemUsageEstimate(memUsage, samplesVisited uint64, totalProps int) uint64 {
	// averageMemUsage * total object props
	return uint64(float32(memUsage) / float32(samplesVisited) * float32(totalProps))
}

// computeSampleStep will take the total items we want to sample from and a sample rate.
// It will use this value to determine the sample step, which indicates how often we need
// to grab a sample. For example, with 100 total items and a 0.2 sample rate, it means
// we want to sample 20% of 100 items, in order to do so we need to pick an item ever 5
// (5 * 20 == 100)
func computeSampleStep(totalItems int, sampleRate float64) int {
	if sampleRate == 0 || totalItems == 0 {
		return 1
	}
	if sampleRate >= 0.5 {
		// We allow a max sample size half of the total items
		sampleRate = 0.5
	}

	totalSamples := float64(totalItems) * sampleRate
	return int(math.Floor(float64(totalItems) / totalSamples))
}

var (
	ErrMaxDepth = errors.New("reached max depth")
)

type MemUsageReporter interface {
	// The newMemUsage value is used to allow tracking significant differences in how
	// we track memory usage vs a more accurate tracking. This will help
	// us evaluate the impact of any memory usage value change.
	MemUsage(ctx *MemUsageContext) (memUsage uint64, err error)
}
