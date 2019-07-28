package timer

import (
	"fmt"
	"time"

	"github.com/TianQinS/fastapi/basic"
	"github.com/TianQinS/fastapi/post"
)

const (
	// The maximum number of times the function was called per second.
	TMAP_CAPACITY = 1024 * 1024
)

var (
	GPost   *post.Post
	TSecond *TimerMap
)

type PostItem struct {
	postFunc interface{}
	postArgs []interface{}
	Second   int64
}

type TimerMap struct {
	itemQueue  *basic.EsQueue
	vals       []interface{}
	lastSecond int64
	dataMap    map[int64][]*PostItem
}

// Imprecise delay call after seconds.
func NewTimerMap(capacity uint64) *TimerMap {
	tm := &TimerMap{
		itemQueue:  basic.NewQueue(capacity),
		vals:       make([]interface{}, capacity),
		lastSecond: 0,
		dataMap:    make(map[int64][]*PostItem, 0),
	}
	return tm
}

func (this *PostItem) Run() {
	if this.postFunc == nil {
		return
	}
	GPost.PutJob(_TIMER_JOB_GROUP, this.postFunc, this.postArgs...)
}

func (this *PostItem) Cancel() {
	this.postFunc = nil
}

func (this *TimerMap) Put(duration int64, f interface{}, postArgs []interface{}) *PostItem {
	item := &PostItem{
		Second:   this.lastSecond + duration,
		postFunc: f,
		postArgs: postArgs,
	}
	ok, quantity := this.itemQueue.Put(item)
	if !ok {
		fmt.Printf("[CallAfterSeconds] put fail quantity=%d", quantity)
	}
	return item
}

func (this *TimerMap) Update() {
	cnt, _ := this.itemQueue.Gets(this.vals)
	for i := uint64(0); i < cnt; i++ {
		val := this.vals[i]
		item := val.(*PostItem)
		key := this.lastSecond + item.Second
		if _, ok := this.dataMap[key]; !ok {
			this.dataMap[key] = make([]*PostItem, 0, 1)
		}
		this.dataMap[key] = append(this.dataMap[key], item)
	}
}

func (this *TimerMap) Tick(args ...interface{}) (res interface{}) {
	this.Update()
	this.lastSecond++
	if itemList, ok := this.dataMap[this.lastSecond]; ok {
		for _, item := range itemList {
			(*item).Run()
		}
		delete(this.dataMap, this.lastSecond)
	}
	return
}

// Delay function for seconds,
// can be called in large quantities (by setting thresholds), but with deviation less than 1 second.
func CallOut(duration int64, callback interface{}, args ...interface{}) *PostItem {
	if duration <= 0 {
		return nil
	}
	return TSecond.Put(duration, callback, args)
}

func init() {
	TSecond = NewTimerMap(TMAP_CAPACITY)
	GPost = post.GPost
	now := time.Now()
	d := time.Second - time.Nanosecond*time.Duration(now.Nanosecond()) + time.Nanosecond
	AddCallback(d, func() {
		AddTimer(time.Second, TSecond.Tick)
	})
}
