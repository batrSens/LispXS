package interpreter

import (
	"fmt"
	ex "lispx/expressions"
)

const (
	ModOr = iota
	ModAnd
	ModIf
	ModEval
	ModExec
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
				return ex.NewSymbol("begin").Cons(ir.newError("quote: must be 1 argument").ToList())
			}

			return ex.NewSymbol("begin").Cons(args[0].ToList())
		},
	},

	"quote": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ir.newError("quote: must be 1 argument")
			}

			return args[0]
		},
		Mod: &Mod{
			Type: ModExec,
			Exec: map[int]struct{}{},
		},
	},

	"define": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 2 {
				return ir.newError("define: must be 2 arguments")
			}

			if args[0].Type != ex.Symbol {
				return ir.newError("define: second argument is not symbol")
			}

			ir.varsEnvironment.CurSymbols[args[0].String] = args[1]
			return args[1]
		},
		Mod: &Mod{
			Type: ModExec,
			Exec: map[int]struct{}{2: {}},
		},
	},

	"lambda": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) < 2 {
				return ir.newError("define: must be at less 2 arguments")
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
				return ir.newError(fmt.Sprintf("if: expected 2 or 3 expressions, got %d", len(args)))
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
			if len(args) != 2 && len(args) != 3 {
				return ir.newError(fmt.Sprintf(">: expected 2 expressions, got %d", len(args)))
			}

			for _, arg := range args {
				if arg.Type != ex.Number {
					return ir.newError(">: expected numbers")
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
			if len(args) != 2 && len(args) != 3 {
				return ir.newError(fmt.Sprintf("<: expected 2 expressions, got %d", len(args)))
			}

			for _, arg := range args {
				if arg.Type != ex.Number {
					return ir.newError("<: expected numbers")
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
				return ir.newError(fmt.Sprintf("=: expected at less 2 expressions, got %d", len(args)))
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

	"+": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			res := 0.0

			for _, arg := range args {
				if arg.Type != ex.Number {
					return ir.newError("+: expected numbers")
				}
				res += arg.Number
			}

			return ex.NewNumber(res)
		},
	},

	"-": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 {
				return ex.NewNumber(0.0)
			}

			if len(args) == 1 {
				if args[0].Type != ex.Number {
					return ir.newError("-: expected numbers")
				}

				return ex.NewNumber(-args[0].Number)
			}

			res := 0.0

			for i, arg := range args {
				if arg.Type != ex.Number {
					return ir.newError("-: expected numbers")
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
					return ir.newError("*: expected numbers")
				}
				res *= arg.Number
			}

			return ex.NewNumber(res)
		},
	},

	"/": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) == 0 {
				return ir.newError("/: expected at least one expression")
			}

			res := args[0].Number

			for _, arg := range args[1:] {

				if arg.Type != ex.Number {
					return ir.newError("/: expected numbers")
				} else if arg.Number == 0 {
					return ir.newError("/: zero division")
				}

				res /= arg.Number
			}

			return ex.NewNumber(res)
		},
	},

	"display": {
		F: func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
			if len(args) != 1 {
				return ir.newError("display: expected one expression")
			}

			arg := args[0]
			switch arg.Type {
			case ex.Symbol, ex.String:
				ir.stdout += arg.String
			case ex.Number:
				ir.stdout += fmt.Sprintf("%f", arg.Number)
			case ex.T:
				ir.stdout += "T"
			case ex.Nil:
				ir.stdout += "nil"
			default:
				ir.stdout += arg.ToString()
			}

			return arg
		},
	},
}
