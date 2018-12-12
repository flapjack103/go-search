package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// Index stores file information and lookup tables that map words to their types
// and their references within the project.
type Index struct {
	fset    *token.FileSet
	fileMgr *FileManager
	// mapping of word to different types of interest
	references map[string][]Reference
	functions  map[string][]*Function
	structs    map[string][]*Struct
}

// BuildIndex constructs the Index by walking the files and parsing their ASTs
func BuildIndex(fm *FileManager) *Index {
	fset := token.NewFileSet()
	idx := &Index{
		fset:       fset,
		fileMgr:    fm,
		references: make(map[string][]Reference),
		functions:  make(map[string][]*Function),
		structs:    make(map[string][]*Struct),
	}

	for _, arg := range idx.fileMgr.files {
		f, err := parser.ParseFile(fset, arg, nil, parser.AllErrors)
		if err != nil {
			fmt.Printf("could not parse %s: %v\n", arg, err)
			continue
		}
		ast.Walk(idx, f)
	}

	idx.scopeReferences()

	return idx
}

// ReferencesByWord returns all references to the given word. Returns true if the word
// was found and false otherwise.
func (x *Index) ReferencesByWord(word string) ([]Reference, bool) {
	refs, ok := x.references[word]
	return refs, ok
}

// Functions returns a list of all Function declarations
func (x *Index) Functions() Functions {
	return x.functions
}

func (x *Index) addReference(word string, ref Reference) {
	x.references[word] = append(x.references[word], ref)
}

func (x *Index) addFunction(name string, body *ast.BlockStmt, recv string) {
	if x.fset == nil || body == nil {
		return
	}
	posStart := x.fset.Position(body.Lbrace)
	posEnd := x.fset.Position(body.Rbrace)
	relPath := x.fileMgr.Rel(posStart.Filename)
	f := &Function{
		Location: &Location{
			File: relPath,
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
	relPath := x.fileMgr.Rel(pos.Filename)
	f := &Function{
		Location: &Location{
			File: relPath,
			Line: pos.Line,
		},
		Name:     name,
		Reciever: recv,
	}
	x.addReference(name, f)
}

func (x *Index) addVariable(name string, n ast.Node, isDecl bool) {
	pos := x.fset.Position(n.Pos())
	relPath := x.fileMgr.Rel(pos.Filename)
	v := &Variable{
		Location: &Location{
			File: relPath,
			Line: pos.Line,
		},
		Name:   name,
		IsDecl: isDecl,
	}

	x.addReference(name, v)
}

func (x *Index) addStruct(name string, n ast.Node) {
	pos := x.fset.Position(n.Pos())
	relPath := x.fileMgr.Rel(pos.Filename)
	s := &Struct{
		Name: name,
		Location: &Location{
			File: relPath,
			Line: pos.Line,
		},
	}

	x.structs[name] = append(x.structs[name], s)
	x.addReference(name, s)
}

// XXX: this function is terrible but it gets the job done
// determines the scopes for the different items parsed from the files
// ex. determine that variable 'idx' is referenced within fn 'main'
func (x *Index) scopeReferences() {
	for _, refs := range x.references {
		for _, ref := range refs {
			loc := ref.GetLocation()
			for _, fns := range x.functions {
				for _, fn := range fns {
					if !fn.Wraps(loc) {
						// loc not within func block
						continue
					}
					loc.Within = fn.Info()
					if fnCall, ok := ref.(*Function); ok {
						fn.Calls = append(fn.Calls, fnCall.Name)
					}
				}
			}
		}
	}
}

// Visit defines what we do when we visit a node in the AST
// XXX: AST parsing is not complete as their are certain expressions that are
// not handled. This handles more basic cases but could be more comprehensive.
func (x *Index) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch d := n.(type) {
	case *ast.IfStmt:
		x.local(d.Cond)
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
	if recv == nil || len(recv.List) == 0 {
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
