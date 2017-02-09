package ratelimit

import (
	"testing"
	"time"
)

func TestNew_createsALimiterWithCapacityOne(t *testing.T) {
	rl := New(time.Duration(1) * time.Second)

	rl.Push(0)
}

func TestNewCapacity_createsALimiterWithCapacity(t *testing.T) {
	rl := NewCapacity(time.Duration(1)*time.Second, 10)

	for i := 0; i < 10; i++ {
		rl.Push("0")
	}
}

func TestLimiter_Push_returnsErrorIfClosed(t *testing.T) {
	rl := New(time.Duration(1))

	if err := rl.Close(); err != nil {
		t.Fail()
	}

	if err := rl.Push(0); err != ErrClosed {
		t.Fail()
	}
}

func TestLimiter_Close_returnsErrorIfClosed(t *testing.T) {
	rl := New(time.Duration(1))

	if err := rl.Close(); err != nil {
		t.Fail()
	}

	if err := rl.Close(); err != ErrClosed {
		t.Fail()
	}
}

func TestLimiter_endToEndWorksForSmallDuration(t *testing.T) {
	rl := New(time.Duration(1))

	go func() {
		for i := 0; i < 10; i++ {
			rl.Push(i)
		}
	}()

	for i := 0; i < 10; i++ {
		v := rl.Pop()
		if v != i {
			t.Fail()
		}
	}
}

func TestLimiter_endToEndWorksReadingUntilClosed(t *testing.T) {
	rl := New(time.Duration(1))

	go func() {
		for i := 0; i < 10; i++ {
			rl.Push(i)
		}
		rl.Close()
	}()

	for i := 0; i < 10; i++ {
		v, ok := rl.PopOk()
		if v != i || !ok {
			t.Fail()
		}
	}

	v, ok := rl.PopOk()
	if v != nil || ok {
		t.Fail()
	}
}

func TestLimiter_endToEndWorksForLargeDurations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping for short")
	}

	d := time.Duration(1) * time.Second
	rl := NewCapacity(d, 2)

	go func() {
		for i := 0; i < 4; i++ {
			rl.Push(i)
		}
		rl.Close()
	}()

	popTimes := make([]time.Time, 0, 10)

	want := 0
	for v, ok := rl.PopOk(); ok; v, ok = rl.PopOk() {
		popTimes = append(popTimes, time.Now())

		if v != want {
			t.Fail()
		}

		want++
	}

	for i := 0; i < len(popTimes)-1; i++ {
		if popTimes[i+1].Sub(popTimes[i]) < d {
			t.Fail()
		}
	}
}
