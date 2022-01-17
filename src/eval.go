package main

type Interpreter struct {
}

func (i *Interpreter) Interprete(expr Expr) (interface{}, error) {
	return expr.Accept(i)
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

func checkBoolOperands(operands ...interface{}) bool {
	for _, operand := range operands {
		if _, ok := operand.(bool); !ok {
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

func (i *Interpreter) VisitLiteral(expr *ExprLiteral) (interface{}, error) {
	return expr.Value.Value(), nil
}

func (i *Interpreter) VisitUnary(expr *ExprUnary) (interface{}, error) {
	right, err := i.Interprete(expr.Expression)
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
	return i.Interprete(expr.Expression)
}

func (i *Interpreter) VisitBinary(expr *ExprBinary) (interface{}, error) {
	left, err := i.Interprete(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := i.Interprete(expr.Right)
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

	// logic
	case AND:
		if checkBoolOperands(left, right) {
			return left.(bool) && right.(bool), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of and must be two bools")
	case OR:
		if checkBoolOperands(left, right) {
			return left.(bool) || right.(bool), nil
		}
		return nil, i.runtimeError(expr.Operator, "operands of or must be two bools")

	default:
		panic("golox error: invalid binary operator type")
	}
}
