// Package timer is a timer for timer task.
// It contains runtime crontab and conventional timer.
package timer

import (
	"container/heap"
	"sync"
	"time"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
	// the minimum time interval before next call.
	TIME_INTERVAL = 10 * time.Millisecond
	// the asynchronous worker key.
	_TIMER_JOB_GROUP = "timer"
)

var (
	// the adding order.
	nextAddSeq    uint
	timerHeap     TimerHeap
	timerHeapLock sync.Mutex
)

type Timer struct {
	fireTime  time.Time
	interval  time.Duration
	asyncFunc interface{}
	params    []interface{}
	repeat    bool
	addseq    uint
}

type TimerHeap struct {
	timers []*Timer
}

/** Timer heap. **/
func (this *TimerHeap) Less(i, j int) bool {
	t1, t2 := this.timers[i].fireTime, this.timers[j].fireTime
	if t1.Before(t2) {
		return true
	}
	if t1.After(t2) {
		return false
	}
	// Making sure Timer with same deadline is fired according to their add order.
	return this.timers[i].addseq < this.timers[j].addseq
}

func (this *TimerHeap) Swap(i, j int) {
	this.timers[i], this.timers[j] = this.timers[j], this.timers[i]
}

func (this *TimerHeap) Push(item interface{}) {
	this.timers = append(this.timers, item.(*Timer))
}

func (this *TimerHeap) Pop() (item interface{}) {
	l := len(this.timers)
	this.timers, item = this.timers[:l-1], this.timers[l-1]
	return
}

func (this *TimerHeap) Len() int {
	return len(this.timers)
}

/** Process timer **/
func (this *Timer) Cancel() {
	this.asyncFunc = nil
}

func (this *Timer) IsActive() bool {
	return this.asyncFunc != nil
}

// Add a callback for the timer, it will be executed asynchronously.
func Add(d time.Duration, f interface{}, repeat bool, args []interface{}) *Timer {
	if d < TIME_INTERVAL {
		d = TIME_INTERVAL
	}
	t := &Timer{
		fireTime:  time.Now().Add(d),
		interval:  d,
		asyncFunc: f,
		params:    args,
		repeat:    repeat,
	}

	timerHeapLock.Lock()
	t.addseq = nextAddSeq
	nextAddSeq++
	heap.Push(&timerHeap, t)
	timerHeapLock.Unlock()
	return t
}

// Add a callback which will be called after specified duration.
func AddCallback(d time.Duration, f interface{}, args ...interface{}) *Timer {
	return Add(d, f, false, args)
}

// Add a timer which calls callback periodly.
func AddTimer(d time.Duration, f interface{}, args ...interface{}) *Timer {
	return Add(d, f, true, args)
}

// Tick once for timers.
func Tick() {
	defer timerHeapLock.Unlock()
	now := time.Now()
	timerHeapLock.Lock()
	for {
		if timerHeap.Len() <= 0 {
			nextAddSeq = 1
			break
		}
		nextFireTime := timerHeap.timers[0].fireTime
		if nextFireTime.After(now) {
			break
		}

		t := heap.Pop(&timerHeap).(*Timer)
		if t.asyncFunc == nil {
			continue
		}
		if t.asyncFunc != nil {
			GPost.PutJob(_TIMER_JOB_GROUP, t.asyncFunc, t.params...)
		}

		if t.repeat {
			t.fireTime = t.fireTime.Add(t.interval)
			if !t.fireTime.After(now) { // Might happen when interval is very small
				t.fireTime = now.Add(t.interval)
			}
			t.addseq = nextAddSeq
			nextAddSeq++
			// Add Timer back to heap
			heap.Push(&timerHeap, t)
		} else {
			t.asyncFunc = nil
		}
	}
}

func selfTickRoutine(tickInterval time.Duration) {
	for {
		time.Sleep(tickInterval)
		Tick()
	}
}

// Start the self-ticking routine, which ticks per tickInterval
// Initialize crontab module.
func StartTicks(tickInterval time.Duration) {
	go selfTickRoutine(tickInterval)
	startCrontab()
}

func init() {
	// Init minimum heap.
	heap.Init(&timerHeap)
	StartTicks(TIME_INTERVAL)
}
