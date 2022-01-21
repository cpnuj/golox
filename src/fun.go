package main

import (
	"errors"
	"time"
)

type LoxCallable interface {
	Arity() int
	Call(*Interpreter, []interface{}) (interface{}, error)
}

// builtin
type BuildinFun struct {
	name  string
	arity int
	call  func(*Interpreter, []interface{}) (interface{}, error)
}

func (b *BuildinFun) Call(i *Interpreter, args []interface{}) (interface{}, error) {
	return b.call(i, args)
}

func (b *BuildinFun) Arity() int {
	return b.arity
}

// clock
var BuildinClock *BuildinFun = &BuildinFun{
	name:  "clock",
	arity: 0,
	call: func(i *Interpreter, args []interface{}) (interface{}, error) {
		return float64(time.Now().Unix()), nil
	},
}

// sleep
var BuildinSleep *BuildinFun = &BuildinFun{
	name:  "sleep",
	arity: 1,
	call: func(i *Interpreter, args []interface{}) (interface{}, error) {
		f, ok := args[0].(float64)
		if !ok {
			return nil, errors.New("sleep only accept float64 as arg")
		}
		time.Sleep(time.Duration(f) * time.Second)
		return nil, nil
	},
}

// user defined funtion
type LoxFunction struct {
	definition *StmtFun
	closure    *Environment
}

func NewLoxFunction(definition *StmtFun, closure *Environment) *LoxFunction {
	return &LoxFunction{
		definition: definition,
		closure:    closure,
	}
}

func (f *LoxFunction) Arity() int {
	return len(f.definition.Params)
}

func (f *LoxFunction) Call(i *Interpreter, args []interface{}) (interface{}, error) {
	env := NewEnvironment(f.closure)
	params := f.definition.Params
	for i := range args {
		env.Define(params[i], args[i])
	}

	// catch return value
	_, err := i.execBlock(f.definition.Body, env)
	if err != nil {
		if retval, ok := err.(*Return); ok {
			return retval.Value(), nil
		} else {
			return nil, err
		}
	}

	return nil, nil
}
