package main

import (
	"github.com/TianQinS/fastapi/hotfix"
	"github.com/TianQinS/fastapi/hotfix/stdlibs"
	"github.com/containous/yaegi/interp"
)

const src = `package foo

import( 
	"fmt"
	"github.com/TianQinS/fastapi/basic"
)

func Exec(cmd string) string {
	fmt.Println(cmd)
	out, err := basic.Exec(cmd)
	if err == nil {
		return string(out)
	}
	return err.Error()
}`

func init() {
	hotfix.NewHotFix(
		"github.com/TianQinS/fastapi/hotfix",
		"github.com/TianQinS/fastapi/basic",
		"github.com/TianQinS/fastapi/basic",
	)
}

func main() {
	i := interp.New(interp.Options{})
	symbols := stdlibs.Symbols
	i.Use(symbols)

	_, err := i.Eval(src)
	if err != nil {
		panic(err)
	}

	v, err := i.Eval("foo.Exec")
	if err != nil {
		panic(err)
	}

	bar := v.Interface().(func(string) string)

	r := bar("ls")
	println(r)
}
