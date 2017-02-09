//Package ratelimit provides a primitive for rate limiting arbitrary values
//with a provided throughput given as a time.Duration.
package ratelimit

import (
	"errors"
	"sync"
	"time"
)

//DefaultCapacity is the capacity used for calls to New.
const DefaultCapacity = 1

//ErrClosed designates that a Limiter is already closed in calls to Push and Close.
var ErrClosed = errors.New("ratelimit: limiter already closed")

//Limiter is a primitive that rate limits values pushed to it.
//A maximum of one value can be popped in the allotted duration.
//
//Limiter acts as a first-in-first-out queue where the next value to pop will not
//be released until at least a given duration has passed since the last value
//has been popped.
type Limiter struct {
	lock     *sync.Mutex
	nextTime time.Time

	d time.Duration

	values chan interface{}
}

//New creates a Limiter with a capacity of DefaultCapacity and throughput duration d.
func New(d time.Duration) *Limiter {
	return NewCapacity(d, DefaultCapacity)
}

//NewCapacity creates a Limiter with capacity and throughput duration d.
func NewCapacity(d time.Duration, capacity int) *Limiter {
	return &Limiter{
		lock:     &sync.Mutex{},
		nextTime: time.Now(),
		d:        d,
		values:   make(chan interface{}, capacity),
	}
}

//Push places value in l to be popped later.
//Push does not return until there is space in l to store value (determined by
//l's capacity).
//
//err will be ErrClosed if l.Close() has already been called.
func (l *Limiter) Push(value interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrClosed
		}
	}()
	l.values <- value
	return
}

//Pop releases a value from l.
//It will not return a value until 1) there is a value in l to pop, and 2) the
//provided duration has passed since the most recent return of Pop.
//
//If l is closed, then the returned value will be nil.
func (l *Limiter) Pop() interface{} {
	v, _ := l.PopOk()
	return v
}

//PopOk releases a value from l.
//It works just like Pop, but has an extra return value that designates if l is
//not closed and therefore legitimate.
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

//Close closes l and prevents any more values from being pushed.
//Note that values not yet popped are still available to receive.
//
//If l is already closed, then ErrClosed is returned, otherwise err is nil.
func (l *Limiter) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrClosed
		}
	}()
	close(l.values)
	return
}
