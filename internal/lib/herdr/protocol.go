package herdr

import (
	"context"
	"fmt"
	"sync"
)

// Protocol is the herdr wire protocol that hive supports.
const Protocol = 22

type protocolGate struct {
	done bool
	err  error
}

var (
	gateMu sync.Mutex
	gates  = map[string]*protocolGate{}
)

// Ping checks that the server answers and uses protocol 22.
func Ping(ctx context.Context, opts ...Option) (int, error) {
	var result struct {
		Type     string `json:"type"`
		Version  string `json:"version"`
		Protocol int    `json:"protocol"`
	}
	if err := Call(ctx, "ping", struct{}{}, &result, opts...); err != nil {
		return 0, err
	}
	if result.Type != "pong" {
		return 0, fmt.Errorf("bad reply: result type %q, want pong", result.Type)
	}
	if result.Protocol != Protocol {
		return result.Protocol, fmt.Errorf("herdr protocol %d, hive needs %d", result.Protocol, Protocol)
	}
	return result.Protocol, nil
}

func ensureProtocol(ctx context.Context, options callOptions) error {
	path, err := socketPath()
	if err != nil {
		return err
	}
	gateMu.Lock()
	defer gateMu.Unlock()
	gate := gates[path]
	if gate == nil {
		gate = &protocolGate{}
		gates[path] = gate
	}
	if gate.done {
		return gate.err
	}
	pingOptions := []Option{WithDeadline(options.deadline)}
	if options.logger != nil {
		pingOptions = append(pingOptions, WithLogger(options.logger))
	}
	_, err = Ping(ctx, pingOptions...)
	gate.err = err
	gate.done = true
	return err
}
