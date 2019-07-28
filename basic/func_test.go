package basic

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCatch(t *testing.T) {
	e1 := Catch(func() {
		panic(fmt.Errorf("bad1"))
	})
	assert.Equal(t, e1.Error(), "bad1")

	e2 := CatchWithParams(func(args ...interface{}) {
		Throw(fmt.Errorf(args[0].(string)))
	}, "bad2")
	assert.Equal(t, e2.Error(), "bad2")

	e3, _ := CatchFunc(func(args ...interface{}) (res interface{}) {
		Throw(fmt.Errorf(args[0].(string)))
		return nil
	}, "bad3")
	assert.Equal(t, e3.Error(), "bad3")
}
