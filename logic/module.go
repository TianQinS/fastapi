package logic

import (
	"github.com/sbinet/go-python"
)

type Module struct {
	Meths []python.PyMethodDef
	Mod   *python.PyObject
	Name  string
}

func (this *Module) Register(name, doc string, flag python.MethodDefFlags, meth func(self, args *python.PyObject) *python.PyObject) {
	this.Meths = append(this.Meths, python.PyMethodDef{
		Name:  name,
		Meth:  meth,
		Flags: flag,
		Doc:   doc,
	})
}

func (this *Module) Update(name string) (*python.PyObject, error) {
	if mod, err := python.Py_InitModule(name, this.Meths); err == nil {
		this.Mod = mod
		this.Name = name
		return mod, nil
	} else {
		this.Name = ""
		return nil, err
	}
}

func (this *Module) SetObjectAttr(name string, value *python.PyObject) int {
	return this.Mod.SetAttrString(name, value)
}

func (this *Module) SetClassAttr(name string, value *python.PyObject) int {
	key := python.PyString_FromString(name)
	return this.Mod.GenericSetAttr(key, value)
}

func (this *Module) IsValid() bool {
	return this.Name != ""
}
