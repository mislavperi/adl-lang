package lexer

import (
	"github.com/mislavperi/gem-lang/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	character    byte
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.advancePastWhitespace()

	switch l.character {
	case '=':
		if l.peekChar() == '=' {
			character := l.character
			l.readChar()
			literal := string(character) + string(l.character)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = token.Token{Type: token.ASSIGN, Literal: string(l.character)}
		}
	case '+':
		tok = token.Token{Type: token.PLUS, Literal: string(l.character)}
	case '-':
		tok = token.Token{Type: token.MINUS, Literal: string(l.character)}
	case '!':
		if l.peekChar() == '=' {
			character := l.character
			l.readChar()
			literal := string(character) + string(l.character)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = token.Token{Type: token.BANG, Literal: string(l.character)}
		}
	case '/':
		tok = token.Token{Type: token.SLASH, Literal: string(l.character)}
	case '*':
		tok = token.Token{Type: token.ASTERISK, Literal: string(l.character)}
	case '<':
		tok = token.Token{Type: token.LT, Literal: string(l.character)}
	case '>':
		tok = token.Token{Type: token.GT, Literal: string(l.character)}
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Literal: string(l.character)}
	case '(':
		tok = token.Token{Type: token.LPAREN, Literal: string(l.character)}
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: string(l.character)}
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: string(l.character)}
	case '{':
		tok = token.Token{Type: token.LBRACE, Literal: string(l.character)}
	case '}':
		tok = token.Token{Type: token.RBRACE, Literal: string(l.character)}
	case '"':
		tok = token.Token{Type: token.STRING, Literal: l.readString()}
	case '[':
		tok = token.Token{Type: token.LBRACKET, Literal: string(l.character)}
	case ']':
		tok = token.Token{Type: token.RBRACKET, Literal: string(l.character)}
	case ':':
		tok = token.Token{Type: token.COLON, Literal: string(l.character)}
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isLetter(l.character) {
			tok.Literal = l.consumeLiteral()
			tok.Type = token.LookupIdentifier(tok.Literal)
			return tok
		} else if isDigit(l.character) {
			tok = token.Token{Type: token.INT, Literal: l.readNumber()}
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.character)}
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) consumeLiteral() string {
	position := l.position
	for isLetter(l.character) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.character = 0
	} else {
		l.character = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.character) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.character == '"' || l.character == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) advancePastWhitespace() {
	for {
		if !isWhitespace(l.character) {
			break
		}
		l.readChar()
	}
}

func isWhitespace(character byte) bool {
	return character == 32 || character == 9 || character == 10 || character == 13
}

func isLetter(character byte) bool {
	return (character >= 65 && character <= 90) || (character >= 97 && character <= 122) || character == 95
}

func isDigit(character byte) bool {
	return character >= 48 && character <= 57
}
