package main

import (
	"fmt"
	"strings"
)

type ExprPrintVisitor struct {
}

var _ ExprVisitor = &ExprPrintVisitor{}

func (p *ExprPrintVisitor) Print(expr Expr) (string, error) {
	result, err := expr.Accept(p)
	if err != nil {
		return "", err
	}
	return stringify(result), nil
}

func parenness(strs ...string) string {
	return "(" + strings.Join(strs, " ") + ")"
}

func stringify(a interface{}) string {
	return fmt.Sprintf("%v", a)
}

func (p *ExprPrintVisitor) VisitLiteral(expr *ExprLiteral) (interface{}, error) {
	return stringify(expr.Value.Value()), nil
}

func (p *ExprPrintVisitor) VisitVariable(expr *ExprVariable) (interface{}, error) {
	return stringify(expr.Name.Value()), nil
}

func (p *ExprPrintVisitor) VisitAssign(expr *ExprAssign) (interface{}, error) {
	value, err := p.Print(expr.Value)
	if err != nil {
		return nil, err
	}
	return parenness("=", stringify(expr.Name.Value()), stringify(value)), nil
}

func (p *ExprPrintVisitor) VisitUnary(expr *ExprUnary) (interface{}, error) {
	operator := expr.UnaryOperator.Value()
	operand, err := expr.Expression.Accept(p)
	if err != nil {
		return nil, err
	}
	return parenness(stringify(operator), stringify(operand)), nil
}

func (p *ExprPrintVisitor) VisitGrouping(expr *ExprGrouping) (interface{}, error) {
	inner, err := expr.Expression.Accept(p)
	if err != nil {
		return nil, err
	}
	return parenness("group", stringify(inner)), nil
}

func (p *ExprPrintVisitor) VisitBinary(expr *ExprBinary) (interface{}, error) {
	operator := expr.Operator.Value()

	left, err := expr.Left.Accept(p)
	if err != nil {
		return nil, err
	}

	right, err := expr.Right.Accept(p)
	if err != nil {
		return nil, err
	}

	return parenness(stringify(operator), stringify(left), stringify(right)), nil
}

func (p *ExprPrintVisitor) VisitLogical(expr *ExprLogical) (interface{}, error) {
	operator := expr.Operator.Value()

	left, err := expr.Left.Accept(p)
	if err != nil {
		return nil, err
	}

	right, err := expr.Right.Accept(p)
	if err != nil {
		return nil, err
	}

	return parenness(stringify(operator), stringify(left), stringify(right)), nil
}
