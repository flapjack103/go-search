package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

// Index keeps file information for fast lookup
type Index struct {
	fset     *token.FileSet
	filename string
	pkgDecl  map[*ast.GenDecl]bool
	wordMap  map[string]References
}

func newIndex(fset *token.FileSet) *Index {
	return &Index{
		fset:    fset,
		wordMap: make(map[string]References),
	}
}

func buildIndex(files []string) *Index {
	fs := token.NewFileSet()
	idx := newIndex(fs)

	for _, arg := range files {
		f, err := parser.ParseFile(fs, arg, nil, parser.AllErrors)
		if err != nil {
			log.Printf("could not parse %s: %v", arg, err)
			continue
		}
		ast.Walk(idx, f)
	}
	return idx
}

// addReference
func (x *Index) addReference(word string, n ast.Node, decl bool) {
	pos := x.fset.Position(n.Pos())
	r := &Reference{
		file:       pos.Filename,
		lineNumber: pos.Line,
		column:     pos.Column,
		node:       n,
		isDecl:     decl,
	}

	if _, ok := x.wordMap[word]; !ok {
		x.wordMap[word] = make(References)
	}
	x.wordMap[word][r.file] = append(x.wordMap[word][r.file], r)
}

// References returns all references to the given word. Returns true if the word
// was found and false otherwise.
func (x *Index) References(word string) (References, bool) {
	refs, ok := x.wordMap[word]
	return refs, ok
}

// Visit defines what we do when we visit a node in the AST
func (x *Index) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch d := n.(type) {
	case *ast.IfStmt:
		x.local(d.Cond)
	case *ast.ForStmt:
		p := d.Cond.Pos()
		pos := x.fset.Position(p)
		fmt.Println(pos.String())
	case *ast.AssignStmt:
		if d.Tok != token.DEFINE {
			return x
		}
		for _, name := range d.Lhs {
			x.local(name)
		}
	case *ast.RangeStmt:
		x.local(d.Key)
		x.local(d.Value)
	case *ast.FuncDecl:
		if d.Recv != nil {
			x.localList(d.Recv.List, token.FUNC)
		}
		x.localList(d.Type.Params.List, token.FUNC)
		if d.Type.Results != nil {
			x.localList(d.Type.Results.List, token.FUNC)
		}
	case *ast.GenDecl:
		if d.Tok != token.VAR {
			return x
		}
		for _, spec := range d.Specs {
			if value, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range value.Names {
					if name.Name == "_" {
						continue
					}
					x.addReference(name.Name, n, true)
				}
			}
		}
	}

	return x
}

func (x *Index) local(n ast.Node) {
	ident, ok := n.(*ast.Ident)
	if !ok {
		return
	}
	if ident.Name == "_" || ident.Name == "" {
		return
	}
	if ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
		x.addReference(ident.Name, n, true)
	} else {
		x.addReference(ident.Name, n, false)
	}
}

func (x *Index) localList(fs []*ast.Field, t token.Token) {
	for _, f := range fs {
		for _, name := range f.Names {
			x.local(name)
		}
	}
}
