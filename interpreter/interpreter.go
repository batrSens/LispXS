package interpreter

import (
	"fmt"
	ex "lispx/expressions"
	"lispx/parser"
	"strconv"
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

func (se *stackExpr) PreLast() *ex.Expr {
	return (*se)[len(*se)-2]
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
	mod             *Mod
	varsEnvironment *ex.Vars
}

type stackCall []call

func (sc *stackCall) Push(control *ex.Expr, argsNum int, mod *Mod) {
	*sc = append(*sc, call{
		control: control,
		argsNum: argsNum,
		mod:     mod,
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

	argsNum         int
	mod             *Mod
	varsEnvironment *ex.Vars

	stdout, stderr string
}

func NewInterpreter(program *ex.Expr) *Interpreter {
	vars := ex.NewRootVars()

	for f := range functions {
		vars.CurSymbols[f] = ex.NewFunction(f)
	}

	return &Interpreter{
		control:         program,
		varsEnvironment: vars,
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

				if ir.argsNum == 0 {
					if name := ir.dataStack.Last().String; ir.dataStack.Last().Type == ex.Function {
						ir.mod = functions[name].Mod
					}
				}

				curExpr := ir.getCurSymbol()
				ir.argsNum++

				if ir.mod != nil {
					switch ir.mod.Type {
					case ModOr:
						panic("unimplemented!")
					case ModAnd:
						panic("unimplemented!")
					case ModIf:
						if ir.argsNum == 2 && ir.dataStack.Last().IsNil() ||
							ir.argsNum == 3 && !ir.dataStack.PreLast().IsNil() ||
							ir.argsNum > 3 {
							ir.dataStack.Push(ex.NewNil())
							continue
						}
					case ModEval:
						panic("unimplemented!")
					case ModExec:
						if _, ok := ir.mod.Exec[ir.argsNum]; !ok {
							ir.dataStack.Push(curExpr)
							continue
						}
					default:
						panic("unexpected mod " + strconv.Itoa(ir.mod.Type))
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
				switch ir.dataStack.Last().Type {
				case ex.Function:
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

				case ex.Closure:
					cl := ir.dataStack.Pop()
					ir.callClosure(cl, args)
					break Inner

				default:
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
	curEnv := ir.varsEnvironment
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

	ir.dataStack.Push(fn.F(ir, args))
}

func (ir *Interpreter) popLastCall() {
	lCall := ir.callStack.Pop()

	if lCall.varsEnvironment != nil {
		ir.varsEnvironment = lCall.varsEnvironment
	}

	ir.argsNum = lCall.argsNum
	ir.mod = lCall.mod
	ir.control = lCall.control
}

func (ir *Interpreter) pushLastCall() {
	ir.callStack.Push(ir.control, ir.argsNum, ir.mod)

	newControl := ir.control.Car()
	if newControl.Type == ex.Error {
		panic(newControl.String)
	}

	ir.control = newControl
	ir.argsNum = 0
	ir.mod = nil
}

func (ir *Interpreter) callClosure(closure *ex.Expr, args []*ex.Expr) {
	vars, err := closure.NewClosureVars(args)
	if err != nil {
		ir.dataStack.Push(ir.newError(err.Error()))
		ir.popLastCall()
		return
	}

	ir.callStack.SetVars(ir.varsEnvironment)
	ir.varsEnvironment = vars
	ir.control = closure.ClosureBody()
	ir.argsNum = 0
}

func (ir *Interpreter) newError(message string) *ex.Expr {
	ir.stderr += message + "\n"

	return ex.NewError(message)
}
