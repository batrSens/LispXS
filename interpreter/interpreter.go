package interpreter

import (
	"bytes"
	"fmt"
	"io"
	"os"

	ex "github.com/batrSens/LispX/expressions"
	"github.com/batrSens/LispX/parser"
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

	outstr, errstr := bytes.NewBufferString(""), bytes.NewBufferString("")

	res := NewInterpreter(exprs, outstr, errstr, os.Stdin).run()

	return &Output{
		Stdout: outstr.String(),
		Stderr: errstr.String(),
		Output: res,
	}, nil
}

func ExecuteStdout(program string) (*ex.Expr, error) {
	prs := parser.NewParser(program)
	exprs, err := prs.Parse()
	if err != nil {
		return nil, err
	}

	res := NewInterpreter(exprs, os.Stdout, os.Stderr, os.Stdin).run()

	return res, nil
}

func ExecuteTo(program string, ioout, ioerr io.Writer, ioin io.Reader) (*ex.Expr, error) {
	prs := parser.NewParser(program)
	exprs, err := prs.Parse()
	if err != nil {
		return nil, err
	}

	res := NewInterpreter(exprs, ioout, ioerr, ioin).run()

	return res, nil
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
		fmt.Println(e.DebugString())
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

func (sc *stackCall) SetMod(mod *Mod) {
	last := (*sc)[len(*sc)-1]
	last.mod = mod
	(*sc)[len(*sc)-1] = last
}

type Interpreter struct {
	callStack stackCall
	dataStack stackExpr
	control   *ex.Expr

	argsNum         int
	mod             *Mod
	varsEnvironment *ex.Vars

	stdout, stderr io.Writer
	stdin          io.Reader
}

func NewInterpreter(program *ex.Expr, stdout, stderr io.Writer, stdin io.Reader) *Interpreter {
	vars := ex.NewRootVars()

	for f := range functions {
		vars.CurSymbols[f] = ex.NewFunction(f)
	}

	vars.CurSymbols["T"] = ex.NewSymbol("T")
	vars.CurSymbols["nil"] = ex.NewNil()

	return &Interpreter{
		control:         program,
		varsEnvironment: vars,
		stderr:          stderr,
		stdout:          stdout,
		stdin:           stdin,
	}
}

func (ir *Interpreter) run() *ex.Expr {
	for {
		if len(ir.dataStack) > 0 && ir.dataStack.Last().Type == ex.Fatal {
			if fatal := ir.fatalFall(); fatal != nil {
				return fatal
			}
		}

		if ir.argsNum == 1 {
			ir.modLoad()
		}

		if ir.argsNum != 0 {
			ir.nextSymbol()
		}

		if !ir.control.IsNil() {
			curExpr := ir.getCurSymbol()
			ir.argsNum++

			if ir.mod != nil && modApply(ir) {
				continue
			}

			switch curExpr.Type {
			case ex.Number, ex.Nil, ex.Fatal, ex.Function, ex.Closure, ex.Macro:
				ir.dataStack.Push(curExpr)
			case ex.Symbol:
				expr := ir.resolveSymbol(curExpr)
				ir.dataStack.Push(expr)
			case ex.Pair:
				ir.pushLastCall()
			default:
				panic(fmt.Sprint("unexpected symbol type ", curExpr.Type))
			}

			// end of list
		} else {
			f, args := ir.popArgs()

			switch f.Type {
			case ex.Function:
				ir.execFunc(f, args)

				if f.String == "eval" {
					ir.control = ir.dataStack.Pop()
					ir.argsNum = 0
					ir.mod = nil
					continue
				}

				if len(ir.callStack) == 0 {
					if len(ir.dataStack) != 1 {
						ir.dataStack.Debug()
						panic("expected 1 value on the stack;")
					}

					return ir.dataStack.Pop()
				}

				ir.popLastCallAndCheckMacro()

			case ex.Closure:
				ir.callClosure(f, args)

			case ex.Macro:
				ir.callMacro(f, args)

			default:
				ir.dataStack.Push(ex.NewFatal(f.DebugString() + " is not a function"))
				ir.popLastCallAndCheckMacro()
			}
		}
	}
}

func (ir *Interpreter) modLoad() {
	switch ir.dataStack.Last().Type {
	case ex.Function:
		name := ir.dataStack.Last().String
		ir.mod = functions[name].Mod

		if name == "try" {
			newEnv := ex.NewVarsWithParent(ir.varsEnvironment)
			ir.setNewVars(newEnv)
		}
	case ex.Macro:
		ir.mod = &Mod{Type: ModExec, Exec: map[int]struct{}{}}
	}
}

func (ir *Interpreter) fatalFall() *ex.Expr {
	fatal := ir.dataStack.Pop()
	var f *ex.Expr

	for i := 0; true; i++ {
		if i > 0 {
			if len(ir.callStack) == 0 {
				_, _ = fmt.Fprint(ir.stderr, fatal.StackTrace())
				return fatal
			}

			if f.Equal(ex.NewFunction("try")) && ir.argsNum == 1 {
				ir.dataStack.Push(f)
				ir.dataStack.Push(fatal)
				ir.argsNum = 2
				return nil
			}

			ir.popLastCall()
		}

		ir.argsNum--
		if ir.argsNum <= 0 {
			fatal.AddTrace(ex.NewSymbol("none"), ir.argsNum)
		} else {
			f, _ = ir.popArgs()
			fatal.AddTrace(f, ir.argsNum)
		}
	}

	panic("unexpected")
}

func (ir *Interpreter) resolveSymbol(symbol *ex.Expr) *ex.Expr {
	curEnv := ir.varsEnvironment
	for curEnv != nil {
		if expr, ok := curEnv.CurSymbols[symbol.String]; ok {
			return expr
		}

		curEnv = curEnv.Parent
	}

	return ex.NewFatal(fmt.Sprintf("symbol '%s' is not defined", symbol.String))
}

func (ir *Interpreter) popArgs() (f *ex.Expr, args []*ex.Expr) {
	var res []*ex.Expr

	for i := 0; i < ir.argsNum-1; i++ {
		res = append([]*ex.Expr{ir.dataStack.Pop()}, res...)
	}

	return ir.dataStack.Pop(), res
}

func (ir *Interpreter) nextSymbol() {
	cdr := ir.control.Cdr()
	if cdr.Type == ex.Fatal {
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

func (ir *Interpreter) setNewVars(vars *ex.Vars) {
	ir.callStack.SetVars(ir.varsEnvironment)
	ir.varsEnvironment = vars
}

func (ir *Interpreter) popLastCallAndCheckMacro() {
	ir.popLastCall()
	if ir.mod != nil && ir.mod.Type == ex.Macro {
		ir.applyMacro()
	}
}

func (ir *Interpreter) applyMacro() {
	ir.callStack.Push(ir.control, ir.argsNum, ir.mod.Old)

	ir.argsNum = 0
	ir.mod = nil

	prog := ir.dataStack.Pop()
	if prog.Type == ex.Pair {
		ir.control = prog
	} else {
		ir.control = ex.NewSymbol("begin").Cons(prog.ToList())
	}
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
	if newControl.Type == ex.Fatal {
		panic(newControl.String)
	}

	ir.control = newControl
	ir.argsNum = 0
	ir.mod = nil
}

func (ir *Interpreter) callClosure(closure *ex.Expr, args []*ex.Expr) {
	vars, err := closure.NewClosureVars(args)
	if err != nil {
		ir.dataStack.Push(ex.NewFatal(err.Error()))
		ir.popLastCall()
		return
	}

	ir.setNewVars(vars)
	ir.control = closure.ClosureBody()
	ir.argsNum = 0
	ir.mod = nil
}

func (ir *Interpreter) callMacro(macro *ex.Expr, args []*ex.Expr) {
	vars, err := macro.NewClosureVars(args)
	if err != nil {
		ir.dataStack.Push(ex.NewFatal(err.Error()))
		ir.popLastCall()
		return
	}

	ir.callStack.SetMod(&Mod{Type: ModMacro, Old: ir.callStack.Last().mod})
	ir.setNewVars(vars)
	ir.control = macro.ClosureBody()
	ir.argsNum = 0
	ir.mod = nil
}
