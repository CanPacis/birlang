package main

import (
	"bufio"
	"os"

	"github.com/canpacis/birlang/src/engine"
)

func main() {
	if len(os.Args) > 1 {
		instance := engine.BirEngine{Path: os.Args[1]}
		instance.Init()
		instance.Run()

		// fmt.Printf("%+v", instance.Uses[0].GetCurrentScope())
	} else {
		repl_caret := "[BIR] "
		instance := engine.BirEngine{Anonymous: true}
		instance.Init()
		scanner := bufio.NewScanner(os.Stdin)
		os.Stdout.WriteString(repl_caret)

		for scanner.Scan() {
			result := instance.Feed(scanner.Text())
			os.Stdout.WriteString(result + "\n")
			os.Stdout.WriteString(repl_caret)
		}
	}
}
