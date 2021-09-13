package goja

import (
	"context"
	"strings"

	"golang.org/x/time/rate"
)

// SetRateLimiter sets the rate limiter
func (self *Runtime) SetRateLimiter(limiter *rate.Limiter) {
	self.limiter = limiter
	if limiter == nil {
		return
	}

	self.fillBucket()
}

// NOTE: we should try to avoid making expensive operations within this
// function since it gets called millions of times per second.
func (self *Runtime) waitOneTick() {
	self.ticks++
	if self.limiter == nil {
		return
	}

	if self.limiterTicksLeft > 0 {
		self.limiterTicksLeft--
		return
	}
	self.fillBucket()

	ctx := self.vm.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if waitErr := self.limiter.WaitN(ctx, self.limiterTicksLeft); waitErr != nil {
		if self.vm.ctx == nil {
			panic(waitErr)
		}
		if ctxErr := self.vm.ctx.Err(); ctxErr != nil {
			panic(ctxErr)
		}
		if strings.Contains(waitErr.Error(), "would exceed") {
			panic(context.DeadlineExceeded)
		}
		panic(waitErr)
	}
}

const burstDivisor = 5

func (self *Runtime) fillBucket() {
	self.limiterTicksLeft = self.limiter.Burst() / burstDivisor
}
