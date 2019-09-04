package hotfix

import (
	"fmt"
	"go/importer"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

type HotFix struct {
	prefix     string
	outDir     string
	outExclude string
	mods       map[string]bool
}

// NewHotFix make a HotFix object that contains relevant information of hotfix.
func NewHotFix(stdlibs, prefix string, mods ...string) *HotFix {
	hot := &HotFix{}
	hot.Init(stdlibs, prefix, mods...)
	return hot
}

// addPkg only filter required dependencies what is necessary.
func (this *HotFix) addPkg(pkgName string) {
	if pkgName == "" {
		return
	}
	if pkgName == this.outExclude {
		return
	}
	if strings.Contains(pkgName, "internal/") {
		return
	}
	if _, ok := this.mods[pkgName]; ok {
		return
	}
	fmt.Println(pkgName)
	if strings.HasPrefix(pkgName, this.prefix) {
		this.mods[pkgName] = true
		this.parsePkg(pkgName)
	} else {
		this.mods[pkgName] = false
	}
}

// parsePkg parse the dependencies of `pkgName` module.
func (this *HotFix) parsePkg(pkgName string) error {
	p, err := importer.For("source", nil).Import(pkgName)
	if err != nil {
		p, err = importer.For("gc", nil).Import(pkgName)
		if err != nil {
			// basic.PackErrorMsg(err, pkgName)
			fmt.Println(err)
			return err
		}
	}
	for _, pkg := range p.Imports() {
		this.addPkg(pkg.Path())
	}
	return nil
}

// loadPkgs update the local Symbol files.
func (this *HotFix) loadPkgs() {
	mods := make([]string, 0)
	for mod, _ := range this.mods {
		mods = append(mods, mod)
	}
	// fmt.Println(this.prefix, this.outDir, mods)
	Parse(this.outDir, mods...)
}

// Init create a stdlib directory in you `outputDir` and make Symbol files for dependent packages,
// these Symbol files will be updated when this function is called.
// Modules in mods with the prefix `modPrefix` will be parsed(dependency chain) with recursion.
func (this *HotFix) Init(outputDir, modPrefix string, mods ...string) {
	this.prefix = modPrefix
	this.outDir = path.Join(strings.Replace(os.Getenv("GOPATH"), "\\", "/", -1), "src", outputDir, "stdlibs")
	this.outExclude = path.Join(outputDir, "stdlibs")
	this.mods = make(map[string]bool, len(mods))
	// os.RemoveAll(this.outDir)
	if _, err := os.Stat(this.outDir); err != nil {
		os.Mkdir(this.outDir, os.ModePerm)
	}
	for _, mod := range mods {
		this.addPkg(mod)
	}
	err := ioutil.WriteFile(path.Join(this.outDir, "stdlibs.go"), []byte(StdContent), 0666)
	if err != nil {
		log.Fatal(err)
	}
	this.loadPkgs()
}
