package register

import (
	"testing"

	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// TestRegisterRefusesReservedNames covers the name refusal of D1.
func TestRegisterRefusesReservedNames(t *testing.T) {
	home := t.TempDir()
	for _, name := range []string{"coordinator", "supervisor"} {
		envelope := Run(home, agentlib.RegisterOptions{Profile: records.ProfileGeneral, Name: name, NameSet: true, PaneID: "w1:p1"})
		if envelope.Outcome != "failed" || envelope.Error != "name is reserved: "+name {
			t.Fatalf("Run(name=%s) = %#v; want the reserved-name refusal", name, envelope)
		}
	}
}
