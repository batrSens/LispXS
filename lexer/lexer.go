package lexer

import (
	"fmt"
	"math"
	"unicode"
)

const (
	TagNumber = iota
	TagSymbol
	TagLPar
	TagRPar
	TagQuote
	TagComma
	TagEOF
)

type Coords struct {
	Cursor, Line, Column int
}

func NewCoords() Coords {
	return Coords{
		Cursor: 0,
		Line:   1,
		Column: 1,
	}
}

type Token struct {
	Coords Coords
	Tag    int
	String string
	Number float64
}

type LexError struct {
	Coords  Coords
	Message string
}

func (le LexError) Error() string {
	return fmt.Sprintf("%s at %+v", le.Message, le.Coords)
}

type Lexer struct {
	text   []rune
	coords Coords
}

func NewLexer(text string) *Lexer {
	runes := append([]rune(text), '\n')

	return &Lexer{
		text:   runes,
		coords: NewCoords(),
	}
}

func (l *Lexer) NextToken() (*Token, error) {
	if l.eof() {
		return l.token(TagEOF), nil
	}

	for !l.eof() && unicode.IsSpace(l.getCurrentChar()) {
		l.moveCursor()
	}

	if l.eof() {
		return l.token(TagEOF), nil
	}

	var res *Token

	switch l.getCurrentChar() {
	case ';':
		{
			for l.getCurrentChar() != '\n' {
				l.moveCursor()
			}

			return l.NextToken()
		}
	case '|':
		{
			sym, err := l.parseStrWithBorder('|')
			if err != nil {
				return nil, err
			}

			return l.tokenString(TagSymbol, sym), nil
		}
	case '(':
		res = l.token(TagLPar)
	case ')':
		res = l.token(TagRPar)
	case '\'':
		res = l.token(TagQuote)
	case ',':
		res = l.token(TagComma)
	default:
		return l.parseSymbolOrNumber()
	}

	l.moveCursor()

	return res, nil
}

func (l *Lexer) parseStrWithBorder(border rune) (string, error) {
	var str []rune

	l.moveCursor()

	c := l.getCurrentChar()
	for c != border && c != '\n' {
		if c == '\\' {
			l.moveCursor()
			c = l.getCurrentChar()
			if c == border || c == '\\' {
				str = append(str, c)
			} else if c == 'n' {
				str = append(str, '\n')
			} else if c == 't' {
				str = append(str, '\t')
			} else {
				return "", l.lexError("unexpected character after '\\'")
			}
		} else {
			str = append(str, c)
		}

		l.moveCursor()
		c = l.getCurrentChar()
	}

	if c == '\n' {
		return "", l.lexError("couldn't find end of string")
	}

	l.moveCursor()

	return string(str), nil
}

func (l *Lexer) parseSymbolOrNumber() (*Token, error) {
	start := l.coords.Cursor
	leftSign := 1.0
	if l.getCurrentChar() == '-' {
		leftSign = -1.0
		l.moveCursor()
	}

	leftPart, leftLen := l.getNumber()
	if leftLen == 0 {
		return l.parseSymbol(start)
	}

	res := leftSign * float64(leftPart)
	if l.getCurrentChar() == '/' {
		l.moveCursor()

		rightPart, rightLen := l.getNumber()
		if rightLen == 0 {
			return l.parseSymbol(start)
		}

		res /= float64(rightPart)
		if l.isWSOrPar() {
			return l.tokenNumber(TagNumber, res), nil
		} else {
			return l.parseSymbol(start)
		}
	}

	if l.getCurrentChar() == '.' {
		l.moveCursor()

		rightPart, rightLen := l.getNumber()
		if rightLen == 0 {
			return l.parseSymbol(start)
		}

		res += leftSign * float64(rightPart) * math.Pow10(-rightLen)
	}

	if l.getCurrentChar() == 'e' {
		l.moveCursor()

		rightSign := 1
		if l.getCurrentChar() == '-' {
			rightSign = -1
			l.moveCursor()
		}

		rightPart, rightLen := l.getNumber()
		if rightLen == 0 {
			return l.parseSymbol(start)
		}

		res *= math.Pow10(rightSign * rightPart)
	}

	if l.isWSOrPar() {
		return l.tokenNumber(TagNumber, res), nil
	}

	return l.parseSymbol(start)
}

func (l *Lexer) parseSymbol(start int) (*Token, error) {
	for !l.isWSOrPar() {
		l.moveCursor()
	}

	res := string(l.text[start:l.coords.Cursor])
	return l.tokenString(TagSymbol, res), nil
}

func (l *Lexer) isWSOrPar() bool {
	c := l.getCurrentChar()
	return unicode.IsSpace(c) || c == '(' || c == ')'
}

func (l *Lexer) getNumber() (int, int) {
	res, length := 0, 0

	for unicode.IsDigit(l.getCurrentChar()) {
		length++
		res *= 10
		res += int(l.getCurrentChar()) - '0'
		l.moveCursor()
	}

	return res, length
}

func (l *Lexer) getCurrentChar() rune {
	return l.text[l.coords.Cursor]
}

func (l *Lexer) eof() bool {
	return len(l.text) <= l.coords.Cursor
}

func (l *Lexer) moveCursor() {
	if len(l.text) <= l.coords.Cursor {
		panic("eof")
	}

	if l.getCurrentChar() == '\n' {
		l.coords.Line++
		l.coords.Column = 1
	} else {
		l.coords.Column++
	}

	l.coords.Cursor++
}

func (l *Lexer) token(tag int) *Token {
	return &Token{
		Coords: l.coords,
		Tag:    tag,
	}
}

func (l *Lexer) tokenString(tag int, str string) *Token {
	return &Token{
		Coords: l.coords,
		Tag:    tag,
		String: str,
	}
}

func (l *Lexer) tokenNumber(tag int, num float64) *Token {
	return &Token{
		Coords: l.coords,
		Tag:    tag,
		Number: num,
	}
}

func (l *Lexer) lexError(msg string) *LexError {
	return &LexError{
		Coords:  l.coords,
		Message: msg,
	}
}
