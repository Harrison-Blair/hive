// Package store reads and writes bounded JSON records under the hive home.
// store.go owns record access and locking; atomic.go owns atomic file replacement.
package store
