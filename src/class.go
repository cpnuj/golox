package main

import "fmt"

type LoxClass struct {
	name string
}

var _ LoxCallable = &LoxClass{}

func NewLoxClass(def *StmtClass) *LoxClass {
	return &LoxClass{
		name: def.Name,
	}
}

func (class *LoxClass) Arity() int {
	return 0
}

func (class *LoxClass) Call(*Interpreter, []interface{}) (interface{}, error) {
	return NewLoxInstance(class), nil
}

type LoxInstance struct {
	class *LoxClass
	// fileds map[string]interface{}
}

func NewLoxInstance(class *LoxClass) *LoxInstance {
	return &LoxInstance{
		class: class,
	}
}

func (i *LoxInstance) String() string {
	return fmt.Sprintf("%s instance", i.class.name)
}
