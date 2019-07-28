package timer

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimer(t *testing.T) {
	INTERVAL := 100 * time.Millisecond
	x := 0
	px := x
	now := time.Now()
	nextTime := now.Add(INTERVAL)
	fmt.Printf("now is %s, next time should be %s\n", time.Now(), nextTime)

	AddTimer(INTERVAL, func(d *int) {
		*d += 1
		fmt.Printf("timer %s x %v px %v\n", time.Now(), d, px)
	}, &x)

	for i := 0; i < 10; i++ {
		time.Sleep(nextTime.Add(INTERVAL / 2).Sub(time.Now()))
		fmt.Printf("Check x %v px %v @ %s\n", x, px, time.Now())
		assert.Equal(t, x, px+1)
		px = x
		nextTime = nextTime.Add(INTERVAL)
		fmt.Printf("now is %s, next time should be %s\n", time.Now(), nextTime)
	}
}

func TestCallbackSeq(t *testing.T) {
	a := 0
	d := time.Second

	for i := 0; i < 100; i++ {
		d0 := i
		AddCallback(d, func(d1, d2 *int) {
			assert.Equal(t, *d1, *d2)
			*d2 += 1
		}, &d0, &a)
	}
	time.Sleep(d + time.Second*1)
}

func TestCancelCallback(t *testing.T) {
	INTERVAL := 20 * time.Millisecond
	x := 0

	timer := AddCallback(INTERVAL, func(d *int) {
		*d = 1
	}, &x)
	assert.Equal(t, timer.IsActive(), true)
	timer.Cancel()
	assert.Equal(t, timer.IsActive(), false)
	time.Sleep(INTERVAL * 2)
	assert.Equal(t, 0, x)
}

func TestCancelTimer(t *testing.T) {
	INTERVAL := 20 * time.Millisecond
	x := 0
	timer := AddTimer(INTERVAL, func(d *int) {
		*d += 1
	}, &x)
	assert.Equal(t, timer.IsActive(), true)
	timer.Cancel()
	assert.Equal(t, timer.IsActive(), false)
	time.Sleep(INTERVAL * 2)
	assert.Equal(t, 0, x)
}

func TestTimerPerformance(t *testing.T) {
	f, err := os.Create("TestTimerPerformance.cpuprof")
	if err != nil {
		panic(err)
	}

	pprof.StartCPUProfile(f)
	duration := 10 * time.Second

	for i := 0; i < 40000; i++ {
		if rand.Float32() < 0.5 {
			d := time.Duration(rand.Int63n(int64(duration)))
			AddCallback(d, func() {})
		} else {
			d := time.Duration(rand.Int63n(int64(time.Second)))
			AddTimer(d, func() {})
		}
	}

	log.Println("Waiting for", duration, "...")
	time.Sleep(duration)
	pprof.StopCPUProfile()
}
