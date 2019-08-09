package post

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	a := 0
	p := NewPost(uint64(1024), 2)
	fmt.Println(&a)
	p.PutQueueStrict(func(args ...interface{}) {
		d := args[0].(*int)
		*d = 1
		fmt.Println(args, d, *d)
	}, &a)
	time.Sleep(15 * time.Millisecond)
	assert.Equal(t, 1, a)

	p.DelOne()
	p.PutQueueStrict(func(args ...interface{}) {
		d := args[0].(*int)
		*d = 0
		fmt.Println(args, d, *d)
	}, &a)
	time.Sleep(15 * time.Millisecond)
	assert.Equal(t, 0, a)

	p.DelOne()
	p.PutQueueStrict(func(args ...interface{}) {
		d := args[0].(*int)
		*d = 1
		fmt.Println(args, d, *d)
	}, &a)
	time.Sleep(15 * time.Millisecond)
	assert.Equal(t, 0, a)

	p.PutQueueSpec(func(d *int) {
		*d = 2
		fmt.Println(d, *d)
	}, &a)
	time.Sleep(15 * time.Millisecond)
	assert.Equal(t, 2, a)

	p.AddOne()
	p.PutQueue(func(d *int) {
		*d = 1
		fmt.Println(d, *d)
	}, &a)
	time.Sleep(15 * time.Millisecond)
	assert.Equal(t, 1, a)
	p.Close()
}
