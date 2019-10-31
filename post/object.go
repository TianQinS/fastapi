package post

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/TianQinS/fastapi/basic"
)

const (
	// max sleep time in ms.
	MAX_SLEEP_TIME = 10000 * time.Microsecond
)

type QueueMsg struct {
	Func            interface{}
	Callback        interface{}
	Params          []interface{}
	CallbackParams  []interface{}
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

func (this *QueueMsg) Init(f, cb interface{}, params, cbParams []interface{}, strict bool) {
	this.Func = f
	this.Callback = cb
	this.Params = params
	this.CallbackParams = cbParams
	this.StrictUnReflect = strict
}

func (this *RpcObject) Init(qSize uint64) {
	this.Functions = map[string]interface{}{}
	this.Queue = basic.NewQueue(qSize)
	this.Vals = make([]interface{}, qSize, qSize)
	this.IsRun = true
}

// Register functions for object, f can be any function type,
// but must be an `func(args ...interface{})` type in strict mode without reflect.
func (this *RpcObject) Register(id string, f interface{}) {
	if _, ok := this.Functions[id]; ok {
		log.Println(fmt.Sprintf("function id %v: already registered", id))
	}
	this.Functions[id] = f
}

func (this *RpcObject) newMsg(f, cb interface{}, params, cbParams []interface{}, strict bool) *QueueMsg {
	item, ok := this.itemPool.Get().(*QueueMsg)
	if ok {
		item.Init(f, cb, params, cbParams, strict)
		return item
	}
	return &QueueMsg{
		Func:            f,
		Callback:        cb,
		Params:          params,
		CallbackParams:  cbParams,
		StrictUnReflect: strict,
	}
}

func (this *RpcObject) releaseMsg(item *QueueMsg) {
	this.itemPool.Put(item)
}

func (this *RpcObject) PutQueue(f interface{}, strictUnReflect bool, params ...interface{}) error {
	ok, quantity := this.Queue.Put(this.newMsg(f, nil, params, nil, strictUnReflect))
	if !ok {
		return fmt.Errorf("Put Fail, quantity:%v\n", quantity)
	}
	return nil
}

func (this *RpcObject) PutQueueWithCallback(f, cb interface{}, strictUnReflect bool, cbParams, params []interface{}) error {
	ok, quantity := this.Queue.Put(this.newMsg(f, cb, params, cbParams, strictUnReflect))
	// fmt.Println(ok, f, cb, quantity, params)
	if !ok {
		return fmt.Errorf("Put Fail, quantity:%v\n", quantity)
	}
	return nil
}

func (this *RpcObject) PutQueueForPost(f interface{}, strictUnReflect bool, params []interface{}) error {
	ok, quantity := this.Queue.Put(this.newMsg(f, nil, params, nil, strictUnReflect))
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
				log.Println("Remote function(%v) not found", f)
				continue LOOP
			}
		default:
			function = f
		}
		_runFunc := func() {
			defer func() {
				info := recover()
				switch info.(type) {
				case error:
					basic.PackErrorMsg(info.(error), msg)
				case string:
					basic.PackErrorMsg(fmt.Errorf("%s->%s", info.(string), runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()), fmt.Sprintf("%+v", msg))
				default:
					if info != nil {
						basic.PackErrorMsg(fmt.Errorf("%+v", info), msg)
					}
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

				rets := _f.Call(in)
				// fmt.Println(cnt, msg.Func, msg.Params, rets)
				// process callback logic.
				if cb := msg.Callback; cb != nil {
					_f = reflect.ValueOf(cb)
					params := msg.CallbackParams
					if params != nil {
						for k := range params {
							rets = append([]reflect.Value{reflect.ValueOf(params[k])}, rets...)
						}
					}
					_f.Call(rets)
				}
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
