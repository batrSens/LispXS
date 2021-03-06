package expressions

import (
	"fmt"
	"strconv"
)

const (
	Symbol = iota
	Pair

	Fatal
	Function
	Closure
	Macro
	Number
	Nil
)

type ExprError struct {
	message string
}

func NewExprError(message string) *ExprError {
	return &ExprError{message: message}
}

func (ee *ExprError) Error() string {
	return ee.message
}

type Vars struct {
	CurSymbols map[string]*Expr
	Parent     *Vars
}

func NewRootVars() *Vars {
	return &Vars{
		CurSymbols: map[string]*Expr{},
		Parent:     nil,
	}
}

func NewVarsWithParent(parent *Vars) *Vars {
	return &Vars{
		CurSymbols: map[string]*Expr{},
		Parent:     parent,
	}
}

func (v *Vars) IsRoot() bool {
	return v.Parent == nil
}

func varsDebug(vars map[string]*Expr) string {
	res := "( "
	for k, _ := range vars {
		res += k + /*" : " + v.ToString() +*/ " "
	}
	return res + ")"
}

func (v *Vars) Debug() {
	str := fmt.Sprintf("%s ", varsDebug(v.CurSymbols))
	cur := v.Parent
	for cur != nil {
		str += fmt.Sprintf("-> %s ", varsDebug(cur.CurSymbols))
		cur = cur.Parent
	}
	fmt.Println(str)
}

type variable struct {
	name               string
	calculatedForMacro bool
}

type closureVars struct {
	variableNumber bool
	vars           []variable
}

type trace struct {
	f   *Expr
	pos int
}

type Expr struct {
	Type               int
	String             string
	Number             float64
	Res, car, cdr      *Expr
	CalculatedForMacro bool

	Vars       closureVars
	ParentVars *Vars
	stackTrace []struct {
		f   *Expr
		pos int
	}
}

func (e *Expr) DebugString() string {
	switch e.Type {
	case Number:
		return fmt.Sprintf("Number(%s)", strconv.FormatFloat(e.Number, 'f', -1, 64))
	case Symbol:
		return fmt.Sprintf("Symbol(%s)", e.String)
	case Fatal:
		return fmt.Sprintf("Fatal(%s)", e.String)
	case Function:
		return fmt.Sprintf("Function(%s)", e.String)
	case Closure:
		return "Closure" + fmt.Sprintf("%v", e.Vars.vars) + e.cdr.ToString()
	case Macro:
		return "Macro" + fmt.Sprintf("%v", e.Vars.vars) + e.cdr.ToString()
	case Nil:
		return "Nil"
	case Pair:
		return fmt.Sprintf("( %s . %s )", e.car.DebugString(), e.cdr.DebugString())
	default:
		return fmt.Sprintf("%+v", e)
	}
}

func (e *Expr) ToString() string {
	switch e.Type {
	case Number:
		return fmt.Sprintf("%s", strconv.FormatFloat(e.Number, 'f', -1, 64))
	case Symbol:
		return e.String
	case Fatal:
		return fmt.Sprintf("Fatal(%s)", e.String)
	case Function:
		return fmt.Sprintf("Function(%s)", e.String)
	case Closure:
		return "Closure" + fmt.Sprintf("%v", e.Vars.vars)
	case Macro:
		return "Macro" + fmt.Sprintf("%v", e.Vars.vars)
	case Nil:
		return "nil"
	case Pair:
		res := "("
		cur := e
		i := 0
		for cur.Type != Nil {
			if i > 0 {
				res += " "
			}
			i++
			res += cur.Car().ToString()
			cur = cur.Cdr()
		}
		return res + ")"
	default:
		return fmt.Sprintf("%+v", e)
	}
}

func (e *Expr) StackTrace() string {
	res := "FATAL: " + e.String + "\n"
	for _, st := range e.stackTrace {
		res += st.f.DebugString() + " [" + strconv.Itoa(st.pos) + "]\n"
	}
	return res
}

func NewSymbol(name string) *Expr {
	return &Expr{
		Type:   Symbol,
		String: name,
	}
}

func NewFatal(tag string, res ...*Expr) *Expr {
	fat := &Expr{
		Type:   Fatal,
		String: tag,
	}

	if len(res) > 1 {
		panic("NewFatal: must be one res")
	}

	if len(res) == 1 {
		fat.Res = res[0]
	} else {
		fat.Res = NewNil() // NewSymbol(tag)
	}

	if fat.Res == nil {
		panic("NewFatal: it shouldn't have happened")
	}

	return fat
}

func NewFunction(name string) *Expr {
	return &Expr{
		Type:   Function,
		String: name,
	}
}

func NewClosure(args *Expr, body []*Expr, parentVars *Vars) *Expr {

	if args.Type != Pair && args.Type != Nil && args.Type != Symbol {
		return NewFatal("lambda: args must be a pair or nil or symbol")
	}

	exists := map[string]struct{}{}
	vars := closureVars{
		variableNumber: false,
		vars:           []variable{},
	}
	if args.Type == Symbol {
		vars = closureVars{
			variableNumber: true,
			vars:           []variable{{name: args.String}},
		}
	} else {
		for !args.IsNil() {
			if args.Car().Type != Symbol {
				return NewFatal("lambda: all args must be a symbols")
			}

			if _, ok := exists[args.Car().String]; ok {
				return NewFatal("lambda: all args must be a different")
			}

			exists[args.Car().String] = struct{}{}
			vars.vars = append(vars.vars, variable{name: args.Car().String})
			args = args.Cdr()
		}
	}

	if len(body) == 0 {
		return NewFatal("lambda: nil body")
	}

	lambdaBody := NewNil()
	for i := len(body) - 1; i >= 0; i-- {
		lambdaBody = body[i].Cons(lambdaBody)
	}

	return &Expr{
		Type:       Closure,
		car:        NewFunction("begin"),
		cdr:        lambdaBody,
		Vars:       vars,
		ParentVars: parentVars,
	}
}

func NewMacro(args *Expr, body []*Expr, parentVars *Vars) *Expr {

	if args.Type != Pair && args.Type != Nil && args.Type != Symbol {
		return NewFatal("defmacro: args must be a pair or nil or symbol")
	}

	exists := map[string]struct{}{}
	vars := closureVars{
		variableNumber: false,
		vars:           []variable{},
	}
	if args.Type == Symbol {
		vars = closureVars{
			variableNumber: true,
			vars:           []variable{{name: args.String, calculatedForMacro: args.CalculatedForMacro}},
		}
	} else {
		for !args.IsNil() {
			if args.Car().Type != Symbol {
				return NewFatal("defmacro: all args must be a symbols")
			}

			if _, ok := exists[args.Car().String]; ok {
				return NewFatal("defmacro: all args must be a different")
			}

			exists[args.Car().String] = struct{}{}
			vars.vars = append(vars.vars, variable{name: args.Car().String, calculatedForMacro: args.Car().CalculatedForMacro})
			args = args.Cdr()
		}
	}

	if len(body) == 0 {
		return NewFatal("defmacro: nil body")
	}

	lambdaBody := NewNil()
	for i := len(body) - 1; i >= 0; i-- {
		lambdaBody = body[i].Cons(lambdaBody)
	}

	return &Expr{
		Type:       Macro,
		car:        NewFunction("begin"),
		cdr:        lambdaBody,
		Vars:       vars,
		ParentVars: parentVars,
	}
}

func (e *Expr) MacroExecMod() map[int]struct{} {
	if e.Vars.variableNumber {
		if e.Vars.vars[0].calculatedForMacro {
			return nil
		}
		return map[int]struct{}{}
	}

	res := map[int]struct{}{}
	for i, v := range e.Vars.vars {
		if v.calculatedForMacro {
			res[i+1] = struct{}{}
		}
	}

	return res
}

func (e *Expr) NewClosureVars(args []*Expr) (*Vars, error) {
	vars := NewRootVars()
	vars.Parent = e.ParentVars

	if e.Vars.variableNumber {
		if len(e.Vars.vars) != 1 {
			panic("expected one, given: " + strconv.Itoa(len(e.Vars.vars)))
		}

		argsList := NewNil()
		for i := len(args) - 1; i >= 0; i-- {
			argsList = args[i].Cons(argsList)
		}

		vars.CurSymbols[e.Vars.vars[0].name] = argsList
	} else {
		if len(e.Vars.vars) != len(args) {
			return nil, NewExprError(fmt.Sprintf("call: expected %d args, got %d args", len(e.Vars.vars), len(args)))
		}

		for i, v := range e.Vars.vars {
			vars.CurSymbols[v.name] = args[i]
		}
	}

	return vars, nil
}

func (e *Expr) ClosureBody() *Expr {
	return e.car.Cons(e.cdr)
}

func NewNumber(num float64) *Expr {
	return &Expr{
		Type:   Number,
		Number: num,
	}
}

func NewT() *Expr {
	return &Expr{
		Type:   Symbol,
		String: "T",
	}
}

func NewNil() *Expr {
	return &Expr{
		Type: Nil,
	}
}

func (e *Expr) AddTrace(f *Expr, pos int) {
	e.stackTrace = append(e.stackTrace, trace{f, pos})
}

func (e *Expr) Cons(cdr *Expr) *Expr {
	if cdr.Type == Pair || cdr.Type == Nil {
		return &Expr{
			Type: Pair,
			car:  e,
			cdr:  cdr,
		}
	}

	return NewFatal("cons: cdr must be a pair or nil")
}

func (e *Expr) Car() *Expr {
	if e.Type == Pair {
		return e.car
	}

	return NewFatal("car: object must be pair")
}

func (e *Expr) Cdr() *Expr {
	if e.Type == Pair {
		return e.cdr
	}

	return NewFatal("cdr: object must be pair: " + e.DebugString())
}

func (e *Expr) Equal(e1 *Expr) bool {
	if e == nil || e1 == nil {
		return e == e1
	}

	if e.Type == Closure {
		return false
	}

	return e.Type == e1.Type && (e.Type == Fatal || e.String == e1.String && e.Number == e1.Number && e.car.Equal(e1.car) && e.cdr.Equal(e1.cdr))
}

func (e *Expr) ToList() *Expr {
	return e.Cons(NewNil())
}

func (e *Expr) IsNil() bool {
	return e.Type == Nil
}

func (e *Expr) Length() int {
	length := 0
	cur := e

	for cur.Type == Pair {
		length++
		cur = cur.cdr
	}

	return length
}

func (e *Expr) Index(i int) *Expr {
	cur := e

	for i > 0 {
		i--
		cur = cur.cdr
	}

	return cur.car
}
