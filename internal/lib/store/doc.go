// Package store reads and writes bounded JSON records under the hive home.
// store.go owns record access; lock.go owns the bounded lock and the locked handle; atomic.go owns atomic file replacement; migrate.go owns the schema migration.
package store
