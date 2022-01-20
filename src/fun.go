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
