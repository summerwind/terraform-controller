package exec

import (
	"bytes"
	"log"
	"os"
	"os/exec"
)

type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

type Command struct {
	Name       string
	Path       string
	WorkingDir string
	Debug      bool
}

func (c *Command) Run(args ...string) (*Result, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	cmd := exec.Command(c.Path, args...)
	cmd.Dir = c.WorkingDir
	cmd.Env = os.Environ()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if c.Debug {
		log.Printf("[%s] %s %v", c.Name, c.Path, args)
		log.Printf("[%s] exit code: %d", c.Name, cmd.ProcessState.ExitCode())
		log.Printf("[%s] stderr: %s", c.Name, stderr.Bytes())
		log.Printf("[%s] stdout: %s", c.Name, stdout.Bytes())

		if err != nil {
			log.Printf("[%s] error: %v", c.Name, err)
		}
	}

	return &Result{
		ExitCode: cmd.ProcessState.ExitCode(),
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
	}, err
}
