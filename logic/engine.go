package logic

import (
	"errors"
	"fmt"
	"log"

	"github.com/sbinet/go-python"
)

func init() {
	err := python.Initialize()
	if err != nil {
		panic(err.Error())
	}
}

func ParseGo(val interface{}) *python.PyObject {
	switch val.(type) {
	case int:
		return python.PyInt_FromLong(val.(int))
	case string:
		return python.PyString_FromString(val.(string))
	case int64:
		return python.PyLong_FromLongLong(val.(int64))
	case int32:
		return python.PyLong_FromLongLong(int64(val.(int32)))
	case uint:
		return python.PyLong_FromUnsignedLong(val.(uint))
	case uint64:
		return python.PyLong_FromUnsignedLongLong(val.(uint64))
	case uint32:
		return python.PyLong_FromUnsignedLongLong(uint64(val.(uint32)))
	case float64:
		return python.PyLong_FromDouble(val.(float64))
	case float32:
		return python.PyLong_FromDouble(float64(val.(float32)))
	case nil:
		return python.Py_None
	}
	return val.(*python.PyObject)
}

func FetchErr() string {
	exc, val, tb := python.PyErr_Fetch()
	return fmt.Sprintf("Exception,exc=%v\nval=%v\ntraceback=%v\n", exc, val, tb)
}

func GetFunc(module, function string) (*python.PyObject, error) {
	m := python.PyImport_ImportModule("sys")
	if m == nil {
		return nil, errors.New("Import module sys error")
	}
	path := m.GetAttrString("path")
	if path == nil {
		return nil, errors.New("Get sys path error")
	}
	// append current diretcory
	currentDir := python.PyString_FromString("")
	python.PyList_Insert(path, 0, currentDir)
	m = python.PyImport_ImportModule(module)
	if m == nil {
		return nil, fmt.Errorf("Import module %s error", module)
	}
	f := m.GetAttrString(function)
	if f == nil {
		return nil, fmt.Errorf("Module %s function %s null", module, function)
	}
	return f, nil
}

func ParsePy(dat *python.PyObject) (ret interface{}, err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()
	ret = dat.Type().String()
	switch ret.(type) {
	case string:
		log.Println(ret)
	}
	return
}

func CallFunc(function *python.PyObject, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	tArgs := python.PyTuple_New(len(args))
	for i, arg := range args {
		python.PyTuple_SetItem(tArgs, i, ParseGo(arg))
	}
	tKwargs := python.PyDict_New()
	for key, val := range kwargs {
		if err := python.PyDict_SetItem(
			tKwargs,
			python.PyString_FromString(key),
			ParseGo(val),
		); err != nil {
			return nil, err
		}
	}
	out := function.Call(tArgs, tKwargs)
	return ParsePy(out)
}
