// Package store reads and writes bounded JSON records under the hive home.
// store.go owns record access and the walk; lock.go owns the bounded lock and the locked handle; atomic.go owns atomic file replacement; namedlock.go owns the named lock for one long operation.
package store
