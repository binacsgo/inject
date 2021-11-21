package main

import (
	"fmt"

	"github.com/binacsgo/inject"
)

// A interface B
type A interface{}

// AImpl implement
type AImpl struct{}

// B interface B
type B interface {
	Work()
}

// BImpl implement
type BImpl struct {
	Ba A `inject-name:"A"`
}

// Work implement the B interface
func (b *BImpl) Work() {
	fmt.Printf("b = %+v, a = %+v\n", b, b.Ba)
}

func main() {
	b := initService()
	b.Work()
}

func initService() B {
	b := BImpl{}
	inject.Regist("B", &b)
	inject.Regist("A", &AImpl{})
	inject.DoInject()
	return &b
}
