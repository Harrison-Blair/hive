// Package goal holds the shared goal rules that several goal operations use.
// goal.go owns the access check, the record paths, and the failed envelope;
// log.go writes goal events; render.go formats shared text. Each goal command
// has its operation in a child package.
package goal
