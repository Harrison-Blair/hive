// Package worktree finds or makes the Git checkout of a branch for the agents
// that share it, and keeps the ownership record of each checkout that hive
// makes. Each worktree command has its operation in a child package.
//
// git.go runs git and reads the repository and its checkouts; checkout.go
// finds or makes the checkout of a branch and reads its ownership record;
// scratch.go makes the scratch place of the project of a directory.
package worktree
