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
and whose result will be returned from `if` in case of result of conditional is not nil, third - else. If the third argument 
is missing - `if` returns nil in case result of conditional is nil.

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

### `begin`

Returns result of last expression (nil in case of zero number of arguments).

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
