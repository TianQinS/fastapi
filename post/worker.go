// Job worker with goroutine, suitable for inactive calls.
package post

import (
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/TianQinS/fastapi/basic"
)

const (
	// job queue's buffer size.
	ASYNC_JOB_QUEUE_MAXLEN = 10000
)

var (
	JobWorkersLock       sync.RWMutex
	numJobWorkersRunning sync.WaitGroup
	// create a default object.
	JobWorkers = map[string]*JobWorker{}
)

type JobWorker struct {
	jobQueue chan QueueMsg
}

// A job worker will create a goroutine with the memory consumption of the jobQueue.
func newJobWorker() *JobWorker {
	worker := &JobWorker{
		jobQueue: make(chan QueueMsg, ASYNC_JOB_QUEUE_MAXLEN),
	}
	numJobWorkersRunning.Add(1)
	go worker.loop()
	return worker
}

// Gets or creates a worker with the name you specify
func getJobWorker(group string) (worker *JobWorker) {
	// The read lock.
	JobWorkersLock.RLock()
	worker, _ = JobWorkers[group]
	JobWorkersLock.RUnlock()

	if worker == nil {
		JobWorkersLock.Lock()
		worker = newJobWorker()
		log.Printf("[NewJobWorker] group=%s\n", group)
		JobWorkers[group] = worker
		JobWorkersLock.Unlock()
	}
	return
}

func (this *JobWorker) loop() {
	defer numJobWorkersRunning.Done()
	for msg := range this.jobQueue {
		_runFunc := func() {
			defer func() {
				info := recover()
				switch info.(type) {
				case error:
					basic.PackErrorMsg(info.(error), msg)
				case string:
					basic.PackErrorMsg(fmt.Errorf("%s", info.(string)), msg)
				default:
					if info != nil {
						basic.PackErrorMsg(fmt.Errorf("%+v", info), msg)
					}
				}
			}()

			if msg.StrictUnReflect {
				msg.Func.(func(args ...interface{}))(msg.Params...)
			} else {
				_f := reflect.ValueOf(msg.Func)
				in := make([]reflect.Value, len(msg.Params))
				for k := range in {
					in[k] = reflect.ValueOf(msg.Params[k])
				}

				rets := _f.Call(in)
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

func (this *JobWorker) appendJob(f interface{}, strictUnReflect bool, params []interface{}) {
	this.jobQueue <- QueueMsg{f, nil, params, nil, strictUnReflect}
}

func (this *JobWorker) appendJobWithCallback(f, cb interface{}, cbParams, params []interface{}) {
	this.jobQueue <- QueueMsg{f, cb, params, cbParams, false}
}

func Close() bool {
	var cleared bool
	// Close the global gorountine pool.
	if GPost != nil {
		GPost.Close()
	}
	// Close all job queue workers
	log.Println("Waiting for all async job workers to be cleared ...")
	JobWorkersLock.Lock()
	if len(JobWorkers) > 0 {
		for group, worker := range JobWorkers {
			close(worker.jobQueue)
			log.Printf("Clear %s\n", group)
		}
		JobWorkers = map[string]*JobWorker{}
		cleared = true
	}
	JobWorkersLock.Unlock()

	// numJobWorkersRunning.Done() as numJobWorkersRunning.Add(-1)
	// wait for all job workers to quit
	numJobWorkersRunning.Wait()
	return cleared
}
