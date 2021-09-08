package implementor

import (
	"fmt"

	"github.com/canpacis/birlang/src/ast"
	"github.com/canpacis/birlang/src/util"
)

type Implementor struct {
	IOBuffer []byte
	Name     string
}

func (implementor Implementor) Interface(arguments []ast.IntPrimitiveExpression, verbs []ast.IntPrimitiveExpression) ast.IntPrimitiveExpression {
	fmt.Println("Hello There", arguments, verbs)
	switch verbs[0].Value {

	}
	return util.GenerateIntPrimitive(-1)
}

func (implementor Implementor) Push(arguments []ast.IntPrimitiveExpression) ast.IntPrimitiveExpression {
	implementor.IOBuffer = append(implementor.IOBuffer, byte(arguments[0].Value))
	return util.GenerateIntPrimitive(-1)
}

func (implementor Implementor) Pull() ast.IntPrimitiveExpression {
	return util.GenerateIntPrimitive(int64(implementor.IOBuffer[0]))
}
