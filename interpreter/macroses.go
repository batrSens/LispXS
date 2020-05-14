package interpreter

import ex "lispx/expressions"

var macrosesUnex = map[string]map[int]struct{}{
	"quote":  {},
	"define": {2: struct{}{}},
	"lambda": {},
}

var macroses = map[string]func(ir *Interpreter, args []*ex.Expr) *ex.Expr{

	"quote": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
		if len(args) != 1 {
			ir.dataStack.Push(ir.newError("quote: must be 1 argument"))
			return nil
		}

		ir.dataStack.Push(args[0])
		return nil
	},

	"define": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
		if len(args) != 2 {
			ir.dataStack.Push(ir.newError("define: must be 2 arguments"))
			return nil
		}

		if args[0].Type != ex.Symbol {
			ir.dataStack.Push(ir.newError("define: second argument is not symbol"))
			return nil
		}

		ir.varsEnviroment.CurSymbols[args[0].String] = args[1]
		ir.dataStack.Push(args[1])
		return nil
	},

	"lambda": func(ir *Interpreter, args []*ex.Expr) *ex.Expr {
		if len(args) < 2 {
			ir.dataStack.Push(ir.newError("define: must be at less 2 arguments"))
			return nil
		}

		ir.dataStack.Push(ex.NewClosure(args[0], args[1:], ir.varsEnviroment))
		return nil
	},
}

//func macrosResult(ir *Interpreter, res *ex.Expr) *ex.Expr {
//	ir.dataStack.Push(res)
//	return nil
//}
