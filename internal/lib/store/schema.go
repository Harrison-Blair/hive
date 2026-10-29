package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
)

// schemaVersion is the only home schema version that this build reads.
const schemaVersion = 4

const schemaName = "schema.json"

type schemaFile struct {
	Version int `json:"version"`
}

// checkSchema writes the schema file of a fresh home and refuses a home at
// any version other than schemaVersion.
func (s *Store) checkSchema() error {
	path := filepath.Join(s.stateDir(), schemaName)
	data, err := readRecordFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s.writeSchema()
	}
	if err != nil {
		return err
	}
	var schema schemaFile
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("schema invalid: %s", path)
	}
	if schema.Version != schemaVersion {
		return fmt.Errorf("home schema %d is not supported; this hive needs %d", schema.Version, schemaVersion)
	}
	return nil
}

func (s *Store) writeSchema() error {
	data, err := marshalRecord(schemaFile{Version: schemaVersion})
	if err != nil {
		return err
	}
	// The write is a store lifecycle event, not the action of the command.
	// It carries no request ID, so a request filter shows only the command
	// events.
	restoreRequestID := hivelog.SetRequestID("")
	defer restoreRequestID()
	writer, err := hivelog.NewWriter(s.home, "store", time.Now)
	if err != nil {
		return fmt.Errorf("state schema log: %w", err)
	}
	defer func() { _ = writer.Close() }()
	logger := slog.New(writer)
	if err := writeRecordFile(filepath.Join(s.stateDir(), schemaName), data); err != nil {
		logger.Error("state schema write failed", "version", schemaVersion, "error", err.Error())
		return err
	}
	logger.Debug("state schema written", "version", schemaVersion)
	return nil
}
