package main

import (
	"testing"
)

func TestPrintVisitor(t *testing.T) {
	expect := "(+ (* 1 2) (/ 3 4))"

	mul := ExprBinary{
		Left: &ExprLiteral{
			Value: Token{lexval: 1},
		},
		Right: &ExprLiteral{
			Value: Token{lexval: 2},
		},
		Operator: Token{
			lexeme: "*",
		},
	}

	div := ExprBinary{
		Left: &ExprLiteral{
			Value: Token{lexval: 3},
		},
		Right: &ExprLiteral{
			Value: Token{lexval: 4},
		},
		Operator: Token{
			lexeme: "/",
		},
	}

	top := ExprBinary{
		Left:  &mul,
		Right: &div,
		Operator: Token{
			lexeme: "+",
		},
	}

	p := &ExprPrintVisitor{}

	get, err := p.Print(&top)
	if err != nil {
		t.Error(err)
	}

	if get != expect {
		t.Fatalf("expect: %s\nget: %s\n", expect, get)
	}
}