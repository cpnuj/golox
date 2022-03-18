package main

import "fmt"

type Resolver struct {
	locals map[Expr]int
	scopes []map[string]bool

	errs    error
	inclass int
}

var (
	_ ExprVisitor = &Resolver{}
	_ StmtVisitor = &Resolver{}
)

func NewResolver() *Resolver {
	return &Resolver{
		locals: make(map[Expr]int),
		scopes: []map[string]bool{make(map[string]bool)},
	}
}

func (r *Resolver) addError(err error) {
	if r.errs == nil {
		r.errs = err
	} else {
		r.errs = fmt.Errorf("%s\n%s", r.errs, err)
	}
}

func (r *Resolver) Resolve(statements []Stmt) (map[Expr]int, error) {
	defer func() {
		e := recover()
		if e != nil {
			r.addError(e.(*LoxError))
		}
	}()

	for _, statement := range statements {
		if _, err := r.resolveStmt(statement); err != nil {
			r.addError(err)
		}
	}

	return r.locals, r.errs
}

func (r *Resolver) resolveExpr(expr Expr) (interface{}, error) {
	return expr.Accept(r)
}

func (r *Resolver) resolveStmt(stmt Stmt) (interface{}, error) {
	return stmt.Accept(r)
}

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *Resolver) declare(name string) {
	r.scopes[len(r.scopes)-1][name] = false
}

func (r *Resolver) define(name string) {
	if _, ok := r.scopes[len(r.scopes)-1][name]; !ok {
		panic("programming error")
	}
	r.scopes[len(r.scopes)-1][name] = true
}

func (r *Resolver) resolveLocal(expr Expr, nameTK Token, mustResolve bool) bool {
	name := nameTK.Value().(string)
	distance := -1
	for i := len(r.scopes) - 1; i >= 0; i-- {
		distance++
		init, ok := r.scopes[i][name]
		if !ok {
			continue
		}
		if !init {
			panic(NewLoxError(ResolveError, nameTK, "resolver: uninitialized variable"))
		}
		r.locals[expr] = distance
		return true
	}
	return !mustResolve
}

func (r *Resolver) VisitLiteral(*ExprLiteral) (interface{}, error) {
	return nil, nil
}

func (r *Resolver) VisitVariable(expr *ExprVariable) (interface{}, error) {
	r.resolveLocal(expr, expr.Name, false /* mustResolve */)
	return nil, nil
}

func (r *Resolver) VisitAssign(expr *ExprAssign) (interface{}, error) {
	r.resolveLocal(expr, expr.Name, false)
	return r.resolveExpr(expr.Value)
}

func (r *Resolver) VisitUnary(expr *ExprUnary) (interface{}, error) {
	return r.resolveExpr(expr.Expression)
}

func (r *Resolver) VisitGrouping(expr *ExprGrouping) (interface{}, error) {
	return r.resolveExpr(expr.Expression)
}

func (r *Resolver) VisitBinary(expr *ExprBinary) (interface{}, error) {
	if _, err := r.resolveExpr(expr.Left); err != nil {
		return nil, err
	}
	if _, err := r.resolveExpr(expr.Right); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *Resolver) VisitLogical(expr *ExprLogical) (interface{}, error) {
	if _, err := r.resolveExpr(expr.Left); err != nil {
		return nil, err
	}
	if _, err := r.resolveExpr(expr.Right); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *Resolver) VisitCall(expr *ExprCall) (interface{}, error) {
	if _, err := r.resolveExpr(expr.Callee); err != nil {
		return nil, err
	}
	for _, arg := range expr.Args {
		if _, err := r.resolveExpr(arg); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (r *Resolver) VisitGet(expr *ExprGet) (interface{}, error) {
	if _, err := r.resolveExpr(expr.Object); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *Resolver) VisitSet(expr *ExprSet) (interface{}, error) {
	if _, err := r.resolveExpr(expr.Object); err != nil {
		return nil, err
	}
	if _, err := r.resolveExpr(expr.Value); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *Resolver) VisitThis(expr *ExprThis) (interface{}, error) {
	if !r.resolveLocal(expr, expr.Keyword, true) {
		panic(NewLoxError(ResolveError, expr.Keyword, "Can't use 'this' out of class"))
	}
	return nil, nil
}

func (r *Resolver) VisitSuper(expr *ExprSuper) (interface{}, error) {
	if r.inclass <= 0 {
		r.addError(
			NewLoxError(ResolveError, expr.Keyword, "Can't use 'super' outside of a class."),
		)
		return nil, nil
	}
	if !r.resolveLocal(expr, expr.Keyword, true) {
		r.addError(
			NewLoxError(ResolveError, expr.Keyword, "Can't use 'super' in a class with no superclass."),
		)
	}
	return nil, nil
}

func (r *Resolver) VisitExpression(stmt *StmtExpression) (interface{}, error) {
	return r.resolveExpr(stmt.Expression)
}

func (r *Resolver) VisitPrint(stmt *StmtPrint) (interface{}, error) {
	return r.resolveExpr(stmt.Expression)
}

func (r *Resolver) VisitVar(stmt *StmtVar) (interface{}, error) {
	name := stmt.Name.Value().(string)
	r.declare(name)

	if stmt.Initializer != nil {
		if _, err := r.resolveExpr(stmt.Initializer); err != nil {
			return nil, err
		}
	}

	r.define(name)
	return nil, nil
}

func (r *Resolver) VisitBlock(stmt *StmtBlock) (interface{}, error) {
	r.beginScope()
	for i := range stmt.Statements {
		if _, err := r.resolveStmt(stmt.Statements[i]); err != nil {
			return nil, err
		}
	}
	r.endScope()
	return nil, nil
}

func (r *Resolver) VisitIf(stmt *StmtIf) (interface{}, error) {
	if _, err := r.resolveExpr(stmt.Cond); err != nil {
		return nil, err
	}
	if _, err := r.resolveStmt(stmt.Then); err != nil {
		return nil, err
	}
	if stmt.Else != nil {
		return r.resolveStmt(stmt.Else)
	}
	return nil, nil
}

func (r *Resolver) VisitWhile(stmt *StmtWhile) (interface{}, error) {
	if _, err := r.resolveExpr(stmt.Cond); err != nil {
		return nil, err
	}
	return r.resolveStmt(stmt.Body)
}

func (r *Resolver) resolveFunction(stmt *StmtFun) (interface{}, error) {
	r.beginScope()
	for _, param := range stmt.Params {
		r.declare(param)
		r.define(param)
	}
	for _, statement := range stmt.Body {
		if _, err := r.resolveStmt(statement); err != nil {
			return nil, err
		}
	}
	r.endScope()
	return nil, nil
}

func (r *Resolver) VisitFun(stmt *StmtFun) (interface{}, error) {
	r.declare(stmt.Name)
	r.define(stmt.Name)
	return r.resolveFunction(stmt)
}

func (r *Resolver) VisitReturn(stmt *StmtReturn) (interface{}, error) {
	if stmt.Value != nil {
		return r.resolveExpr(stmt.Value)
	}
	return nil, nil
}

func (r *Resolver) enterClass() {
	r.inclass++
}

func (r *Resolver) endClass() {
	r.inclass--
}

func (r *Resolver) VisitClass(stmt *StmtClass) (interface{}, error) {
	r.enterClass()

	if stmt.Superclass != nil {
		if stmt.Superclass.Name.lexeme == stmt.Name {
			r.addError(NewLoxError(
				ResolveError, stmt.Superclass.Name, "A class can't inherit from itself.",
			))
		}
		if _, err := r.resolveExpr(stmt.Superclass); err != nil {
			return nil, err
		}
	}

	r.declare(stmt.Name)
	r.define(stmt.Name)

	r.beginScope()
	r.declare("this")
	r.define("this")

	if stmt.Superclass != nil {
		r.declare("super")
		r.define("super")
	}

	r.beginScope()
	for _, method := range stmt.Methods {
		if _, err := r.VisitFun(method); err != nil {
			return nil, err
		}
	}
	r.endScope()

	r.endScope()

	r.endClass()
	return nil, nil
}
