package main

import (
	"ezgo-cli/run"
	"flag"
	"fmt"
)

const (
	CmdRun = "run"
)

func main() {
	cmd := flag.String("c", "", "运行命令")
	flag.Parse()
	switch *cmd {
	case CmdRun:
		run.Start()
	default:
		fmt.Println("Invalid command", cmd)
	}
}
