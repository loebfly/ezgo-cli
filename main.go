package main

import (
	"ezgo-cli/new"
	"ezgo-cli/run"
	"fmt"
	"os"
)

const (
	CommandRun = "run" // 运行项目
	CommandNew = "new" // 初始化项目
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}
	switch os.Args[1] {
	case CommandNew:
		new.Exec()
	case CommandRun:
		run.Exec()
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Usage: ezgo-cli [COMMAND] [options]")
	fmt.Println("Commands:")
	fmt.Println("  new: 生成基于ezgin脚手架的模板项目")
	fmt.Println("  run: 运行项目")
	fmt.Println("Run 'ezgo-cli [COMMAND] -help' get options for command")
}
