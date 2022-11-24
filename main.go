package main

import (
	"ezgo-cli/run"
	"fmt"
	"os"
)

const (
	CommandRun = "run"
)

func main() {
	fmt.Println("ezgo-cli v1.0.0")
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}
	switch os.Args[1] {
	case CommandRun:
		run.Exec()
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Usage: ezgo-cli [COMMAND] [options]")
	fmt.Println("Commands:")
	fmt.Println("  run: 运行项目")
	fmt.Println("Run 'ezgo-cli [COMMAND] -help' get options for command")
}
