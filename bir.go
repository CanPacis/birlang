package main

import (
	"bufio"
	"os"

	"github.com/canpacis/birlang/src/engine"
)

func main() {
	if len(os.Args) > 1 {
		instance := engine.NewEngine(os.Args[1], false, false, 0)
		instance.Run()

		// v, _ := json.MarshalIndent(instance.GetCurrentScope(), "", "  ")
		// fmt.Println(string(v))
	} else {
		repl_caret := "[BIR] "
		instance := engine.NewEngine("", true, false, 0)
		scanner := bufio.NewScanner(os.Stdin)
		os.Stdout.WriteString(repl_caret)

		for scanner.Scan() {
			result := instance.Feed(scanner.Text())
			os.Stdout.WriteString(result + "\n")
			os.Stdout.WriteString(repl_caret)
		}
	}
}
