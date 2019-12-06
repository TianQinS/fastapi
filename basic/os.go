package basic

import (
	"bytes"
	"log"
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

func Exec(cmd string) ([]byte, error) {
	var err error
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd1 := "/bin/sh"
	cmd2 := "-c"
	switch runtime.GOOS {
	case "windows":
		{
			cmd1 = "cmd"
			cmd2 = "/C"
		}
	case "freebsd":
		{
			// cmd1 = "/bin/csh"
		}
	}
	p := exec.Command(cmd1, cmd2, cmd)
	// out, err = p.Output()

	p.Stdout = &out
	p.Stderr = &stderr
	if err = p.Run(); err != nil {
		log.Println(stderr.String())
		return stderr.Bytes(), err
	}
	return out.Bytes(), err
}
