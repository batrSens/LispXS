package interpreter

import (
	"fmt"
	ex "lispx/expressions"
	"lispx/lexer"
	"strconv"
)

const (
	ModOr = iota
	ModAnd
	ModIf
	ModExec
	ModTry
)

type Mod struct {
	Type int
	Exec map[int]struct{}
}

type Func struct {
	F   func(ir *Interpreter, args []*ex.Expr) *ex.Expr
	Mod *Mod
}

var functions = map[string]Func{

	"eval": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewSymbol("begin").Cons(ex.NewFatal("quote: must be 1 argument").ToList())
			}

			return ex.NewSymbol("begin").Cons(args[0].ToList())
		},
	},

	"quote": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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

	"try": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 && len(args) != 3 {
				return ex.NewFatal("try: must be 1 or 2 arguments")
			}

			if args[0].Type != ex.Fatal {
				return args[0]
			}

			if len(args) == 1 {
				return ex.NewNil()
			}

			return args[2]
		},
		Mod: &Mod{
			Type: ModTry,
		},
	},

	"panic!": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 || args[0].Type != ex.String {
				return ex.NewFatal("panic: must be one string")
			}

			return ex.NewFatal(args[0].String)
		},
	},

	"car": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("car: must be 1 argument")
			}

			return args[0].Car()
		},
	},

	"cdr": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("cdr: must be 1 argument")
			}

			return args[0].Cdr()
		},
	},

	"cons": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 2 {
				return ex.NewFatal("cons: must be 2 arguments")
			}

			return args[0].Cons(args[1])
		},
	},

	"define": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 2 {
				return ex.NewFatal("define: must be 2 arguments")
			}

			if args[0].Type != ex.Symbol {
				return ex.NewFatal("define: second argument is not symbol")
			}

			ir.varsEnvironment.CurSymbols[args[0].String] = args[1]
			return args[1]
		},
		Mod: &Mod{
			Type: ModExec,
			Exec: map[int]struct{}{2: {}},
		},
	},

	"set!": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) < 2 {
				return ex.NewFatal("define: must be at less 2 arguments")
			}

			return ex.NewClosure(args[0], args[1:], ir.varsEnvironment)
		},
		Mod: &Mod{
			Type: ModExec,
			Exec: map[int]struct{}{},
		},
	},

	"begin": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 {
				return ex.NewNil()
			}
			return args[len(args)-1]
		},
	},

	"or": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("not: must be 1 argument")
			}

			if args[0].IsNil() {
				return ex.NewT()
			}

			return ex.NewNil()
		},
	},

	"atom?": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("atom?: must be 1 argument")
			}

			if args[0].Type == ex.Pair {
				return ex.NewNil()
			}

			return ex.NewT()
		},
	},

	"list?": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("list?: must be 1 argument")
			}

			if args[0].Type == ex.Pair || args[0].IsNil() {
				return ex.NewT()
			}

			return ex.NewNil()
		},
	},

	"number?": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("number?: must be 1 argument")
			}

			if args[0].Type == ex.Number {
				return ex.NewT()
			}

			return ex.NewNil()
		},
	},

	"string?": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("string?: must be 1 argument")
			}

			if args[0].Type == ex.String {
				return ex.NewT()
			}

			return ex.NewNil()
		},
	},

	"symbol?": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("len: must be 1 argument")
			}

			if args[0].Type != ex.String {
				return ex.NewFatal("len: must be a string")
			}

			return ex.NewNumber(float64(len([]rune(args[0].String))))
		},
	},

	"string->symbol": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("string->symbol: must be 1 argument")
			}

			if args[0].Type != ex.String {
				return ex.NewFatal("string->symbol: must be a string")
			}

			return ex.NewSymbol(args[0].String)
		},
	},

	"symbol->string": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("symbol->string: must be 1 argument")
			}

			if args[0].Type != ex.Symbol {
				return ex.NewFatal("symbol->string: must be a symbol")
			}

			return ex.NewString(args[0].String)
		},
	},

	"string->number": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr { // todo: norm
			if len(args) != 1 {
				return ex.NewFatal("string->number: must be 1 argument")
			}

			if args[0].Type != ex.String {
				return ex.NewFatal("string->number: must be a string")
			}

			tok, err := lexer.NewLexer(args[0].String).NextToken()
			if err != nil || tok.Tag != lexer.TagNumber {
				return ex.NewFatal("string->number: incorrect string")
			}

			return ex.NewNumber(tok.Number)
		},
	},

	"number->string": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr { // todo: norm
			if len(args) != 1 {
				return ex.NewFatal("number->string: must be 1 argument")
			}

			if args[0].Type != ex.Number {
				return ex.NewFatal("number->string: must be a number")
			}

			return ex.NewString(strconv.FormatFloat(args[0].Number, 'f', -1, 64))
		},
	},

	"+": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 {
				return ex.NewNumber(0.0)
			}

			switch args[0].Type {
			case ex.Number:
				res := 0.0
				for _, arg := range args {
					if arg.Type != ex.Number {
						return ex.NewFatal("+: expected numbers")
					}
					res += arg.Number
				}
				return ex.NewNumber(res)

			case ex.String:
				res := ""
				for _, arg := range args {
					if arg.Type != ex.String {
						return ex.NewFatal("+: expected strings")
					}
					res += arg.String
				}
				return ex.NewString(res)
			default:
				return ex.NewFatal("+: expected numbers or strings")
			}
		},
	},

	"-": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 {
				return ex.NewNumber(0.0)
			}

			if len(args) == 1 {
				if args[0].Type != ex.Number {
					return ex.NewFatal("-: expected numbers")
				}

				return ex.NewNumber(-args[0].Number)
			}

			if args[0].Type == ex.String {
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

				return ex.NewString(string(runes[int(args[1].Number):int(args[2].Number)]))
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
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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

	"display": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ex.NewFatal("display: expected one expression")
			}

			arg := args[0]
			switch arg.Type {
			case ex.Symbol, ex.String:
				ir.stdout += arg.String
			case ex.Number:
				ir.stdout += fmt.Sprintf("%f", arg.Number)
			case ex.Nil:
				ir.stdout += "nil"
			default:
				ir.stdout += arg.ToString()
			}

			return arg
		},
	},
}
