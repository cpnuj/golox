package main

import "fmt"

type ExprPrintVisitor struct {
}

func parenness(s string) string {
	return "(" + s + ")"
}

func (p *ExprPrintVisitor) VisitLiteral(expr *ExprLiteral) interface{} {
	return fmt.Sprintf("%v", expr.Value.lexval)
}

func (p *ExprPrintVisitor) VisitUnary(expr *ExprUnary) interface{} {
	return parenness(
		fmt.Sprintf("%v %v", expr.UnaryOperator.Value(), expr.Expression.Accept(p)),
	)
}

func (p *ExprPrintVisitor) VisitGrouping(expr *ExprGrouping) interface{} {
	return parenness(
		fmt.Sprintf("%v %v", "group", expr.Expression.Accept(p)),
	)
}

func (p *ExprPrintVisitor) VisitBinary(expr *ExprBinary) interface{} {
	return parenness(
		fmt.Sprintf("%v %v %v", expr.Operator.Value(), expr.Left.Accept(p), expr.Right.Accept(p)),
	)
}
