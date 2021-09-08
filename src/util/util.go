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

func UUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	uuid := hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[0:4])

	return uuid
}
