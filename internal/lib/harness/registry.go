package harness

var profiles = map[string]Profile{
	"claude":   claudeProfile{},
	"codex":    codexProfile{},
	"opencode": opencodeProfile{},
	"pi":       piProfile{},
}
