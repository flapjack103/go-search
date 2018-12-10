package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// CallStackRoot wraps the root CallStack item for the function tree
type CallStackRoot struct {
	Root *CallStack `json:"downward"`
}

// CallStack represents a node in the function tree. It holds a function name,
// and its children are the functions (other CallStack nodes) invoked by the
// function.
type CallStack struct {
	Name     string `json:"name"`
	Depth    int
	Children []*CallStack `json:"children"`
}

// Write generates the callstack as a json file
func (cs *CallStackRoot) Write(outpath string) error {
	data, err := json.Marshal(cs)
	if err != nil {
		return fmt.Errorf("Error creating callstack json: %s", err)
	}

	if err := ioutil.WriteFile(outpath, data, 0755); err != nil {
		return fmt.Errorf("Error writing json data: %s", err)
	}

	return nil
}

// Functions is a map of function names to Function objects
type Functions map[string][]*Function

// BuildCallStack generates a function tree from the defined Functions.
// It starts at the main function since Go programs begin execution at func main()
// recursively builds the function call tree. Functions not invoked in the program
// will not appear in the CallStack.
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
			// function is defined externally
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
