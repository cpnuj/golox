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
	definition    *StmtFun
	closure       *Environment
	isInitializer bool // is class initializer
}

func NewLoxFunction(definition *StmtFun, closure *Environment, isInitializer bool) *LoxFunction {
	return &LoxFunction{
		definition:    definition,
		closure:       closure,
		isInitializer: isInitializer,
	}
}

func (f *LoxFunction) Arity() int {
	return len(f.definition.Params)
}

func (f *LoxFunction) Call(i *Interpreter, args []interface{}) (ret interface{}, err error) {
	env := NewEnvironment(f.closure)
	params := f.definition.Params
	for i := range args {
		env.Define(params[i], args[i])
	}

	// return this if f is init function
	defer func() {
		if f.isInitializer {
			ret, _ = f.closure.Get("this", 1)
		}
	}()

	// catch return value
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		retval, isReturn := r.(*Return)
		if !isReturn {
			panic(r)
		}

		ret = retval.Value
	}()

	i.execBlock(f.definition.Body, env)

	return
}
