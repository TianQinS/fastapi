package post

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	a := 0
	GPost.PutJobStrict("testGroup", func(args ...interface{}) {
		d := args[0].(*int)
		*d = 1
		fmt.Println(args, d, *d)
	}, &a)
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, 1, a)
	GPost.PutJob("testGroup", func(d *int) {
		*d = 0
		fmt.Println(d, *d)
	}, &a)
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, 0, a)
	Close()
}
