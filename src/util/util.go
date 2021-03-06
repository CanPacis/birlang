package util

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/canpacis/birlang/src/ast"
)

func GenerateIntPrimitive(value int64) ast.IntPrimitiveExpression {
	return ast.IntPrimitiveExpression{
		Operation: "primitive",
		Value:     value,
		Type:      "int",
		Position: ast.Position{
			Line: 0,
			Col:  0,
		},
	}
}

func GenerateIntFromBool(condition bool) ast.IntPrimitiveExpression {
	var value int64
	if condition {
		value = 1
	} else {
		value = 0
	}

	return ast.IntPrimitiveExpression{
		Operation: "primitive",
		Value:     value,
		Type:      "int",
		Position: ast.Position{
			Line: 0,
			Col:  0,
		},
	}
}

func GenerateIdentifier(name string) ast.Identifier {
	return ast.Identifier{
		Operation: "identifier",
		Negative:  false,
		Value:     name,
		Position:  ast.Position{Line: 0, Col: 0},
	}
}

func IsPowerOfTen(input int64) bool {
	if input%10 != 0 || input == 0 {
		return false
	}

	if input == 10 {
		return true
	}

	return IsPowerOfTen(input / 10)
}

func GenerateNativeFunction(name string, body ast.NativeFunction) ast.BlockDeclarationStatement {
	return ast.BlockDeclarationStatement{
		Name:         GenerateIdentifier(name),
		Native:       true,
		Verbs:        []ast.Identifier{},
		Arguments:    []ast.Identifier{},
		Implementing: false,
		Implements:   ast.Identifier{},
		Populate:     nil,
		Position:     ast.Position{Line: 1, Col: 0},
		Instance:     nil,
		Body:         body,
	}
}

func GenerateNativeFunctionReturn(_error bool, _warn bool, message string, value int64) ast.NativeFunctionReturn {
	return ast.NativeFunctionReturn{
		Error:   _error,
		Warn:    _warn,
		Message: message,
		Value:   GenerateIntPrimitive(value),
	}
}

func UUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	uuid := hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[0:4])

	return uuid
}

type Color struct {
	Red     string `json:"red"`
	Yellow  string `json:"yellow"`
	Cyan    string `json:"cyan"`
	Grey    string `json:"grey"`
	Default string `json:"default"`
}

func NewColor(colored bool) Color {
	if colored {
		return Color{
			Red:     "\033[1;31m",
			Yellow:  "\033[1;33m",
			Cyan:    "\033[0;36m",
			Grey:    "\033[1;30m",
			Default: "\033[0m",
		}
	} else {
		return Color{
			Red:     "",
			Yellow:  "",
			Cyan:    "",
			Grey:    "",
			Default: "",
		}
	}
}

func (color Color) OutputRed(message string) string {
	return color.Red + message + color.Default
}

func (color Color) OutputYellow(message string) string {
	return color.Yellow + message + color.Default
}

func (color Color) OutputCyan(message string) string {
	return color.Cyan + message + color.Default
}

func (color Color) OutputGrey(message string) string {
	return color.Grey + message + color.Default
}
