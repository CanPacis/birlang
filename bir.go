package main

import (
	"bufio"
	"os"

	"github.com/canpacis/birlang/src/engine"
)

func main() {
	std_path := os.Getenv("BirStd")
	if std_path != "" {
		if len(os.Args) > 1 {
			instance := engine.NewEngine(os.Args[1], std_path, false, false, 0)
			instance.Init()
			instance.Run()

			// v, _ := json.MarshalIndent(instance.GetCurrentScope().Frame, "", "  ")
			// fmt.Println(string(v))

			// v, _ := json.MarshalIndent(instance.GetCurrentScope().Blocks[1], "", "  ")
			// fmt.Println(string(v))

			// for _, block := range instance.GetCurrentScope().Blocks {
			// 	fmt.Printf("%+v\n", block.Instance)
			// }

		} else {
			repl_caret := "> "
			instance := engine.NewEngine("", std_path, true, false, 1)
			instance.Init()
			scanner := bufio.NewScanner(os.Stdin)
			os.Stdout.WriteString("Bir v0.1.1\n")
			os.Stdout.WriteString("Exit using ctrl+c\n")
			os.Stdout.WriteString(repl_caret)

			for scanner.Scan() {
				result := instance.Feed(scanner.Text())
				os.Stdout.WriteString(result + "\n")
				os.Stdout.WriteString(repl_caret)
			}
		}
	} else {
		os.Stdout.WriteString("Could not find bir standard path (BirStd) in your environment variables")
	}
}
