// Package agent holds the agent records and the code that the agent commands
// share. Each agent command has its operation in a child package.
//
// record.go keeps one JSON record for each agent at <home>/agents/<name>.json;
// caller.go finds the caller record and holds the shared helpers; slug.go
// turns a directory or a branch name into a part of an agent name; layout.go
// holds the layout lock of spawn, stop, worktree, and coordinator names.
package agent
