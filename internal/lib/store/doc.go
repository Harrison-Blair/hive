// Package store reads and writes bounded JSON records under the hive home.
// store.go owns record access; lock.go owns the bounded lock of WriteIn; hold.go owns the named locks of Hold; atomic.go owns atomic file replacement; schema.go owns the home schema check.
package store
