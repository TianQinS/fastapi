package basic

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHooks(t *testing.T) {
	// Easy use.
	a := false
	startTime := time.Now().Format("2006-01-02 15:04:05")
	endTime := time.Now().Add(1000 * time.Millisecond).Format("2006-01-02 15:04:05")
	HookMgr.Register("test1", startTime, endTime, func(args ...interface{}) {
		fmt.Println(args)
		a = args[0].(bool)
	})
	HookMgr.Fire("test1", true)
	assert.Equal(t, true, a)
	HookMgr.Fire("test1", false)
	assert.Equal(t, false, a)
	time.Sleep(1000 * time.Millisecond)
	HookMgr.Fire("test1", true)
	assert.Equal(t, false, a)
}
