// Package basic contains only the most basic functions.
package basic

import (
	"fmt"
	"reflect"
	"runtime/debug"
)

// The callback function.
type FuncCallback func(args ...interface{})

// The post function.
type Func func(args ...interface{}) (res interface{})

// Check and throw the error.
func Throw(err error) {
	if err != nil {
		panic(err)
	}
}

// Catch the error throw by the function if it paniced.
func Catch(f func()) (err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()

	f()
	return
}

// Pack runtime error msg for log.
func PackErrorMsg(err error, args interface{}) map[string]interface{} {
	msg := make(map[string]interface{}, 3)
	msg["err"] = err.Error()
	msg["args"] = args
	msg["trace"] = string(debug.Stack())
	// mail.SendMsg(msg)
	fmt.Println(msg)
	return msg
}

// Catch the error throw by the function with interface arguments.
func CatchWithParams(f FuncCallback, args ...interface{}) (err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
			PackErrorMsg(e, args)
		}
	}()

	f(args...)
	return
}

func CatchWithReflect(f interface{}, args ...interface{}) (err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
			PackErrorMsg(e, args)
		}
	}()
	_f := reflect.ValueOf(f)
	in := make([]reflect.Value, len(args))
	for k := range in {
		in[k] = reflect.ValueOf(args[k])
	}
	_f.Call(in)
	return
}

// Catch Func for post worker.
func CatchFunc(f Func, args ...interface{}) (err error, res interface{}) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
			res = PackErrorMsg(e, args)
		}
	}()

	res = f(args...)
	return
}
