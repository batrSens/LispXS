package parser

import (
	"fmt"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestParser(t *testing.T) {
	debugT(t, "2 3")
	debugT(t, "() nil 2 (+ 2 3) \"end\" (cons 8 '(3 4))")
}

func debugT(t *testing.T, text string) {
	prs := NewParser(text)

	res, err := prs.Parse()
	assert.Equal(t, err, nil)

	fmt.Println(res.ToString())
}
