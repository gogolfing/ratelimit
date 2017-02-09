package ratelimit

import (
	"errors"
	"sync"
	"time"
)

const DefaultCapacity = 1

var ErrClosed = errors.New("ratelimit: limiter already closed")

type Limiter struct {
	lock     *sync.Mutex
	nextTime time.Time

	d time.Duration

	values chan interface{}
}

func New(d time.Duration) *Limiter {
	return NewCapacity(d, DefaultCapacity)
}

func NewCapacity(d time.Duration, capacity int) *Limiter {
	return &Limiter{
		lock:     &sync.Mutex{},
		nextTime: time.Now(),
		d:        d,
		values:   make(chan interface{}, capacity),
	}
}

func (l *Limiter) Push(value interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrClosed
		}
	}()
	l.values <- value
	return
}

func (l *Limiter) Pop() interface{} {
	v, _ := l.PopOk()
	return v
}

func (l *Limiter) PopOk() (interface{}, bool) {
	v, ok := <-l.values
	if !ok {
		return nil, ok
	}

	l.waitAndBumpNextTime()

	return v, ok
}

func (l *Limiter) waitAndBumpNextTime() {
	l.lock.Lock()
	defer l.lock.Unlock()

	time.Sleep(l.nextTime.Sub(time.Now()))

	l.nextTime = time.Now().Add(l.d)
}

func (l *Limiter) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrClosed
		}
	}()
	close(l.values)
	return
}
