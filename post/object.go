package post

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/TianQinS/fastapi/basic"
)

const (
	// max sleep time in ms.
	MAX_SLEEP_TIME = 10 * time.Millisecond
)

type QueueMsg struct {
	Func            interface{}
	Params          []interface{}
	StrictUnReflect bool
}

type RpcObject struct {
	Functions map[string]interface{}
	// high performance lock-free queue, better performance than Chan at high load.
	Queue *basic.EsQueue
	// for batch extraction of queue data.
	Vals     []interface{}
	IsRun    bool
	itemPool sync.Pool
}

func (this *QueueMsg) Init(f interface{}, params []interface{}, strict bool) {
	this.Func = f
	this.Params = params
	this.StrictUnReflect = strict
}

func (this *RpcObject) Init(qSize uint64) {
	this.Functions = map[string]interface{}{}
	this.Queue = basic.NewQueue(qSize)
	this.Vals = make([]interface{}, qSize)
	this.IsRun = true
}

// Register functions for object, f can be any function type,
// but must be an `func(args ...interface{})` type in strict mode without reflect.
func (this *RpcObject) Register(id string, f interface{}) {
	if _, ok := this.Functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}
	this.Functions[id] = f
}

func (this *RpcObject) newMsg(f interface{}, params []interface{}, strict bool) *QueueMsg {
	item, ok := this.itemPool.Get().(*QueueMsg)
	if ok {
		item.Init(f, params, strict)
		return item
	}
	return &QueueMsg{
		Func:            f,
		Params:          params,
		StrictUnReflect: strict,
	}
}

func (this *RpcObject) releaseMsg(item *QueueMsg) {
	this.itemPool.Put(item)
}

func (this *RpcObject) PutQueue(f interface{}, strictUnReflect bool, params ...interface{}) error {
	ok, quantity := this.Queue.Put(this.newMsg(f, params, strictUnReflect))
	if !ok {
		return fmt.Errorf("Put Fail, quantity:%v\n", quantity)
	}
	return nil
}

func (this *RpcObject) PutQueueForPost(f interface{}, strictUnReflect bool, params []interface{}) error {
	ok, quantity := this.Queue.Put(this.newMsg(f, params, strictUnReflect))
	if !ok {
		return fmt.Errorf("Put Fail, quantity:%v\n", quantity)
	}
	return nil
}

func (this *RpcObject) executeEvent(cnt uint64, vals *[]interface{}) {
	var ok bool
	var function interface{}
LOOP:
	for i := uint64(0); i < cnt; i++ {
		val := (*vals)[i]
		msg := val.(*QueueMsg)
		f := msg.Func

		switch f.(type) {
		case string:
			if function, ok = this.Functions[f.(string)]; !ok {
				fmt.Printf("Remote function(%v) not found\n", f)
				continue LOOP
			}
		default:
			function = f
		}
		_runFunc := func() {
			defer func() {
				if e, ok := recover().(error); ok {
					basic.PackErrorMsg(e, msg)
				}
			}()

			if msg.StrictUnReflect {
				function.(func(args ...interface{}))(msg.Params...)
			} else {
				_f := reflect.ValueOf(function)
				in := make([]reflect.Value, len(msg.Params))
				for k := range in {
					in[k] = reflect.ValueOf(msg.Params[k])
				}
				_f.Call(in)
			}
		}
		_runFunc()
	}
}

// Can only be executed in one gorountine.
// This function returns number of events which can be used for dynamic sleep.
func (this *RpcObject) ExecuteEvent() uint64 {
	cnt, _ := this.Queue.Gets(this.Vals)
	this.executeEvent(cnt, &this.Vals)
	return cnt
}

// Can be executed concurrently but not commonly used.
func (this *RpcObject) ExecuteEventSafe() uint64 {
	qSize := this.Queue.Quantity()
	vals := make([]interface{}, 2*qSize)
	cnt, _ := this.Queue.Gets(vals)
	this.executeEvent(cnt, &vals)
	return cnt
}

// The main loop of RpcObject.
func (this *RpcObject) Loop() {
	for this.IsRun {
		start := time.Now()
		this.ExecuteEvent()
		delta := MAX_SLEEP_TIME - time.Now().Sub(start)
		if delta > 0 {
			time.Sleep(delta)
		} else {
			runtime.Gosched()
		}
	}
}

func (this *RpcObject) Close() {
	this.IsRun = false
	this.ExecuteEventSafe()
}
