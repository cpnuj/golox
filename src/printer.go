package main

import (
	"errors"
	"fmt"
	"strings"
)

// GoTree: code from https://github.com/d6o/GoTree/blob/master/gotree.go

const (
	newLine      = "\n"
	emptySpace   = "    "
	middleItem   = "├── "
	continueItem = "│   "
	lastItem     = "└── "
)

type (
	tree struct {
		text  string
		items []Tree
	}

	// Tree is tree interface
	Tree interface {
		Add(text string) Tree
		AddTree(tree Tree)
		Items() []Tree
		Text() string
		Print() string
	}

	printer struct {
	}

	// Printer is printer interface
	Printer interface {
		Print(Tree) string
	}
)

//NewTree returns a new GoTree.Tree
func NewTree(text string) Tree {
	return &tree{
		text:  text,
		items: []Tree{},
	}
}

//Add adds a node to the tree
func (t *tree) Add(text string) Tree {
	n := NewTree(text)
	t.items = append(t.items, n)
	return n
}

//AddTree adds a tree as an item
func (t *tree) AddTree(tree Tree) {
	t.items = append(t.items, tree)
}

//Text returns the node's value
func (t *tree) Text() string {
	return t.text
}

//Items returns all items in the tree
func (t *tree) Items() []Tree {
	return t.items
}

//Print returns an visual representation of the tree
func (t *tree) Print() string {
	return newPrinter().Print(t)
}

func newPrinter() Printer {
	return &printer{}
}

//Print prints a tree to a string
func (p *printer) Print(t Tree) string {
	return t.Text() + newLine + p.printItems(t.Items(), []bool{})
}

func (p *printer) printText(text string, spaces []bool, last bool) string {
	var result string
	for _, space := range spaces {
		if space {
			result += emptySpace
		} else {
			result += continueItem
		}
	}

	indicator := middleItem
	if last {
		indicator = lastItem
	}

	var out string
	lines := strings.Split(text, "\n")
	for i := range lines {
		text := lines[i]
		if i == 0 {
			out += result + indicator + text + newLine
			continue
		}
		if last {
			indicator = emptySpace
		} else {
			indicator = continueItem
		}
		out += result + indicator + text + newLine
	}

	return out
}

func (p *printer) printItems(t []Tree, spaces []bool) string {
	var result string
	for i, f := range t {
		last := i == len(t)-1
		result += p.printText(f.Text(), spaces, last)
		if len(f.Items()) > 0 {
			spacesChild := append(spaces, last)
			result += p.printItems(f.Items(), spacesChild)
		}
	}
	return result
}

// GoTree End

// Our printer:

type AstPrinter struct {
}

var (
	_ ExprVisitor = &AstPrinter{}
	_ StmtVisitor = &AstPrinter{}
)

func NewAstPrinter() *AstPrinter {
	return &AstPrinter{}
}

func (p *AstPrinter) Print(statements []Stmt) {
	for _, statement := range statements {
		t, err := p.BuildStmt(statement)
		if err != nil {
			logger.EPrintf("%v", err)
			return
		}
		logger.DPrintf(parsedebug, "%s\n", t.Print())
	}
}

func (p *AstPrinter) BuildExpr(expr Expr) (Tree, error) {
	result, err := expr.Accept(p)
	if err != nil {
		return nil, err
	}

	t, ok := result.(Tree)
	if !ok {
		return nil, errors.New("golox error: invalid return type from gotree")
	}

	return t, nil
}

func (p *AstPrinter) BuildStmt(stmt Stmt) (Tree, error) {
	result, err := stmt.Accept(p)
	if err != nil {
		return nil, err
	}

	t, ok := result.(Tree)
	if !ok {
		return nil, errors.New("golox error: invalid return type from gotree")
	}

	return t, nil
}

func stringify(a interface{}) string {
	return fmt.Sprintf("%v", a)
}

func (p *AstPrinter) VisitLiteral(expr *ExprLiteral) (interface{}, error) {
	return NewTree(stringify(expr.Value)), nil
}

func (p *AstPrinter) VisitVariable(expr *ExprVariable) (interface{}, error) {
	return NewTree(stringify(expr.Name.Value())), nil
}

func (p *AstPrinter) VisitAssign(expr *ExprAssign) (interface{}, error) {
	value, err := p.BuildExpr(expr.Value)
	if err != nil {
		return nil, err
	}

	t := NewTree("=")
	t.Add(stringify(expr.Name.Value()))
	t.AddTree(value)

	return t, nil
}

func (p *AstPrinter) VisitUnary(expr *ExprUnary) (interface{}, error) {
	operator := expr.UnaryOperator.Value()
	operand, err := p.BuildExpr(expr.Expression)
	if err != nil {
		return nil, err
	}

	t := NewTree(stringify(operator))
	t.AddTree(operand)

	return t, nil
}

func (p *AstPrinter) VisitGrouping(expr *ExprGrouping) (interface{}, error) {
	inner, err := p.BuildExpr(expr.Expression)
	if err != nil {
		return nil, err
	}

	t := NewTree("group")
	t.AddTree(inner)

	return t, nil
}

func (p *AstPrinter) VisitBinary(expr *ExprBinary) (interface{}, error) {
	operator := expr.Operator.Value()

	left, err := p.BuildExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := p.BuildExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	t := NewTree(stringify(operator))
	t.AddTree(left)
	t.AddTree(right)

	return t, nil
}

func (p *AstPrinter) VisitLogical(expr *ExprLogical) (interface{}, error) {
	operator := expr.Operator.Value()

	left, err := p.BuildExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := p.BuildExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	t := NewTree(stringify(operator))
	t.AddTree(left)
	t.AddTree(right)

	return t, nil
}

func (p *AstPrinter) VisitCall(expr *ExprCall) (interface{}, error) {
	t := NewTree("call")

	callee, err := p.BuildExpr(expr.Callee)
	if err != nil {
		return nil, err
	}

	t.Add("callee").AddTree(callee)

	args := t.Add("args")

	for i := range expr.Args {
		arg, err := p.BuildExpr(expr.Args[i])
		if err != nil {
			return nil, err
		}
		args.AddTree(arg)
	}

	return t, nil
}

func (p *AstPrinter) VisitGet(expr *ExprGet) (interface{}, error) {
	t := NewTree("Get")

	obj, err := p.BuildExpr(expr.Object)
	if err != nil {
		return nil, err
	}
	t.AddTree(obj)

	t.Add(expr.Field.Value().(string))

	return t, nil
}

func (p *AstPrinter) VisitSet(expr *ExprSet) (interface{}, error) {
	t := NewTree("Set")

	obj, err := p.BuildExpr(expr.Object)
	if err != nil {
		return nil, err
	}
	t.AddTree(obj)

	t.Add(expr.Field.Value().(string))

	value, err := p.BuildExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	t.AddTree(value)

	return t, nil
}

func (p *AstPrinter) VisitThis(expr *ExprThis) (interface{}, error) {
	return NewTree("this"), nil
}

func (p *AstPrinter) VisitExpression(stmt *StmtExpression) (interface{}, error) {
	t, err := p.BuildExpr(stmt.Expression)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (p *AstPrinter) VisitPrint(stmt *StmtPrint) (interface{}, error) {
	t := NewTree("print")
	expr, err := p.BuildExpr(stmt.Expression)
	if err != nil {
		return nil, err
	}
	t.AddTree(expr)
	return t, nil
}

func (p *AstPrinter) VisitVar(stmt *StmtVar) (interface{}, error) {
	t := NewTree("var")
	t.Add(stringify(stmt.Name.Value()))

	if stmt.Initializer == nil {
		return t, nil
	}

	initTree := t.Add("initializer")
	init, err := p.BuildExpr(stmt.Initializer)
	if err != nil {
		return nil, err
	}
	initTree.AddTree(init)

	return t, nil
}

func (p *AstPrinter) VisitBlock(stmt *StmtBlock) (interface{}, error) {
	t := NewTree("block")
	for i := range stmt.Statements {
		statement, err := p.BuildStmt(stmt.Statements[i])
		if err != nil {
			return nil, err
		}
		t.AddTree(statement)
	}
	return t, nil
}

func (p *AstPrinter) VisitIf(stmt *StmtIf) (interface{}, error) {
	t := NewTree("if")
	cond, err := p.BuildExpr(stmt.Cond)
	if err != nil {
		return nil, err
	}
	t.AddTree(cond)

	thenTree := t.Add("then")
	thenBranch, err := p.BuildStmt(stmt.Then)
	if err != nil {
		return nil, err
	}
	thenTree.AddTree(thenBranch)

	if stmt.Else == nil {
		return t, nil
	}

	elseTree := t.Add("else")
	elseBranch, err := p.BuildStmt(stmt.Else)
	if err != nil {
		return nil, err
	}
	elseTree.AddTree(elseBranch)

	return t, nil
}

func (p *AstPrinter) VisitWhile(stmt *StmtWhile) (interface{}, error) {
	t := NewTree("while")

	condTree := t.Add("cond")
	cond, err := p.BuildExpr(stmt.Cond)
	if err != nil {
		return nil, err
	}
	condTree.AddTree(cond)

	bodyTree := t.Add("body")
	body, err := p.BuildStmt(stmt.Body)
	if err != nil {
		return nil, err
	}
	bodyTree.AddTree(body)

	return t, nil
}

func (p *AstPrinter) VisitFun(stmt *StmtFun) (interface{}, error) {
	t := NewTree("fun")
	t.Add(stmt.Name)

	params := t.Add("params")
	for i := range stmt.Params {
		params.Add(stmt.Params[i])
	}

	body := t.Add("body")
	for i := range stmt.Body {
		statement, err := p.BuildStmt(stmt.Body[i])
		if err != nil {
			return nil, err
		}
		body.AddTree(statement)
	}

	return t, nil
}

func (p *AstPrinter) VisitReturn(stmt *StmtReturn) (interface{}, error) {
	t := NewTree("return")
	value, err := p.BuildExpr(stmt.Value)
	if err != nil {
		return nil, err
	}
	t.AddTree(value)
	return t, nil
}

func (p *AstPrinter) VisitClass(stmt *StmtClass) (interface{}, error) {
	t := NewTree("class")
	t.Add(stmt.Name)

	methods := t.Add("methods")
	for _, method := range stmt.Methods {
		fun, err := p.BuildStmt(method)
		if err != nil {
			return nil, err
		}
		methods.AddTree(fun)
	}

	return t, nil
}
