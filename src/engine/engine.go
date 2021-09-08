package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/canpacis/birlang/src/ast"
	"github.com/canpacis/birlang/src/scope"
	"github.com/canpacis/birlang/src/thrower"
	"github.com/canpacis/birlang/src/util"
	"github.com/mitchellh/mapstructure"
)

type BirEngine struct {
	ID                   string                 `json:"id"`
	Anonymous            bool                   `json:"anonymous"`
	Path                 string                 `json:"path"`
	URI                  string                 `json:"uri"`
	Filename             string                 `json:"filename"`
	Directory            string                 `json:"directory"`
	Content              string                 `json:"content"`
	VerbosityLevel       int                    `json:"verbosity_level"`
	Parsed               map[string]interface{} `json:"parsed"`
	MaximumCallstackSize uint16                 `json:"maximum_callstack_size"`
	Callstack            []Callstack            `json:"callstack"`
	Scopestack           scope.Scopestack       `json:"scopestack"`
	StdPath              string                 `json:"std_path"`
	Uses                 []BirEngine            `json:"uses"`
	Thrower              thrower.Thrower        `json:"thrower"`
	ColoredOutput        bool                   `json:"colored_output"`
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

func (engine BirEngine) HandleError(err error, position ast.Position) {
	if err != nil {
		engine.Thrower.Throw(err.Error()+"\nThis error is caused by an engine bug", position)
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
	engine.MaximumCallstackSize = 8000
	engine.Thrower = thrower.Thrower{Owner: engine, Color: util.NewColor(engine.ColoredOutput)}
	cwd, _ := os.Getwd()
	engine.StdPath = path.Join(cwd, "std")

	if !engine.Anonymous {
		dir, file := path.Split(engine.Path)
		engine.Directory = dir
		engine.Filename = file
		raw, err := ioutil.ReadFile(engine.Path)

		engine.HandleAnonymousError(err)

		engine.Content = string(raw)
		result := ast.ParserResult{}
		out, err := exec.Command("C:\\Users\\tmwwd\\go\\src\\bir\\bin\\parser\\bir-parser-win.exe", string(engine.Content)).Output()
		engine.HandleAnonymousError(err)

		json.Unmarshal(out, &result)
		if result.Error {
			content := ast.ErrorContent{}
			engine.HandleAnonymousError(mapstructure.Decode(result.Content, &content))
			engine.Thrower.ThrowAnonymous(content.Message)
		} else {
			engine.HandleAnonymousError(mapstructure.Decode(result.Content, &engine.Parsed))
			engine.Scopestack.PushScope(scope.Scope{})
			if engine.Parsed["program"] != nil {
				engine.Callstack = engine.PushCallstack(Callstack{Label: "main", Identifier: "main", Stack: engine.Parsed["program"].([]interface{})})
			} else {
				engine.Thrower.ThrowAnonymous("Syntax error ¯\\_(ツ)_/¯. I actually don't know what's wrong with this parser")
			}

			for _, use := range engine.Parsed["imports"].([]interface{}) {
				var statement ast.UseStatement
				engine.HandleAnonymousError(mapstructure.Decode(use, &statement))

				use_path := path.Join(engine.StdPath, statement.Source.Value+".bir")
				if _, err := os.Stat(use_path); os.IsNotExist(err) {
					engine.Thrower.Throw("Import '"+statement.Source.Value+"' is not included in the standard library", statement.Position)
				}

				use_engine := BirEngine{Path: use_path}
				use_engine.Init()
				use_engine.Run()

				s := use_engine.Scopestack.GetCurrentScope()
				use_scope := scope.Scope{}
				use_scope.Blocks = s.Blocks
				use_scope.Frame = s.Frame
				use_scope.Foreign = true
				use_scope.Immutable = true
				engine.Scopestack.ShiftScope(use_scope)
				engine.Uses = append(engine.Uses, use_engine)
			}
		}
	}
}

func (engine *BirEngine) Feed(input string) string {
	result := ast.ParserResult{}
	out, err := exec.Command("C:\\Users\\tmwwd\\go\\src\\bir\\bin\\parser\\bir-parser-win.exe", input).Output()
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
				engine.Thrower.Throw("Top level return statements are not allowed", position)
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
				engine.Thrower.Throw("Top level throw statements are not allowed", position)
			} else {
				engine.Thrower.Throw("Bir process has thrown error with value '"+strconv.Itoa(int(value.Value))+"'", position)
			}
			engine.Callstack = engine.PopCallstack()
			return value
		case "block_declaration":
			result := ast.BlockDeclarationStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveBlockDeclaration(result)
		case "native_block_declaration":
			result := ast.NativeBlockDeclarationStatement{}
			engine.HandleError(mapstructure.Decode(statement, &result), statement_position)
			engine.ResolveNativeBlockDeclaration(result)
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
			value := engine.ResolveBlockCall(result, false)
			engine.Callstack = engine.PopCallstack()
			return value
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
	return util.GenerateIntPrimitive(-1)
}

func (engine *BirEngine) ResolveAssignStatement(statement ast.AssignStatement) {
	right := engine.ResolveExpression(statement.Right)
	uptable_report := engine.Scopestack.IsVariableUpdatable(statement.Left.Value)

	switch uptable_report {
	case 0:
		engine.Scopestack.UpdateVariable(statement.Left.Value, right)
	case 1:
		engine.Thrower.Throw("Could not assign to a variable that does not exist", statement.Position)
	case 2:
		engine.Thrower.Throw("Could not assign to an immutable variable", statement.Position)
	case 3:
		engine.Thrower.Throw("Could not assign to a foreign variable", statement.Position)
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
			engine.Thrower.Throw("Could not modify a variable that does not exist", statement.Position)
		case 2:
			engine.Thrower.Throw("Could not modify an immutable variable", statement.Position)
		case 3:
			engine.Thrower.Throw("Could not modify a foreign variable", statement.Position)
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

		engine.ResolveCallstack(engine.GetCurrentCallStack())
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

		return engine.ResolveCallstack(engine.GetCurrentCallStack())
	}

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
				return util.GenerateIntPrimitive(-1)
			}
		}
		return util.GenerateIntPrimitive(-1)
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

		return engine.ResolveCallstack(engine.GetCurrentCallStack())
	} else {
		if statement.Default.Body != nil {
			engine.Callstack = engine.PushCallstack(Callstack{
				Label:      "switch-default-block " + engine.GetAnonymousIndex(statement.Position),
				Identifier: "switch-default-block",
				Stack:      statement.Default.Body,
			})

			return engine.ResolveCallstack(engine.GetCurrentCallStack())
		} else {
			return util.GenerateIntPrimitive(-1)
		}
	}
}

// TODO
func (engine *BirEngine) ResolveBlockDeclaration(statement ast.BlockDeclarationStatement) {
	statement.Owner = engine.ID

	if engine.Scopestack.BlockExists(statement.Name.Value) {
		engine.Thrower.Throw("Could not redeclare an existing block", statement.Position)
	} else {
		if statement.Implementing {
			if engine.Scopestack.BlockExists(statement.Implements.Value) {

			} else {
				engine.Thrower.Throw("Could not implement '"+statement.Implements.Value+"', block is non-existant", statement.Implements.Position)
			}
		} else {
			if statement.Body.Init != nil {
				engine.Scopestack.PushScope(scope.Scope{})
				engine.Callstack = engine.PushCallstack(Callstack{
					Identifier: statement.Name.Value,
					Label:      statement.Name.Value + ":init",
					Stack:      statement.Body.Init,
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

// TODO
func (engine *BirEngine) ResolveNativeBlockDeclaration(statement ast.NativeBlockDeclarationStatement) {
	statement.Owner = engine.ID

	// if engine.Scopestack.BlockExists(statement.Name.Value) {
	// 	engine.Thrower.Throw("Could not redeclare an existing block", statement.Position)
	// } else {
	// 	engine.Scopestack.AddBlock(statement)
	// }
}

func (engine *BirEngine) ResolveVariableDeclaration(statement ast.VariableDeclarationStatement) {
	key := statement.Left
	value := engine.ResolveExpression(statement.Right)

	variable := engine.Scopestack.FindVariable(key.Value)
	if engine.Scopestack.VariableExists(key.Value) && !variable.OuterScope {
		engine.Thrower.Throw("Could not redeclare an existing variable", statement.Position)
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
		return engine.ResolveBlockCall(result, false)
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

func (engine BirEngine) PushArguments(expression ast.BlockCallExpression, block ast.BlockDeclarationStatement) {
	if len(expression.Arguments) == len(block.Arguments) {
		for i, argument := range block.Arguments {
			value := engine.ResolveExpression(expression.Arguments[i])
			engine.Scopestack.AddVariable(scope.Value{Key: argument, Value: value, Kind: "const"})
		}
	} else {
		engine.Thrower.Warn("Expected "+strconv.Itoa(len(block.Arguments))+" argument(s), found "+strconv.Itoa(len(expression.Arguments))+" while calling '"+expression.Name.Value+"'", expression.Position)
		for _, argument := range block.Arguments {
			engine.Scopestack.AddVariable(scope.Value{Key: argument, Value: util.GenerateIntPrimitive(-1), Kind: "const"})
		}
	}
}

func (engine BirEngine) PushVerbs(expression ast.BlockCallExpression, block ast.BlockDeclarationStatement) {
	if len(expression.Verbs) == len(block.Verbs) {
		for i, verb := range block.Verbs {
			value := engine.ResolveExpression(expression.Verbs[i])
			engine.Scopestack.AddVariable(scope.Value{Key: verb, Value: value, Kind: "const"})
		}
	} else {
		engine.Thrower.Warn("Expected "+strconv.Itoa(len(block.Verbs))+" verb(s), found "+strconv.Itoa(len(expression.Verbs))+" while calling '"+expression.Name.Value+"'", expression.Position)
		for _, verb := range block.Verbs {
			engine.Scopestack.AddVariable(scope.Value{Key: verb, Value: util.GenerateIntPrimitive(-1), Kind: "const"})
		}
	}
}

// TODO
func (engine *BirEngine) ResolveBlockCall(raw map[string]interface{}, incoming bool) ast.IntPrimitiveExpression {
	expression := ast.BlockCallExpression{}
	engine.HandleError(mapstructure.Decode(raw, &expression), expression.Position)

	if engine.Scopestack.BlockExists(expression.Name.Value) {
		result := engine.Scopestack.FindBlock(expression.Name.Value)

		if result.Foreign {
			owner := engine.FindOwner(result.Block.Owner, expression)
			fmt.Printf("Block is foreign %s is calling the block...\n", owner.ID)
			// owner.ResolveBlockCall(raw, true)
			return util.GenerateIntPrimitive(-1)
		} else {
			if result.Block.Implementing {
				return util.GenerateIntPrimitive(-1)
			} else {
				instance := result.Block.Instance.(*scope.Scope)
				engine.Scopestack.PushScope(*instance)
				engine.PushArguments(expression, *result.Block)
				engine.PushVerbs(expression, *result.Block)
				engine.Callstack = engine.PushCallstack(Callstack{
					Label:      expression.Name.Value,
					Identifier: expression.Name.Value,
					Stack:      result.Block.Body.Program,
				})
				value := engine.ResolveCallstack(engine.GetCurrentCallStack())
				engine.Scopestack.PopScope()
				return value
			}
		}
	} else {
		engine.Thrower.Throw("Could not find block '"+expression.Name.Value+"'", expression.Position)
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
		engine.Thrower.Throw("Could not find variable '"+expression.Value+"' in the frame", expression.Position)
		return util.GenerateIntPrimitive(-1)
	}
}

// TODO
func (engine *BirEngine) ResolveScopeMutaterExpression(raw map[string]interface{}) ast.IntPrimitiveExpression {
	return util.GenerateIntPrimitive(0)
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
		return util.GenerateIntPrimitive(-10)
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
		engine.Thrower.Throw("Could not find block '"+expression.Name.Value+"'", expression.Position)
		return owner
	}
}

func NewEngine(path string, anonymous bool, colored_output bool, verbosity_level int) BirEngine {
	engine := BirEngine{
		Path:           path,
		Anonymous:      anonymous,
		ColoredOutput:  colored_output,
		VerbosityLevel: verbosity_level,
	}
	engine.Init()
	return engine
}
