package implementor

import (
	"bufio"
	"fmt"
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
	StdOut
	StdFile
	StdDone
	StdIn
)

func (implementor Implementor) Interface(verbs []ast.IntPrimitiveExpression, arguments []ast.IntPrimitiveExpression) ast.NativeFunctionReturn {
	if len(verbs) > 0 {
		fmt.Println(verbs[0])
		switch verbs[0].Value {
		case StdPush:
			return implementor.Push(arguments)
		case StdPull:
			return implementor.Pull()
		case StdRead:
			return implementor.Read(arguments)
		case StdWrite:
			return implementor.Write(arguments)
		}
		return util.GenerateNativeFunctionReturn(false, false, "", -1)
	} else {
		return util.GenerateNativeFunctionReturn(true, false, "Native 'bir' block needs at least 1 verb", -1)
	}
}

func (implementor Implementor) Push(arguments []ast.IntPrimitiveExpression) ast.NativeFunctionReturn {
	if len(arguments) > 0 {
		fmt.Println(arguments)
		io_buffer = append(io_buffer, byte(arguments[0].Value))
		return util.GenerateNativeFunctionReturn(false, false, "", -1)
	} else {
		return util.GenerateNativeFunctionReturn(true, false, "Native 'bir' block's 'push' verb needs at least 1 argument", -1)
	}
}

func (implementor Implementor) Pull() ast.NativeFunctionReturn {
	var element int64
	fmt.Println(io_buffer)

	if len(io_buffer) > 0 {
		element = int64(io_buffer[0])
		io_buffer = io_buffer[1:]
	} else {
		element = StdDone
	}
	fmt.Println("Pulled value: ", element)
	return util.GenerateNativeFunctionReturn(false, false, "", element)
}

func (implementor Implementor) Read(arguments []ast.IntPrimitiveExpression) ast.NativeFunctionReturn {
	if len(arguments) > 0 {
		switch arguments[0].Value {
		case StdOut:
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			text := scanner.Text()
			io_buffer = append(io_buffer, []byte(text)...)
		case StdFile:
		}
		io_buffer = []byte{}
		return util.GenerateNativeFunctionReturn(false, false, "", -1)
	} else {
		return util.GenerateNativeFunctionReturn(true, false, "Native 'bir' block's 'read' verb needs at least 1 argument", -1)
	}
}

func (implementor Implementor) Write(arguments []ast.IntPrimitiveExpression) ast.NativeFunctionReturn {
	if len(arguments) > 0 {
		switch arguments[0].Value {
		case StdOut:
			os.Stdout.WriteString(string(io_buffer))
		case StdFile:
		}
		io_buffer = []byte{}
		return util.GenerateNativeFunctionReturn(false, false, "", -1)
	} else {
		return util.GenerateNativeFunctionReturn(true, false, "Native 'bir' block's 'write' verb needs at least 1 argument", -1)
	}
}
