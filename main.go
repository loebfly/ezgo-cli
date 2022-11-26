package main

import (
	"ezgo-cli/cmd"
	"ezgo-cli/idea"
	"ezgo-cli/new"
	"ezgo-cli/run"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}
	switch os.Args[1] {
	case cmd.New:
		new.Exec()
	case cmd.Run:
		run.Exec()
	case cmd.Idea:
		if len(os.Args) < 3 {
			printHelp()
			return
		}
		idea.Exec()
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Usage: ezgo-cli [COMMAND] [options]")
	fmt.Println("Commands:")
	fmt.Println("  new: 生成基于ezgin脚手架的模板项目")
	fmt.Println("  run: 运行项目")
	fmt.Println("  idea: idea配置")
	fmt.Println("Run 'ezgo-cli [COMMAND] -help' get options for command")
}
