package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/batrSens/LispX/interpreter"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	prog := ""
	line := "  "
	var err error

	for len(line) != 1 {
		prog += line

		line, err = reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
	}

	res, err := interpreter.ExecuteStdout(prog)
	if err != nil {
		panic(err)
		return
	}

	fmt.Println(">", res.ToString())
}
