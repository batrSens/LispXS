# LispXS

Minimal expandable and ~~embedded~~ lisp.

## Installation

build (go1.13.7):
```shell script
$ git clone git@github.com:batrSens/LispX.git
$ cd ./LispX
$ go build
```

tests running:
```shell script
$ go test {lispxs_directory}/...
```

run:
```shell script
$ {lispxs_directory}/LispX
```

## Procedures

### `quote`

Returns expression without calculation. Expects one argument. 
Following entries are equivalent: `(quote {EXPR})`, `'{EXPR}`.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(quote (one two))
</pre></td><td><pre>
(one two)
</pre></td></tr>

<tr><td><pre>
(quote 23)
</pre></td><td><pre>
23
</pre></td></tr>

</table>
</details>

---

### `eval`

Evaluates result of expression. Expects one argument.
Following entries are equivalent: `(eval (quote {EXPR}))`, `{EXPR}`.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(eval '(+ 23 32))
</pre></td><td><pre>
55
</pre></td></tr>

<tr><td><pre>
(eval 23)
</pre></td><td><pre>
23
</pre></td></tr>

</table>
</details>

---

### `define`

Defines variable in current scope. 
Expected two variables: first - symbol, second - an expression whose result will be saved and returned from `define`.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(define a 2)
</pre></td><td><pre>
2
</pre></td></tr>

<tr><td><pre>
(define a '|some string|)
</pre></td><td><pre>
some string
</pre></td></tr>

<tr><td><pre>
(define a (+ 20 3)) 
(+ a 32)
</pre></td><td><pre>
55
</pre></td></tr>

</table>
</details>

---

### `set!`

Redefines existed variable in nearest scope. 
Expected two variables: first - symbol, second - an expression whose result will be saved and returned from `define`.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(define a 2) 
(set! a 3)
</pre></td><td><pre>
3
</pre></td></tr>

<tr><td><pre>
(define a 5) 
((lambda (b) (set! a (+ a b)) 50) 
a
</pre></td><td><pre>
55
</pre></td></tr>

</table>
</details>

---

### `lambda`

Returns new closure with current parent scope. When it closure will be called, a new scope is created.
Expected at least two variables: first - list with symbols that means arguments or symbol that means list of arguments,
second and subsequent - body of closure. Closure returns result of last expression of body.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(define a (lambda (b) (+ b b))) 
(a 3)
</pre></td><td><pre>
6
</pre></td></tr>

<tr><td><pre>
(define list (lambda args args))
(list 3 (+ 5 4) 0)
</pre></td><td><pre>
(3 9 0)
</pre></td></tr>

<tr><td><pre>
(define a 5) 
((lambda (b) (set! a (+ a b)) 50) 
a
</pre></td><td><pre>
55
</pre></td></tr>

</table>
</details>

---

### `defmacro`

Defines macro in current scope. When it macro will be called, new code will be created and then executed.
Expected at least three variables: first - symbol, second - list with symbols that means arguments or 
symbol that means list of arguments, third and subsequent - body of macro. Macro executes result of last expression of body.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(define list (lambda args args))
(define a 2) 
(defmacro set10! (b) (list 'set! b 10))
(set10! a)
a
</pre></td><td><pre>
10
</pre></td></tr>

<tr><td><pre>
(defmacro apply s (define f (car s)) (define args (eval (car (cdr s)))) (cons f args))
(apply + '(3 4 5))
</pre></td><td><pre>
12
</pre></td></tr>

</table>
</details>

---

### `if`

Conditional operator. Expected two or three arguments: first - conditional, second - expression that will be calculated
and whose result will be returned from `if` in case of result of conditional is not `nil`, third - else. If the third argument 
is missing - `if` returns `nil` in case result of conditional is `nil`.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(if (> 2 3) 'two 'three)
</pre></td><td><pre>
three
</pre></td></tr>

<tr><td><pre>
(if (> 2 1) 'two)
</pre></td><td><pre>
two
</pre></td></tr>

<tr><td><pre>
(if (> 2 3) 'two)
</pre></td><td><pre>
nil
</pre></td></tr>

</table>
</details>

---

### `or`

Calculates expressions until it meats not `nil` value. Returns this value. If all results of expressions are `nil` then returns `nil`.
Returns `nil` in case of zero number of arguments.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(or 2 3)
</pre></td><td><pre>
2
</pre></td></tr>

<tr><td><pre>
(define a 3)
(or nil (define a 5) (define a 10))
a
</pre></td><td><pre>
5
</pre></td></tr>

<tr><td><pre>
(or (cdr '(2)) nil)
</pre></td><td><pre>
nil
</pre></td></tr>

<tr><td><pre>
(or)
</pre></td><td><pre>
nil
</pre></td></tr>

</table>
</details>

---

### `and`

Calculates expressions until it meats `nil` value. If one of results of expressions is `nil` then returns `nil`. Result of last
expression otherwise. Returns `T` in case of zero number of arguments.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(and 2 3)
</pre></td><td><pre>
3
</pre></td></tr>

<tr><td><pre>
(define a 3)
(and nil (define a 5) (define a 10))
a
</pre></td><td><pre>
3
</pre></td></tr>

<tr><td><pre>
(and (cdr '(3 3)) 'YY)
</pre></td><td><pre>
YY
</pre></td></tr>

<tr><td><pre>
(and)
</pre></td><td><pre>
T
</pre></td></tr>

</table>
</details>

---

### `write`

Writes string representation of expression's result to output channel. Returns it result. Expected one argument.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td><td>out</td></tr>

<tr><td><pre>
(write '(2))
</pre></td><td><pre>
(2)
</pre></td><td><pre>
(2)
</pre></td></tr>

<tr><td><pre>
(write (if T 'ss 3))
</pre></td><td><pre>
ss
</pre></td><td><pre>
ss
</pre></td></tr>

</table>
</details>

---

### `read`

Reads string representation of expression from output channel and returns this expression. Expected zero number of arguments.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td><td>in</td></tr>

<tr><td><pre>
(read)
</pre></td><td><pre>
(2)
</pre></td><td><pre>
(2)
</pre></td></tr>

<tr><td><pre>
(write (if T 'ss 3))
</pre></td><td><pre>
ss
</pre></td><td><pre>
ss
</pre></td></tr>

</table>
</details>

---

### `begin`

Returns result of last expression (`nil` in case of zero number of arguments).

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(begin)
</pre></td><td><pre>
nil
</pre></td></tr>

<tr><td><pre>
(begin 4 (+ 4 5) 'qwerqwe (/ 3 (- 1 2)))
</pre></td><td><pre>
-3
</pre></td></tr>

</table>
</details>

---

### `cons`

Returns new pair. Expected two arguments: first will be 'car' of new pair, second - 'cdr'. Second argument must be a pair or `nil`.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(cons 2 nil)
</pre></td><td><pre>
(2)
</pre></td></tr>

<tr><td><pre>
(cons (+ 5 6) '(12 13 14))
</pre></td><td><pre>
(11 12 13 14)
</pre></td></tr>

</table>
</details>

---

### `car`

Returns 'car' of pair. Expected one argument that must be a pair.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(car '(2))
</pre></td><td><pre>
2
</pre></td></tr>

<tr><td><pre>
(car '((4 5) 6 7))
</pre></td><td><pre>
(4 5)
</pre></td></tr>

</table>
</details>

---

### `cdr`

Returns 'cdr' of pair. Expected one argument that must be a pair.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(cdr '(2))
</pre></td><td><pre>
nil
</pre></td></tr>

<tr><td><pre>
(cdr '((4 5) 6 7))
</pre></td><td><pre>
(6 7)
</pre></td></tr>

</table>
</details>

---

### `symbol->number`

Converts symbol to number. Expected one argument that must be a symbol that name equal to string representation of any number.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(symbol->number '|  23 |)
</pre></td><td><pre>
23
</pre></td></tr>

<tr><td><pre>
(symbol->number '|6/3|)
</pre></td><td><pre>
2
</pre></td></tr>

<tr><td><pre>
(symbol->number '|1234.56e-2|)
</pre></td><td><pre>
12.3456
</pre></td></tr>

<tr><td><pre>
(symbol->number '|  -23.4|)
</pre></td><td><pre>
-23.4
</pre></td></tr>

</table>
</details>

---

### `number->symbol`

Converts number to symbol with name that equal to string representation of number. Expected one argument that must be a number.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(define |234| (lambda (a) (+ a 1)))
((eval (number->symbol 234)) 5)
</pre></td><td><pre>
6
</pre></td></tr>

</table>
</details>

---

### `symbol?`

Returns `T` if argument is a symbol and `nil` otherwise. Expected one argument.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(symbol? '|2|)
</pre></td><td><pre>
T
</pre></td></tr>

<tr><td><pre>
(symbol? 2)
</pre></td><td><pre>
nil
</pre></td></tr>

</table>
</details>

---

### `number?`

Returns `T` if argument is a number and `nil` otherwise. Expected one argument.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(number? 2)
</pre></td><td><pre>
T
</pre></td></tr>

<tr><td><pre>
(number? '|2|)
</pre></td><td><pre>
nil
</pre></td></tr>

</table>
</details>

---

### `pair?`

Returns `T` if argument is a pair and `nil` otherwise. Expected one argument.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(pair? '(2 3 4 5))
</pre></td><td><pre>
T
</pre></td></tr>

<tr><td><pre>
(pair? nil)
</pre></td><td><pre>
nil
</pre></td></tr>

</table>
</details>

---

### `not`

Returns `T` if argument is `nil` and `nil` otherwise. Expected one argument.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(not nil)
</pre></td><td><pre>
T
</pre></td></tr>

<tr><td><pre>
(not 1234)
</pre></td><td><pre>
nil
</pre></td></tr>

</table>
</details>

---

### `=`

Returns `T` if argument are equivalent and `nil` otherwise. Expected at least two arguments.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(= 's2 (begin 's2) (if nil 2 's2) (+ 's '|2|))
</pre></td><td><pre>
T
</pre></td></tr>

<tr><td><pre>
(= '(2 3) (cons 2 '(3)))
</pre></td><td><pre>
T
</pre></td></tr>

<tr><td><pre>
(= 2 '|2|)
</pre></td><td><pre>
nil
</pre></td></tr>

<tr><td><pre>
(= 2 2 2 2 2 3)
</pre></td><td><pre>
nil
</pre></td></tr>

</table>
</details>

---

### `>`

Returns `T` if first argument more than second and `nil` otherwise. Expected two numbers.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(> 3 2)
</pre></td><td><pre>
T
</pre></td></tr>

<tr><td><pre>
(> 3 3)
</pre></td><td><pre>
nil
</pre></td></tr>

</table>
</details>

---

### `<`

Returns `T` if first argument less than second and `nil` otherwise. Expected two numbers.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(< -7 2)
</pre></td><td><pre>
T
</pre></td></tr>

<tr><td><pre>
(< 4 2)
</pre></td><td><pre>
nil
</pre></td></tr>

</table>
</details>

---

### `len`

Returns length `T` of symbol's name in characters. Expected one symbol.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(len '12345)
</pre></td><td><pre>
5
</pre></td></tr>

<tr><td><pre>
(len '漢字!)
</pre></td><td><pre>
3
</pre></td></tr>

</table>
</details>

---

### `+`

Returns sum of numbers or symbol that name is concatenation of names all symbols in arguments. 
Expected any quantity of numbers or at least one symbol.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(+ 2 3 4)
</pre></td><td><pre>
9
</pre></td></tr>

<tr><td><pre>
(+ 'hello, '| world!|)
</pre></td><td><pre>
hello, world!
</pre></td></tr>

<tr><td><pre>
(+)
</pre></td><td><pre>
0
</pre></td></tr>

</table>
</details>

---

### `-`

Returns difference of numbers or symbol that name is substring of symbol's name ( [second_arg, third_arg) ). 
Expected any quantity of numbers or least one symbol and two numbers.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(- 2)
</pre></td><td><pre>
-2
</pre></td></tr>

<tr><td><pre>
(- 2 3 4)
</pre></td><td><pre>
-5
</pre></td></tr>

<tr><td><pre>
(- 'thisstringishuge 4 10)
</pre></td><td><pre>
string
</pre></td></tr>

<tr><td><pre>
(-)
</pre></td><td><pre>
0
</pre></td></tr>

</table>
</details>

---

### `*`

Returns product of numbers. Expected any quantity of numbers.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(* 2 3)
</pre></td><td><pre>
6
</pre></td></tr>

<tr><td><pre>
(*)
</pre></td><td><pre>
1
</pre></td></tr>

</table>
</details>

---

### `*`

Returns division of numbers. Expected at least one number.

<details>
<summary>examples</summary>

<table><tr><td>usage</td><td>result</td></tr>

<tr><td><pre>
(/ 6 2 3)
</pre></td><td><pre>
0.75
</pre></td></tr>

<tr><td><pre>
(/ 7)
</pre></td><td><pre>
7
</pre></td></tr>

</table>
</details>
