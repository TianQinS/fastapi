// The pool for rountine.
package post

import (
	"math/rand"
	"runtime"
	"sync"
	"time"
)

const (
	// max sleep time in ms.
	MAX_SLEEP_TIME = 10
	// EsQueue Capacity.
	ITEM_QUEUE_CAPACITY = 1024 * 1024
	// the initial numbers of goroutine.
	ORI_ROUTINE_NUM = 6
)

var (
	GPost *Post
)

type Post struct {
	objects []*RpcObject
	qSize   uint64
	index   int
	lock    *sync.Mutex
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
	}

	go func() {
		for o.IsRun {
			n := o.ExecuteEvent()
			n = MAX_SLEEP_TIME - n
			if n > 0 {
				time.Sleep(time.Duration(n) * time.Millisecond)
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

func (this *Post) PutQueue(f interface{}, strictUnReflect bool, params ...interface{}) error {
	index := this.index
	if index > 0 {
		o := this.objects[rand.Intn(this.index)]
		return o.PutQueueForPost(f, strictUnReflect, params)
	}
	return nil
}
