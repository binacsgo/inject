package main

import (
	"fmt"
	"testing"
)

// inplement the A interface
type FakeA struct{}

func newFakeA() A {
	return &FakeA{}
}

func (fakeA *FakeA) Hello() {
	fmt.Printf("Hello I'm FakeA used for unittest!")
}

func TestBImpl_Work(t *testing.T) {
	type fields struct {
		Ba A
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test",
			fields: fields{
				Ba: newFakeA(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BImpl{
				Ba: tt.fields.Ba,
			}
			b.Hello()
		})
	}
}
