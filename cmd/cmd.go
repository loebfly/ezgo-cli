package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
)

// ExecInDir 在指定目录下执行命令
func ExecInDir(dir string, cmd string, args ...string) (*exec.Cmd, error) {
	command := exec.Command(cmd, args...)
	stdout := &bytes.Buffer{}
	command.Stdout = stdout
	stderr := &bytes.Buffer{}
	command.Stderr = stderr
	command.Dir = dir
	err := command.Start()
	if err != nil {
		return command, err
	}

	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println(line)
	}

	err = command.Wait()
	return command, err
}

// ExecWithPreCmd 执行命令并将上一个命令的输出作为输入
func ExecWithPreCmd(preCmd *exec.Cmd, cmd string, args ...string) (*exec.Cmd, error) {
	command := exec.Command(cmd, args...)
	command.Stdin = preCmd.Stdout.(*bytes.Buffer)
	stdout := &bytes.Buffer{}
	command.Stdout = stdout
	stderr := &bytes.Buffer{}
	command.Stderr = stderr
	err := command.Start()
	if err != nil {
		return command, err
	}
	err = command.Wait()
	return command, err
}
