package harness

var profiles = map[string]Profile{
	"claude":   claudeProfile{},
	"codex":    codexProfile{},
	"cursor":   cursorProfile{},
	"opencode": opencodeProfile{},
	"pi":       piProfile{},
}
