package clock

import (
	"time"
)

type (
	Interface interface {
		Now() time.Time
	}

	RealClock struct{}

	FrozenClock struct {
		time time.Time
	}
)

func NewRealClock() RealClock {
	return RealClock{}
}

func NewFrozenClock(time time.Time) FrozenClock {
	return FrozenClock{time: time}
}

func (_ RealClock) Now() time.Time {
	return time.Now()
}

func (f FrozenClock) Now() time.Time {
	return f.time
}
