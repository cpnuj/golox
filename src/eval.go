package main

type Interpreter struct {
}

func eval(expr Expr) interface{} {
	return expr.Accept(&Interpreter{})
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

func (p *Interpreter) VisitLiteral(expr *ExprLiteral) interface{} {
	return expr.Value.Value()
}

func (p *Interpreter) VisitUnary(expr *ExprUnary) interface{} {
	right := eval(expr)

	switch expr.UnaryOperator.Type() {
	case MINUS:
		if num, ok := right.(float64); ok {
			return -num
		}
	case BANG:
		return !isTruthy(right)
	}

	return nil
}

func (p *Interpreter) VisitGrouping(expr *ExprGrouping) interface{} {
	return eval(expr.Expression)
}

func (p *Interpreter) VisitBinary(expr *ExprBinary) interface{} {
	left, right := eval(expr.Left), eval(expr.Right)

	switch expr.Operator.Type() {
	// number
	case PLUS:
		// number?
		ln, lok := left.(float64)
		rn, rok := right.(float64)
		if lok && rok {
			return ln + rn
		}
		// string?
		ls, lok := left.(string)
		rs, rok := right.(string)
		if lok && rok {
			return ls + rs
		}
	case MINUS:
		l, lok := left.(float64)
		r, rok := right.(float64)
		if lok && rok {
			return l - r
		}
	case STAR:
		l, lok := left.(float64)
		r, rok := right.(float64)
		if lok && rok {
			return l * r
		}
	case SLASH:
		l, lok := left.(float64)
		r, rok := right.(float64)
		if lok && rok {
			return l / r
		}
	case GREATER:
		l, lok := left.(float64)
		r, rok := right.(float64)
		if lok && rok {
			return l > r
		}
	case GREATER_EQUAL:
		l, lok := left.(float64)
		r, rok := right.(float64)
		if lok && rok {
			return l >= r
		}
	case LESS:
		l, lok := left.(float64)
		r, rok := right.(float64)
		if lok && rok {
			return l < r
		}
	case LESS_EQUAL:
		l, lok := left.(float64)
		r, rok := right.(float64)
		if lok && rok {
			return l <= r
		}

	// FIXME: Support == and != operator for other objects
	case EQUAL_EQUAL:
		l, lok := left.(float64)
		r, rok := right.(float64)
		if lok && rok {
			return l == r
		}
	case BANG_EQUAL:
		l, lok := left.(float64)
		r, rok := right.(float64)
		if lok && rok {
			return l != r
		}

	// logic
	case AND:
		l, lok := left.(bool)
		r, rok := right.(bool)
		if lok && rok {
			return l && r
		}
	case OR:
		l, lok := left.(bool)
		r, rok := right.(bool)
		if lok && rok {
			return l || r
		}
	}

	return nil
}
