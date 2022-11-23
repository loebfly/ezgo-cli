package main

import (
	"ezgo-cli/config"
	"ezgo-cli/let"
	"ezgo-cli/run"
	"flag"
	"fmt"
)

func main() {
	fmt.Printf("欢迎使用ezgo-cli\n")

	config.Init()

	cmd := flag.String("app", "", "运行命令")
	flag.Parse()
	switch *cmd {
	case let.CmdRun:
		run.Exec()
	default:
		fmt.Println("Invalid command", cmd)
	}
}
