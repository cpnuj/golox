package main

type Expr interface {
	Type() ExprType
	Accept(ExprVisitor) interface{}
}

type ExprType int32
const (
	ExprTypeIdle = iota
	ExprTypeLiteral
	ExprTypeUnary
	ExprTypeGrouping
	ExprTypeBinary
)

type ExprVisitor interface {
	VisitLiteral(*ExprLiteral) interface{}
	VisitUnary(*ExprUnary) interface{}
	VisitGrouping(*ExprGrouping) interface{}
	VisitBinary(*ExprBinary) interface{}
}

type ExprLiteral struct{
	Value Token
}

func(node *ExprLiteral) Type() ExprType {
	return ExprTypeLiteral
}

func(node *ExprLiteral) Accept(v ExprVisitor) interface{} {
	return v.VisitLiteral(node)
}

type ExprUnary struct{
	UnaryOperator Token
	Expression Expr
}

func(node *ExprUnary) Type() ExprType {
	return ExprTypeUnary
}

func(node *ExprUnary) Accept(v ExprVisitor) interface{} {
	return v.VisitUnary(node)
}

type ExprGrouping struct{
	Expression Expr
}

func(node *ExprGrouping) Type() ExprType {
	return ExprTypeGrouping
}

func(node *ExprGrouping) Accept(v ExprVisitor) interface{} {
	return v.VisitGrouping(node)
}

type ExprBinary struct{
	Left Expr
	Operator Token
	Right Expr
}

func(node *ExprBinary) Type() ExprType {
	return ExprTypeBinary
}

func(node *ExprBinary) Accept(v ExprVisitor) interface{} {
	return v.VisitBinary(node)
}

