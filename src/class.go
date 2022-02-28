package main

import (
	"fmt"
)

type LoxClass struct {
	name    string
	methods map[string]*LoxFunction
}

var _ LoxCallable = &LoxClass{}

func NewLoxClass(def *StmtClass) *LoxClass {
	return &LoxClass{
		name:    def.Name,
		methods: make(map[string]*LoxFunction),
	}
}

func (class *LoxClass) Arity() int {
	return 0
}

func (class *LoxClass) Call(*Interpreter, []interface{}) (interface{}, error) {
	return NewLoxInstance(class), nil
}

func (class *LoxClass) DefineMethod(name string, fn *LoxFunction) {
	class.methods[name] = fn
}

func (class *LoxClass) FindMethod(name string) *LoxFunction {
	if fn, ok := class.methods[name]; ok {
		return fn
	}
	return nil
}

func (class *LoxClass) String() string {
	return class.name
}

type LoxInstance struct {
	class  *LoxClass
	fileds map[string]interface{}
}

func NewLoxInstance(class *LoxClass) *LoxInstance {
	return &LoxInstance{
		class:  class,
		fileds: make(map[string]interface{}),
	}
}

func (i *LoxInstance) String() string {
	return fmt.Sprintf("%s instance", i.class.name)
}
