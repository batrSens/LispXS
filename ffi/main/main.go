// go build -o lispxs.so -buildmode=c-shared main.go
// gcc -o lispxsc ./main.c ./lispxs.so
package main

import "C"
import (
	"strconv"
	"unsafe"

	ex "github.com/batrSens/LispXS/expressions"
	lispxs "github.com/batrSens/LispXS/interpreter"
)

const (
	ERROR = iota
	NUMBER
	SYMBOL
	NIL
	FATAL
)

//export execute
func execute(prog *C.char) (expr unsafe.Pointer, stdout, stderr, error *C.char) {
	program := C.GoString(prog)

	res, err := lispxs.Execute(program)
	if err != nil {
		return nil, cNil(), cNil(), cError(err)
	}

	return cExpr(res.Output), C.CString(res.Stdout), C.CString(res.Stderr), cNil()
}

//export execute_stdout
func execute_stdout(prog *C.char) (expr unsafe.Pointer, error *C.char) {
	program := C.GoString(prog)

	res, err := lispxs.ExecuteStdout(program)
	if err != nil {
		return nil, cError(err)
	}

	return cExpr(res), cNil()
}

//export library_load
func library_load(path *C.char) (library unsafe.Pointer, str *C.char) {
	lib, err := lispxs.LoadLibrary(C.GoString(path))
	if err != nil {
		return nil, cError(err)
	}

	return cLib(lib), cNil()
}

//export library_call
func library_call(lib, call unsafe.Pointer) (expr unsafe.Pointer, error *C.char) {
	callGo := goCall(call)

	res, err := goLib(lib).Call(callGo.fn, callGo.args...)
	if err != nil {
		return nil, cError(err)
	}

	return cExpr(res), cNil()
}

//export call_new
func call_new(fn *C.char) unsafe.Pointer {
	return cCall(&call{fn: C.GoString(fn), args: []interface{}{}})
}

//export call_add_number
func call_add_number(c unsafe.Pointer, number float64) unsafe.Pointer {
	callGo := goCall(c)
	callGo.args = append(callGo.args, number)
	return cCall(callGo)
}

//export call_add_symbol
func call_add_symbol(c unsafe.Pointer, symbol *C.char) unsafe.Pointer {
	callGo := goCall(c)
	callGo.args = append(callGo.args, C.GoString(symbol))
	return cCall(callGo)
}

//export call_add_symbol
func call_add_list(c, list unsafe.Pointer) unsafe.Pointer {
	callGo := goCall(c)
	callGo.args = append(callGo.args, goCall(list).args)
	return cCall(callGo)
}

//export expr_is_pair
func expr_is_pair(expr unsafe.Pointer) bool {
	return goExpr(expr).Type == ex.Pair
}

//export expr_length
func expr_length(expr unsafe.Pointer) int {
	return goExpr(expr).Length()
}

//export expr_index
func expr_index(expr unsafe.Pointer, i int) unsafe.Pointer {
	return cExpr(goExpr(expr).Index(i))
}

//export expr_atom
func expr_atom(expr unsafe.Pointer) (typ int, number float64, str *C.char) {
	return cAtom(goExpr(expr))
}

func cAtom(res *ex.Expr) (typ int, number float64, str *C.char) {
	switch res.Type {
	case ex.Nil:
		return NIL, 0, cNil()
	case ex.Number:
		return NUMBER, res.Number, cNil()
	case ex.Symbol:
		return SYMBOL, 0, C.CString(res.String)
	case ex.Fatal:
		return FATAL, 0, C.CString(res.String)
	default:
		return ERROR, 0, C.CString("wrong type " + strconv.Itoa(res.Type))
	}
}

func cError(err error) *C.char {
	return C.CString(err.Error())
}

func cNil() *C.char {
	return C.CString("")
}

func cExpr(expr *ex.Expr) unsafe.Pointer {
	exprAlloc := C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))))
	p := (*[1]*ex.Expr)(exprAlloc)
	p[0] = &(*(*ex.Expr)(unsafe.Pointer(expr)))
	return exprAlloc
}

func goExpr(expr unsafe.Pointer) *ex.Expr {
	return (*[1]*ex.Expr)(expr)[0]
}

func cLib(lib *lispxs.Library) unsafe.Pointer {
	libAlloc := C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))))
	p := (*[1]*lispxs.Library)(libAlloc)
	p[0] = &(*(*lispxs.Library)(unsafe.Pointer(lib)))
	return libAlloc
}

func goLib(lib unsafe.Pointer) *lispxs.Library {
	return (*[1]*lispxs.Library)(lib)[0]
}

func cCall(c *call) unsafe.Pointer {
	callAlloc := C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))))
	p := (*[1]*call)(callAlloc)
	p[0] = &(*(*call)(unsafe.Pointer(c)))
	return callAlloc
}

func goCall(c unsafe.Pointer) *call {
	return (*[1]*call)(c)[0]
}

type call struct {
	fn   string
	args []interface{}
}

func main() {}
