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
	init := class.FindMethod("init")
	if init == nil {
		return 0
	}
	return init.Arity()
}

func (class *LoxClass) Call(i *Interpreter, args []interface{}) (interface{}, error) {
	instance := NewLoxInstance(class)
	init := class.FindMethod("init")
	if init == nil {
		return instance, nil
	}
	bind(init, instance)
	return init.Call(i, args)
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
