package main

import (
	"go/ast"
	"os"

	"github.com/jedib0t/go-pretty/table"
)

// Reference to a word in a file
type Reference struct {
	file       string
	lineNumber int
	column     int
	node       ast.Node
	isDecl     bool
}

// func (r *Reference) Line() string {
//   return r.node.Pos()
// }

// References is a map of file names to Reference objects
type References map[string][]*Reference

// Print Results list in a table view
func (r References) Print() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "File", "Line", "Column"})
	for file, refs := range r {
		for _, ref := range refs {
			t.AppendRow([]interface{}{t.Length() + 1, file, ref.lineNumber, ref.column})
		}
	}
	t.Render()
}
