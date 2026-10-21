package tui

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	for _, key := range []string{"HERDR_ENV", "HERDR_PANE_ID", "HERDR_SOCKET_PATH", "HIVE_NO_STARTUP"} {
		if err := os.Unsetenv(key); err != nil {
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}
