package ast

type ParserResult struct {
	Error   bool        `json:"error"`
	Content interface{} `json:"content"`
}

type ErrorContent struct {
	Message  string   `json:"message"`
	Position Position `json:"position"`
}

type Program struct {
	Imports []UseStatement `json:"imports"`
	Program []interface{}  `json:"program"`
}

type UseStatement struct {
	Source   StringPrimitiveExpression `json:"source"`
	Position Position                  `json:"position"`
}

// export type Main = Statement | BlockCallExpression | Comment;
type Main []interface{}

type VariableDeclarationStatement struct {
	Operation string                 `json:"operation"`
	Kind      string                 `json:"kind"`
	Left      Identifier             `json:"left"`
	Right     map[string]interface{} `json:"right"`
	Position  Position               `json:"position"`
}

type BlockDeclarationStatement struct {
	Operation    string       `json:"operation"`
	Owner        string       `json:"owner"`
	Name         Identifier   `json:"name"`
	Verbs        []Identifier `json:"verbs"`
	Arguments    []Identifier `json:"arguments"`
	Body         interface{}  `json:"body,omitempty"`
	Implementing bool         `json:"implementing"`
	Implements   Identifier   `json:"implements"`
	Populate     interface{}  `json:"populate"`
	Position     Position     `json:"position"`
	Instance     interface{}  `json:"instance"`
	Native       bool         `json:"native"`
}

type NativeFunction func(arguments []IntPrimitiveExpression, verbs []IntPrimitiveExpression) NativeFunctionReturn
type NativeFunctionReturn struct {
	Value   IntPrimitiveExpression `json:"value"`
	Error   bool                   `json:"error"`
	Warn    bool                   `json:"warn"`
	Message string                 `json:"string"`
}

type ForStatement struct {
	Operation   string                 `json:"operation"`
	Statement   map[string]interface{} `json:"statement"`
	Placeholder string                 `json:"placeholder"`
	Body        []interface{}          `json:"body"`
	Position    Position               `json:"position"`
}

type SwitchStatement struct {
	Operation string                 `json:"operation"`
	Condition map[string]interface{} `json:"condition"`
	Cases     []SwitchCase           `json:"cases"`
	Default   SwitchCase             `json:"default"`
	Position  Position               `json:"position"`
}

type SwitchCase struct {
	Case map[string]interface{} `json:"case"`
	Body []interface{}          `json:"body"`
}

type WhileStatement struct {
	Operation string                 `json:"operation"`
	Statement map[string]interface{} `json:"statement"`
	Body      []interface{}          `json:"body"`
	Position  Position               `json:"position"`
}

type IfStatement struct {
	Operation string                 `json:"operation"`
	Condition map[string]interface{} `json:"condition"`
	Body      []interface{}          `json:"body"`
	Else      []interface{}          `json:"else"`
	Elifs     []Elif                 `json:"elifs"`
	Position  Position               `json:"position"`
}

type Elif struct {
	Condition map[string]interface{} `json:"condition"`
	Body      []interface{}          `json:"body"`
}

type ReturnStatement struct {
	Operation  string     `json:"operation"`
	Expression Expression `json:"expression"`
	Position   Position   `json:"position"`
}

type ThrowStatement struct {
	Operation  string     `json:"operation"`
	Expression Expression `json:"expression"`
	Position   Position   `json:"position"`
}

type AssignStatement struct {
	Operation string                 `json:"operation"`
	Left      Identifier             `json:"leftomitemptyomitempty"`
	Right     map[string]interface{} `json:"right"`
	Position  Position               `json:"position"`
}

type QuantityModifierStatement struct {
	Operation string                 `json:"operation"`
	Type      string                 `json:"type"`
	Statement map[string]interface{} `json:"statement"`
	Right     map[string]interface{} `json:"right"`
	Position  Position               `json:"position"`
}

type BlockBody struct {
	Init    []interface{} `json:"init"`
	Program []interface{} `json:"program"`
}

type Expression struct {
	Operation string   `json:"operation"`
	Position  Position `json:"position"`
}

type ConditionExpression struct {
	Operation string     `json:"operation"`
	Type      string     `json:"type"`
	Left      Expression `json:"left"`
	Right     Expression `json:"right"`
	Position  Position   `json:"position"`
}

type ReferenceExpression struct {
	Operation string   `json:"operation"`
	Negative  bool     `json:"negative"`
	Value     string   `json:"value"`
	Position  Position `json:"position"`
}

type ArithmeticExpression struct {
	Operation string     `json:"operation"`
	Type      string     `json:"type"`
	Left      Expression `json:"left"`
	Right     Expression `json:"right"`
	Position  Position   `json:"position"`
}

type BlockCallExpression struct {
	Operation string                   `json:"operation"`
	Name      Identifier               `json:"name"`
	Verbs     []map[string]interface{} `json:"verbs"`
	Arguments []map[string]interface{} `json:"arguments"`
	Position  Position                 `json:"position"`
}

type ScopeMutaterExpression struct {
	Operation string                   `json:"operation"`
	Mutater   MutaterKeyword           `json:"mutater"`
	Arguments []map[string]interface{} `json:"arguments"`
	Position  Position                 `json:"position"`
}

type MutaterKeyword struct {
	Operation string   `json:"operation"`
	Negative  bool     `json:"negative"`
	Value     string   `json:"value"`
	Position  Position `json:"position"`
}

type Identifier struct {
	Operation string   `json:"operation"`
	Negative  bool     `json:"negative"`
	Value     string   `json:"value"`
	Position  Position `json:"position"`
}

type PrimitiveExpression struct {
}

type StringPrimitiveExpression struct {
	Operation string   `json:"operation"`
	Type      string   `json:"type"`
	Value     string   `json:"value"`
	Position  Position `json:"position"`
}

type IntPrimitiveExpression struct {
	Operation string   `json:"operation"`
	Type      string   `json:"type"`
	Value     int64    `json:"value"`
	Position  Position `json:"position"`
}

type ArrayPrimitiveExpression struct {
	Operation string                   `json:"operation"`
	Type      string                   `json:"type"`
	Values    []map[string]interface{} `json:"values"`
	Position  Position                 `json:"position"`
}

type Position struct {
	Line uint32 `json:"line"`
	Col  uint32 `json:"col"`
}

type Comment struct {
	Operation string   `json:"operation"`
	Value     string   `json:"value"`
	Position  Position `json:"position"`
}
