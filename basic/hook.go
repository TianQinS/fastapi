// A basic hookmgr for hook functions.
package basic

import (
	"sync"
	"time"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
)

var (
	lock    sync.Mutex
	HookMgr = make(Hooks)
)

// A hook will be fired when the function Fire is called.
type Hook interface {
	Timeout() bool
	Fire(args ...interface{})
}

// Internal type for storing the hooks on a special instance.
type Hooks map[string][]Hook

// Fire all the valid hooks for the given key, a mutex lock is needed in Fire function against actual conditions.
func (this Hooks) Fire(key string, args ...interface{}) (err error) {
	if hooks, ok := this[key]; ok {
		for _, hook := range hooks {
			if hook.Timeout() {
				continue
			}
			hook.Fire(args...)
		}
	}
	return
}

// Add a hook to an instance of Hooks. This is called with
// `Hooks.Add(new(MyHook))` where `MyHook` implements the `Hook` interface.
func (this Hooks) Add(key string, hook Hook) {
	if hook.Timeout() {
		return
	}
	defer lock.Unlock()
	lock.Lock()
	this[key] = append(this[key], hook)
}

// Easy use for hookmgr.
func (this Hooks) Register(key, start, end string, fire func(args ...interface{})) (err error) {
	hook := new(BasicHook)
	if err = hook.SetTimeout(start, end); err == nil {
		hook.callfunc = fire
		this.Add(key, hook)
	}
	return
}

// Base class used to be inherited for hook object.
type BasicHook struct {
	start    time.Time
	end      time.Time
	callfunc func(args ...interface{})
}

func (this *BasicHook) SetTimeout(start, end string) (err error) {
	this.start, err = time.ParseInLocation(TIME_FORMAT, start, time.Local)
	if err == nil {
		this.end, err = time.ParseInLocation(TIME_FORMAT, end, time.Local)
	}
	return err
}

// Check if the hook obj is valid.
func (this *BasicHook) Timeout() bool {
	now := time.Now()
	if now.After(this.end) || now.Before(this.start) {
		return true
	}
	return false
}

// Used by `HookMgr` to fire the given key's hooks.
func (this *BasicHook) Fire(args ...interface{}) {
	if this.callfunc != nil {
		CatchWithParams(this.callfunc, args...)
	}
}
