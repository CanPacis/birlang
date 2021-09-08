package thrower

import (
	"os"
	"strconv"
	"strings"

	"github.com/canpacis/birlang/src/ast"
	"github.com/mitchellh/mapstructure"
)

type Thrower struct {
	Owner interface{}
}

func (thrower *Thrower) Throw(message string, position ast.Position) {
	var engine map[string]interface{}
	mapstructure.Decode(thrower.Owner, &engine)

	color_cyan := ""
	color_yellow := ""
	color_red := ""

	if engine["ColoredOutput"].(bool) {
		color_cyan = "\033[0;36m"
		color_yellow = "\033[1;33m"
		color_red = "\033[1;31m"
	}

	if !engine["Anonymous"].(bool) {
		os.Stdout.WriteString(message + " at " + color_cyan + strconv.Itoa(int(position.Line)) + ":" + strconv.Itoa(int(position.Col)) + "\033[0m in " + color_yellow + engine["Filename"].(string) + "\033[0m\n")
		os.Stdout.WriteString("\n" + thrower.GetSnippet(position) + "\n")
		os.Stdout.WriteString("\nCallstack:\n\t" + thrower.GetCallstack() + "\n")
		os.Stdout.WriteString("\nFile:\n\t" + color_red + engine["URI"].(string) + "\033[0m\n")
		os.Exit(1)
	} else {
		os.Stdout.WriteString(message + " in " + color_yellow + "[REPL]" + "\033[0m\n")
	}
}

func (thrower *Thrower) ThrowAnonymous(message string) {
	var engine map[string]interface{}
	mapstructure.Decode(thrower.Owner, &engine)

	os.Stdout.WriteString(message + "\n")

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

	color_cyan := ""
	color_grey := ""

	if engine["ColoredOutput"].(bool) {
		color_cyan = "\033[0;36m"
		color_grey = "\033[1;30m"
	}

	var callstack []map[string]interface{}
	mapstructure.Decode(engine["Callstack"], &callstack)

	result := []string{}
	for _, stack := range callstack {
		result = append(result, color_cyan+stack["Label"].(string)+"\033[0m"+color_grey+" ()\033[0m")
	}

	return strings.Join(result, "\n\t")
}
