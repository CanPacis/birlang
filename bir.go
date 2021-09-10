package main

import (
	"bufio"
	"os"

	"github.com/canpacis/birlang/src/config"
	"github.com/canpacis/birlang/src/engine"
)

func main() {
	if len(os.Args) > 1 {
		instance := engine.NewEngine(os.Args[1], false, false, 0)
		instance.Init()
		config.HandleConfig(&instance)
		instance.Run()

		// v, _ := json.MarshalIndent(instance.GetCurrentScope().Frame, "", "  ")
		// fmt.Println(string(v))
		// for _, block := range instance.GetCurrentScope().Blocks {
		// 	fmt.Printf("%+v\n", block.Instance)
		// }

	} else {
		repl_caret := "> "
		instance := engine.NewEngine("", true, false, 1)
		instance.Init()
		scanner := bufio.NewScanner(os.Stdin)
		os.Stdout.WriteString("Bir v0.1.0\n")
		os.Stdout.WriteString("Exit using ctrl+c\n")
		os.Stdout.WriteString(repl_caret)

		for scanner.Scan() {
			result := instance.Feed(scanner.Text())
			os.Stdout.WriteString(result + "\n")
			os.Stdout.WriteString(repl_caret)
		}
	}
}
