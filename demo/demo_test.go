package demo

import (
	"fmt"
	"github.com/u2go/go-tiny-supervisor/lib/fn"
	"testing"
)

func TestByte(t *testing.T) {
	b := []byte("")
	t.Log(len(b), b == nil)
}

func TestJson(t *testing.T) {
	fmt.Println(fn.JsonPretty(map[string]interface{}{}))
	fmt.Println("----")
}
