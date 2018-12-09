package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

// Index keeps file information for fast lookup
type Index struct {
	fset       *token.FileSet
	files      []string
	references map[string][]Reference
	functions  map[string][]*Function
	structs    map[string][]*Struct
}

func BuildIndex(files []string) *Index {
	fset := token.NewFileSet()
	idx := &Index{
		fset:       fset,
		references: make(map[string][]Reference),
		functions:  make(map[string][]*Function),
		structs:    make(map[string][]*Struct),
	}
	idx.files = files

	for _, arg := range files {
		f, err := parser.ParseFile(fset, arg, nil, parser.AllErrors)
		if err != nil {
			log.Printf("could not parse %s: %v", arg, err)
			continue
		}
		ast.Walk(idx, f)
	}
	idx.createFuncDeps()
	return idx
}

func (x *Index) createFuncDeps() {
	for _, refs := range x.references {
		for _, ref := range refs {
			fnCall, ok := ref.(*Function)
			if !ok || fnCall.IsDecl {
				continue
			}
			loc := fnCall.GetLocation()
			for _, fns := range x.functions {
				for _, fn := range fns {
					if !fn.Wraps(loc) {
						// loc not within func block
						continue
					}
					loc.Within = fn
					fn.Calls = append(fn.Calls, fnCall.Name)
				}
			}
		}
	}
}

func (x *Index) addFunction(name string, body *ast.BlockStmt, recv string) {
	posStart := x.fset.Position(body.Lbrace)
	posEnd := x.fset.Position(body.Rbrace)
	f := &Function{
		Location: &Location{
			File: posStart.Filename,
			Line: posStart.Line,
		},
		Name:     name,
		IsDecl:   true,
		Size:     posEnd.Line - posStart.Line + 1,
		Reciever: recv,
	}

	x.functions[name] = append(x.functions[name], f)
	x.addReference(name, f)
}

func (x *Index) addFunctionCall(name string, n ast.Node, recv string) {
	pos := x.fset.Position(n.Pos())
	f := &Function{
		Location: &Location{
			File: pos.Filename,
			Line: pos.Line,
		},
		Name:     name,
		Reciever: recv,
	}
	x.addReference(name, f)
}

func (x *Index) addVariable(name string, n ast.Node, isDecl bool) {
	pos := x.fset.Position(n.Pos())
	v := &Variable{
		Location: &Location{
			File: pos.Filename,
			Line: pos.Line,
		},
		Name:   name,
		IsDecl: isDecl,
	}

	x.addReference(name, v)
}

func (x *Index) addStruct(name string, n ast.Node) {
	pos := x.fset.Position(n.Pos())

	s := &Struct{
		Name: name,
		Location: &Location{
			File: pos.Filename,
			Line: pos.Line,
		},
	}

	x.structs[name] = append(x.structs[name], s)
	x.addReference(name, s)
}

// addReference
func (x *Index) addReference(word string, ref Reference) {
	x.references[word] = append(x.references[word], ref)

}

// References returns all references to the given word. Returns true if the word
// was found and false otherwise.
func (x *Index) References(word string) ([]Reference, bool) {
	refs, ok := x.references[word]
	return refs, ok
}

// Functions returns a list of all Function declarations
func (x *Index) Functions() Functions {
	return x.functions
}

// Visit defines what we do when we visit a node in the AST
func (x *Index) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch d := n.(type) {
	case *ast.IfStmt:
		x.local(d.Cond)
	case *ast.StructType:
		// if len(d.Fields.List) > 0 {
		// 	spew.Dump(d.Fields.List[0])
		// }
	case *ast.AssignStmt:
		if d.Tok != token.DEFINE {
			break
		}
		for _, name := range d.Lhs {
			x.local(name)
		}
	case *ast.CallExpr:
		for _, arg := range d.Args {
			x.local(arg)
		}
		switch fun := d.Fun.(type) {
		case *ast.Ident:
			x.addFunctionCall(fun.Name, n, "")
		case *ast.SelectorExpr:
			var obj string
			if x, ok := fun.X.(*ast.Ident); ok {
				obj = x.Name
			}
			x.addFunctionCall(fun.Sel.Name, n, obj)
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
		recv := parseFuncReceiver(d.Recv)
		x.addFunction(d.Name.Name, d.Body, recv)
	case *ast.GenDecl:
		if d.Tok == token.VAR {
			for _, spec := range d.Specs {
				if value, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range value.Names {
						if name.Name == "_" {
							continue
						}
						x.addVariable(name.Name, n, true)
					}
				}
			}
		} else if d.Tok == token.TYPE {
			for _, spec := range d.Specs {
				if value, ok := spec.(*ast.TypeSpec); ok {
					x.addStruct(value.Name.String(), n)
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
		x.addVariable(ident.Name, n, true)
	} else {
		x.addVariable(ident.Name, n, false)
	}
}

func (x *Index) localList(fs []*ast.Field, t token.Token) {
	for _, f := range fs {
		for _, name := range f.Names {
			x.local(name)
		}
	}
}

func parseFuncReceiver(recv *ast.FieldList) string {
	if recv == nil {
		return ""
	}

	var recvStr string
	switch x := recv.List[0].Type.(type) {
	case *ast.Ident:
		recvStr = x.Name
	case *ast.StarExpr:
		n := x.X.(*ast.Ident)
		recvStr = n.Name
	}

	return recvStr
}
