// Job worker with goroutine, suitable for inactive calls.
package post

import (
	"fmt"
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
				if e, ok := recover().(error); ok {
					basic.PackErrorMsg(e, msg)
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
				_f.Call(in)
			}
		}
		_runFunc()
	}
}

func (this *JobWorker) appendJob(f interface{}, strictUnReflect bool, params []interface{}) {
	this.jobQueue <- QueueMsg{f, params, strictUnReflect}
}

func Close() bool {
	var cleared bool
	// Close all job queue workers
	fmt.Println("Waiting for all async job workers to be cleared ...")
	JobWorkersLock.Lock()
	if len(JobWorkers) > 0 {
		for group, worker := range JobWorkers {
			close(worker.jobQueue)
			fmt.Printf("\tclear %s\n", group)
		}
		JobWorkers = map[string]*JobWorker{}
		cleared = true
	}
	JobWorkersLock.Unlock()

	// wait for all job workers to quit
	numJobWorkersRunning.Wait()
	return cleared
}
