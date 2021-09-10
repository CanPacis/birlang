package implementor

import (
	"os"

	"github.com/canpacis/birlang/src/ast"
	"github.com/canpacis/birlang/src/util"
)

var io_buffer []byte

type Implementor struct {
	Name string `json:"name"`
}

const (
	StdPush = iota + 1000000
	StdPull
	StdRead
	StdWrite
)

func (implementor Implementor) Interface(verbs []ast.IntPrimitiveExpression, arguments []ast.IntPrimitiveExpression) ast.NativeFunctionReturn {
	if len(verbs) > 0 {
		switch verbs[0].Value {
		case StdPush:
			return implementor.Push(arguments)
		case StdPull:
			return implementor.Pull()
		case StdRead:
			return implementor.Read(arguments)
		case StdWrite:
			return implementor.Write()
		}
		return util.GenerateNativeFunctionReturn(false, false, "", -1)
	} else {
		return util.GenerateNativeFunctionReturn(true, false, "Native 'bir' block needs at least 1 verb", -1)
	}
}

func (implementor Implementor) Push(arguments []ast.IntPrimitiveExpression) ast.NativeFunctionReturn {
	if len(arguments) > 0 {
		io_buffer = append(io_buffer, byte(arguments[0].Value))
		return util.GenerateNativeFunctionReturn(false, false, "", -1)
	} else {
		return util.GenerateNativeFunctionReturn(true, false, "Native 'bir' block's 'push' verb needs at least 1 argument", -1)
	}
}

func (implementor Implementor) Pull() ast.NativeFunctionReturn {
	return util.GenerateNativeFunctionReturn(false, false, "", -100)
}

func (implementor Implementor) Read(arguments []ast.IntPrimitiveExpression) ast.NativeFunctionReturn {
	return util.GenerateNativeFunctionReturn(false, false, "", -1)
}

func (implementor Implementor) Write() ast.NativeFunctionReturn {
	os.Stdout.WriteString(string(io_buffer))
	io_buffer = []byte{}
	return util.GenerateNativeFunctionReturn(false, false, "", -1)
}
