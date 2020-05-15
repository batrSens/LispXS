package main

import (
	"fmt"

	"github.com/batrSens/LispX/interpreter"
)

func main() {
	res, err := interpreter.ExecuteStdout("(define println (lambda (x) (display x)(display \"\\n\") ) ) (println 5) (println 6) (/ 4 0) (println 2) 9")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//

	fmt.Println(res.ToString())
}
