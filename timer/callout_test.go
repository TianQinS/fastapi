package timer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCallOut(t *testing.T) {
	count := 0
	CallOut(3, func(d *int) {
		t.Logf("callout 3 seconds")
		*d += 1
	}, &count)
	time.Sleep(2 * time.Second)
	assert.Equal(t, 0, count)
	time.Sleep(2 * time.Second)
	assert.Equal(t, 1, count)
}

func TestCancelCallOut(t *testing.T) {
	count := 0
	item := CallOut(3, func(d *int) {
		t.Logf("callout 3 seconds")
		*d += 1
	}, &count)
	assert.Equal(t, 0, count)
	item.Cancel()
	time.Sleep(4 * time.Second)
	assert.Equal(t, 0, count)
}
