package lib

import (
	"time"
)

type Timer interface {
	ResetTimer()
	IsPast() bool
	GetElapsed() time.Duration
	GetPastCount() uint64
}

type timerImpl struct {
	start       time.Time
	last_update time.Time
	interval    time.Duration
}

func NewTimer(interval time.Duration) Timer {
	return &timerImpl{time.Now(), time.Now(), interval}
}

func (t *timerImpl) ResetTimer() {
	t.last_update = time.Now()
}

func (t *timerImpl) IsPast() bool {
	sub := time.Now().Sub(t.last_update)
	if sub > t.interval {
		t.ResetTimer()
		return true
	}
	return false
}

func (t *timerImpl) GetElapsed() time.Duration {
	return time.Now().Sub(t.last_update)
}

func (t *timerImpl) GetPastCount() uint64 {
	sub := time.Now().Sub(t.start)
	return uint64(sub / t.interval)
}
