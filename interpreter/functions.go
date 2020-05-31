package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	ex "github.com/batrSens/LispXS/expressions"
	"github.com/batrSens/LispXS/lexer"
	"github.com/batrSens/LispXS/parser"
)

const (
	ModOr = iota
	ModAnd
	ModIf
	ModExec
	ModTry
	ModMacro
)

type Mod struct {
	Type int
	Exec map[int]struct{}
	Old  *Mod
}

func modApply(ir *interpreter) bool {
	switch ir.mod.Type {
	case ModOr:
		if ir.argsNum > 2 && !ir.dataStack.Last().IsNil() {
			ir.dataStack.Push(ex.NewT())
			return true
		}
	case ModAnd:
		if ir.argsNum > 2 && ir.dataStack.Last().IsNil() {
			ir.dataStack.Push(ex.NewNil())
			return true
		}
	case ModIf:
		if ir.argsNum == 3 && ir.dataStack.Last().IsNil() ||
			ir.argsNum == 4 && !ir.dataStack.PreLast().IsNil() || ir.argsNum > 4 {
			ir.dataStack.Push(ex.NewNil())
			return true
		}
	case ModExec:
		if _, ok := ir.mod.Exec[ir.argsNum-1]; !ok {
			ir.dataStack.Push(ir.getCurSymbol())
			return true
		}
	case ModTry:
		if ir.argsNum < 2 {
			panic("it shouldn't have happened " + strconv.Itoa(ir.argsNum))
		} else if ir.argsNum > 2 {
			ir.dataStack.Push(ex.NewNil())
			return true
		}
	default:
		panic("unexpected mod " + strconv.Itoa(ir.mod.Type))
	}
	return false
}

type Func struct {
	F   func(ir *interpreter, args []*ex.Expr) *ex.Expr
	Mod *Mod
}

var functions = map[string]Func{

	"eval": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFunction("begin").Cons(ex.NewFatal("quote: must be 1 argument").ToList())
			}

			return ex.NewFunction("begin").Cons(args[0].ToList())
		},
	},

	"quote": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("quote: must be 1 argument")
			}

			return args[0]
		},
		Mod: &Mod{
			Type: ModExec,
			Exec: map[int]struct{}{},
		},
	},

	"catch": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 {
				return ex.NewFatal("catch: must be at least one argument")
			}

			if args[0].Type == ex.Fatal {
				panic("catch: it shouldn't have happened")
			}

			return args[0]
		},
		Mod: &Mod{
			Type: ModTry,
		},
	},

	"throw": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 && len(args) != 2 {
				return ex.NewFatal("throw: must be one or two arguments")
			}

			if args[0].Type != ex.Symbol {
				return ex.NewFatal("throw: first argument must be a symbol")
			}

			if len(args) == 2 {
				return ex.NewFatal(args[0].String, args[1])
			}

			return ex.NewFatal(args[0].String)
		},
	},

	"car": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("car: must be 1 argument")
			}

			return args[0].Car()
		},
	},

	"cdr": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("cdr: must be 1 argument")
			}

			return args[0].Cdr()
		},
	},

	"cons": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 2 {
				return ex.NewFatal("cons: must be 2 arguments")
			}

			return args[0].Cons(args[1])
		},
	},

	"define": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 2 {
				return ex.NewFatal("define: must be 2 arguments")
			}

			if args[0].Type != ex.Symbol {
				return ex.NewFatal("define: first argument is not a symbol")
			}

			ir.varsEnvironment.CurSymbols[args[0].String] = args[1]
			return args[1]
		},
		Mod: &Mod{
			Type: ModExec,
			Exec: map[int]struct{}{2: {}},
		},
	},

	"defmacro": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) < 3 {
				return ex.NewFatal("defmacro: must be at less 3 arguments")
			}

			if args[0].Type != ex.Symbol {
				return ex.NewFatal("defmacro: second argument is not a symbol")
			}

			macro := ex.NewMacro(args[1], args[2:], ir.varsEnvironment)
			ir.varsEnvironment.CurSymbols[args[0].String] = macro

			return macro
		},
		Mod: &Mod{
			Type: ModExec,
			Exec: map[int]struct{}{},
		},
	},

	"set!": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 2 {
				return ex.NewFatal("set!: must be 2 arguments")
			}

			if args[0].Type != ex.Symbol {
				return ex.NewFatal("set!: second argument is not symbol")
			}

			curEnv := ir.varsEnvironment
			for curEnv != nil {
				if _, ok := curEnv.CurSymbols[args[0].String]; ok {
					curEnv.CurSymbols[args[0].String] = args[1]
					return args[1]
				}
				curEnv = ir.varsEnvironment.Parent
			}

			return ex.NewFatal("set!: symbol '" + args[0].String + "' is not defined")
		},
		Mod: &Mod{
			Type: ModExec,
			Exec: map[int]struct{}{2: {}},
		},
	},

	"lambda": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) < 2 {
				return ex.NewFatal("lambda: must be at less 2 arguments")
			}

			return ex.NewClosure(args[0], args[1:], ir.varsEnvironment)
		},
		Mod: &Mod{
			Type: ModExec,
			Exec: map[int]struct{}{},
		},
	},

	"begin": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 {
				return ex.NewNil()
			}
			return args[len(args)-1]
		},
	},

	"or": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			for _, arg := range args {
				if !arg.IsNil() {
					return arg
				}
			}

			return ex.NewNil()
		},
		Mod: &Mod{
			Type: ModOr,
		},
	},

	"and": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 {
				return ex.NewT()
			}

			return args[len(args)-1]
		},
		Mod: &Mod{
			Type: ModAnd,
		},
	},

	"if": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 2 && len(args) != 3 {
				return ex.NewFatal(fmt.Sprintf("if: expected 2 or 3 expressions, got %d", len(args)))
			}

			if !args[0].IsNil() {
				return args[1]
			}

			if len(args) == 2 {
				return ex.NewNil()
			}

			return args[2]
		},
		Mod: &Mod{
			Type: ModIf,
		},
	},

	">": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 2 {
				return ex.NewFatal(fmt.Sprintf(">: expected 2 expressions, got %d", len(args)))
			}

			for _, arg := range args {
				if arg.Type != ex.Number {
					return ex.NewFatal(">: expected numbers")
				}
			}

			if args[0].Number > args[1].Number {
				return ex.NewT()
			}

			return ex.NewNil()
		},
	},

	"<": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 2 {
				return ex.NewFatal(fmt.Sprintf("<: expected 2 expressions, got %d", len(args)))
			}

			for _, arg := range args {
				if arg.Type != ex.Number {
					return ex.NewFatal("<: expected numbers")
				}
			}

			if args[0].Number < args[1].Number {
				return ex.NewT()
			}

			return ex.NewNil()
		},
	},

	"=": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) < 2 {
				return ex.NewFatal(fmt.Sprintf("=: expected at less 2 expressions, got %d", len(args)))
			}

			cur := args[0]
			for _, arg := range args[1:] {
				if !cur.Equal(arg) {
					return ex.NewNil()
				}
			}

			return ex.NewT()
		},
	},

	"not": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("not: must be 1 argument")
			}

			if args[0].IsNil() {
				return ex.NewT()
			}

			return ex.NewNil()
		},
	},

	"pair?": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("pair?: must be 1 argument")
			}

			if args[0].Type == ex.Pair {
				return ex.NewT()
			}

			return ex.NewNil()
		},
	},

	"number?": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("number?: must be 1 argument")
			}

			if args[0].Type == ex.Number {
				return ex.NewT()
			}

			return ex.NewNil()
		},
	},

	"symbol?": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("symbol?: must be 1 argument")
			}

			if args[0].Type == ex.Symbol {
				return ex.NewT()
			}

			return ex.NewNil()
		},
	},

	"len": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("len: must be 1 argument")
			}

			if args[0].Type != ex.Symbol {
				return ex.NewFatal("len: must be a symbol")
			}

			return ex.NewNumber(float64(len([]rune(args[0].String))))
		},
	},

	"symbol->number": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("symbol->number: must be 1 argument")
			}

			if args[0].Type != ex.Symbol {
				return ex.NewFatal("symbol->number: must be a symbol")
			}

			lex := lexer.NewLexer(args[0].String)

			tok, err := lex.NextToken()
			if err != nil || tok.Tag != lexer.TagNumber {
				return ex.NewFatal("string->number: incorrect string")
			}

			tok2, err := lex.NextToken()
			if err != nil || tok2.Tag != lexer.TagEOF {
				return ex.NewFatal("string->number: incorrect string")
			}

			return ex.NewNumber(tok.Number)
		},
	},

	"number->symbol": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("number->symbol: must be 1 argument")
			}

			if args[0].Type != ex.Number {
				return ex.NewFatal("number->symbol: must be a number")
			}

			return ex.NewSymbol(strconv.FormatFloat(args[0].Number, 'f', -1, 64))
		},
	},

	"+": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 {
				return ex.NewNumber(0.0)
			}

			switch args[0].Type {
			case ex.Number:
				res := 0.0
				for _, arg := range args {
					if arg.Type != ex.Number {
						return ex.NewFatal("+: expected numbers, given " + arg.ToString())
					}
					res += arg.Number
				}
				return ex.NewNumber(res)

			case ex.Symbol:
				res := ""
				for _, arg := range args {
					if arg.Type != ex.Symbol {
						return ex.NewFatal("+: expected symbols")
					}
					res += arg.String
				}
				return ex.NewSymbol(res)
			default:
				return ex.NewFatal("+: expected numbers or symbols, given: " + args[0].ToString())
			}
		},
	},

	"-": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 {
				return ex.NewNumber(0.0)
			}

			if len(args) == 1 {
				if args[0].Type != ex.Number {
					return ex.NewFatal("-: expected numbers")
				}

				return ex.NewNumber(-args[0].Number)
			}

			if args[0].Type == ex.Symbol {
				if len(args) != 3 {
					return ex.NewFatal("-: expected 3 arguments")
				}

				if args[1].Type != ex.Number || args[2].Type != ex.Number {
					return ex.NewFatal("-: expected 2 last numbers")
				}

				if args[1].Number > args[2].Number {
					return ex.NewFatal("-: first number must be less or equal than second")
				}

				runes := []rune(args[0].String)

				if args[1].Number < 0 || args[2].Number > float64(len(runes)) {
					return ex.NewFatal("-: incorrect range")
				}

				return ex.NewSymbol(string(runes[int(args[1].Number):int(args[2].Number)]))
			}

			res := 0.0

			for i, arg := range args {
				if arg.Type != ex.Number {
					return ex.NewFatal("-: expected numbers")
				}
				if i == 0 {
					res = arg.Number
				} else {
					res -= arg.Number
				}
			}

			return ex.NewNumber(res)
		},
	},

	"*": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			res := 1.0

			for _, arg := range args {
				if arg.Type != ex.Number {
					return ex.NewFatal("*: expected numbers")
				}
				res *= arg.Number
			}

			return ex.NewNumber(res)
		},
	},

	"/": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 || args[0].Type != ex.Number {
				return ex.NewFatal("/: expected at least one number")
			}

			res := args[0].Number

			for _, arg := range args[1:] {

				if arg.Type != ex.Number {
					return ex.NewFatal("/: expected numbers")
				} else if arg.Number == 0 {
					return ex.NewFatal("/: zero division")
				}

				res /= arg.Number
			}

			return ex.NewNumber(res)
		},
	},

	"write": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("write: expected one expression")
			}

			_, err := fmt.Fprint(ir.stdout, args[0].ToString())
			if err != nil {
				return ex.NewFatal(err.Error())
			}

			return args[0]
		},
	},

	"read": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 0 {
				return ex.NewFatal("read: expected zero expressions")
			}

			var expr *ex.Expr
			var exprStr string
			for {
				str, err := bufio.NewReader(ir.stdin).ReadString('\n')
				if err != nil && err != io.EOF {
					return ex.NewFatal("read: " + err.Error())
				}

				exprStr += str

				expr, err = parser.NewParser(exprStr).Parse()
				if err != nil {
					if pErr, ok := err.(*parser.ParseError); ok && pErr.Got == lexer.TagEOF {
						continue
					}
					return ex.NewFatal("read: " + err.Error())
				} else {
					break
				}
			}

			return expr
		},
	},

	"load": {
		F: func(ir *interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("load: expected one expression")
			}

			if args[0].Type != ex.Symbol {
				return ex.NewFatal("load: expected zero expressions")
			}

			file, err := ioutil.ReadFile(args[0].String)
			if err != nil {
				return ex.NewFatal("load: " + err.Error())
			}

			str := string(file)

			expr, err := parser.NewParser(str).Parse()
			if err != nil {
				return ex.NewFatal("load: " + err.Error())
			}

			return expr
		},
	},
}
