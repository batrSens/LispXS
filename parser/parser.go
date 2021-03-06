package parser

import (
	"fmt"

	ex "github.com/batrSens/LispXS/expressions"
	"github.com/batrSens/LispXS/lexer"
)

type ParseError struct {
	Got, want int
	message   string
	coords    lexer.Coords
}

func NewParseErr(got, want int, message string, coords lexer.Coords) *ParseError {
	return &ParseError{
		Got:     got,
		want:    want,
		message: message,
		coords:  coords,
	}
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("got: %d, want: %d, message: %s, coords: %+v", pe.Got, pe.want, pe.message, pe.coords)
}

// PROGRAM ::= INNER eof
// LIST    ::= ( INNER )
// INNER   ::= ELEM INNER | .
// ELEM    ::= ' ELEM | , ELEM | number | symbol | LIST

type Parser struct {
	curToken *lexer.Token
	lexer    *lexer.Lexer
}

func NewParser(text string) *Parser {
	return &Parser{
		lexer: lexer.NewLexer(text),
	}
}

// PROGRAM ::= INNER eof
func (p *Parser) Parse() (*ex.Expr, error) {
	err := p.nextToken()
	if err != nil {
		return nil, err
	}

	prog, err := p.parseInner()
	if err != nil {
		return nil, err
	}

	err = p.expect(lexer.TagEOF)
	if err != nil {
		return nil, err
	}

	return prog, nil
}

// LIST ::= ( INNER )
func (p *Parser) parseList() (*ex.Expr, error) {
	err := p.expect(lexer.TagLPar)
	if err != nil {
		return nil, err
	}

	res, err := p.parseInner()
	if err != nil {
		return nil, err
	}

	err = p.expect(lexer.TagRPar)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// INNER ::= ELEM INNER | .
func (p *Parser) parseInner() (*ex.Expr, error) {
	if p.curToken.Tag != lexer.TagRPar && p.curToken.Tag != lexer.TagEOF {
		elem, err := p.parseElem()
		if err != nil {
			return nil, err
		}

		list, err := p.parseInner()
		if err != nil {
			return nil, err
		}

		res := elem.Cons(list)
		if res.Type == ex.Fatal {
			return nil, ex.NewExprError(res.String)
		}

		return res, nil
	}

	return ex.NewNil(), nil
}

// ELEM ::= ' ELEM | , ELEM | number | symbol | LIST
func (p *Parser) parseElem() (*ex.Expr, error) {
	var res *ex.Expr

	switch p.curToken.Tag {
	case lexer.TagQuote:
		err := p.expect(lexer.TagQuote)
		if err != nil {
			return nil, err
		}

		expr, err := p.parseElem()
		if err != nil {
			return nil, err
		}

		res := ex.NewFunction("quote").Cons(expr.ToList())
		if res.Type == ex.Fatal {
			return nil, ex.NewExprError(res.String)
		}

		return res, nil
	case lexer.TagComma:
		err := p.expect(lexer.TagComma)
		if err != nil {
			return nil, err
		}

		expr, err := p.parseElem()
		if err != nil {
			return nil, err
		}

		expr.CalculatedForMacro = true

		return expr, nil
	case lexer.TagNumber:
		res = ex.NewNumber(p.curToken.Number)
	case lexer.TagSymbol:
		res = ex.NewSymbol(p.curToken.String)
	case lexer.TagLPar:
		return p.parseList()
	default:
		return nil, NewParseErr(p.curToken.Tag, -1, "multi unexpected", p.curToken.Coords)
	}

	err := p.nextToken()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *Parser) expect(expected int) error {
	if p.curToken.Tag != expected {
		return NewParseErr(p.curToken.Tag, expected, "unexpected", p.curToken.Coords)
	}

	return p.nextToken()
}

func (p *Parser) nextToken() error {
	curTok, err := p.lexer.NextToken()
	if err != nil {
		return err
	}

	p.curToken = curTok
	return nil
}
