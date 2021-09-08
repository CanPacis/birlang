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

func GenerateNativeFunction(name string, body ast.NativeFunction) ast.BlockDeclarationStatement {
	return ast.BlockDeclarationStatement{
		Name:         GenerateIdentifier(name),
		Native:       true,
		Verbs:        []ast.Identifier{},
		Arguments:    []ast.Identifier{},
		Implementing: false,
		Implements:   ast.Identifier{},
		Popluate:     nil,
		Position:     ast.Position{Line: 1, Col: 0},
		Instance:     nil,
		Body:         body,
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
	Red     string
	Yellow  string
	Cyan    string
	Grey    string
	Default string
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
