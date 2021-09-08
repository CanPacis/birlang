package implementor

import (
	"github.com/canpacis/birlang/src/ast"
)

type Implementor struct {
	IOBuffer []byte
}

func (implementor Implementor) Push(arguments []ast.IntPrimitiveExpression) {
	implementor.IOBuffer = append(implementor.IOBuffer, byte(arguments[0].Value))
}
