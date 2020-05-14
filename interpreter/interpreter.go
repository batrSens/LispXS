package interpreter

import (
	"fmt"
	ex "lispx/expressions"
	"lispx/parser"
)

type Output struct {
	Stdout, Stderr string
	Output         *ex.Expr
}

func Execute(program string) (*Output, error) {
	prs := parser.NewParser(program)
	exprs, err := prs.Parse()
	if err != nil {
		return nil, err
	}

	intrp := NewInterpreter(exprs)
	return intrp.run(), nil
}

type stackExpr []*ex.Expr

func (se *stackExpr) Push(expr *ex.Expr) {
	*se = append(*se, expr)
}

func (se *stackExpr) Pop() *ex.Expr {
	res := (*se)[len(*se)-1]
	*se = (*se)[:len(*se)-1]
	return res
}

func (se *stackExpr) Last() *ex.Expr {
	return (*se)[len(*se)-1]
}

func (se *stackExpr) Debug() {
	fmt.Println("STACK =======")
	for _, e := range *se {
		fmt.Println(e.ToString())
	}
	fmt.Println("END =========")
}

type call struct {
	control *ex.Expr

	argsNum         int
	argsUnex        map[int]struct{}
	varsEnvironment *ex.Vars
}

type stackCall []call

func (sc *stackCall) Push(control *ex.Expr, argsNum int, argsUnex map[int]struct{}) {
	*sc = append(*sc, call{
		control:  control,
		argsNum:  argsNum,
		argsUnex: argsUnex,
	})
}

func (sc *stackCall) Pop() call {
	res := (*sc)[len(*sc)-1]
	*sc = (*sc)[:len(*sc)-1]
	return res
}

func (sc *stackCall) Last() call {
	return (*sc)[len(*sc)-1]
}

func (sc *stackCall) SetVars(vars *ex.Vars) {
	last := (*sc)[len(*sc)-1]
	last.varsEnvironment = vars
	(*sc)[len(*sc)-1] = last
}

type Interpreter struct {
	callStack stackCall
	dataStack stackExpr
	control   *ex.Expr

	argsNum        int
	argsUnex       map[int]struct{}
	varsEnviroment *ex.Vars

	stdout, stderr string
}

func NewInterpreter(program *ex.Expr) *Interpreter {
	vars := ex.NewRootVars()

	for f := range functions {
		vars.CurSymbols[f] = ex.NewFunction(f)
	}

	for m := range macroses {
		vars.CurSymbols[m] = ex.NewMacros(m)
	}

	return &Interpreter{
		control:        program,
		varsEnviroment: vars,
	}
}

func (ir *Interpreter) run() *Output {
	for {
		if ir.argsNum == 0 {
			switch ir.getCurSymbol().Type {
			case ex.String, ex.Number, ex.Error, ex.T, ex.Nil:
				ir.dataStack.Push(ir.getCurSymbol())
			case ex.Symbol:
				expr, _ := ir.resolveSymbol(ir.getCurSymbol())

				if expr.Type == ex.Macro {
					ir.argsUnex = macrosesUnex[expr.String]
				}

				ir.dataStack.Push(expr)

			case ex.Pair:
				ir.pushLastCall()
				continue
			default:
				panic(fmt.Sprint("unexpected symbol type", ir.getCurSymbol().Type))
			}
		}

	Inner:
		for {
			ir.nextSymbol()
			if !ir.control.IsNil() {
				curExpr := ir.getCurSymbol()
				ir.argsNum++

				if ir.argsUnex != nil {
					if _, ok := ir.argsUnex[ir.argsNum]; !ok {
						ir.dataStack.Push(curExpr)
						continue
					}
				}

				switch curExpr.Type {
				case ex.String, ex.Number, ex.Error, ex.T, ex.Nil:
					ir.dataStack.Push(curExpr)
				case ex.Symbol:
					expr, _ := ir.resolveSymbol(curExpr)
					ir.dataStack.Push(expr)
				case ex.Pair:
					ir.pushLastCall()
					break Inner
				default:
					panic(fmt.Sprint("unexpected symbol type", curExpr.Type))
				}

				// end of list
			} else {
				args := ir.popArgs()
				if ir.dataStack.Last().Type == ex.Function {
					f := ir.dataStack.Pop()
					ir.execFunc(f, args)

					if len(ir.callStack) == 0 {
						if len(ir.dataStack) != 1 {
							ir.dataStack.Debug()
							panic("expected 1 value on the stack;")
						}

						return &Output{
							Stdout: ir.stdout,
							Stderr: ir.stderr,
							Output: ir.dataStack.Pop(),
						}
					}

					ir.popLastCall()

				} else if ir.dataStack.Last().Type == ex.Closure {
					cl := ir.dataStack.Pop()
					ir.callClosure(cl, args)
					break Inner

				} else if ir.dataStack.Last().Type == ex.Macro {
					m := ir.dataStack.Pop()
					ir.execMacros(m, args)

					if ir.control == nil {
						ir.popLastCall()
					} else {
						panic("unimplemented!")
						//ir.control = res
					}
				} else {
					obj := ir.dataStack.Pop()
					ir.dataStack.Push(ir.newError(obj.ToString() + " is not a function"))
					ir.popLastCall()
				}
			}
		}
	}
}

// todo: not ex.Error, default to symbols
func (ir *Interpreter) resolveSymbol(symbol *ex.Expr) (*ex.Expr, bool) {
	curEnv := ir.varsEnviroment
	for curEnv != nil {
		if expr, ok := curEnv.CurSymbols[symbol.String]; ok {
			return expr, true
		}

		curEnv = curEnv.Parent
	}

	return ir.newError(fmt.Sprintf("symbol '%s' is not defined", symbol.String)), false
}

func (ir *Interpreter) popArgs() []*ex.Expr {
	var res []*ex.Expr

	for i := 0; i < ir.argsNum; i++ {
		res = append([]*ex.Expr{ir.dataStack.Pop()}, res...)
	}

	return res
}

func (ir *Interpreter) nextSymbol() {
	cdr := ir.control.Cdr()
	if cdr.Type == ex.Error {
		ir.dataStack.Debug()
		panic(cdr.String)
	}

	ir.control = cdr
}

func (ir *Interpreter) getCurSymbol() *ex.Expr {
	car := ir.control.Car()

	return car
}

func (ir *Interpreter) execFunc(f *ex.Expr, args []*ex.Expr) {
	fn, ok := functions[f.String]
	if !ok {
		panic("unexpected func " + f.String)
	}

	ir.dataStack.Push(fn(ir, args))
}

func (ir *Interpreter) execMacros(m *ex.Expr, args []*ex.Expr) {
	mac, ok := macroses[m.String]
	if !ok {
		panic("unexpected macros " + m.String)
	}

	ir.control = mac(ir, args)
}

func (ir *Interpreter) popLastCall() {
	lCall := ir.callStack.Pop()

	if lCall.varsEnvironment != nil {
		ir.varsEnviroment = lCall.varsEnvironment
	}

	ir.argsNum = lCall.argsNum
	ir.argsUnex = lCall.argsUnex
	ir.control = lCall.control
}

func (ir *Interpreter) pushLastCall() {
	ir.callStack.Push(ir.control, ir.argsNum, ir.argsUnex)

	newControl := ir.control.Car()
	if newControl.Type == ex.Error {
		panic(newControl.String)
	}

	ir.control = newControl
	ir.argsNum = 0
	ir.argsUnex = nil
}

func (ir *Interpreter) callClosure(closure *ex.Expr, args []*ex.Expr) {
	vars, err := closure.NewClosureVars(args)
	if err != nil {
		ir.dataStack.Push(ir.newError(err.Error()))
		ir.popLastCall()
		return
	}

	ir.callStack.SetVars(ir.varsEnviroment)
	ir.varsEnviroment = vars
	ir.control = closure.ClosureBody()
	ir.argsNum = 0
}

func (ir *Interpreter) newError(message string) *ex.Expr {
	ir.stderr += message + "\n"

	return ex.NewError(message)
}
