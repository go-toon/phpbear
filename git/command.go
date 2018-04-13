package git

import (
	"fmt"
	"strings"
	"time"
	"io"
	"os/exec"
	"os"
	"bytes"
)

type Command struct {
	name string
	args []string
	envs []string
}

func (c *Command) String() string {
	if len(c.args) == 0 {
		return c.name
	}

	return fmt.Sprintf("%s %s", c.name, strings.Join(c.args, " "))
}

// NewCommand creates and returns a new Git Command based on given command and args.
func NewCommand(args ...string) *Command {
	return &Command{
		name: "git",
		args: args,
	}
}

// NewCommand creates and returns a new Composer Command based on given command and args.
func NewComposerCommand(args ...string) *Command {
	return &Command{
		name: "composer",
		args: args,
	}
}

// AddArguments adds new argument(s) to the command.
func (c *Command) AddArguments(args ...string) *Command {
	c.args = append(c.args, args...)
	return c
}

// AddEnvs adds new environment variables to the command.
func (c *Command) AddEnvs(envs ...string) *Command {
	c.envs = append(c.envs, envs...)
	return c
}

const DEFAULT_TIMEOUT = 60 * time.Second

// RunInDirTimeoutPipeline executes the command in given directory with given timeout,
// it pipes stdout and stderr to given io.Writer.
func (c *Command) RunInDirTimeoutPipeline(timeout time.Duration, dir string, stdout, stderr io.Writer) error {
	if timeout == -1 {
		timeout = DEFAULT_TIMEOUT
	}

	cmd := exec.Command(c.name, c.args...)
	if c.envs != nil {
		cmd.Env = append(os.Environ(), c.envs...)
	}
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	var err error
	select {
	case <-time.After(timeout):
		if cmd.Process != nil && cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
			if err := cmd.Process.Kill(); err != nil {
				return fmt.Errorf("faild to kill process: %v", err)
			}
		}
		<-done
		return ErrExecTimeout{timeout}
	case err = <- done:
	}

	return err
}

// RunInDirTimeout executes the command in given directory with given timeout,
// and returns stdout in []byte and error (combined with stderr).
func (c *Command) RunInDirTimeout(timeout time.Duration, dir string) ([]byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if err := c.RunInDirTimeoutPipeline(timeout, dir, stdout, stderr); err != nil {
		return nil, fmt.Errorf("%v - %s", err, stderr)
	}

	return stdout.Bytes(), nil
}

// RunInDirPipeline executes the command in given directory,
// it pipes stdout and stderr to given io.Writer.
func (c *Command) RunInDirPipeline(dir string, stdout, stderr io.Writer) error {
	return c.RunInDirTimeoutPipeline(-1, dir, stdout, stderr)
}

// RunInDir executes the command in given directory
// and returns stdout in string and error (combined with stderr).
func (c *Command) RunInDir(dir string) (string, error) {
	stdout, err := c.RunInDirTimeout(-1, dir)
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

// RunTimeout executes the command in defualt working directory with given timeout,
// and returns stdout in string and error (combined with stderr).
func (c *Command) RunTimeout(timeout time.Duration) (string, error) {
	stdout, err := c.RunInDirTimeout(timeout, "")
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

// Run executes the command in defualt working directory
// and returns stdout in string and error (combined with stderr).
func (c *Command) Run() (string, error) {
	return c.RunTimeout(-1)
}

// AddChange marks local changes to be ready for commit.
func AddChanges(repoPath string, all bool, files ...string) error {
	cmd := NewCommand("add")
	if all {
		cmd.AddArguments("--all")
	}
	_, err := cmd.AddArguments(files...).RunInDir(repoPath)
	return err
}

// 提交commit
func CommitChanges(repoPath, message string) error {
	cmd := NewCommand("commit")

	if message == "" {
		message = "Init Project"
	}
	cmd.AddArguments("-m", message)
	_, err := cmd.RunInDir(repoPath)

	if err != nil && err.Error() == "exit status 1" {
		return nil
	}

	return err
}

// Push 代码
func Push(repoPath, remote, branch string) error {
	_, err := NewCommand("push", remote, branch).RunInDir(repoPath)
	return err
}