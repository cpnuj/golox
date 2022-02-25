package main

type Resolver struct {
	locals map[Expr]int
	scopes []map[string]bool
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

func (r *Resolver) Resolve(statements []Stmt) (map[Expr]int, error) {
	for _, statement := range statements {
		if _, err := r.resolveStmt(statement); err != nil {
			return nil, err
		}
	}
	return r.locals, nil
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

func (r *Resolver) resolveLocal(expr Expr, nameTK Token) error {
	name := nameTK.Value().(string)
	distance := -1
	for i := len(r.scopes) - 1; i >= 0; i-- {
		distance++
		init, ok := r.scopes[i][name]
		if !ok {
			continue
		}
		if !init {
			return logger.NewError(nameTK.row, nameTK.col, "resolver: uninitialized variable")
		}
		r.locals[expr] = distance
		return nil
	}
	// variable not defined by script, find in global
	return nil
}

func (r *Resolver) VisitLiteral(*ExprLiteral) (interface{}, error) {
	return nil, nil
}

func (r *Resolver) VisitVariable(expr *ExprVariable) (interface{}, error) {
	return nil, r.resolveLocal(expr, expr.Name)
}

func (r *Resolver) VisitAssign(expr *ExprAssign) (interface{}, error) {
	if err := r.resolveLocal(expr, expr.Name); err != nil {
		return nil, err
	}
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
	r.resolveExpr(expr.Object)
	return nil, nil
}

func (r *Resolver) VisitSet(expr *ExprSet) (interface{}, error) {
	r.resolveExpr(expr.Object)
	r.resolveExpr(expr.Value)
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

func (r *Resolver) VisitFun(stmt *StmtFun) (interface{}, error) {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	// resolve function
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

func (r *Resolver) VisitReturn(stmt *StmtReturn) (interface{}, error) {
	return r.resolveExpr(stmt.Value)
}

func (r *Resolver) VisitClass(stmt *StmtClass) (interface{}, error) {
	r.declare(stmt.Name)
	r.define(stmt.Name)
	return nil, nil
}
