package goja

import (
	"context"
	"strings"
)

func (self *Runtime) waitOneTick(ticks int) {
	self.ticks++
	if self.Limiter == nil {
		return
	}
	ctx := self.vm.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if waitErr := self.Limiter.Wait(ctx); waitErr != nil {
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
