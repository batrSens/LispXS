package expressions

import "fmt"

const (
	Symbol = iota
	Fatal
	Pair

	Macro
	Function
	Closure
	String
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

func (v *Vars) IsRoot() bool {
	return v.Parent == nil
}

type closureVars []string

type Expr struct {
	Type       int
	String     string
	Number     float64
	car, cdr   *Expr
	Vars       closureVars
	ParentVars *Vars
}

func (e *Expr) ToString() string {
	switch e.Type {
	case Number:
		return fmt.Sprintf("Number(%f)", e.Number)
	case String:
		return fmt.Sprintf("String(%s)", e.String)
	case Symbol:
		return fmt.Sprintf("Symbol(%s)", e.String)
	case Fatal:
		return fmt.Sprintf("Fatal(%s)", e.String)
	case Function:
		return fmt.Sprintf("Function(%s)", e.String)
	case Closure:
		return fmt.Sprintf("Closure")
	case Macro:
		return fmt.Sprintf("Macro(%s)", e.String)
	case Nil:
		return "Nil"
	case Pair:
		return fmt.Sprintf("( %s . %s )", e.car.ToString(), e.cdr.ToString())
	default:
		return fmt.Sprintf("%+v", e)
	}
}

func NewSymbol(name string) *Expr {
	return &Expr{
		Type:   Symbol,
		String: name,
	}
}

func NewFatal(msg string) *Expr {
	return &Expr{
		Type:   Fatal,
		String: msg,
	}
}

func NewFunction(name string) *Expr {
	return &Expr{
		Type:   Function,
		String: name,
	}
}

func NewMacros(name string) *Expr {
	return &Expr{
		Type:   Macro,
		String: name,
	}
}

func NewClosure(args *Expr, body []*Expr, parentVars *Vars) *Expr {

	if args.Type != Pair && args.Type != Nil {
		return NewFatal("lambda: args must be a pair or nil")
	}

	exists := map[string]struct{}{}
	vars := closureVars{}
	for !args.IsNil() {
		if args.Car().Type != Symbol {
			return NewFatal("lambda: all args must be a symbols")
		}

		if _, ok := exists[args.Car().String]; ok {
			return NewFatal("lambda: all args must be a different")
		}

		exists[args.Car().String] = struct{}{}
		vars = append(vars, args.Car().String)
		args = args.Cdr()
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
		car:        NewSymbol("begin"),
		cdr:        lambdaBody,
		Vars:       vars,
		ParentVars: parentVars,
	}
}

func (e *Expr) NewClosureVars(args []*Expr) (*Vars, error) {
	vars := NewRootVars()
	vars.Parent = e.ParentVars

	if len(e.Vars) != len(args) {
		return nil, NewExprError(fmt.Sprintf("expected %d args, got %d args", len(e.Vars), len(args)))
	}

	for i, v := range e.Vars {
		vars.CurSymbols[v] = args[i]
	}

	return vars, nil
}

func (e *Expr) ClosureBody() *Expr {
	return e.car.Cons(e.cdr)
}

func NewString(str string) *Expr {
	return &Expr{
		Type:   String,
		String: str,
	}
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

	return NewFatal("cdr: object must be pair: " + e.ToString())
}

//func (e *Expr) ListLength() int {
//	cur := e.cdr
//	n := 1
//
//	for cur.Type != Nil {
//		cur = cur.cdr
//		n++
//	}
//
//	return n
//}

func (e *Expr) Equal(e1 *Expr) bool {
	if e == nil || e1 == nil {
		return e == e1
	}

	if e.Type == Closure {
		return false
	}

	//if e.Type != e1.Type {
	//	fmt.Println("qwe1")
	//}
	//
	//if e.String != e1.String {
	//	fmt.Println("qwe2")
	//}
	//
	//if e.Number != e1.Number {
	//	fmt.Println("qwe3", e.Number, e1.Number)
	//}
	//
	//if !e.car.Equal(e1.car) {
	//	fmt.Println("qwe4")
	//}
	//
	//if !e.cdr.Equal(e1.cdr) {
	//	fmt.Println("qwe5")
	//}

	return e.Type == e1.Type && (e.Type == Fatal || e.String == e1.String && e.Number == e1.Number && e.car.Equal(e1.car) && e.cdr.Equal(e1.cdr))
}

func (e *Expr) ToList() *Expr {
	return e.Cons(NewNil())
}

func (e *Expr) IsNil() bool {
	return e.Type == Nil
}
