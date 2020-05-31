package interpreter

import (
	"math"
	"strconv"
	"testing"

	ex "github.com/batrSens/LispXS/expressions"

	"github.com/magiconair/properties/assert"
)

func TestInterpreter(t *testing.T) {
	//ress, err := Execute(`
	//	(define list (lambda args args))
	//	((lambda ()
	//		(define temp set!)
	//		(defmacro settemp (sym val)
	//			(if (= sym '+) (throw '|couldn't redefine '+' func|))
	//			(list temp sym (eval val)))
	//		(set! set! settemp)))
	//	(set! + >)
	//	(+ 3 2)
	//`)
	//assert.Equal(t, err, nil)
	//fmt.Printf("%+v\n%s\n", ress, ress.Output.ToString())
	//
	//lib, err := LoadLibrary("../path_to_file")
	//assert.Equal(t, err, nil)
	//
	//res1, err := lib.Call("--", 9.0)
	//assert.Equal(t, err, nil)
	//fmt.Println("resss", res1.ToString())
	//
	//res2, err := lib.Call("-", "this is string", 1, 6)
	//assert.Equal(t, err, nil)
	//fmt.Println("resss", res2.ToString())

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
	res, err = Execute("(= 'sd 'sd)")
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

	test++ // 34 catch
	res, err = Execute("(define a (catch (/ 6 0) (/ (+ 3 4)))) a")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 35 catch catch
	res, err = Execute("(define a (catch (/ 6 1) 7)) a")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(6)), true, "test#"+strconv.Itoa(test))

	test++ // 36 catch
	res, err = Execute("(define a (catch (/ 6 0) (+ 5) (/))) a")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNil()), true, "test#"+strconv.Itoa(test))

	test++ // 37 catch
	res, err = Execute("(define a (symbol? (catch (/ 6 0) (default 'symsym)))) a")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewT()), true, "test#"+strconv.Itoa(test))

	test++ // 38 catch
	res, err = Execute("(catch (/ 6 0) error-description) error-description")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewFatal("")), true, "test#"+strconv.Itoa(test))

	test++ // 39 panic
	res, err = Execute("(if nil 2 (throw PANIC!))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewFatal("")), true, "test#"+strconv.Itoa(test))

	test++ // 40 catch for panic
	res, err = Execute("(if nil 2 (catch (throw 'PANIC! '|Don't panic|) (PANIC!)))")
	assert.Equal(t, res.Output.Equal(ex.NewSymbol("Don't panic")), true, "test#"+strconv.Itoa(test))

	test++ // 41 catch
	res, err = Execute("(begin 2 3 4 5 (catch (d 'd 32) (/ 9) (default (* (+ 1 2) 2))))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(6)), true, "test#"+strconv.Itoa(test))

	test++ // 42 catch
	res, err = Execute("(begin 2 3 4 5 (catch (/ 2 3 4 5 (- 34 4 (/ 2 0)) 32) (/ 7)))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 43 catch
	res, err = Execute("(begin 2 3 4 5 (catch (+ 8 9 (/ 1 6/12)) 7))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(19)), true, "test#"+strconv.Itoa(test))

	test++ // 44 catch
	res, err = Execute("(begin 2 3 4 5 (catch (d 'd 32) (default 7)))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 45 catch
	res, err = Execute("(begin 2 3 4 5 (catch ((/ 2 0) 'd 32) (/ 7)))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(7)), true, "test#"+strconv.Itoa(test))

	test++ // 46 variable number of arguments
	res, err = Execute("(define f (lambda args (if (= args nil) nil (cons (+ (car args) 100) (eval (cons 'f (cdr args))))))) (f 8 3 4)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(108).Cons(ex.NewNumber(103).Cons(ex.NewNumber(104).ToList()))), true, "test#"+strconv.Itoa(test))

	test++ // 47 macro apply
	res, err = Execute("(define list (lambda args args)) (defmacro apply (f ,args) (cons f args)) (apply list '(4 5 6 'gtgtg))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(4).Cons(ex.NewNumber(5).
		Cons(ex.NewNumber(6).Cons(ex.NewSymbol("gtgtg").ToList())))), true, "test#"+strconv.Itoa(test))

	test++ // 48 macro apply
	res, err = Execute("(define list (lambda args args)) (defmacro apply (f ,args) (cons f args)) (apply - '(4 5 6))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(-7)), true, "test#"+strconv.Itoa(test))

	test++ // 49 macro set10
	res, err = Execute("(define list (lambda args args)) (defmacro set10 (s) (list 'set! s 10)) (define qwe 303) (set10 qwe) qwe")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(10)), true, "test#"+strconv.Itoa(test))

	test++ // 50 error
	res, err = Execute("((/ (+ 2 9) 0) 4)")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewFatal("")), true, "test#"+strconv.Itoa(test))

	test++ // 51 error
	res, err = Execute("(define list (lambda args args)) (defmacro mac s (list (car s) (car (car (cdr s))) (car (cdr (car (cdr s)))))) (mac list (3))")
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewFatal("")), true, "test#"+strconv.Itoa(test))

	test++ // 52 sqrt
	res, err = Execute(`
		(define <= (lambda (a b) (or (< a b) (= a b)) ))
		(define sqrt (lambda (x) 
			(define findi (lambda (i)
				(if (<= (* i i) x)
					(findi (+ i 1))
					(- i 1))))
			(define i (findi 0))
			(define p (/ (- x (* i i)) (* 2 i)))
			(define a (+ i p))
			(- a (/ (* p p) (* 2 a))))) 
		(sqrt 21.0681)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, math.Abs(res.Output.Number-4.59) < 0.01, true, "test#"+strconv.Itoa(test))

	test++ // 53 quadric resolve
	res, err = Execute(`
		(define list (lambda args args))

		(define <= (lambda (a b) (or (< a b) (= a b)) ))

		(define sqrt (lambda (x) 
			(define findi (lambda (i)
				(if (<= (* i i) x)
					(findi (+ i 1))
					(- i 1))))
			(define i (findi 0))
			(define p (/ (- x (* i i)) (* 2 i)))
			(define a (+ i p))
			(- a (/ (* p p) (* 2 a)))))

		(define resolve (lambda (a b c) 
			(define disc (lambda (a b c) (- (* b b) (* 4 a c)))) 
			(define d (disc a b c)) 
			(if (> d 0) 
				(list (/ (- (- b) (sqrt d)) (* 2 a)) (/ (+ (- b) (sqrt d)) (* 2 a))) 
				(if (= d 0) 
					(list (/ (- b) (* 2 a))) 
					'()))))

		(resolve 1 -1 -2)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(-1).Cons(ex.NewNumber(2).ToList())), true, "test#"+strconv.Itoa(test))

	test++ // 53 get func
	res, err = Execute(`
		(define get (lambda (l n)
			(if (= n 0)
				(car l)
				(get (cdr l) (- n 1)))))

		(get '(s trtrt 5 laa kooo r 4) 3)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewSymbol("laa")), true, "test#"+strconv.Itoa(test))

	test++ // 54 struct definition via macros
	res, err = Execute(`
		(defmacro apply (f ,args) (cons f args))

		(define list (lambda args args))

		(define pow2 (lambda (x) (* x x)))

		(define get (lambda (l n)
			(if (= n 0)
				(car l)
				(get (cdr l) (- n 1)))))

		(define <= (lambda (a b) (or (< a b) (= a b)) ))

		(define sqrt (lambda (x) 
			(define findi (lambda (i)
				(if (<= (* i i) x)
					(findi (+ i 1))
					(- i 1))))
			(define i (findi 0))
			(define p (/ (- x (* i i)) (* 2 i)))
			(define a (+ i p))
			(- a (/ (* p p) (* 2 a)))))

		(defmacro defstruct args
			(define structname (car args))
			(define funcname (lambda (str) (+ structname '- str)))
			(define methods (lambda (args i)
				(if (not args)
					nil
					(cons
						(list 'define (funcname (+ 'get- (car args))) (list 'lambda '(s) (list 'get 's i)))
						(methods (cdr args) (+ i 1))))))
			(cons 
				'begin 
				(cons 
					(list 'define (funcname 'new) (list 'lambda (cdr args) (cons 'list (cons (list 'quote structname) (cdr args))))) 
					(cons 
						(list 'define (funcname '?) (list 'lambda '(s) (list '= '(car s) (list 'quote structname))))
						(methods (cdr args) 1)))))

		(defstruct point x y)

		(define dist (lambda (p1 p2) 
			(if (not (and (point-? p1) (point-? p2))) (throw '|points expected|))
			(sqrt (+ (pow2 (- (point-get-x p2) (point-get-x p1))) (pow2 (- (point-get-y p2) (point-get-y p1)))))))
		
		(dist (point-new 1 2) (point-new -2 6))`)
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(5)), true, "test#"+strconv.Itoa(test))

	test++ // 55 setl! macros
	res, err = Execute(`
		(define list (lambda args args))

		(define <= (lambda (a b) (or (< a b) (= a b)) ))

		(define get (lambda (l n)
			(if (= n 0)
				(car l)
				(get (cdr l) (- n 1)))))
		
		(defmacro setl! (l pos val)
			(define pos (eval pos))
			(define val (eval val))
			(define mut (lambda (l i v)
				(if (<= i 0) 
					(cons v (cdr l))
					(cons (car l) (mut (cdr l) (- i 1) v)))))
			(list 'set! l (list mut l pos (list 'quote val))))

		(define lst '(s trtrt 5 laa kooo r 4))
		(setl! lst 3 'opo2323p)
		(get lst 3)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewSymbol("opo2323p")), true, "test#"+strconv.Itoa(test))

	test++ // 56 mutable struct definition via macros
	res, err = Execute(`
		(defmacro apply (f ,args) (cons f args))
	
		(define list (lambda args args))
	
		(define pow2 (lambda (x) (* x x)))
	
		(define get (lambda (l n)
			(if (= n 0)
				(car l)
				(get (cdr l) (- n 1)))))
	
		(define <= (lambda (a b) (or (< a b) (= a b)) ))
	
		(define sqrt (lambda (x)
			(define findi (lambda (i)
				(if (<= (* i i) x)
					(findi (+ i 1))
					(- i 1))))
			(define i (findi 0))
			(define p (/ (- x (* i i)) (* 2 i)))
			(define a (+ i p))
			(- a (/ (* p p) (* 2 a)))))
	
		(defmacro setl! (l pos val)
			(define mut (lambda (l i v)
				(if (<= i 0)
					(cons v (cdr l))
					(cons (car l) (mut (cdr l) (- i 1) v)))))
			(list 'set! l (list mut l pos (list 'quote val))))
	
		(defmacro defstruct args
			(define structname (car args))
			(define funcname (lambda (str) (+ structname '- str)))
			(define methods (lambda (args i)
				(if args
					(cons
						(list 'define (funcname (+ 'get- (car args))) (list 'lambda '(s) (list 'get 's i)))
						(cons
							(write (list 'defmacro (funcname (+ 'set- (car args))) '(s v) (list 'list ''setl! 's i 'v)))
							(methods (cdr args) (+ i 1)))))))
			(cons
				'begin
				(cons
					(list 'define (funcname 'new) (list 'lambda (cdr args) (cons 'list (cons (list 'quote structname) (cdr args)))))
					(cons
						(list 'define (funcname '?) (list 'lambda '(s) (list '= '(car s) (list 'quote structname))))
						(methods (cdr args) 1)))))
	
		(defstruct point x y)
	
		(define dist (lambda (p1 p2)
			(if (not (and (point-? p1) (point-? p2))) (throw '|points expected|))
			(sqrt (+ (pow2 (- (point-get-x p2) (point-get-x p1))) (pow2 (- (point-get-y p2) (point-get-y p1)))))))
	
		(define pt1 (point-new 4 2))
		(define pt2 (point-new -2 6))
		(point-set-y pt1 -2)
	
		(dist pt2 pt1)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, res.Output.Equal(ex.NewNumber(10)), true, "test#"+strconv.Itoa(test))

}
