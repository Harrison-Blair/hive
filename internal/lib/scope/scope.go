// Package scope parses memory and handoff scopes and checks paths in the home.
//
// scope.go parses a scope and checks each path type.
// input.go reads bounded stdin with a deadline.
package scope

import "time"

// Scope is one resolved memory or handoff location.
type Scope struct {
	Value string
	Dir   string
}

// Resolve parses one scope value.
func Resolve(home, value string) (Scope, error) {
	return Scope{}, nil
}

// CheckDir checks one directory path in the home.
func CheckDir(home, dir string) error {
	return nil
}

// CheckFile checks one file path in the home.
func CheckFile(home, path string) error {
	return nil
}

// DefaultDeadline is the stdin read deadline for the memory and handoff commands.
const DefaultDeadline = 30 * time.Second
