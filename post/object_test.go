package post

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	o         RpcObject
	benchTest bool
)

func init() {
	benchTest = false
	o = RpcObject{}
	o.Init(1024)
	o.Register("test1", func(d *int, str string) {
		*d = 1
	})
	if benchTest {
		return
	}
	go func() {
		for o.IsRun {
			n := o.ExecuteEvent()
			n = 10 - n
			if n > 0 {
				time.Sleep(time.Duration(n) * time.Millisecond)
			} else {
				runtime.Gosched()
			}
		}
	}()
}

func func1(args ...interface{}) {
	d := args[0].(*int)
	_ = args[1].(string)
	*d = 1
}

func func2(d *int, str string) {
	*d = 1
}

func TestObject(t *testing.T) {
	if benchTest {
		return
	}
	d1, d2 := 0, 0
	err := o.PutQueue("test1", false, &d1, "test")
	assert.Equal(t, err, nil)
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 1, d1)
	err = o.PutQueue(func1, true, &d2, "test")
	assert.Equal(t, err, nil)
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 1, d2)
}

func BenchmarkTest1(b *testing.B) {
	d := 0
	for i := 0; i < b.N; i++ {
		o.PutQueue(func1, true, &d, "test")
		o.ExecuteEvent()
	}
}

func BenchmarkTest2(b *testing.B) {
	d := 0
	for i := 0; i < b.N; i++ {
		o.PutQueue(func2, false, &d, "test")
		o.ExecuteEvent()
	}
}
