package interpreter

import (
	"fmt"
	ex "lispx/expressions"
	"strconv"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestInterpreter(t *testing.T) {
	res, err := Execute("(define fact (lambda (n) (if (> n 1) (* n (fact (- n 1))) 1))) (fact 5)")
	//res, err := Execute("(define s (lambda () (+ 9 2))) (s)")
	assert.Equal(t, err, nil)
	fmt.Printf("%+v\n%s\n", res, res.Output.ToString())

	test := 0 // nil program
	res, err = Execute("   ")
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
	assert.Equal(t, res.Output.Equal(ex.NewError("")), true, "test#"+strconv.Itoa(test))

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
	assert.Equal(t, res.Output.Equal(ex.NewError("")), true, "test#"+strconv.Itoa(test))

	test++ // 12 lambda call with nil arguments
	res, err = Execute("(define s (lambda () (+ 9 2))) (s)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(11)), true, "test#"+strconv.Itoa(test))

	test++ // 13 recursion
	res, err = Execute("(define fact (lambda (n) (if (> n 1) (* n (fact (- n 1))) 1))) (fact 5)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(120)), true, "test#"+strconv.Itoa(test))

}
