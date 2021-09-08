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

func (scopestack *Scopestack) PopScope() *Scope {
	result := Scopestack{}
	scope := scopestack.Scopes[len(scopestack.Scopes)-1]
	result.Scopes = append(result.Scopes, scopestack.Scopes[:len(scopestack.Scopes)-1]...)

	scopestack.Scopes = result.Scopes
	return &scope
}

func (scopestack *Scopestack) AddVariable(value Value) {
	scopestack.GetCurrentScope().Frame = append(scopestack.GetCurrentScope().Frame, value)
}

func (scopestack *Scopestack) IsVariableUpdatable(key string) UpdateReport {
	if !scopestack.VariableExists(key) {
		return 1
	} else {
		scope_value := scopestack.FindVariable(key)
		if scope_value.Immutable || scope_value.Value.Kind == "const" {
			return 2
		} else {
			if scope_value.OuterScope {
				return 0
			}
			if scope_value.Foreign {
				return 3
			} else {
				return 0
			}
		}
	}
}

func (scopestack *Scopestack) UpdateVariable(key string, value ast.IntPrimitiveExpression) {
	var i int
	var j int
	for k, scope := range scopestack.Reverse() {
		for v, value := range scope.Frame {
			if value.Key.Value == key {
				i = k
				j = v
			}
		}
	}

	scopestack.Scopes[i].Frame[j].Value = value
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
	for i, scope := range scopestack.Reverse() {
		for _, value := range scope.Frame {
			if value.Key.Value == key {
				return ScopeValue{
					Value:      &value,
					Foreign:    scope.Foreign,
					Immutable:  scope.Immutable,
					OuterScope: i != 0,
				}
			}
		}
	}

	return ScopeValue{
		Value:      nil,
		Foreign:    false,
		Immutable:  false,
		OuterScope: false,
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
	for i, scope := range scopestack.Reverse() {
		for _, value := range scope.Blocks {
			if value.Name.Value == key {
				return ScopeBlock{
					Block:      &value,
					Foreign:    scope.Foreign,
					Immutable:  scope.Immutable,
					OuterScope: i != 0,
				}
			}
		}
	}

	return ScopeBlock{
		Block:      nil,
		Foreign:    false,
		Immutable:  false,
		OuterScope: false,
	}
}

func (scopestack *Scopestack) GetCurrentScope() *Scope {
	return &scopestack.Scopes[len(scopestack.Scopes)-1]
}

type ScopeBlock struct {
	Block      *ast.BlockDeclarationStatement `json:"block"`
	Foreign    bool                           `json:"foreign"`
	Immutable  bool                           `json:"immutable"`
	OuterScope bool                           `json:"outer_scope"`
}

type ScopeValue struct {
	Value      *Value `json:"value"`
	Foreign    bool   `json:"foreign"`
	Immutable  bool   `json:"immutable"`
	OuterScope bool   `json:"outer_scope"`
}

type Value struct {
	Key   ast.Identifier             `json:"key"`
	Value ast.IntPrimitiveExpression `json:"value"`
	Kind  string                     `json:"kind"`
}

type Scope struct {
	Immutable bool                            `json:"immutable"`
	Foreign   bool                            `json:"foreign"`
	Frame     []Value                         `json:"frame"`
	Blocks    []ast.BlockDeclarationStatement `json:"blocks"`
}

type UpdateReport int
