package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/batrSens/LispXS/interpreter"
)

func main() {
	newlines := flag.Bool("n", false, "waiting for double newline (\"\\n\\n\")")
	eof := flag.Bool("e", false, "waiting for EOF")
	flag.Parse()

	var prog string
	var err error
	reader := bufio.NewReader(os.Stdin)

	if *newlines {
		line := "  "
		var err error

		for len(line) != 1 {
			prog += line

			line, err = reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
		}
	} else if *eof {
		prog, err = reader.ReadString(0)
		if err != nil && err != io.EOF {
			panic(err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "LispXS v0.1.2")
		_, _ = interpreter.ExecuteStdout(`
            (define repl nil)
            ((lambda ()
              (define define define) (define lambda lambda) (define defmacro defmacro)
              (define if if) (define cons cons) (define car car) (define cdr cdr) (define nil nil)
              (define write write) (define begin begin) (define eval eval) (define read read)
              (define list (lambda args args))
              (defmacro map (f1 ,args1)
                (define helper (lambda (f args)
                  (if args
                    (cons (list f (list quote (car args))) (helper f (cdr args))))))
                (cons list (helper f1 args1)))
              (define writeln (lambda (sym) (write sym) (write '|\n|)))
              (defmacro repl1 ()
                (list begin (list write ''|> |) (list map writeln (list map eval (list read))) (list repl1)))
              (set! repl repl1)))
            (repl)`)
		return
	}

	res, err := interpreter.ExecuteStdout(prog)
	if err != nil {
		panic(err)
		return
	}

	fmt.Println(">", res.ToString())
}
