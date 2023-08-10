package goja

import (
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

const burstDivisor = 5

func (self *Runtime) fillBucket() {
	self.limiterTicksLeft = self.limiter.Burst() / burstDivisor
}
