package main

import (
	"encoding/json"
	"fmt"
)

// Location is a thing
type Location struct {
	File   string    `json:"file"`
	Line   int       `json:"line"`
	Within *Function `json:"within"`
}

func (l *Location) String() string {
	return fmt.Sprintf("%s:%d", l.File, l.Line)
}

// Reference to a word in a file
type Reference interface {
	GetID() int64
	GetLocation() *Location
	ToJSON() ([]byte, error)
}

type Variable struct {
	*Location `'json:"location"`
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsDecl    bool   `json:"is_decl"`
}

func (v *Variable) GetID() int64 {
	return v.ID
}

func (v *Variable) GetLocation() *Location {
	return v.Location
}

func (v *Variable) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

type Function struct {
	*Location `json:"location"`
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Reciever  string   `json:"receiver"`
	Size      int      `json:"size"`
	IsDecl    bool     `json:"is_decl"`
	Calls     []string `json:"fn_calls"`
}

func (f *Function) GetID() int64 {
	return f.ID
}

func (f *Function) GetLocation() *Location {
	return f.Location
}

func (f *Function) ToJSON() ([]byte, error) {
	return json.Marshal(f)
}

func (f *Function) Info() string {
	if f.Reciever != "" {
		return fmt.Sprintf("%s.%s (%s:%d)", f.Reciever, f.Name, f.File, f.Line)
	}
	return fmt.Sprintf("%s (%s:%d)", f.Name, f.File, f.Line)
}

func (f *Function) Wraps(loc *Location) bool {
	if !f.IsDecl {
		// not a function body
		return false
	}
	if f.File != loc.File {
		// not even in the same file bro
		return false
	}
	return loc.Line >= f.Line && loc.Line <= f.Line+f.Size
}

type Struct struct {
	*Location `json:"location"`
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Fields    []string `json:"fields"`
}

func (s *Struct) GetID() int64 {
	return s.ID
}

func (s *Struct) GetLocation() *Location {
	return s.Location
}

func (s *Struct) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

type Result struct {
	Word      string `json:"word"`
	Type      string `json:"type"`
	Reference string `json:"reference"`
	IsDecl    string `json:"is_decl"`
	WithinFn  string `json:"within_fn"`
}

type References []Reference

func (r References) Format() []*Result {
	results := make([]*Result, 0, len(r))
	for _, ref := range r {
		switch d := ref.(type) {
		case *Function:
			res := &Result{
				Word:      d.Name,
				Type:      "function",
				Reference: d.Location.String(),
				IsDecl:    "no",
				WithinFn:  "global",
			}
			if d.IsDecl {
				res.IsDecl = "yes"
			}
			if d.Within != nil {
				res.WithinFn = d.Within.Info()
			}
			results = append(results, res)
		case *Variable:
			res := &Result{
				Word:      d.Name,
				Type:      "variable",
				Reference: d.Location.String(),
				IsDecl:    "no",
				WithinFn:  "global",
			}
			if d.IsDecl {
				res.IsDecl = "yes"
			}
			if d.Within != nil {
				res.WithinFn = d.Within.Info()
			}
			results = append(results, res)
		case *Struct:
			res := &Result{
				Word:      d.Name,
				Type:      "struct",
				Reference: d.Location.String(),
				IsDecl:    "no",
				WithinFn:  "global",
			}
			results = append(results, res)
		default:

		}
	}
	return results
}

// Functions is a map of file names to Function objects
type Functions map[string][]*Function

func (f Functions) BuildCallStack() *CallStackRoot {
	fns, ok := f["main"]
	if !ok || len(fns) == 0 {
		return nil
	}

	entryFn := fns[0]
	cs := &CallStack{}
	f.callStackHelper(entryFn, cs, []string{}, 0)

	return &CallStackRoot{cs}
}

func (f Functions) callStackHelper(fn *Function, cs *CallStack, seen []string, depth int) {
	seen = append(seen, fn.Info())
	cs.Name = fn.Info()
	cs.Depth = depth
	for _, fnName := range fn.Calls {
		fnDefs, ok := f[fnName]
		if !ok || len(fnDefs) == 0 {
			// fn is external
			cs.Children = append(cs.Children,
				&CallStack{Name: fmt.Sprintf("%s (external)", fnName), Depth: depth + 1})
			continue
		}

		childFn := fnDefs[0]
		if alreadySeen(seen, childFn) {
			// avoid loops
			cs.Children = append(cs.Children, &CallStack{Name: childFn.Info(), Depth: depth + 1})
			continue
		}

		childCS := &CallStack{}
		f.callStackHelper(childFn, childCS, seen, depth+1)
		cs.Children = append(cs.Children, childCS)
	}
}

func alreadySeen(seen []string, fn *Function) bool {
	for _, s := range seen {
		if s == fn.Info() {
			return true
		}
	}
	return false
}
