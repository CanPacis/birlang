package thrower

import (
	"os"
	"strconv"
	"strings"

	"github.com/canpacis/birlang/src/ast"
	"github.com/canpacis/birlang/src/util"
	"github.com/mitchellh/mapstructure"
)

type Thrower struct {
	Owner interface{}
	Color util.Color
}

func (thrower *Thrower) Throw(message string, position ast.Position) {
	var engine map[string]interface{}
	mapstructure.Decode(thrower.Owner, &engine)

	if !engine["Anonymous"].(bool) {
		os.Stdout.WriteString(thrower.Color.OutputRed("[ERROR]") + " " + message + " at " + thrower.Color.OutputCyan(strconv.Itoa(int(position.Line))+":"+strconv.Itoa(int(position.Col))) + " in " + thrower.Color.OutputYellow(engine["Filename"].(string)) + "\n")
		os.Stdout.WriteString("\n" + thrower.GetSnippet(position) + "\n")
		os.Stdout.WriteString("\nCallstack:\n\t" + thrower.GetCallstack() + "\n")
		os.Stdout.WriteString("\nFile:\n\t" + thrower.Color.OutputRed(engine["URI"].(string)) + "\n")
		os.Exit(1)
	} else {
		os.Stdout.WriteString(thrower.Color.OutputRed("[ERROR]") + " " + message + " in " + thrower.Color.OutputYellow("[REPL]") + "\n")
	}
}

func (thrower *Thrower) ThrowAnonymous(message string) {
	var engine map[string]interface{}
	mapstructure.Decode(thrower.Owner, &engine)

	os.Stdout.WriteString(thrower.Color.OutputRed("[ERROR]") + " " + message + "\n")

	if !engine["Anonymous"].(bool) {
		os.Exit(1)
	}
}

func (thrower *Thrower) GetSnippet(position ast.Position) string {
	var engine map[string]interface{}
	mapstructure.Decode(thrower.Owner, &engine)
	lines := strings.Split(engine["Content"].(string), "\n")

	dummy := make([]string, position.Col)
	result := lines[position.Line-1] + "\n" + strings.Join(dummy, " ") + "^"
	return result
}

func (thrower *Thrower) GetCallstack() string {
	var engine map[string]interface{}
	mapstructure.Decode(thrower.Owner, &engine)

	var callstack []map[string]interface{}
	mapstructure.Decode(engine["Callstack"], &callstack)

	result := []string{}
	for _, stack := range callstack {
		result = append(result, thrower.Color.OutputCyan(stack["Label"].(string))+thrower.Color.OutputGrey(" ()"))
	}

	return strings.Join(result, "\n\t")
}

func (thrower *Thrower) Warn(message string, position ast.Position) {
	var engine map[string]interface{}
	mapstructure.Decode(thrower.Owner, &engine)

	if engine["VerbosityLevel"].(int) == 1 || engine["VerbosityLevel"].(int) == 2 {
		if !engine["Anonymous"].(bool) {
			os.Stdout.WriteString(thrower.Color.OutputYellow("[WARNING]") + " " + message + " at " + thrower.Color.OutputCyan(strconv.Itoa(int(position.Line))+":"+strconv.Itoa(int(position.Col))) + " in " + thrower.Color.OutputYellow(engine["Filename"].(string)) + "\n")
			if engine["VerbosityLevel"].(int) == 2 {
				os.Stdout.WriteString("\n" + thrower.GetSnippet(position) + "\n")
				os.Stdout.WriteString("\nCallstack:\n\t" + thrower.GetCallstack() + "\n")
				os.Stdout.WriteString("\nFile:\n\t" + thrower.Color.OutputRed(engine["URI"].(string)) + "\n")
			}
		} else {
			os.Stdout.WriteString(thrower.Color.OutputYellow("[WARNING]") + " " + message + " in " + thrower.Color.OutputYellow("[REPL]") + "\n")
		}
	}
}
