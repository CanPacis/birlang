@{%

import m from "https://dev.jspm.io/moo"
import n from "https://dev.jspm.io/nearley";

const nearley = n as Nearley;
const moo = m as { compile: Function; keywords: Function }

interface Nearley {
  Parser: { new (grammar: any): any };
  Grammar: { fromCompiled: (grammar: Grammar) => any };
  Rule: { new (): any };
}

export class BirParser {
  parser: any;

  constructor() {
    this.parser = new nearley.Parser(nearley.Grammar.fromCompiled(grammar));
  }

  parse(input: string) {
    this.parser.feed(input);

    if (this.parser.results.length > 1) {
      console.error("Grammar is ambigious.");
    }

    return this.parser.results[0];
  }
}

const lexer = moo.compile({
  WhiteSpace: { match: /[ \t\n\r]+/, lineBreaks: true },
  NumberLiteral: {
    match: /-?[0-9]+(?:\.[0-9]+)?/,
  },
  BinaryLiteral: {
    match: /-?@b[0-1]+/
  },
  HexLiteral: {
    match: /-?@x[0-9a-fA-F]+/
  },
  OctalLiteral: {
    match: /-?@o[0-7]+/
  },
  Dot: ".",
  Colon: ":",
  Comma: ",",
  LeftParens: "(",
  RightParens: ")",
  LeftCurlyBrackets: "{",
  RightCurlyBrackets: "}",
  LeftBrackets: "[",
  RightBrackets: "]",
  Plus: "+",
  Multiplier: "*",
  QuestionMark: "?",
  Caret: "^",
  Apostrophe: "'",
  Percent: "%",
  ConditionSign: {
    match: /!?(?:&&|\|\||<=|>=|<|>|==)/,
  },
  Ampersand: "&",
  EqualsTo: "=",
	Pound: "#",
  StringLiteral: {
    match: /"(?:[^\n\\"]|\\["\\ntbfr])*"/,
    value: (s: string) => JSON.parse(s)
  },
	Comment: {
    match: /#[^\n]*/,
    value: (s: string) => s
  },
  Divider: "/",
  Identifier: {
    match: /-?[a-zA-Z_][a-zA-Z_0-9]*/,
    type: moo.keywords({
      Use: "use",
      Const: "const",
      Let: "let",
      Init: "init",
      Return: "return",
      Throw: "throw",
			Debugger: "debugger",
			Implements: "implements",
  		If: "if",
      Elif: "elif",
      Else: "else",
  		Switch: "switch",
  		Default: "default",
  		For: "for",
  		While: "while",
  		As: "as",
      Case: "case",
      Log: "log",
      Read: "Read",
      Write: "Write",
      Delete: "Delete",
    })
  },
  Minus: "-"
})

function getCondition(condition: string) {
  switch(condition) {
    case "&&":
      return "and";
    case "||":
      return "or";
    case "<":
      return "less_than";
    case ">":
      return "greater_than";
    case "<=":
      return "less_than_equals";
    case ">=":
      return "greater_than_equals";
    case "==":
      return "equals";
    case "!&&":
      return "nand";
    case "!||":
      return "nor";
    case "!<":
      return "not_less_than";
    case "!>":
      return "not_greater_than";
    case "!<=":
      return "not_less_than_equals";
    case "!>=":
      return "not_greater_than_equals";
    case "!==":
      return "not_equals";
  }
}

function arithmetic(type: any, data: any) {
  return { 
    operation: "arithmetic", 
    type, 
    left: data[0], 
    right: data[4] || { type: "int", value: 0, operation: "primitive", position: { col: 0, line: 0 } }, 
    position: data[0].position 
  }
}

function position(data: any) {
  return { line: data.line, col: data.col }
}

%}

@preprocessor typescript
@lexer lexer

Program 
  -> _ (Use _ {% id %}):*
  (Main _ {% id %}):* 
  {% d => ({ imports: d[1], program: d[2] }) %}

Use -> "use" __ string
  {% d => ({ 
    source: d[2],
    position: position(d[0])
  }) %}

Main 
	-> Statement {% id %}
  | BlockCall {% id %}
  | ScopeMutaterExpression {%id%}
	| Comment {% id %}



Statement 
	-> VariableDeclarationStatement {%id%}
	| BlockDeclarationStatement {%id%}
	| ForStatement {%id%}
	| SwitchStatement {%id%}
	| WhileStatement {%id%}
	| IfStatement {%id%}
	| ReturnStatement {%id%}
  | ThrowStatement {%id%}
  | AssignStatement {%id%}
  | QuantityModifierStatement {%id%}

VariableDeclarationStatement -> ("const" {%id%} | "let" {%id%}) __ identifier _ "=" _ Expression
  {% d => ({ operation: "variable_declaration", kind: d[0].value, left: d[2], right: d[6], position: { line: d[0].line, col: d[0].col } }) %}

SwitchStatement -> "switch" __ Expression _ 
  "{" _ ("case" __ Expression _ CodeBlock _ 
  {% d => ({ case: d[2], body: d[4] }) %}):* ("default" _ CodeBlock _ 
  {% d => ({ body: d[2] }) %}):? "}"
  {% d => ({ 
    operation: "switch_statement", 
    condition: d[2],
    cases: d[6],
    default: d[7],
    position: position(d[0])
  }) %}

ForStatement -> "for" __ Expression __ "as" __ identifier _ CodeBlock
  {% d => ({ 
    operation: "for_statement", 
    statement: d[2], 
    placeholder: d[6].value, 
    body: d[8],
    position: position(d[0]) 
  }) %}

WhileStatement -> "while" __ Expression _ CodeBlock
  {% d => ({ 
    operation: "while_statement", 
    statement: d[2], 
    body: d[4],
    position: position(d[0]) 
  }) %}

IfStatement -> "if" __ Expression _ CodeBlock 
  (_ "elif" __ Expression _ CodeBlock 
  {% d => ({ condition: d[3], body: d[5] }) %}):*
  (_ "else" _ CodeBlock {% d => d[3] %}):?
  {% d => ({ 
    operation: "if_statement", 
    condition: d[2], 
    body: d[4],
    elifs: d[5],
    else: d[6],
    position: position(d[0])
  }) %}

QuantityModifierStatement
  -> Mutatable "+" "+" 
    {% d => ({ operation: "quantity_modifier_statement", type: "increment", statement: d[0], position: d[0].position }) %}
  | Mutatable "-" "-"
    {% d => ({ operation: "quantity_modifier_statement", type: "decrement", statement: d[0], position: d[0].position }) %}
  | Mutatable _ "+" "=" _ Expression
    {% d => ({ operation: "quantity_modifier_statement", type: "add", statement: d[0], right: d[5], position: d[0].position }) %}
  | Mutatable _ "-" "=" _ Expression
    {% d => ({ operation: "quantity_modifier_statement", type: "subtract", statement: d[0], right: d[5], position: d[0].position }) %}
  | Mutatable _ "*" "=" _ Expression
    {% d => ({ operation: "quantity_modifier_statement", type: "multiply", statement: d[0], right: d[5], position: d[0].position }) %}
  | Mutatable _ "/" "=" _ Expression
    {% d => ({ operation: "quantity_modifier_statement", type: "divide", statement: d[0], right: d[5], position: d[0].position }) %}

AssignStatement -> Mutatable _ "=" _ Expression
  {% d => ({ 
    operation: "assign_statement", 
    left: d[0], 
    right: d[4],
    position: d[0].position 
  }) %}

ReturnStatement -> "return" __ Expression {% d => ({ operation: "return_statement", expression: d[2], position: position(d[0]) }) %}
ThrowStatement -> "throw" __ Expression {% d => ({ operation: "throw_statement", expression: d[2], position: position(d[0]) }) %}

BlockDeclarationStatement -> identifier BlockVerb:* __ "[" _ (ArgumentList _ {% id %}):? "]" _ BlockContent
	{% d => ({ operation: "block_declaration", name: d[0], verbs: d[1], arguments: d[5], body: d[8], position: d[0].position, implementing: false, initialized: false }) %}
  | identifier _ "implements" _ identifier (__ "{" _ (string {% id %}| array {% id %}) _ "}" {% d => d[3] %}):?
  {% d => ({ operation: "block_declaration", name: d[0], implements: d[4], position: d[0].position, implementing: true, populate: d[5], initialized: false }) %}

BlockVerb -> ":" Expression {% d => d[1] %}

MutaterKeyword -> ("Read" {%id%} | "Write" {%id%} | "Delete" {%id%}) 
  {% d => ({ 
    negative: d[0].value[0] === "-", 
    operation: "identifier", 
    value: d[0].value, 
    position: { line: d[0].line, col: d[0].col } 
  }) %}

BlockContent -> "{" _ (BlockInit _ {% id %}):?
  (Main _ {% id %}):* "}" 
  {% d => ({ init: d[2], program: d[3] }) %} 

BlockInit -> "init" _ "{" _ (Main _ {% id %}):* "}" 
  {% d => d[4] %} 


Expression
  -> Conditions {% id %}
  | Arithmetic {% id %}

Arithmetic
  -> Arithmetic _ "+" _ MultDiv {% d => arithmetic("addition", d) %}
  | Arithmetic _ "-" _ MultDiv {% d => arithmetic("subtraction", d) %}
  | MultDiv {% id %}

MultDiv 
  -> MultDiv _ "*" _ Exponent {% d => arithmetic("multiplication", d) %}
  | MultDiv _ "/" _ Exponent {% d => arithmetic("division", d) %}
  | Exponent {% id %}

Exponent 
  -> Caller _ "^" _ Exponent {% d => arithmetic("exponent", d) %}
  | Caller _ "'" _ Exponent {% d => arithmetic("root", d) %}
  | Caller _ "%" _ Exponent {% d => arithmetic("modulus", d) %}
  | Caller _ "log" {% d => arithmetic("log10", d) %}
  | Caller {% id %}

Caller 
  -> BlockCall {% id %}
  | ScopeMutaterExpression {% id %}

ScopeMutaterExpression -> "[" _ MutaterKeyword (__ ArgumentList _ {% d => d[1] %}):? "]"
  {% d => ({ 
    operation: "scope_mutater_expression", 
    mutater: d[2], 
    arguments: d[3],
    position: position(d[0]) 
  }) %}
  | SubExpression {% id %}

BlockCall 
  -> Caller BlockVerb:* _ "(" _ (ArgumentList _ {% id %}):? ")"
    {% d => ({ 
      operation: "block_call", 
      name: d[0], 
      verbs: d[1],
      arguments: d[5] || [],
      position: d[0].position
    }) %}

Conditions -> Expression _ %ConditionSign _ Arithmetic
  {% d => ({ 
    operation: "condition", 
    type: getCondition(d[2].value), 
    left: d[0], 
    right: d[4],
    position: d[0].position 
  }) %}

SubExpression
  -> Primitive {% id %}
  | VariableReference {% id %}
  | Grouping {% id %}

ArgumentList 
  -> Expression _ "," _ ArgumentList {% d => [d[0], ...d[4]] %}
  | Expression {% d => [d[0]] %}

Grouping -> "{" Expression "}"
  {% d => d[1] %}

VariableReference -> identifier 
  {% d => ({ 
    operation: "reference", 
    value: d[0].value,
    negative: d[0].negative,
    position: { line: d[0].position.line, col: d[0].position.col }
  }) %}

Primitive 
  -> number {% id %} 


Comment -> %Comment
  {% d => ({ 
    operation: "comment", 
    value: d[0].value,
    position: position(d[0])
  }) %}


CodeBlock -> "{" _ (Main _ {% id %}):* "}"
  {% d => d[2] %}

Mutatable 
  -> VariableReference {% id %}

array -> "[" (ArgumentList _ {% id %}):? "]" 
  {% d => ({
    operation: "primitive",
    type: "array", 
    values: d[1]?.flat(Number.POSITIVE_INFINITY) || [],
    position: { line: d[0].line, col: d[0].col }
  }) %}

number -> %NumberLiteral {% d => ({ 
    operation: "primitive",
    value: parseInt(d[0].value, 10),
    type: "int",
    position: position(d[0])
  })  %} | %BinaryLiteral {% d => ({ 
    operation: "primitive",
    value: d[0].value[0] === "-" ? -(parseInt(d[0].value.substr(3), 2)) : parseInt(d[0].value.substr(2), 2),
    type: "int",
    position: position(d[0])
  })  %} | %HexLiteral {% d => ({ 
    operation: "primitive",
    value: d[0].value[0] === "-" ? -(parseInt(d[0].value.substr(3), 16)) : parseInt(d[0].value.substr(2), 16),
    type: "int",
    position: position(d[0])
  })  %} | %OctalLiteral {% d => ({ 
    operation: "primitive",
    value: d[0].value[0] === "-" ? -(parseInt(d[0].value.substr(3), 8)) : parseInt(d[0].value.substr(2), 8),
    type: "int",
    position: position(d[0])
  })  %}

string -> %StringLiteral {% d => ({ operation: "primitive", value: d[0].value,position: position(d[0]), type: "string" }) %}

identifier -> %Identifier {% d => ({ negative: d[0].value[0] === "-", operation: "identifier", value: d[0].value[0] === "-" ? d[0].value.substr(1) : d[0].value, position: { line: d[0].line, col: d[0].col } }) %}

_ -> [\s]:*     {% (d) =>  null %}
__ -> [\s]:+     {% (d) =>  null %}