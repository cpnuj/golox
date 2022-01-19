package main

type Expr interface {
	Type() ExprType
	Accept(ExprVisitor) (interface{}, error)
}

type ExprType int32
const (
	ExprTypeIdle = iota
	ExprTypeLiteral
	ExprTypeVariable
	ExprTypeAssign
	ExprTypeUnary
	ExprTypeGrouping
	ExprTypeBinary
)

type ExprVisitor interface {
	VisitLiteral(*ExprLiteral) (interface{}, error)
	VisitVariable(*ExprVariable) (interface{}, error)
	VisitAssign(*ExprAssign) (interface{}, error)
	VisitUnary(*ExprUnary) (interface{}, error)
	VisitGrouping(*ExprGrouping) (interface{}, error)
	VisitBinary(*ExprBinary) (interface{}, error)
}

type ExprLiteral struct{
	Value Token
}

func(node *ExprLiteral) Type() ExprType {
	return ExprTypeLiteral
}

func(node *ExprLiteral) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitLiteral(node)
}

type ExprVariable struct{
	Name Token
}

func(node *ExprVariable) Type() ExprType {
	return ExprTypeVariable
}

func(node *ExprVariable) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitVariable(node)
}

type ExprAssign struct{
	Name Token
	Value Expr
}

func(node *ExprAssign) Type() ExprType {
	return ExprTypeAssign
}

func(node *ExprAssign) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitAssign(node)
}

type ExprUnary struct{
	UnaryOperator Token
	Expression Expr
}

func(node *ExprUnary) Type() ExprType {
	return ExprTypeUnary
}

func(node *ExprUnary) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitUnary(node)
}

type ExprGrouping struct{
	Expression Expr
}

func(node *ExprGrouping) Type() ExprType {
	return ExprTypeGrouping
}

func(node *ExprGrouping) Accept(v ExprVisitor) (interface{}, error) {
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

func(node *ExprBinary) Accept(v ExprVisitor) (interface{}, error) {
	return v.VisitBinary(node)
}

type Stmt interface {
	Type() StmtType
	Accept(StmtVisitor) (interface{}, error)
}

type StmtType int32
const (
	StmtTypeIdle = iota
	StmtTypeExpression
	StmtTypePrint
	StmtTypeVar
	StmtTypeBlock
	StmtTypeIf
)

type StmtVisitor interface {
	VisitExpression(*StmtExpression) (interface{}, error)
	VisitPrint(*StmtPrint) (interface{}, error)
	VisitVar(*StmtVar) (interface{}, error)
	VisitBlock(*StmtBlock) (interface{}, error)
	VisitIf(*StmtIf) (interface{}, error)
}

type StmtExpression struct{
	Expression Expr
}

func(node *StmtExpression) Type() StmtType {
	return StmtTypeExpression
}

func(node *StmtExpression) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitExpression(node)
}

type StmtPrint struct{
	Expression Expr
}

func(node *StmtPrint) Type() StmtType {
	return StmtTypePrint
}

func(node *StmtPrint) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitPrint(node)
}

type StmtVar struct{
	Name Token
	Initializer Expr
}

func(node *StmtVar) Type() StmtType {
	return StmtTypeVar
}

func(node *StmtVar) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitVar(node)
}

type StmtBlock struct{
	Statements []Stmt
}

func(node *StmtBlock) Type() StmtType {
	return StmtTypeBlock
}

func(node *StmtBlock) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitBlock(node)
}

type StmtIf struct{
	Cond Expr
	Then Stmt
	Else Stmt
}

func(node *StmtIf) Type() StmtType {
	return StmtTypeIf
}

func(node *StmtIf) Accept(v StmtVisitor) (interface{}, error) {
	return v.VisitIf(node)
}

