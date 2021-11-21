package main

import (
	"fmt"

	"github.com/binacsgo/inject"
)

// A interface B
type A interface {
	Hello()
}

// AImpl implement
type AImpl struct{}

// Work implement the B interface
func (a *AImpl) Hello() {
	fmt.Printf("Hello I'm A\n")
}

// B interface B
type B interface {
	Hello()
}

// BImpl implement
type BImpl struct {
	Ba A `inject-name:"A"`
}

// Work implement the B interface
func (b *BImpl) Hello() {
	b.Ba.Hello()
}

func main() {
	b := initService()
	b.Hello()
}

func initService() B {
	b := BImpl{}
	inject.Regist("B", &b)
	inject.Regist("A", &AImpl{})
	inject.DoInject()
	return &b
}
