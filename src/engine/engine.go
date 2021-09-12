package engine

import (
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/canpacis/birlang/src/ast"
	"github.com/canpacis/birlang/src/implementor"
	"github.com/canpacis/birlang/src/scope"
	"github.com/canpacis/birlang/src/thrower"
	"github.com/canpacis/birlang/src/util"
	"github.com/mitchellh/mapstructure"
)

func ApplyConfig(options map[string]interface{}, instance *BirEngine) {
	var config map[string]interface{}
	mapstructure.Decode(options, &config)

	instance.ColoredOutput = config["ColoredOutput"].(bool)
	instance.VerbosityLevel = config["VerbosityLevel"].(int)

	if config["MaximumCallstackSize"] != 0 {
		instance.MaximumCallstackSize = config["MaximumCallstackSize"].(int)
	}
	instance.Thrower = thrower.Thrower{Owner: instance, Color: util.NewColor(instance.ColoredOutput)}

	for _, use := range instance.Uses {
		ApplyConfig(instance.Config, &use)
	}

}

type BirEngine struct {
	ID                   string                    `json:"id"`
	Anonymous            bool                      `json:"anonymous"`
	NamespaceAllowed     bool                      `json:"namespace_allowed"`
	Path                 string                    `json:"path"`
	URI                  string                    `json:"uri"`
	Filename             string                    `json:"filename"`
	Directory            string                    `json:"directory"`
	Content              string                    `json:"content"`
	VerbosityLevel       int                       `json:"verbosity_level"`
	Parsed               map[string]interface{}    `json:"parsed"`
	MaximumCallstackSize int                       `json:"maximum_callstack_size"`
	Callstack            []Callstack               `json:"callstack"`
	Scopestack           scope.Scopestack          `json:"scopestack"`
	StdPath              string                    `json:"std_path"`
	Uses                 []BirEngine               `json:"uses"`
	Thrower              thrower.Thrower           `json:"thrower"`
	ColoredOutput        bool                      `json:"colored_output"`
	Implementors         []implementor.Implementor `json:"implementors"`
	Config               map[string]interface{}    `json:"config"`
}

type Callstack struct {
	Label      string        `json:"label"`
	Identifier string        `json:"identifier"`
	Stack      []interface{} `json:"stack"`
}

func (engine *BirEngine) PushCallstack(callstack Callstack) []Callstack {
	result := engine.Callstack
	result = append(result, callstack)

	return result
}

func (engine *BirEngine) PopCallstack() []Callstack {
	result := make([]Callstack, 0)
	result = append(result, engine.Callstack[:len(engine.Callstack)-1]...)

	return result
}

func (engine *BirEngine) ReverseCallstack() []Callstack {
	a := make([]Callstack, len(engine.Callstack))
	copy(a, engine.Callstack)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}

func (engine BirEngine) HandleError(err error, position ast.Position) {
	if err != nil {
		engine.Thrower.Throw(err.Error()+"\nThis error is caused by an engine bug", position, engine.Callstack)
	}
}

func (engine BirEngine) HandleAnonymousError(err error) {
	if err != nil {
		engine.Thrower.ThrowAnonymous(err.Error() + "\nThis error is caused by an engine bug")
	}
}

func (engine *BirEngine) Init() {
	engine.URI = "file://" + engine.Path
	engine.ID = util.UUID()
	engine.MaximumCallstackSize = 10
	engine.Thrower = thrower.Thrower{Owner: engine, Color: util.NewColor(engine.ColoredOutput)}
	cwd, _ := os.Getwd()
	engine.StdPath = path.Join(cwd, "std")
	engine.Scopestack.PushScope(scope.Scope{})

	for _, i := range engine.Implementors {
		engine.Scopestack.AddBlock(util.GenerateNativeFunction(i.Name, i.Interface))
	}

	if !engine.Anonymous {
		dir, file := path.Split(engine.Path)
		engine.Directory = dir
		engine.Filename = file
		raw, err := os.ReadFile(engine.Path)

		engine.HandleAnonymousError(err)

		engine.Content = string(raw)
		result := ast.ParserResult{}
		out, err := exec.Command("node", "C:\\Users\\tmwwd\\go\\src\\birlang\\bin\\parser\\parser", string(engine.Content)).Output()
		engine.HandleAnonymousError(err)

		json.Unmarshal(out, &result)
		if result.Error {
			content := ast.ErrorContent{}
			engine.HandleAnonymousError(mapstructure.Decode(result.Content, &content))
			engine.Thrower.ThrowAnonymous(content.Message)
		} else {
			engine.HandleAnonymousError(mapstructure.Decode(result.Content, &engine.Parsed))

			if engine.Parsed["program"] != nil {
				engine.Callstack = engine.PushCallstack(Callstack{Label: "main", Identifier: "main", Stack: engine.Parsed["program"].([]interface{})})
			} else {
				engine.Thrower.ThrowAnonymous("Syntax error ¯\\_(ツ)_/¯. I actually don't know what's wrong with this parser")
			}

			engine.AddImports(engine.Parsed)
		}
	}
}

func (engine *BirEngine) AddImports(stack map[string]interface{}) {
	var is_standard bool
	for _, use := range stack["imports"].([]interface{}) {
		var statement ast.UseStatement
		engine.HandleAnonymousError(mapstructure.Decode(use, &statement))

		var use_path string
		if strings.HasPrefix(statement.Source.Value, "std:") {
			is_standard = true
			use_path = path.Join(engine.StdPath, strings.Split(statement.Source.Value, "std:")[1]+".bir")
			if _, err := os.Stat(use_path); os.IsNotExist(err) {
				engine.Thrower.Throw("Import '"+statement.Source.Value+"' is not included in the standard library", statement.Position, engine.Callstack)
			}
		} else if strings.HasPrefix(statement.Source.Value, "module:") {
			is_standard = false
			use_path = path.Join(engine.Directory, strings.Split(statement.Source.Value, "module:")[1])
			if _, err := os.Stat(use_path); os.IsNotExist(err) {
				engine.Thrower.Throw("Import '"+statement.Source.Value+"' could not be found", statement.Position, engine.Callstack)
			}
		} else {
			is_standard = false
			engine.Thrower.Throw("Uknown use prefix '"+strings.Split(statement.Source.Value, ":")[0]+"'", statement.Position, engine.Callstack)
		}

		use_engine := BirEngine{Path: use_path}
		use_engine.Implementors = engine.Implementors
		use_engine.Init()
		if is_standard {
			use_engine.NamespaceAllowed = true
		}
		use_engine.Run()

		s := use_engine.Scopestack.GetCurrentScope()
		use_scope := scope.Scope{}
		use_scope.Blocks = s.Blocks
		use_scope.Frame = s.Frame
		use_scope.Foreign = true
		use_scope.Immutable = true
		engine.Scopestack.ShiftScope(use_scope)
		engine.Scopestack.Namespaces = append(engine.Scopestack.Namespaces, use_engine.Scopestack.Namespaces...)
		engine.Uses = append(engine.Uses, use_engine)
	}
}

func (engine *BirEngine) Feed(input string) string {
	result := ast.ParserResult{}
	out, err := exec.Command("node", "C:\\Users\\tmwwd\\go\\src\\birlang\\bin\\parser\\parser.js", input).Output()
	engine.HandleAnonymousError(err)
	json.Unmarshal(out, &result)

	if result.Error {
		content := ast.ErrorContent{}
		engine.HandleAnonymousError(mapstructure.Decode(result.Content, &content))
		return content.Message
	} else {
		var stack map[string]interface{}
		engine.HandleAnonymousError(mapstructure.Decode(result.Content, &stack))
		engine.Scopestack.PushScope(scope.Scope{})

		engine.AddImports(stack)

		if stack["program"] != nil {
			engine.Callstack = engine.PushCallstack(Callstack{Label: "main", Identifier: "main", Stack: stack["program"].([]interface{})})
			result := engine.ResolveCallstack(engine.GetCurrentCallStack())
			return strconv.Itoa(int(result.Value))
		} else {
			engine.Thrower.ThrowAnonymous("Syntax error ¯\\_(ツ)_/¯. I actually don't know what's wrong with this parser ಠ_ಠ")
			return ""
		}
	}
}

func (engine BirEngine) GetCurrentCallStack() Callstack {
	if len(engine.Callstack) > 0 {
		return engine.Callstack[len(engine.Callstack)-1]
	}
	return Callstack{Label: "Root Stack", Identifier: "root_stack", Stack: []interface{}{}}
}

func (engine BirEngine) GetCurrentScope() scope.Scope {
	return *engine.Scopestack.GetCurrentScope()
}

func (engine *BirEngine) Run() {
	if len(engine.Callstack) > 0 {
		engine.ResolveCallstack(engine.GetCurrentCallStack())
	}
}

func (engine *BirEngine) ResolveCallstack(callstack Callstack) ast.IntPrimitiveExpression {
	var value ast.IntPrimitiveExpression

	for _, statement := range callstack.Stack {
		operation := statement.(map[string]interface{})["operation"]
		pos := statement.(map[string]interface{})["position"]
		var statement_position ast.Position
		engine.HandleAnonymousError(mapstructure.Decode(pos, &statement_position))
		switch operation {
		case "variable_declaration":
			result := ast.VariableDeclarationStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveVariableDeclaration(result)
		case "return_statement":
			var result map[string]interface{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			value := engine.ResolveExpression(result["expression"].(map[string]interface{}))
			var position ast.Position
			engine.HandleError(mapstructure.Decode(result["position"], &position), statement_position)

			if engine.GetCurrentCallStack().Identifier == "main" {
				engine.Thrower.Throw("Top level return statements are not allowed", position, engine.Callstack)
			}
			engine.Callstack = engine.PopCallstack()
			return value
		case "throw_statement":
			var result map[string]interface{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			value := engine.ResolveExpression(result["expression"].(map[string]interface{}))
			var position ast.Position
			engine.HandleError(mapstructure.Decode(result["position"], &position), statement_position)

			if engine.GetCurrentCallStack().Identifier == "main" {
				engine.Thrower.Throw("Top level throw statements are not allowed", position, engine.Callstack)
			} else {
				engine.Thrower.Throw("Bir process has thrown error with value '"+strconv.Itoa(int(value.Value))+"'", position, engine.Callstack)
			}
			engine.Callstack = engine.PopCallstack()
			return value
		case "block_declaration":
			result := ast.BlockDeclarationStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveBlockDeclaration(result)
		case "namespace_declaration":
			result := ast.NamespaceDeclarationStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveNamespaceDeclaration(result)
		case "namespace_indexer":
			var result map[string]interface{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveNamespaceIndexerExpression(result)
		case "quantity_modifier_statement":
			result := ast.QuantityModifierStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveQuantityModifierStatement(result)
		case "assign_statement":
			result := ast.AssignStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveAssignStatement(result)
		case "block_call":
			var result map[string]interface{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			value = engine.ResolveBlockCall(result, "")
		case "scope_mutater_expression":
			var result map[string]interface{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveScopeMutaterExpression(result)
		case "for_statement":
			result := ast.ForStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveForStatement(result)
		case "while_statement":
			result := ast.WhileStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveWhileStatement(result)
		case "if_statement":
			result := ast.IfStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			value := engine.ResolveIfStatement(result)
			engine.Callstack = engine.PopCallstack()
			return value
		case "switch_statement":
			result := ast.SwitchStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			value := engine.ResolveSwitchStatement(result)
			engine.Callstack = engine.PopCallstack()
			return value
		default:
		}
	}

	engine.Callstack = engine.PopCallstack()
	if value.Position.Line != 0 {
		return value
	} else {
		return util.GenerateIntPrimitive(-1)
	}
}

func (engine *BirEngine) ResolveAssignStatement(statement ast.AssignStatement) {
	right := engine.ResolveExpression(statement.Right)
	uptable_report := engine.Scopestack.IsVariableUpdatable(statement.Left.Value)

	switch uptable_report {
	case 0:
		engine.Scopestack.UpdateVariable(statement.Left.Value, right)
	case 1:
		engine.Thrower.Throw("Could not assign to a variable that does not exist", statement.Position, engine.Callstack)
	case 2:
		engine.Thrower.Throw("Could not assign to an immutable variable", statement.Position, engine.Callstack)
	case 3:
		engine.Thrower.Throw("Could not assign to a foreign variable", statement.Position, engine.Callstack)
	}
}

func (engine *BirEngine) ResolveQuantityModifierStatement(statement ast.QuantityModifierStatement) {
	var new_value int64
	reference := engine.ResolveExpression(statement.Statement)

	switch statement.Type {
	case "increment":
		new_value = reference.Value + 1
	case "decrement":
		new_value = reference.Value - 1
	case "add":
		right := engine.ResolveExpression(statement.Right)
		new_value = reference.Value + right.Value
	case "subtract":
		right := engine.ResolveExpression(statement.Right)
		new_value = reference.Value - right.Value
	case "multiply":
		right := engine.ResolveExpression(statement.Right)
		new_value = reference.Value * right.Value
	case "divide":
		right := engine.ResolveExpression(statement.Right)
		new_value = reference.Value / right.Value
	}

	if statement.Statement["operation"] == "reference" {
		uptable_report := engine.Scopestack.IsVariableUpdatable(statement.Statement["value"].(string))

		switch uptable_report {
		case 0:
			engine.Scopestack.UpdateVariable(statement.Statement["value"].(string), util.GenerateIntPrimitive(new_value))
		case 1:
			engine.Thrower.Throw("Could not modify a variable that does not exist", statement.Position, engine.Callstack)
		case 2:
			engine.Thrower.Throw("Could not modify an immutable variable", statement.Position, engine.Callstack)
		case 3:
			engine.Thrower.Throw("Could not modify a foreign variable", statement.Position, engine.Callstack)
		}
	}
}

func (engine *BirEngine) ResolveWhileStatement(statement ast.WhileStatement) {
	condition := engine.ResolveExpression(statement.Statement)

	for condition.Value == 1 {
		engine.Callstack = engine.PushCallstack(Callstack{
			Label:      "while-block " + engine.GetAnonymousIndex(statement.Position),
			Identifier: "while-block",
			Stack:      statement.Body,
		})

		engine.Scopestack.PushScope(scope.Scope{})
		engine.ResolveCallstack(engine.GetCurrentCallStack())
		engine.Scopestack.PopScope()
		condition = engine.ResolveExpression(statement.Statement)
	}
}

func (engine *BirEngine) ResolveForStatement(statement ast.ForStatement) {
	iterator := engine.ResolveExpression(statement.Statement)

	for i := 0; i < int(iterator.Value); i++ {
		_scope := scope.Scope{}
		engine.Scopestack.PushScope(_scope)
		engine.Scopestack.AddVariable(scope.Value{Key: util.GenerateIdentifier(statement.Placeholder), Value: util.GenerateIntPrimitive(int64(i)), Kind: "const"})
		engine.Callstack = engine.PushCallstack(Callstack{
			Label:      "for-block " + engine.GetAnonymousIndex(statement.Position),
			Identifier: "for-block",
			Stack:      statement.Body,
		})
		engine.ResolveCallstack(engine.GetCurrentCallStack())
		engine.Scopestack.PopScope()
	}
}

func (engine *BirEngine) ResolveIfStatement(statement ast.IfStatement) ast.IntPrimitiveExpression {
	condition := engine.ResolveExpression(statement.Condition)

	runBlock := func(name string, block []interface{}) ast.IntPrimitiveExpression {
		engine.Callstack = engine.PushCallstack(Callstack{
			Label:      name + "-block " + engine.GetAnonymousIndex(statement.Position),
			Identifier: name + "-block",
			Stack:      block,
		})

		result := engine.ResolveCallstack(engine.GetCurrentCallStack())
		engine.Scopestack.PopScope()
		return result
	}

	engine.Scopestack.PushScope(scope.Scope{})
	if condition.Value == 1 {
		return runBlock("if", statement.Body)
	} else {
		if statement.Elifs != nil {
			var selectedElif ast.Elif

			for _, elif := range statement.Elifs {
				elifCondition := engine.ResolveExpression(elif.Condition)
				if elifCondition.Value == 1 {
					selectedElif = elif
				}
			}

			if selectedElif.Body != nil {
				return runBlock("elif", selectedElif.Body)
			} else {
				if statement.Else != nil {
					return runBlock("else", statement.Else)
				}
				result := util.GenerateIntPrimitive(-1)
				engine.Scopestack.PopScope()
				return result

			}
		}
		result := util.GenerateIntPrimitive(-1)
		engine.Scopestack.PopScope()
		return result
	}
}

func (engine *BirEngine) ResolveSwitchStatement(statement ast.SwitchStatement) ast.IntPrimitiveExpression {
	condition := engine.ResolveExpression(statement.Condition)
	var body []interface{}

	for _, _c := range statement.Cases {
		_case := engine.ResolveExpression(_c.Case)

		if _case.Value == condition.Value {
			body = _c.Body
		}
	}

	if body != nil {
		engine.Callstack = engine.PushCallstack(Callstack{
			Label:      "switch-case-block " + engine.GetAnonymousIndex(statement.Position),
			Identifier: "switch-case-block",
			Stack:      body,
		})

		engine.Scopestack.PushScope(scope.Scope{})
		result := engine.ResolveCallstack(engine.GetCurrentCallStack())
		engine.Scopestack.PopScope()
		return result
	} else {
		if statement.Default.Body != nil {
			engine.Callstack = engine.PushCallstack(Callstack{
				Label:      "switch-default-block " + engine.GetAnonymousIndex(statement.Position),
				Identifier: "switch-default-block",
				Stack:      statement.Default.Body,
			})

			engine.Scopestack.PushScope(scope.Scope{})
			result := engine.ResolveCallstack(engine.GetCurrentCallStack())
			engine.Scopestack.PopScope()
			return result
		} else {
			return util.GenerateIntPrimitive(-1)
		}
	}
}

func (engine *BirEngine) ResolveBlockDeclaration(statement ast.BlockDeclarationStatement) {
	statement.Owner = engine.ID

	if engine.Scopestack.BlockExists(statement.Name.Value) {
		engine.Thrower.Throw("Could not redeclare an existing block", statement.Position, engine.Callstack)
	} else {
		if statement.Implementing {
			if engine.Scopestack.BlockExists(statement.Implements.Value) {
				implemented := engine.Scopestack.FindBlock(statement.Implements.Value)
				statement.Instance = implemented.Block.Instance

				if statement.Populate != nil {
					_type := statement.Populate.(map[string]interface{})["type"]

					engine.Scopestack.PushScope(*statement.Instance.(*scope.Scope))
					switch _type {
					case "string":
						var populate ast.StringPrimitiveExpression
						engine.HandleAnonymousError(mapstructure.Decode(statement.Populate, &populate))

						for i, value := range populate.Value {
							engine.Scopestack.AddVariable(scope.Value{
								Key:   util.GenerateIdentifier("value_" + strconv.Itoa(i)),
								Value: util.GenerateIntPrimitive(int64(value)),
								Kind:  "const",
							})
						}

						if engine.Scopestack.VariableExists("index") {
							engine.Scopestack.UpdateVariable("index", util.GenerateIntPrimitive(int64(len(populate.Value))))
						}
					case "array":
						var populate ast.ArrayPrimitiveExpression
						engine.HandleAnonymousError(mapstructure.Decode(statement.Populate, &populate))

						for i, value := range populate.Values {
							v := engine.ResolveExpression(value)
							engine.Scopestack.AddVariable(scope.Value{
								Key:   util.GenerateIdentifier("value_" + strconv.Itoa(i)),
								Value: v,
								Kind:  "const",
							})
						}

						if engine.Scopestack.VariableExists("index") {
							engine.Scopestack.UpdateVariable("index", util.GenerateIntPrimitive(int64(len(populate.Values))))
						}
					}
					statement.Instance = engine.Scopestack.PopScope()
				}

			} else {
				engine.Thrower.Throw("Could not implement '"+statement.Implements.Value+"', block is non-existant", statement.Implements.Position, engine.Callstack)
			}
		} else {
			var body ast.BlockBody
			engine.HandleAnonymousError(mapstructure.Decode(statement.Body, &body))
			if body.Init != nil {
				engine.Scopestack.PushScope(scope.Scope{})
				engine.Callstack = engine.PushCallstack(Callstack{
					Identifier: statement.Name.Value,
					Label:      statement.Name.Value + ":init",
					Stack:      body.Init,
				})
				engine.ResolveCallstack(engine.GetCurrentCallStack())
				statement.Instance = engine.Scopestack.PopScope()
			} else {
				statement.Instance = &scope.Scope{}
			}
		}
		engine.Scopestack.AddBlock(statement)
	}
}

func (engine *BirEngine) ResolveVariableDeclaration(statement ast.VariableDeclarationStatement) {
	key := statement.Left
	value := engine.ResolveExpression(statement.Right)

	variable := engine.Scopestack.FindVariable(key.Value)
	if engine.Scopestack.VariableExists(key.Value) && !variable.OuterScope {
		engine.Thrower.Throw("Could not redeclare an existing variable", statement.Position, engine.Callstack)
	} else {
		engine.Scopestack.AddVariable(scope.Value{Key: key, Value: value, Kind: statement.Kind})
	}
}

func (engine *BirEngine) ResolveExpression(raw map[string]interface{}) ast.IntPrimitiveExpression {
	var position ast.Position
	engine.HandleAnonymousError(mapstructure.Decode(raw["position"], &position))
	switch raw["operation"] {
	case "primitive":
		expression := ast.IntPrimitiveExpression{}
		engine.HandleError(mapstructure.Decode(raw, &expression), position)
		result := ast.IntPrimitiveExpression{}
		engine.HandleError(mapstructure.Decode(expression, &result), position)
		return result
	case "block_call":
		var result map[string]interface{}
		engine.HandleError(mapstructure.Decode(raw, &result), position)
		return engine.ResolveBlockCall(result, "")
	case "namespace_indexer":
		var result map[string]interface{}
		engine.HandleError(mapstructure.Decode(raw, &result), position)
		return engine.ResolveNamespaceIndexerExpression(result)
	case "scope_mutater_expression":
		var result map[string]interface{}
		engine.HandleError(mapstructure.Decode(raw, &result), position)
		return engine.ResolveScopeMutaterExpression(result)
	case "arithmetic":
		return engine.ResolveArithmeticExpression(raw)
	case "condition":
		return engine.ResolveConditionExpression(raw)
	case "reference":
		return engine.ResolveReferenceExpression(raw)
	default:
		return util.GenerateIntPrimitive(-1)
	}
}

func (engine BirEngine) PushArguments(expression ast.BlockCallExpression, block ast.BlockDeclarationStatement, incoming string) []scope.Value {
	result := []scope.Value{}

	if len(expression.Arguments) == len(block.Arguments) {
		for i, argument := range block.Arguments {
			value := engine.ResolveExpression(expression.Arguments[i])
			result = append(result, scope.Value{Key: argument, Value: value, Kind: "const"})
		}
	} else {
		engine.Thrower.Warn("Expected "+strconv.Itoa(len(block.Arguments))+" argument(s), found "+strconv.Itoa(len(expression.Arguments))+" while calling '"+expression.Name.Value+"'", expression.Position, engine.Callstack)
		for _, argument := range block.Arguments {
			result = append(result, scope.Value{Key: argument, Value: util.GenerateIntPrimitive(-1), Kind: "const"})
		}
	}

	return result
}

func (engine BirEngine) PushVerbs(expression ast.BlockCallExpression, block ast.BlockDeclarationStatement, incoming string) []scope.Value {
	result := []scope.Value{}

	if len(expression.Verbs) == len(block.Verbs) {
		for i, verb := range block.Verbs {
			value := engine.ResolveExpression(expression.Verbs[i])
			result = append(result, scope.Value{Key: verb, Value: value, Kind: "const"})
		}
	} else {
		engine.Thrower.Warn("Expected "+strconv.Itoa(len(block.Verbs))+" verb(s), found "+strconv.Itoa(len(expression.Verbs))+" while calling '"+expression.Name.Value+"'", expression.Position, engine.Callstack)
		for _, verb := range block.Verbs {
			result = append(result, scope.Value{Key: verb, Value: util.GenerateIntPrimitive(-1), Kind: "const"})
		}
	}

	return result
}

func (engine *BirEngine) ResolveBlockCall(raw map[string]interface{}, incoming string) ast.IntPrimitiveExpression {
	expression := ast.BlockCallExpression{}
	engine.HandleError(mapstructure.Decode(raw, &expression), expression.Position)

	if engine.Scopestack.BlockExists(expression.Name.Value) {
		result := engine.Scopestack.FindBlock(expression.Name.Value)

		if len(engine.Callstack) <= engine.MaximumCallstackSize {
			if result.Foreign {
				owner := engine.FindOwner(result.Block.Owner, expression)
				owner.Scopestack.PushScope(engine.GetCurrentScope())

				local_scope := []scope.Value{}

				local_scope = append(local_scope, engine.PushArguments(expression, *result.Block, incoming)...)
				local_scope = append(local_scope, engine.PushVerbs(expression, *result.Block, incoming)...)

				for _, value := range local_scope {
					owner.Scopestack.AddVariable(value)
				}

				value := owner.ResolveBlockCall(raw, owner.ID)
				owner.Scopestack.PopScope()
				return value
			} else {
				if result.Block.Native {
					var body ast.NativeFunction
					engine.HandleAnonymousError(mapstructure.Decode(result.Block.Body, &body))

					arguments := []ast.IntPrimitiveExpression{}
					verbs := []ast.IntPrimitiveExpression{}

					for _, argument := range expression.Arguments {
						arguments = append(arguments, engine.ResolveExpression(argument))
					}
					for _, verb := range expression.Verbs {
						verbs = append(verbs, engine.ResolveExpression(verb))
					}

					engine.Callstack = engine.PushCallstack(Callstack{
						Label:      expression.Name.Value,
						Identifier: "$" + expression.Name.Value,
						Stack:      []interface{}{},
					})
					native_function_return := body(verbs, arguments)
					if native_function_return.Error {
						engine.Thrower.Throw(native_function_return.Message, expression.Position, engine.Callstack)
						engine.Callstack = engine.PopCallstack()
						return native_function_return.Value
					} else if native_function_return.Warn {
						engine.Thrower.Warn(native_function_return.Message, expression.Position, engine.Callstack)
						engine.Callstack = engine.PopCallstack()
						return native_function_return.Value
					} else {
						engine.Callstack = engine.PopCallstack()
						return native_function_return.Value
					}
				}
				if result.Block.Implementing {
					implemented := engine.Scopestack.FindBlock(result.Block.Implements.Value)
					instance := result.Block.Instance.(*scope.Scope)

					if implemented.Foreign {
						owner := engine.FindOwner(implemented.Block.Owner, expression)
						engine.Scopestack.PushScope(*instance)
						raw["name"].(map[string]interface{})["value"] = implemented.Block.Name.Value

						local_scope := []scope.Value{}

						local_scope = append(local_scope, engine.PushArguments(expression, *implemented.Block, incoming)...)
						local_scope = append(local_scope, engine.PushVerbs(expression, *implemented.Block, incoming)...)

						for _, value := range local_scope {
							owner.Scopestack.AddVariable(value)
						}

						value := owner.ResolveBlockCall(raw, owner.ID)
						owner.Scopestack.PopScope()
						return value
					}

					engine.Scopestack.PushScope(*instance)
					local_scope := []scope.Value{}

					local_scope = append(local_scope, engine.PushArguments(expression, *implemented.Block, incoming)...)
					local_scope = append(local_scope, engine.PushVerbs(expression, *implemented.Block, incoming)...)

					for _, value := range local_scope {
						engine.Scopestack.AddVariable(value)
					}

					var body map[string][]interface{}
					engine.HandleError(mapstructure.Decode(implemented.Block.Body, &body), expression.Position)
					engine.Callstack = engine.PushCallstack(Callstack{
						Label:      result.Block.Name.Value + "->" + implemented.Block.Name.Value,
						Identifier: "$" + result.Block.Name.Value,
						Stack:      body["program"],
					})

					value := engine.ResolveCallstack(engine.GetCurrentCallStack())
					engine.Scopestack.PopScope()
					return value
				} else {
					var body ast.BlockBody
					engine.HandleAnonymousError(mapstructure.Decode(result.Block.Body, &body))
					instance := result.Block.Instance.(*scope.Scope)
					engine.Scopestack.PushScope(*instance)

					if incoming == "" {
						local_scope := []scope.Value{}

						local_scope = append(local_scope, engine.PushArguments(expression, *result.Block, incoming)...)
						local_scope = append(local_scope, engine.PushVerbs(expression, *result.Block, incoming)...)

						for _, value := range local_scope {
							engine.Scopestack.AddVariable(value)
						}
					}

					engine.Callstack = engine.PushCallstack(Callstack{
						Label:      expression.Name.Value,
						Identifier: "$" + expression.Name.Value,
						Stack:      body.Program,
					})
					value := engine.ResolveCallstack(engine.GetCurrentCallStack())
					engine.Scopestack.PopScope()
					return value
				}
			}
		} else {
			var callstack []Callstack

			if len(engine.Callstack) >= 10 {
				callstack = engine.Callstack[:10]
			} else {
				callstack = engine.Callstack
			}

			engine.Thrower.Throw("Bir process has overflown the maximum callstack size", expression.Position, callstack)
			return util.GenerateIntPrimitive(-1)
		}
	} else {
		engine.Thrower.Throw("Could not find block '"+expression.Name.Value+"'", expression.Position, engine.Callstack)
		return util.GenerateIntPrimitive(-1)
	}
}

func (engine *BirEngine) ResolveReferenceExpression(raw map[string]interface{}) ast.IntPrimitiveExpression {
	result := engine.Scopestack.FindVariable(raw["value"].(string))
	expression := ast.ReferenceExpression{}
	engine.HandleError(mapstructure.Decode(raw, &expression), expression.Position)

	if result.Value != nil {
		if raw["negative"].(bool) {
			return util.GenerateIntPrimitive(-result.Value.Value.Value)
		}
		return result.Value.Value
	} else {
		engine.Thrower.Throw("Could not find variable '"+expression.Value+"' in the frame", expression.Position, engine.Callstack)
		return util.GenerateIntPrimitive(-1)
	}
}

func (engine BirEngine) FindUpperBlockScope() (*scope.Scope, int) {
	selected_stack_index := -1

	for i, cs := range engine.ReverseCallstack() {
		if strings.HasPrefix(cs.Identifier, "$") {
			selected_stack_index = i
			break
		}
	}

	if selected_stack_index < 0 {
		return &scope.Scope{}, -1
	} else {
		selected_scope := (engine.Scopestack.Reverse())[selected_stack_index]
		return &selected_scope, selected_stack_index
	}
}

func (engine *BirEngine) ResolveScopeMutaterExpression(raw map[string]interface{}) ast.IntPrimitiveExpression {
	expression := ast.ScopeMutaterExpression{}
	engine.HandleAnonymousError(mapstructure.Decode(raw, &expression))

	WriteToScope := func() ast.IntPrimitiveExpression {
		if len(expression.Arguments) < 2 {
			engine.Thrower.Throw("Scope mutation with 'Write' operation needs at least 2 arguments but '"+strconv.Itoa(len(expression.Arguments))+"' is given", expression.Position, engine.Callstack)
			return util.GenerateIntPrimitive(-1)
		}
		arguments := []ast.IntPrimitiveExpression{}

		for _, value := range expression.Arguments {
			arguments = append(arguments, engine.ResolveExpression(value))
		}

		selected_scope, index := engine.FindUpperBlockScope()
		selected_scope.AddVariable(scope.Value{
			Key:   util.GenerateIdentifier("value_" + strconv.Itoa(int(arguments[0].Value))),
			Value: arguments[1],
			Kind:  "const",
		})
		engine.Scopestack.SwapAtIndex(index, *selected_scope)

		return util.GenerateIntPrimitive(-1)
	}

	ReadFromScope := func() ast.IntPrimitiveExpression {
		if len(expression.Arguments) < 1 {
			engine.Thrower.Throw("Scope mutation with 'Read' operation needs at least 1 argument but '"+strconv.Itoa(len(expression.Arguments))+"' is given", expression.Position, engine.Callstack)
			return util.GenerateIntPrimitive(-1)
		}
		arguments := []ast.IntPrimitiveExpression{}

		for _, value := range expression.Arguments {
			arguments = append(arguments, engine.ResolveExpression(value))
		}

		selected_scope, i := engine.FindUpperBlockScope()
		var selected_value ast.IntPrimitiveExpression

		if i < 0 {
			engine.Thrower.Throw("IDK", expression.Position, engine.Callstack)
		}

		for _, value := range selected_scope.Frame {
			if value.Key.Value == "value_"+strconv.Itoa(int(arguments[0].Value)) {
				selected_value = value.Value
			}
		}

		if selected_value.Type == "" {
			engine.Thrower.Throw("Could not read index '"+strconv.Itoa(int(arguments[0].Value))+"', the value is non-existant", expression.Position, engine.Callstack)
			return util.GenerateIntPrimitive(-1)
		}
		return selected_value
	}

	// TODO: Delete index from scope
	DeleteFromScope := func() ast.IntPrimitiveExpression {
		return util.GenerateIntPrimitive(100)
	}

	if engine.GetCurrentCallStack().Identifier == "main" {
		engine.Thrower.Throw("Top level scope mutater expressions are not allowed", expression.Position, engine.Callstack)
		return util.GenerateIntPrimitive(-1)
	}

	switch expression.Mutater.Value {
	case "Write":
		return WriteToScope()
	case "Read":
		return ReadFromScope()
	case "Delete":
		return DeleteFromScope()
	default:
		engine.Thrower.Throw("Unknown mutater '"+expression.Mutater.Value+"'", expression.Mutater.Position, engine.Callstack)
		return util.GenerateIntPrimitive(1)
	}
}

// TODO
func (engine *BirEngine) ResolveNamespaceDeclaration(statement ast.NamespaceDeclarationStatement) {
	if engine.NamespaceAllowed {
		engine.Scopestack.PushScope(scope.Scope{})
		for _, raw := range statement.Body {
			var sub_statement ast.VariableDeclarationStatement
			engine.HandleError(mapstructure.Decode(raw, &sub_statement), statement.Position)
			engine.ResolveVariableDeclaration(sub_statement)
		}
		namespace_scope := engine.Scopestack.PopScope()
		engine.Scopestack.PushNamespace(statement.Name.Value, *namespace_scope)
	} else {
		engine.Thrower.Throw("Namespaces are not allowed in this file", statement.Position, engine.Callstack)
	}
}

// TODO
func (engine *BirEngine) ResolveNamespaceIndexerExpression(raw map[string]interface{}) ast.IntPrimitiveExpression {
	expression := ast.NamespaceIndexerExpression{}
	engine.HandleAnonymousError(mapstructure.Decode(raw, &expression))
	namespace_exists := engine.Scopestack.NamespaceExists(expression.Namespace.Value)

	if namespace_exists {
		value := engine.Scopestack.FindInNamespace(expression.Namespace.Value, expression.Index.Value)

		if value.Value.Type != "" {
			return value.Value
		}

		engine.Thrower.Throw("Could not find variable '"+expression.Index.Value+"' in the namespace '"+expression.Namespace.Value+"'", expression.Position, engine.Callstack)
		return util.GenerateIntPrimitive(-1)
	}

	engine.Thrower.Throw("Could not find namespace '"+expression.Namespace.Value+"'", expression.Position, engine.Callstack)
	return util.GenerateIntPrimitive(-1)
}

func (engine *BirEngine) ResolveConditionExpression(raw map[string]interface{}) ast.IntPrimitiveExpression {
	left := engine.ResolveExpression(raw["left"].(map[string]interface{}))
	right := engine.ResolveExpression(raw["right"].(map[string]interface{}))

	switch raw["type"] {
	case "equals":
		return util.GenerateIntFromBool(left.Value == right.Value)
	case "not_equals":
		return util.GenerateIntFromBool(left.Value != right.Value)
	case "less_than":
		return util.GenerateIntFromBool(left.Value < right.Value)
	case "less_than_equals":
		return util.GenerateIntFromBool(left.Value <= right.Value)
	case "not_less_than":
		return util.GenerateIntFromBool(!(left.Value < right.Value))
	case "not_less_than_equals":
		return util.GenerateIntFromBool(!(left.Value <= right.Value))
	case "greater_than":
		return util.GenerateIntFromBool(left.Value > right.Value)
	case "greater_than_equals":
		return util.GenerateIntFromBool(left.Value >= right.Value)
	case "not_greater_than":
		return util.GenerateIntFromBool(!(left.Value > right.Value))
	case "not_greater_than_equals":
		return util.GenerateIntFromBool(!(left.Value >= right.Value))
	default:
		return util.GenerateIntPrimitive(-1)
	}
}

func (engine *BirEngine) ResolveArithmeticExpression(raw map[string]interface{}) ast.IntPrimitiveExpression {
	left := engine.ResolveExpression(raw["left"].(map[string]interface{}))
	right := engine.ResolveExpression(raw["right"].(map[string]interface{}))

	switch raw["type"] {
	case "addition":
		return util.GenerateIntPrimitive(left.Value + right.Value)
	case "subtraction":
		return util.GenerateIntPrimitive(left.Value - right.Value)
	case "multiplication":
		return util.GenerateIntPrimitive(left.Value * right.Value)
	case "division":
		return util.GenerateIntPrimitive(left.Value / right.Value)
	case "exponent":
		return util.GenerateIntPrimitive(int64(math.Pow(float64(left.Value), float64(right.Value))))
	case "modulus":
		return util.GenerateIntPrimitive(left.Value % right.Value)
	case "root":
		return util.GenerateIntPrimitive(int64(math.Pow(float64(left.Value), float64(1/right.Value))))
	case "log10":
		return util.GenerateIntPrimitive(-100)
	default:
		return util.GenerateIntPrimitive(-10)
	}
}

func (engine BirEngine) GetAnonymousIndex(position ast.Position) string {
	return "[" + engine.Filename + "->" + strconv.Itoa(int(position.Line)) + ":" + strconv.Itoa(int(position.Col)) + "]"
}

func (engine BirEngine) FindOwner(id string, expression ast.BlockCallExpression) *BirEngine {
	var owner *BirEngine

	for _, use := range engine.Uses {
		if use.ID == id {
			owner = &use
			break
		}
	}

	if owner != nil {
		return owner
	} else {
		engine.Thrower.Throw("Could not find block '"+expression.Name.Value+"'", expression.Position, engine.Callstack)
		return owner
	}
}

func NewEngine(path string, anonymous bool, colored_output bool, verbosity_level int) BirEngine {
	engine := BirEngine{
		Path:             path,
		NamespaceAllowed: false,
		Anonymous:        anonymous,
		ColoredOutput:    colored_output,
		VerbosityLevel:   verbosity_level,
		Implementors:     []implementor.Implementor{{Name: "bir"}},
	}
	return engine
}
