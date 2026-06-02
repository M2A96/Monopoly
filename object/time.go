package object

//go:generate mockgen -destination=../test/v2/time.go -package=test -mock_names=Timer=MockTime . Timer

import "time"

type (
	// Timer is an interface.
	Timer interface {
		// NowUTC is a function.
		NowUTC() time.Time
		// Since is a function.
		Since(
			time.Time,
		) time.Duration
	}

	// GetTimer is an interface.
	GetTimer interface {
		// GetTimer is a function.
		GetTimer() Timer
	}

	timeV2 struct {
		time.Time
	}
)

var _ Timer = (*timeV2)(nil)

// NewTime is a function.
func NewTime() *timeV2 {
	return &timeV2{
		Time: time.Time{},
	}
}

// NowUTC is a function.
func (*timeV2) NowUTC() time.Time {
	return time.Now().UTC()
}

// Since is a function.
func (*timeV2) Since(
	timeTime time.Time,
) time.Duration {
	return time.Since(timeTime)
}
