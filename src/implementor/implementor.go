package implementor

import (
	"fmt"

	"github.com/canpacis/birlang/src/ast"
	"github.com/canpacis/birlang/src/util"
)

type Implementor struct {
	IOBuffer []byte `json:"io_buffer"`
	Name     string `json:"name"`
}

const (
	StdPush = iota + 1000000
	StdPull
	StdRead
	StdWrite
)

func (implementor Implementor) Interface(verbs []ast.IntPrimitiveExpression, arguments []ast.IntPrimitiveExpression) ast.IntPrimitiveExpression {
	fmt.Println("Hello There", verbs, arguments)
	// switch verbs[0].Value {

	// }
	return util.GenerateIntPrimitive(-1)
}

func (implementor Implementor) Push(arguments []ast.IntPrimitiveExpression) ast.IntPrimitiveExpression {
	implementor.IOBuffer = append(implementor.IOBuffer, byte(arguments[0].Value))
	return util.GenerateIntPrimitive(-1)
}

func (implementor Implementor) Pull() ast.IntPrimitiveExpression {
	return util.GenerateIntPrimitive(int64(implementor.IOBuffer[0]))
}

func (implementor Implementor) Read(arguments []ast.IntPrimitiveExpression) ast.IntPrimitiveExpression {
	return util.GenerateIntPrimitive(-1)
}

func (implementor Implementor) Write() ast.IntPrimitiveExpression {
	return util.GenerateIntPrimitive(-1)
}
