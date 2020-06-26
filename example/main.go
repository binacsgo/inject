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
type B interface{}

// BImpl implement
type BImpl struct {
	Ba A `inject-name:"A"`
}

func main() {
	fmt.Println("Example program:")
	b := initService()
	fmt.Printf("b = %+v, a = %+v\n", b, b.Ba)
}

func initService() *BImpl {
	b := BImpl{}

	inject.Regist("A", &AImpl{})
	inject.Regist("B", &b)

	err := inject.DoInject()
	if err != nil {
		panic(err.Error())
	}
	return &b
}
