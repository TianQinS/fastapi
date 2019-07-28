package basic

import (
	"os"
	"os/exec"
	"path"
	"runtime"
)

// Create a file handler for log without close.
func NewFile(filePath string) (*os.File, error) {
	dir := path.Dir(filePath)
	os.Mkdir(dir, 0777)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func Exec(cmd string) (out []byte, err error) {
	cmd1 := "/bin/sh"
	cmd2 := "-c"
	if runtime.GOOS == "windows" {
		cmd1 = "cmd"
		cmd2 = "/C"
	}
	p := exec.Command(cmd1, cmd2, cmd)
	out, err = p.Output()
	return
}
