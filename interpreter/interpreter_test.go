package interpreter

import (
	"strconv"
	"testing"

	ex "github.com/batrSens/LispX/expressions"

	"github.com/magiconair/properties/assert"
)

func TestInterpreter(t *testing.T) {
	//res, err := Execute("(display 5) (display 6) (try(/ 4 0)) '(display d \"f\" (()) 2)")
	//assert.Equal(t, err, nil)
	//fmt.Printf("%+v\n%s\n", res, res.Output.ToString())

	test := 0 // nil program
	res, err := Execute("   ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNil()), true, "test#"+strconv.Itoa(test))

	test++ // 1 function that was received from expression
	res, err = Execute(" ((if (> 2 3) + -) 5 4)  ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(1)), true, "test#"+strconv.Itoa(test))

	test++ // 2 function that was received from expression
	res, err = Execute(" ((if (> 4 3) + -) 5 4)  ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(9)), true, "test#"+strconv.Itoa(test))

	test++ // 3 lambda definition
	res, err = Execute(" (define a (lambda (a b) (+ a (/ b 2/4)))) (a 9 3) ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(15)), true, "test#"+strconv.Itoa(test))

	test++ // 4 incorrect argument
	res, err = Execute(" (+ 2 3 '() 3) ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewFatal("")), true, "test#"+strconv.Itoa(test))

	test++ // 5 lambda with accessing to outer variable; redefinition of default symbol
	res, err = Execute(" (define - (lambda (b) (+ a b)) ) (define a 3) (- 4) ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 6 call of lambda that is result of other lambda
	res, err = Execute(" (define num+ (lambda (num) (lambda (a) (+ a num)))) (define 4+ (num+ 4)) (4+ 9) ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(13)), true, "test#"+strconv.Itoa(test))

	test++ // 7 anon lambda call
	res, err = Execute(" ((lambda (a) (+ a 3)) 4)  ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 8 anon lambda call
	res, err = Execute(" ((lambda (a) (+ a 3)) 4)  ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 9 anon lambda call with nil arguments
	res, err = Execute(" ((lambda () (+ 8 3)))  ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(11)), true, "test#"+strconv.Itoa(test))

	test++ // 10 list quote
	res, err = Execute(" '( 2 3 4) ")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(2).Cons(ex.NewNumber(3).Cons(ex.NewNumber(4).ToList()))), true, "test#"+strconv.Itoa(test))

	test++ // 11 lambda call with incorrect number of arguments
	res, err = Execute("(define s (lambda (a) (+ ww a))) (s 2 3)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewFatal("")), true, "test#"+strconv.Itoa(test))

	test++ // 12 lambda call with nil arguments
	res, err = Execute("(define s (lambda () (+ 9 2))) (s)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(11)), true, "test#"+strconv.Itoa(test))

	test++ // 13 =
	res, err = Execute("(= 2 3 4)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNil()), true, "test#"+strconv.Itoa(test))

	test++ // 14 =
	res, err = Execute("(= 4 4 4 4)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewT()), true, "test#"+strconv.Itoa(test))

	test++ // 15 =
	res, err = Execute("(= \"s\" \"s\")")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewT()), true, "test#"+strconv.Itoa(test))

	test++ // 16 recursion
	res, err = Execute("(define fact (lambda (n) (if (> n 1) (* n (fact (- n 1))) 1))) (fact 5)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(120)), true, "test#"+strconv.Itoa(test))

	test++ // 17 double recursion
	res, err = Execute("(define fib (lambda (num) (if (= num 0) 0 (if (= num 1) 1 (+ (fib (- num 2)) (fib (- num 1))))))) (fib 10)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(55)), true, "test#"+strconv.Itoa(test))

	test++ // 18 mutual recursion
	res, err = Execute("(define is_even (lambda (num) (if (= num 0) T (is_odd (- num 1))))) (define is_odd (lambda (num) (if (= num 0) nil (is_even (- num 1))))) (is_even 12)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewT()), true, "test#"+strconv.Itoa(test))

	test++ // 19 mutual recursion
	res, err = Execute("(define is_even (lambda (num) (if (= num 0) T (is_odd (- num 1))))) (define is_odd (lambda (num) (if (= num 0) nil (is_even (- num 1))))) (is_even 11)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNil()), true, "test#"+strconv.Itoa(test))

	test++ // 20 and
	res, err = Execute("(and 2 T 4 5)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(5)), true, "test#"+strconv.Itoa(test))

	test++ // 21 and
	res, err = Execute("(and 2 T nil 4 5)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNil()), true, "test#"+strconv.Itoa(test))

	test++ // 22 and
	res, err = Execute("(and)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewT()), true, "test#"+strconv.Itoa(test))

	test++ // 23 or
	res, err = Execute("(or nil 2 T 4 5)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(2)), true, "test#"+strconv.Itoa(test))

	test++ // 24 or
	res, err = Execute("(or nil nil nil)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNil()), true, "test#"+strconv.Itoa(test))

	test++ // 25 or
	res, err = Execute("(or)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNil()), true, "test#"+strconv.Itoa(test))

	test++ // 26 eval
	res, err = Execute("(eval '(+ 2 4 3))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(9)), true, "test#"+strconv.Itoa(test))

	test++ // 27 eval
	res, err = Execute("(define a '(define b (+ 4 80))) (eval a) (+ b 22)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(106)), true, "test#"+strconv.Itoa(test))

	test++ // 28 set! in lambda
	res, err = Execute("(define r 4) ((lambda () (set! r 5))) r")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(5)), true, "test#"+strconv.Itoa(test))

	test++ // 29 define in lambda
	res, err = Execute("(define r 4) ((lambda () (define r 5))) r")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(4)), true, "test#"+strconv.Itoa(test))

	test++ // 30 define in lambda
	res, err = Execute("(define r 4) ((lambda () (define r 5) r))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(5)), true, "test#"+strconv.Itoa(test))

	test++ // 31 set! in lambda
	res, err = Execute("(define r 4) ((lambda () (set! r 5) r))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(5)), true, "test#"+strconv.Itoa(test))

	test++ // 32 T
	res, err = Execute("(if T 2 3)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(2)), true, "test#"+strconv.Itoa(test))

	test++ // 33 nil
	res, err = Execute("(if nil 2 3)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(3)), true, "test#"+strconv.Itoa(test))

	test++ // 34 try
	res, err = Execute("(define a (try (/ 6 0) 7)) a")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 35 try catch
	res, err = Execute("(define a (try (/ 6 1) 7)) a")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(6)), true, "test#"+strconv.Itoa(test))

	test++ // 36 try
	res, err = Execute("(define a (try (/ 6 0) )) a")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNil()), true, "test#"+strconv.Itoa(test))

	test++ // 37 try
	res, err = Execute("(define a (string? (try (/ 6 0) error-description))) a")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewT()), true, "test#"+strconv.Itoa(test))

	test++ // 38 try
	res, err = Execute("(try (/ 6 0) error-description) error-description")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewFatal("")), true, "test#"+strconv.Itoa(test))

	test++ // 39 panic
	res, err = Execute("(if nil 2 (panic! \"PANIC!\"))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewFatal("")), true, "test#"+strconv.Itoa(test))

	test++ // 40 try for panic
	res, err = Execute("(if nil 2 (try (panic! \"PANIC!\") \"Don't panic\"))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewString("Don't panic")), true, "test#"+strconv.Itoa(test))

	test++ // 41 try
	res, err = Execute("(begin 2 3 4 5 (try (d 'd 32) 7))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 42 try
	res, err = Execute("(begin 2 3 4 5 (try (/ 2 3 4 5 (- 34 4 (/ 2 0)) 32) 7))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 43 try
	res, err = Execute("(begin 2 3 4 5 (try (+ 8 9 (/ 1 6/12)) 7))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(19)), true, "test#"+strconv.Itoa(test))

	test++ // 44 try
	res, err = Execute("(begin 2 3 4 5 (try (d 'd 32) 7))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 45 try
	res, err = Execute("(begin 2 3 4 5 (try ((/ 2 0) 'd 32) 7))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

}
