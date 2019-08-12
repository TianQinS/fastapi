// modified by esQueue.
package basic

import (
	"fmt"
	"math"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	OVERFLOW_CHECK_NUM = uint64(2 << 60)
	UINT64_MAX_NUM     = math.MaxUint64 // ^uint64(0)
)

type esCache struct {
	putNo uint64
	getNo uint64
	value interface{}
}

// lock free queue
type EsQueue struct {
	capacity uint64
	capMod   uint64
	putPos   uint64
	getPos   uint64
	cache    []esCache
}

func NewQueue(capacity uint64) *EsQueue {
	if capacity > OVERFLOW_CHECK_NUM {
		panic("Capacity overflow.")
	}
	q := new(EsQueue)
	q.capacity = minQuantity(capacity)
	q.capMod = q.capacity - 1
	q.putPos = 0
	q.getPos = 0
	q.cache = make([]esCache, q.capacity)
	for i := range q.cache {
		cache := &q.cache[i]
		cache.getNo = uint64(i)
		cache.putNo = uint64(i)
	}
	cache := &q.cache[0]
	cache.getNo = q.capacity
	cache.putNo = q.capacity
	return q
}

func (q *EsQueue) String() string {
	getPos := atomic.LoadUint64(&q.getPos)
	putPos := atomic.LoadUint64(&q.putPos)
	return fmt.Sprintf("Queue{capacity: %v, capMod: %v, putPos: %v, getPos: %v}",
		q.capacity, q.capMod, putPos, getPos)
}

func (q *EsQueue) Capacity() uint64 {
	return q.capacity
}

func (q *EsQueue) Quantity() uint64 {
	var putPos, getPos uint64
	var quantity uint64
	getPos = atomic.LoadUint64(&q.getPos)
	putPos = atomic.LoadUint64(&q.putPos)

	if putPos >= getPos {
		quantity = putPos - getPos
	} else if getPos-putPos > OVERFLOW_CHECK_NUM {
		quantity = UINT64_MAX_NUM + (putPos - getPos) + 1
	} else {
		quantity = 0
	}

	return quantity
}

// put queue functions
func (q *EsQueue) Put(val interface{}) (ok bool, quantity uint64) {
	var putPos, putPosNew, getPos, posCnt uint64
	var cache *esCache
	capMod := q.capMod

	getPos = atomic.LoadUint64(&q.getPos)
	putPos = atomic.LoadUint64(&q.putPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else if getPos-putPos > OVERFLOW_CHECK_NUM {
		posCnt = UINT64_MAX_NUM + (putPos - getPos) + 1
	} else {
		posCnt = 0
	}

	if posCnt >= capMod-1 {
		runtime.Gosched()
		return false, posCnt
	}

	putPosNew = putPos + 1
	if !atomic.CompareAndSwapUint64(&q.putPos, putPos, putPosNew) {
		runtime.Gosched()
		return false, posCnt
	}

	cache = &q.cache[putPosNew&capMod]

	for {
		getNo := atomic.LoadUint64(&cache.getNo)
		putNo := atomic.LoadUint64(&cache.putNo)
		if putPosNew == putNo && getNo == putNo {
			cache.value = val
			atomic.AddUint64(&cache.putNo, q.capacity)
			return true, posCnt + 1
		} else {
			// runtime.Gosched()
			time.Sleep(time.Millisecond)
		}
	}
}

// puts queue functions
func (q *EsQueue) Puts(values []interface{}) (puts, quantity uint64) {
	var putPos, putPosNew, getPos, posCnt, putCnt uint64
	capMod := q.capMod

	getPos = atomic.LoadUint64(&q.getPos)
	putPos = atomic.LoadUint64(&q.putPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else if getPos-putPos > OVERFLOW_CHECK_NUM {
		posCnt = UINT64_MAX_NUM + (putPos - getPos) + 1
	} else {
		posCnt = 0
	}

	if posCnt >= capMod-1 {
		runtime.Gosched()
		return 0, posCnt
	}

	if capPuts, size := q.capacity-posCnt, uint64(len(values)); capPuts >= size {
		putCnt = size
	} else {
		putCnt = capPuts
	}
	putPosNew = putPos + putCnt

	if !atomic.CompareAndSwapUint64(&q.putPos, putPos, putPosNew) {
		runtime.Gosched()
		return 0, posCnt
	}

	for posNew, v := putPos+1, uint64(0); v < putCnt; posNew, v = posNew+1, v+1 {
		var cache *esCache = &q.cache[posNew&capMod]
		for {
			getNo := atomic.LoadUint64(&cache.getNo)
			putNo := atomic.LoadUint64(&cache.putNo)
			if posNew == putNo && getNo == putNo {
				cache.value = values[v]
				atomic.AddUint64(&cache.putNo, q.capacity)
				break
			} else {
				runtime.Gosched()
			}
		}
	}
	return putCnt, posCnt + putCnt
}

// get queue functions
func (q *EsQueue) Get() (val interface{}, ok bool, quantity uint64) {
	var putPos, getPos, getPosNew, posCnt uint64
	var cache *esCache
	capMod := q.capMod

	putPos = atomic.LoadUint64(&q.putPos)
	getPos = atomic.LoadUint64(&q.getPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else if getPos-putPos > OVERFLOW_CHECK_NUM {
		posCnt = UINT64_MAX_NUM + (putPos - getPos) + 1
	} else {
		posCnt = 0
	}

	if posCnt < 1 {
		runtime.Gosched()
		return nil, false, posCnt
	}

	getPosNew = getPos + 1
	if !atomic.CompareAndSwapUint64(&q.getPos, getPos, getPosNew) {
		runtime.Gosched()
		return nil, false, posCnt
	}

	cache = &q.cache[getPosNew&capMod]

	for {
		getNo := atomic.LoadUint64(&cache.getNo)
		putNo := atomic.LoadUint64(&cache.putNo)
		if getPosNew == getNo && getNo == putNo-q.capacity {
			val = cache.value
			cache.value = nil
			atomic.AddUint64(&cache.getNo, q.capacity)
			return val, true, posCnt - 1
		} else {
			// runtime.Gosched()
			time.Sleep(time.Millisecond)
		}
	}
}

// gets queue functions
func (q *EsQueue) Gets(values []interface{}) (gets, quantity uint64) {
	var putPos, getPos, getPosNew, posCnt, getCnt uint64
	capMod := q.capMod

	putPos = atomic.LoadUint64(&q.putPos)
	getPos = atomic.LoadUint64(&q.getPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else if getPos-putPos > OVERFLOW_CHECK_NUM {
		posCnt = UINT64_MAX_NUM + (putPos - getPos) + 1
	} else {
		posCnt = 0
	}

	if posCnt < 1 {
		runtime.Gosched()
		return 0, posCnt
	}

	if size := uint64(len(values)); posCnt >= size {
		getCnt = size
	} else {
		getCnt = posCnt
	}
	getPosNew = getPos + getCnt

	if !atomic.CompareAndSwapUint64(&q.getPos, getPos, getPosNew) {
		runtime.Gosched()
		return 0, posCnt
	}

	for posNew, v := getPos+1, uint64(0); v < getCnt; posNew, v = posNew+1, v+1 {
		var cache *esCache = &q.cache[posNew&capMod]
		for {
			getNo := atomic.LoadUint64(&cache.getNo)
			putNo := atomic.LoadUint64(&cache.putNo)
			if posNew == getNo && getNo == putNo-q.capacity {
				values[v] = cache.value
				cache.value = nil
				getNo = atomic.AddUint64(&cache.getNo, q.capacity)
				break
			} else {
				runtime.Gosched()
			}
		}
	}

	return getCnt, posCnt - getCnt
}

// round 到最近的2的倍数
func minQuantity(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	v++
	return v
}
