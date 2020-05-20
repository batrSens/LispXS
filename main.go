package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/batrSens/LispX/interpreter"
)

func main() {
	newlines := flag.Bool("n", false, "wait for double newline (\"\\n\\n\") instead of EOF")
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
	} else {
		prog, err = reader.ReadString(0)
		if err != nil && err != io.EOF {
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
