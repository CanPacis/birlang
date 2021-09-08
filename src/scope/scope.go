package scope

import (
	"github.com/canpacis/birlang/src/ast"
)

type Scopestack struct {
	Scopes []Scope `json:"scopes"`
}

func (scopestack *Scopestack) Reverse() []Scope {
	a := make([]Scope, len(scopestack.Scopes))
	copy(a, scopestack.Scopes)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}

func (scopestack *Scopestack) ShiftScope(scope Scope) {
	scopestack.Scopes = append([]Scope{scope}, scopestack.Scopes...)
}

func (scopestack *Scopestack) PushScope(scope Scope) {
	scopestack.Scopes = append(scopestack.Scopes, scope)
}

func (scopestack *Scopestack) UnshiftScope() {
	result := Scopestack{}
	result.Scopes = append(result.Scopes, scopestack.Scopes[:len(scopestack.Scopes)-1]...)

	scopestack.Scopes = result.Scopes
}

func (scopestack *Scopestack) PopScope() {
	result := Scopestack{}
	result.Scopes = append(result.Scopes, scopestack.Scopes[:len(scopestack.Scopes)-1]...)

	scopestack.Scopes = result.Scopes
}

func (scopestack *Scopestack) AddVariable(value Value) {
	scopestack.GetCurrentScope().Frame = append(scopestack.GetCurrentScope().Frame, value)
}

func (scopestack *Scopestack) IsVariableUpdatable(key string) UpdateReport {
	if !scopestack.VariableExists(key) {
		return 1
	} else {
		scope_value := scopestack.FindVariable(key)
		if scope_value.Immutable {
			return 2
		} else {
			if scope_value.Foreign {
				return 3
			} else {
				if scope_value.Value.Kind == "const" {
					return 2
				} else {
					return 0
				}
			}
		}
	}
}

func (scopetack *Scopestack) UpdateVariable(key string, value ast.IntPrimitiveExpression) {
	variable := scopetack.FindVariable(key)

	variable.Value.Value = value
}

func (scopestack *Scopestack) VariableExists(key string) bool {
	for _, scope := range scopestack.Reverse() {
		for _, value := range scope.Frame {
			if value.Key.Value == key {
				return true
			}
		}
	}

	return false
}

func (scopestack *Scopestack) FindVariable(key string) ScopeValue {
	for _, scope := range scopestack.Reverse() {
		for _, value := range scope.Frame {
			if value.Key.Value == key {
				return ScopeValue{
					Value:     &value,
					Foreign:   scope.Foreign,
					Immutable: scope.Immutable,
				}
			}
		}
	}

	return ScopeValue{
		Value:     nil,
		Foreign:   false,
		Immutable: false,
	}
}

func (scopestack *Scopestack) AddBlock(block ast.BlockDeclarationStatement) {
	scopestack.GetCurrentScope().Blocks = append(scopestack.GetCurrentScope().Blocks, block)
}

func (scopestack *Scopestack) BlockExists(key string) bool {
	for _, scope := range scopestack.Reverse() {
		for _, value := range scope.Blocks {
			if value.Name.Value == key {
				return true
			}
		}
	}

	return false
}

func (scopestack *Scopestack) FindBlock(key string) ScopeBlock {
	for _, scope := range scopestack.Reverse() {
		for _, value := range scope.Blocks {
			if value.Name.Value == key {
				return ScopeBlock{
					Block:     &value,
					Foreign:   scope.Foreign,
					Immutable: scope.Immutable,
				}
			}
		}
	}

	return ScopeBlock{
		Block:     nil,
		Foreign:   false,
		Immutable: false,
	}
}

func (scopestack *Scopestack) GetCurrentScope() *Scope {
	return &scopestack.Scopes[len(scopestack.Scopes)-1]
}

type ScopeBlock struct {
	Block     *ast.BlockDeclarationStatement
	Foreign   bool
	Immutable bool
}

type ScopeValue struct {
	Value     *Value
	Foreign   bool
	Immutable bool
}

type Scope struct {
	Immutable bool                            `json:"immutable"`
	Foreign   bool                            `json:"foreign"`
	Frame     []Value                         `json:"frame"`
	Blocks    []ast.BlockDeclarationStatement `json:"blocks"`
}

type Value struct {
	Key   ast.Identifier             `json:"key"`
	Value ast.IntPrimitiveExpression `json:"value"`
	Kind  string                     `json:"kind"`
}

type UpdateReport int
