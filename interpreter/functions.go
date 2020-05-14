package interpreter

import (
	"fmt"
	ex "lispx/expressions"
)

var functions = map[string]func(ir *Interpreter, args []*ex.Expr) *ex.Expr{

	"begin": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
		if len(args) == 0 {
			return ex.NewNil()
		}
		return args[len(args)-1]
	},

	"if": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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

	">": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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

	"+": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
		res := 0.0

		for _, arg := range args {
			if arg.Type != ex.Number {
				return ir.newError("+: expected numbers")
			}
			res += arg.Number
		}

		return ex.NewNumber(res)
	},

	"-": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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

	"*": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
		res := 1.0

		for _, arg := range args {
			if arg.Type != ex.Number {
				return ir.newError("*: expected numbers")
			}
			res *= arg.Number
		}

		return ex.NewNumber(res)
	},

	"/": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
		if len(args) == 0 {
			return ir.newError("/: expected at least one expression")
		}

		res := args[0].Number

		for i := 1; i < len(args); i++ {
			arg := args[i]

			if arg.Type != ex.Number {
				return ir.newError("/: expected numbers")
			} else if arg.Number == 0 {
				return ir.newError("/: zero division")
			}

			res /= arg.Number
		}

		return ex.NewNumber(res)
	},

	"display": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
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

		return ex.NewT()
	},

	//"eval": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
	//	if len(args) != 1 {
	//		return ir.newError("eval: expected one expression")
	//	}
	//
	//	ir.callStack.Push(ex.NewNil().Cons(args[0].ToList()), 1)
	//
	//	return ex.NewFunction("begin")
	//},
}
