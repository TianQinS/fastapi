// The pool for gorountine, suitable for high load.
package post

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

const (
	// EsQueue Capacity.
	ITEM_QUEUE_CAPACITY = 1024 * 1024
	// the initial numbers of goroutine.
	ORI_ROUTINE_NUM = 4
	// max sleep time in ms.
	MAX_SLEEP_TIME = 10 * time.Millisecond
)

var (
	GPost *Post
)

type Post struct {
	objects   []*RpcObject
	Functions map[string]interface{}
	qSize     uint64
	index     int
	lock      *sync.Mutex
}

func init() {
	GPost = NewPost(ITEM_QUEUE_CAPACITY, ORI_ROUTINE_NUM)
}

func NewPost(queueCapacity uint64, oriNum int) *Post {
	p := &Post{
		index:   0,
		qSize:   queueCapacity,
		objects: make([]*RpcObject, 0, oriNum),
		lock:    new(sync.Mutex),
	}
	p.AddObjects(oriNum)
	return p
}

func (this *Post) Size() int {
	return len(this.objects)
}

func (this *Post) Register(id string, f interface{}) {
	if _, ok := this.Functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}
	this.Functions[id] = f
}

func (this *Post) AddOne() *RpcObject {
	defer this.lock.Unlock()
	this.lock.Lock()

	var o *RpcObject
	if this.index < this.Size() && this.index >= 0 {
		o = this.objects[this.index]
		if !o.IsRun {
			o.IsRun = true
		}
	} else {
		o = &RpcObject{}
		o.Init(this.qSize)
		o.Functions = this.Functions
	}

	go func() {
		for o.IsRun {
			start := time.Now()
			o.ExecuteEvent()
			delta := MAX_SLEEP_TIME - time.Now().Sub(start)
			if delta > 0 {
				time.Sleep(delta)
			} else {
				runtime.Gosched()
			}
		}
	}()
	this.objects = append(this.objects, o)
	this.index++
	return o
}

func (this *Post) AddObjects(num int) {
	for num > 0 {
		this.AddOne()
		num--
	}
}

func (this *Post) DelOne() {
	defer this.lock.Unlock()
	this.lock.Lock()
	index := this.index - 1
	if index >= 0 {
		o := this.objects[index]
		o.IsRun = false
		o.ExecuteEventSafe()
		this.index = index
	}
}

// Close all rountines for pre shutdown.
func (this *Post) Close() {
	defer this.lock.Unlock()
	this.lock.Lock()
	if this.Size() > 0 {
		for _, o := range this.objects {
			this.index--
			o.IsRun = false
			o.ExecuteEventSafe()
		}
	}
}

func (this *Post) PutQueue(f interface{}, params ...interface{}) error {
	index := this.index
	if index > 0 {
		o := this.objects[rand.Intn(this.index)]
		return o.PutQueueForPost(f, false, params)
	}
	return nil
}

// Call a function with routine pool in high load situations.
func (this *Post) PutQueueStrict(f interface{}, params ...interface{}) error {
	index := this.index
	if index > 0 {
		o := this.objects[rand.Intn(this.index)]
		return o.PutQueueForPost(f, true, params)
	}
	return nil
}

// Append an asynchronous task, new worker will be created dynamically by the group.
func (this *Post) PutJob(group string, f interface{}, params ...interface{}) {
	worker := getJobWorker(group)
	worker.appendJob(f, false, params)
}

func (this *Post) PutJobStrict(group string, f interface{}, params ...interface{}) {
	worker := getJobWorker(group)
	worker.appendJob(f, true, params)
}
