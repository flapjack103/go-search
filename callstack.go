package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type CallStackRoot struct {
	Root *CallStack `json:"downward"`
}

type CallStack struct {
	Name     string `json:"name"`
	Depth    int
	Children []*CallStack `json:"children"`
}

func (cs *CallStackRoot) Write(filepath string) error {
	data, err := json.Marshal(cs)
	if err != nil {
		return fmt.Errorf("Error creating callstack json: %s\n", err)
	}

	if err := ioutil.WriteFile("static/tree.json", data, 0755); err != nil {
		return fmt.Errorf("Error writing json data: %s\n", err)
	}

	return nil
}
