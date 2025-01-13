<Program> ::= <StatementList>

<StatementList> ::= <Statement> <StatementList>
| ε

<Statement> ::= <LetStatement>
| <ReturnStatement>
| <ExpressionStatement>

<LetStatement> ::= "let" <Identifier> "=" <Expression> ";"

<ReturnStatement> ::= "return" <Expression> ";"

<ExpressionStatement> ::= <Expression> ";"

<Expression> ::= <PrefixExpression>
| <InfixExpression>
| <GroupedExpression>
| <IfExpression>
| <FunctionLiteral>
| <IntegerLiteral>
| <Identifier>
| <Boolean>
| <StringLiteral>
| <ArrayLiteral>
| <HashLiteral>
| <IndexExpression>
| <CallExpression>

<PrefixExpression> ::= <PrefixOperator> <Expression>

<PrefixOperator> ::= "!" | "-"

<InfixExpression> ::= <Expression> <InfixOperator> <Expression>

<InfixOperator> ::= "+" | "-" | "\*" | "/" | "==" | "!=" | "<" | ">"

<GroupedExpression> ::= "(" <Expression> ")"

<IfExpression> ::= "if" "(" <Expression> ")" <BlockStatement> ["else" <BlockStatement>]

<FunctionLiteral> ::= "fn" "(" <Parameters> ")" <BlockStatement>

<Parameters> ::= <Identifier> ["," <Identifier>]\*
| ε

<IntegerLiteral> ::= <DIGIT>+

<Identifier> ::= <LETTER> <IdentifierPart>\*

<IdentifierPart> ::= <LETTER> | <DIGIT> | "\_"

<Boolean> ::= "true" | "false"

<StringLiteral> ::= "\"" <CHAR>\* "\""

<ArrayLiteral> ::= "[" <ExpressionList> "]"

<HashLiteral> ::= "{" <HashContent> "}"

<HashContent> ::= <Expression> ":" <Expression> ["," <Expression> ":" <Expression>]\*
| ε

<ExpressionList> ::= <Expression> ["," <Expression>]\*
| ε

<IndexExpression> ::= <Expression> "[" <Expression> "]"

<CallExpression> ::= <Expression> "(" <ExpressionList> ")"

<BlockStatement> ::= "{" <StatementList> "}"

<DIGIT> ::= "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9"

<LETTER> ::= "a" | "b" | ... | "z" | "A" | "B" | ... | "Z"

<CHAR> ::= any character except '"'
