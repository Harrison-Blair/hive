// Package land delivers a verified task branch to its project.
// land.go owns the preconditions, the named lock, the record write, and the
// delivery result; merge.go owns the merge mode; pr.go owns the pull request
// mode; render.go formats the result.
package land
