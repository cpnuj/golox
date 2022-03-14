package main

import (
	"fmt"
)

type Environment struct {
	parent *Environment
	values map[string]interface{}
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent: parent,
		values: make(map[string]interface{}),
	}
}

func (env *Environment) Define(name string, value interface{}) {
	env.values[name] = value
}

func (env *Environment) Set(name string, depth int, value interface{}) bool {
	if depth > 0 {
		return env.parent.Set(name, depth-1, value)
	}

	if _, found := env.values[name]; found {
		env.values[name] = value
		return true
	}

	return false
}

func (env *Environment) Get(name string, depth int) (interface{}, bool) {
	if depth > 0 {
		return env.parent.Get(name, depth-1)
	}

	if value, found := env.values[name]; found {
		return value, true
	}

	return nil, false
}

type Interpreter struct {
	globalEnv *Environment
	localEnv  *Environment
	locals    map[Expr]int
}

var (
	_ ExprVisitor = &Interpreter{}
	_ StmtVisitor = &Interpreter{}
)

func NewInterpreter() *Interpreter {
	global := NewEnvironment(nil)
	global.Define("clock", BuildinClock)
	global.Define("sleep", BuildinSleep)

	return &Interpreter{
		globalEnv: global,
		localEnv:  NewEnvironment(nil),
	}
}

func (i *Interpreter) Interprete(statements []Stmt) error {
	for _, statement := range statements {
		err := i.execute(statement)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) SetLocals(locals map[Expr]int) {
	i.locals = locals
}

func (i *Interpreter) execute(statement Stmt) error {
	_, err := statement.Accept(i)
	return err
}

func (i *Interpreter) runtimeError(token Token, msg string) error {
	row, col := token.Pos()
	return logger.NewError(row, col, msg)
}

func checkNumOperands(operands ...interface{}) bool {
	for _, operand := range operands {
		if _, ok := operand.(float64); !ok {
			return false
		}
	}
	return true
}

func checkStringOperands(operands ...interface{}) bool {
	for _, operand := range operands {
		if _, ok := operand.(string); !ok {
			return false
		}
	}
	return true
}

func isTruthy(obj interface{}) bool {
	if obj == nil {
		return false
	}
	if b, ok := obj.(bool); ok {
		return b
	}
	return true
}

func (i *Interpreter) eval(expr Expr) (interface{}, error) {
	return expr.Accept(i)
}

func (i *Interpreter) VisitLiteral(expr *ExprLiteral) (interface{}, error) {
	return expr.Value, nil
}

func (i *Interpreter) VisitVariable(expr *ExprVariable) (interface{}, error) {
	name := expr.Name.Value().(string)

	// defined in local
	if depth, ok := i.locals[expr]; ok {
		value, ok := i.localEnv.Get(name, depth)
		if !ok {
			return nil, i.runtimeError(expr.Name, "undefined variable "+name)
		}
		return value, nil
	}

	// search in global
	if value, ok := i.globalEnv.Get(name, 0); ok {
		return value, nil
	}

	return nil, i.runtimeError(expr.Name, "undefined variable "+name)
}

func (i *Interpreter) VisitAssign(expr *ExprAssign) (interface{}, error) {
	name := expr.Name.Value().(string)
	value, err := i.eval(expr.Value)
	if err != nil {
		return nil, err
	}

	// defined in local
	if depth, ok := i.locals[expr]; ok {
		if ok := i.localEnv.Set(name, depth, value); !ok {
			return nil, i.runtimeError(expr.Name, "Lox Error: undefined variable "+name)
		}
		return value, nil
	}

	// search in global
	if ok := i.globalEnv.Set(name, 0, value); ok {
		return value, nil
	}

	return nil, i.runtimeError(expr.Name, "Lox Error: undefined variable "+name)
}

func (i *Interpreter) VisitUnary(expr *ExprUnary) (interface{}, error) {
	right, err := i.eval(expr.Expression)
	if err != nil {
		return nil, err
	}

	switch expr.UnaryOperator.Type() {
	case MINUS:
		if !checkNumOperands(right) {
			return nil, i.runtimeError(expr.UnaryOperator, "operand of - must be a number")
		}
		return -right.(float64), nil
	case BANG:
		return !isTruthy(right), nil
	default:
		panic("golox error: invalid unary operator type")
	}
}

func (i *Interpreter) VisitGrouping(expr *ExprGrouping) (interface{}, error) {
	return i.eval(expr.Expression)
}

func (i *Interpreter) VisitBinary(expr *ExprBinary) (interface{}, error) {
	left, err := i.eval(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := i.eval(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type() {
	case PLUS:
		if checkNumOperands(left, right) {
			return left.(float64) + right.(float64), nil
		}
		if checkStringOperands(left, right) {
			return left.(string) + right.(string), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of + must be two strings or two numbers")
	case MINUS:
		if checkNumOperands(left, right) {
			return left.(float64) - right.(float64), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of - must be two numbers")
	case STAR:
		if checkNumOperands(left, right) {
			return left.(float64) * right.(float64), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of * must be two numbers")
	case SLASH:
		if checkNumOperands(left, right) {
			return left.(float64) / right.(float64), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of / must be two numbers")
	case GREATER:
		if checkNumOperands(left, right) {
			return left.(float64) > right.(float64), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of > must be two numbers")
	case GREATER_EQUAL:
		if checkNumOperands(left, right) {
			return left.(float64) >= right.(float64), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of >= must be two numbers")
	case LESS:
		if checkNumOperands(left, right) {
			return left.(float64) < right.(float64), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of < must be two numbers")
	case LESS_EQUAL:
		if checkNumOperands(left, right) {
			return left.(float64) <= right.(float64), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of <= must be two numbers")

	// FIXME: Support == and != operator for other objects
	case EQUAL_EQUAL:
		if checkNumOperands(left, right) {
			return left.(float64) == right.(float64), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of == must be two numbers")
	case BANG_EQUAL:
		if checkNumOperands(left, right) {
			return left.(float64) != right.(float64), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of != must be two numbers")

	default:
		panic("golox error: invalid binary operator type")
	}
}

func (i *Interpreter) VisitLogical(expr *ExprLogical) (interface{}, error) {
	left, err := i.eval(expr.Left)
	if err != nil {
		return nil, err
	}

	if expr.Operator.Type() == OR {
		if isTruthy(left) {
			return left, nil
		}
	} else {
		if !isTruthy(left) {
			return left, nil
		}
	}

	return i.eval(expr.Right)
}

func (i *Interpreter) VisitCall(expr *ExprCall) (interface{}, error) {
	callee, err := i.eval(expr.Callee)
	if err != nil {
		return nil, err
	}

	function, callable := callee.(LoxCallable)
	if !callable {
		return nil, i.runtimeError(expr.Paren, "totally not a function")
	}

	args := make([]interface{}, 0)
	for _, arg := range expr.Args {
		value, err := i.eval(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, value)
	}

	if len(args) > function.Arity() {
		return nil, i.runtimeError(expr.Paren, "too many arguments")
	} else if len(args) < function.Arity() {
		return nil, i.runtimeError(expr.Paren, "too few arguments")
	}

	return function.Call(i, args)
}

// bind binds a class method to an instance of the class
func bind(method *LoxFunction, instance *LoxInstance) bool {
	// The this variable must be defined in the closure's parent environment
	// if the method is a valid class method.
	return method.closure.Set("this", 1, instance)
}

func (i *Interpreter) VisitGet(expr *ExprGet) (interface{}, error) {
	value, err := i.eval(expr.Object)
	if err != nil {
		return nil, err
	}

	obj, ok := value.(*LoxInstance)
	if !ok {
		return nil, i.runtimeError(expr.Dot, "only instances have fields")
	}

	filed := expr.Field.Value().(string)
	if ret, ok := obj.fileds[filed]; ok {
		return ret, nil
	}
	if fn := obj.class.FindMethod(filed); fn != nil {
		if !bind(fn, obj) {
			return nil, i.runtimeError(expr.Dot, "Lox error: cannot bind method and instance")
		}
		return fn, nil
	}
	return nil, i.runtimeError(expr.Field, "no such field")
}

func (i *Interpreter) VisitSet(expr *ExprSet) (interface{}, error) {
	value, err := i.eval(expr.Object)
	if err != nil {
		return nil, err
	}

	obj, ok := value.(*LoxInstance)
	if !ok {
		return nil, i.runtimeError(expr.Dot, "only instances have fields")
	}

	ret, err := i.eval(expr.Value)
	if err != nil {
		return nil, err
	}
	filed := expr.Field.Value().(string)
	obj.fileds[filed] = ret
	return ret, nil
}

func (i *Interpreter) VisitThis(expr *ExprThis) (interface{}, error) {
	err := i.runtimeError(expr.Keyword, "Lox error: cannot resolve this")
	// get in local env
	if depth, ok := i.locals[expr]; ok {
		if this, ok := i.localEnv.Get("this", depth); ok {
			return this, nil
		}
	}
	return nil, err
}

func (i *Interpreter) VisitExpression(statement *StmtExpression) (interface{}, error) {
	return i.eval(statement.Expression)
}

func (i *Interpreter) VisitPrint(statement *StmtPrint) (interface{}, error) {
	value, err := i.eval(statement.Expression)
	if err != nil {
		return nil, err
	}
	fmt.Println(value)
	return nil, nil
}

func (i *Interpreter) VisitVar(statement *StmtVar) (interface{}, error) {
	var name string
	var initializer interface{}
	var err error

	name = statement.Name.Value().(string)

	if statement.Initializer != nil {
		initializer, err = i.eval(statement.Initializer)
		if err != nil {
			return nil, err
		}
	}

	i.localEnv.Define(name, initializer)

	return nil, nil
}

func (i *Interpreter) execBlock(statements []Stmt, env *Environment) (interface{}, error) {
	// enter new environment
	previous := i.localEnv
	i.localEnv = env

	// back to old environment
	defer func() {
		i.localEnv = previous
	}()

	for _, statement := range statements {
		if err := i.execute(statement); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (i *Interpreter) VisitBlock(statement *StmtBlock) (interface{}, error) {
	return i.execBlock(statement.Statements, NewEnvironment(i.localEnv))
}

func (i *Interpreter) VisitIf(statement *StmtIf) (interface{}, error) {
	cond, err := i.eval(statement.Cond)
	if err != nil {
		return nil, err
	}
	if isTruthy(cond) {
		return nil, i.execute(statement.Then)
	} else if statement.Else != nil {
		return nil, i.execute(statement.Else)
	}
	return nil, nil
}

func (i *Interpreter) VisitWhile(statement *StmtWhile) (interface{}, error) {
	for {
		cond, err := i.eval(statement.Cond)
		if err != nil {
			return nil, err
		}
		if !isTruthy(cond) {
			break
		}
		err = i.execute(statement.Body)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) VisitFun(statement *StmtFun) (interface{}, error) {
	fn := NewLoxFunction(statement, i.localEnv, false)
	i.localEnv.Define(statement.Name, fn)
	return fn, nil
}

// Return implements the error interface, and throwed in VisitReturn.
// Should be caught at LoxFunction.Call
type Return struct {
	value interface{}
}

func (r *Return) Value() interface{} {
	return r.value
}

func (r *Return) Error() string {
	return ""
}

func (i *Interpreter) VisitReturn(statement *StmtReturn) (interface{}, error) {
	if statement.Value != nil {
		value, err := i.eval(statement.Value)
		if err != nil {
			return nil, err
		}
		return nil, &Return{value: value}
	}
	return nil, nil
}

func (i *Interpreter) defineMethod(class *LoxClass, statement *StmtFun) {
	var fn *LoxFunction
	if statement.Name == "init" {
		fn = NewLoxFunction(statement, i.localEnv, true /* isInitializer */)
	} else {
		fn = NewLoxFunction(statement, i.localEnv, false)
	}
	class.DefineMethod(statement.Name, fn)
}

func (i *Interpreter) VisitClass(statement *StmtClass) (interface{}, error) {
	var cls *LoxClass

	if statement.Superclass == nil {
		cls = NewLoxClass(statement, nil)
	} else {
		superclass, err := i.VisitVariable(statement.Superclass)
		if err != nil {
			return nil, err
		}
		cls = NewLoxClass(statement, superclass.(*LoxClass))
	}

	i.localEnv.Define(statement.Name, cls)

	// define a new environment to store pointer this
	preEnv := i.localEnv
	i.localEnv = NewEnvironment(i.localEnv)
	i.localEnv.Define("this", nil)

	// yet another environment to store class methods
	// TODO: remove this new env create
	i.localEnv = NewEnvironment(i.localEnv)
	for _, method := range statement.Methods {
		i.defineMethod(cls, method)
	}

	// quit to origin env
	i.localEnv = preEnv
	return cls, nil
}
