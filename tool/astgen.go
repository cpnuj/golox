package main

import (
	"fmt"
	"io"
	"os"
)

var writer io.Writer

func writef(format string, a ...interface{}) {
	fmt.Fprintf(writer, format, a...)
}

type Field struct {
	tname string // type name
	fname string // field name
}

type Type struct {
	typename string
	fields   []Field
}

const head = `package main

`

func defineAST(baseName string, types []Type) {
	// write common interface
	writef("type %s interface {\n", baseName)
	writef("\tType() %sType\n", baseName)
	writef("\tAccept(%sVisitor) (interface{}, error)\n", baseName)
	writef("}\n\n")

	// write type enum and add acceptors
	var acceptors []string

	writef("type %sType int32\n", baseName)
	writef("const (\n")
	writef("\t%sTypeIdle = iota\n", baseName)
	for i := range types {
		writef("\t%sType%s\n", baseName, types[i].typename)
		acceptors = append(acceptors, types[i].typename)
	}
	writef(")\n\n")

	// write visitor interface
	writef("type %sVisitor interface {\n", baseName)
	for _, acceptor := range acceptors {
		writef("\tVisit%s(*%s%s) (interface{}, error)\n", acceptor, baseName, acceptor)
	}
	writef("}\n\n")

	// write type definition
	for i := range types {
		typeName := baseName + types[i].typename
		writef("type %s struct{\n", typeName)
		for _, field := range types[i].fields {
			writef("\t%s %s\n", field.fname, field.tname)
		}
		writef("}\n\n")

		writef("func(node *%s) Type() %sType {\n", typeName, baseName)
		writef("\treturn %sType%s\n", baseName, types[i].typename)
		writef("}\n\n")

		writef("func(node *%s) Accept(v %sVisitor) (interface{}, error) {\n", typeName, baseName)
		writef("\treturn v.Visit%s(node)\n", types[i].typename)
		writef("}\n\n")
	}
}

func main() {
	writer = os.Stdout
	writef(head)

	var types []Type

	types = append(types, Type{
		typename: "Literal",
		fields: []Field{
			{"Token", "Value"},
		},
	})

	types = append(types, Type{
		typename: "Variable",
		fields: []Field{
			{"Token", "Name"},
		},
	})

	types = append(types, Type{
		typename: "Assign",
		fields: []Field{
			{"Token", "Name"},
			{"Expr", "Value"},
		},
	})

	types = append(types, Type{
		typename: "Unary",
		fields: []Field{
			{"Token", "UnaryOperator"},
			{"Expr", "Expression"},
		},
	})

	types = append(types, Type{
		typename: "Grouping",
		fields: []Field{
			{"Expr", "Expression"},
		},
	})

	types = append(types, Type{
		typename: "Binary",
		fields: []Field{
			{"Expr", "Left"},
			{"Token", "Operator"},
			{"Expr", "Right"},
		},
	})

	types = append(types, Type{
		typename: "Logical",
		fields: []Field{
			{"Expr", "Left"},
			{"Token", "Operator"},
			{"Expr", "Right"},
		},
	})

	defineAST("Expr", types)

	types = []Type{}

	types = append(types, Type{
		typename: "Expression",
		fields: []Field{
			{"Expr", "Expression"},
		},
	})

	types = append(types, Type{
		typename: "Print",
		fields: []Field{
			{"Expr", "Expression"},
		},
	})

	types = append(types, Type{
		typename: "Var",
		fields: []Field{
			{"Token", "Name"},
			{"Expr", "Initializer"},
		},
	})

	types = append(types, Type{
		typename: "Block",
		fields: []Field{
			{"[]Stmt", "Statements"},
		},
	})

	types = append(types, Type{
		typename: "If",
		fields: []Field{
			{"Expr", "Cond"},
			{"Stmt", "Then"},
			{"Stmt", "Else"},
		},
	})

	defineAST("Stmt", types)
}
