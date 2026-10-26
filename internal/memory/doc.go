// Package memory holds the shared memory rules: the memory files, the index,
// and the atomic file write that handoff uses.
//
// index.go reads memory files, builds the index, and writes one file atomically.
// frontmatter.go parses the memory frontmatter block.
// errors.go returns the user text and the safe log text of an error.
// The memory write operation is in the child package write.
package memory
