package main

import (
	"encoding/json"
	"fmt"
)

// Location defines where something exists in the project
type Location struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Within string `json:"within"` //function identifier that wraps the reference if any
}

func (l *Location) String() string {
	return fmt.Sprintf("%s:%d", l.File, l.Line)
}

// Reference is an interface type to represent a word in a Go file
type Reference interface {
	GetLocation() *Location
	ToJSON() ([]byte, error)
}

// Variable implements Reference and represents a variable type in the Go code
type Variable struct {
	*Location `'json:"location"`
	Name      string `json:"name"`
	IsDecl    bool   `json:"is_decl"`
}

// GetLocation returns the Location of the Variable
func (v *Variable) GetLocation() *Location {
	return v.Location
}

// ToJSON marshalls the Variable to JSON
func (v *Variable) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

// Function implements Reference and represents a function type in the Go code.
// This can be a function call or function declaration.
type Function struct {
	*Location `json:"location"`
	Name      string   `json:"name"`
	Reciever  string   `json:"receiver"`
	Size      int      `json:"size"`
	IsDecl    bool     `json:"is_decl"`
	Calls     []string `json:"fn_calls"`
}

// GetLocation returns the Location of the Function
func (f *Function) GetLocation() *Location {
	return f.Location
}

// ToJSON marshalls the Function to JSON
func (f *Function) ToJSON() ([]byte, error) {
	return json.Marshal(f)
}

// Info outputs a formatted string about the Function
// eg. "Index.Create (index.go:200)"
func (f *Function) Info() string {
	if f.Reciever != "" {
		return fmt.Sprintf("%s.%s (%s:%d)", f.Reciever, f.Name, f.File, f.Line)
	}
	return fmt.Sprintf("%s (%s:%d)", f.Name, f.File, f.Line)
}

// Wraps returns true if the given location is "wrapped" by the function, eg.
// it exists within the function body. Returns false otherwise.
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

// Struct implements Reference and represents a struct type in the Go code.
type Struct struct {
	*Location `json:"location"`
	Name      string   `json:"name"`
	Fields    []string `json:"fields"` // XXX: not implemented
}

// GetLocation returns the Location of the Struct
func (s *Struct) GetLocation() *Location {
	return s.Location
}

// ToJSON marshalls the Struct to JSON
func (s *Struct) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// References is a list of Reference interfaces
type References []Reference

// SmartSort orders results based on what developers might consider most
// relevant when searching through code. This is elaborated upon in the README.md
type SmartSort []Reference

// This sort implementation is a smarter way to rank results. It assumes the
// following characteristics about code discovery:
// 1. function references are more important than variable or struct references
// 2. function declarations are more important than function calls
// 3. a function is more important if it is invoked more
// 4. structs are more interesting than variables
// 5. between variables, if one reference is a declaration, it is more important
//		otherwise use the names of the variable to break the tie.
func (r SmartSort) Len() int      { return len(r) }
func (r SmartSort) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r SmartSort) Less(i, j int) bool {
	switch r1 := r[i].(type) {
	case *Function:
		if r2, ok := r[j].(*Function); ok {
			// if both results are functions, favor the declaration
			if r1.IsDecl && r2.IsDecl {
				// if both are declarations, favor the one used more
				return len(r1.Calls) > len(r2.Calls)
			}
			// r1 < r2 if r2 is a function declaration
			return r2.IsDecl
		}
		// r2 < r1 since r1 is a function reference
		return false
	case *Variable:
		if _, ok := r[j].(*Function); ok {
			// r1 < r2 because r2 is a function ref
			return true
		}
		if _, ok := r[j].(*Struct); ok {
			// r1 < r2 because r2 is a struct ref
			return true
		}
		// both are variables, break tie with name
		r2, _ := r[j].(*Variable)
		if r1.Name == r2.Name {
			// same name too!
			return r2.IsDecl
		}
		return r1.Name > r2.Name
	case *Struct:
		if _, ok := r[j].(*Function); ok {
			// r1 < r2 because r2 is a function ref
			return true
		}
		if _, ok := r[j].(*Variable); ok {
			// r1 > r2 because r2 is only a variable
			return false
		}
		// both are structs, break tie with name
		r2, _ := r[j].(*Struct)
		return r1.Name > r2.Name
	}
	// we should never hit this
	return true
}

// Result is the JSON response type for a reference in the code
type Result struct {
	Word      string `json:"word"`
	Type      string `json:"type"`
	Reference string `json:"reference"`
	IsDecl    string `json:"is_decl"`
	WithinFn  string `json:"within_fn"`
}

// Format code References to be Result types
func (r References) Format() []*Result {
	results := make([]*Result, 0, len(r))
	for _, ref := range r {
		var res *Result
		switch d := ref.(type) {
		case *Function:
			res = &Result{
				Word:      d.Name,
				Type:      "function",
				Reference: d.Location.String(),
				IsDecl:    "no",
				WithinFn:  "global",
			}
			if d.IsDecl {
				res.IsDecl = "yes"
			}
			if d.Within != "" {
				res.WithinFn = d.Within
			}
		case *Variable:
			res = &Result{
				Word:      d.Name,
				Type:      "variable",
				Reference: d.Location.String(),
				IsDecl:    "no",
				WithinFn:  "global",
			}
			if d.IsDecl {
				res.IsDecl = "yes"
			}
			if d.Within != "" {
				res.WithinFn = d.Within
			}
		case *Struct:
			res = &Result{
				Word:      d.Name,
				Type:      "struct",
				Reference: d.Location.String(),
				IsDecl:    "yes",
				WithinFn:  "global",
			}
		default:
			fmt.Printf("Unknown Reference type %v\n", d)
		}
		if res != nil {
			results = append(results, res)
		}
	}
	return results
}
