package retry

import (
	"math"
	"math/rand"
	"time"
)

func init() {
	RetriesAttemptedCelling = int(math.Log2(float64(math.MaxInt)))
}

var RetriesAttemptedCelling int

type EqualJitterBackoff struct {
	baseDelay  time.Duration
	maxBackoff time.Duration
}

func NewEqualJJitterBackoff(baseDelay time.Duration, maxBackoff time.Duration) *EqualJitterBackoff {
	return &EqualJitterBackoff{
		baseDelay:  baseDelay,
		maxBackoff: maxBackoff,
	}
}

func (j *EqualJitterBackoff) BackoffDelay(attempt int, err error) (time.Duration, error) {
	// ceil = min(2 ^ attempts * baseDealy, maxBackoff)
	// ceil/2 + [0.0, 1.0) *(ceil/2 + 1)
	if attempt > RetriesAttemptedCelling {
		attempt = RetriesAttemptedCelling
	}
	delayDuration := j.baseDelay * (1 << attempt)
	if delayDuration > j.maxBackoff {
		delayDuration = j.maxBackoff
	}
	half := delayDuration.Seconds() / 2
	return floatSecondsDuration(half + rand.Float64()*float64(half+1)), nil
}

type FullJitterBackoff struct {
	baseDelay  time.Duration
	maxBackoff time.Duration
}

func NewFullJitterBackoff(baseDelay time.Duration, maxBackoff time.Duration) *FullJitterBackoff {
	return &FullJitterBackoff{
		baseDelay:  baseDelay,
		maxBackoff: maxBackoff,
	}
}

func (j *FullJitterBackoff) BackoffDelay(attempt int, err error) (time.Duration, error) {
	// [0.0, 1.0) * min(2 ^ attempts * baseDealy, maxBackoff)
	if attempt > RetriesAttemptedCelling {
		attempt = RetriesAttemptedCelling
	}
	delayDuration := j.baseDelay * (1 << attempt)
	if delayDuration > j.maxBackoff {
		delayDuration = j.maxBackoff
	}
	return floatSecondsDuration(rand.Float64() * float64(delayDuration.Seconds())), nil
}

type FixedDelayBackoff struct {
	fixedBackoff time.Duration
}

func NewFixedDelayBackoff(fixedBackoff time.Duration) *FixedDelayBackoff {
	return &FixedDelayBackoff{
		fixedBackoff: fixedBackoff,
	}
}

func (j *FixedDelayBackoff) BackoffDelay(attempt int, err error) (time.Duration, error) {
	return j.fixedBackoff, nil
}

func floatSecondsDuration(v float64) time.Duration {
	return time.Duration(v * float64(time.Second))
}
