package main

import "fmt"

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

func (env *Environment) Set(name string, value interface{}) bool {
	if _, found := env.values[name]; found {
		env.values[name] = value
		return true
	}

	if env.parent != nil {
		return env.parent.Set(name, value)
	}

	return false
}

func (env *Environment) Get(name string) (interface{}, bool) {
	if value, found := env.values[name]; found {
		return value, true
	}

	if env.parent != nil {
		return env.parent.Get(name)
	}

	return nil, false
}

type Interpreter struct {
	environment *Environment
}

var (
	_ ExprVisitor = &Interpreter{}
	_ StmtVisitor = &Interpreter{}
)

func NewInterpreter() *Interpreter {
	return &Interpreter{
		environment: NewEnvironment(nil),
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
	return expr.Value.Value(), nil
}

func (i *Interpreter) VisitVariable(expr *ExprVariable) (interface{}, error) {
	name := expr.Name.Value().(string)
	value, ok := i.environment.Get(name)
	if !ok {
		return nil, i.runtimeError(expr.Name, "undefined variable "+name)
	}
	return value, nil
}

func (i *Interpreter) VisitAssign(expr *ExprAssign) (interface{}, error) {
	name := expr.Name.Value().(string)
	value, err := i.eval(expr.Value)
	if err != nil {
		return nil, err
	}

	if ok := i.environment.Set(name, value); !ok {
		return nil, i.runtimeError(expr.Name, "undefined variable "+name)
	}

	return value, nil
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

	i.environment.Define(name, initializer)

	return nil, nil
}

func (i *Interpreter) VisitBlock(statement *StmtBlock) (interface{}, error) {
	// enter new environment
	previous := i.environment
	i.environment = NewEnvironment(previous)

	// back to old environment
	defer func() {
		i.environment = previous
	}()

	for _, stmt := range statement.Statements {
		if err := i.execute(stmt); err != nil {
			return nil, err
		}
	}

	return nil, nil
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
