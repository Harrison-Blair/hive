package herdr

import (
	"context"
	"fmt"
	"sync"
)

// Protocol is the herdr wire protocol that hive supports.
const Protocol = 22

type protocolGate struct {
	done chan struct{}
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
	protocol := 0
	err := call(ctx, "ping", struct{}{}, &result, func() error {
		if result.Type != "pong" {
			return fmt.Errorf("bad reply: result type %q, want pong", result.Type)
		}
		protocol = result.Protocol
		if result.Protocol != Protocol {
			return fmt.Errorf("herdr protocol %d, hive needs %d", result.Protocol, Protocol)
		}
		return nil
	}, opts...)
	return protocol, err
}

func ensureProtocol(ctx context.Context, options callOptions) error {
	path, err := socketPath()
	if err != nil {
		return err
	}
	gateMu.Lock()
	gate, found := gates[path]
	if !found {
		gate = &protocolGate{done: make(chan struct{})}
		gates[path] = gate
		gateMu.Unlock()
		err := handshake(ctx, options)
		gateMu.Lock()
		gate.err = err
		close(gate.done)
		gateMu.Unlock()
		return err
	}
	gateMu.Unlock()
	select {
	case <-gate.done:
		return gate.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func handshake(ctx context.Context, options callOptions) error {
	pingOptions := []Option{WithDeadline(options.deadline)}
	if options.logger != nil {
		pingOptions = append(pingOptions, WithLogger(options.logger))
	}
	_, err := Ping(ctx, pingOptions...)
	return err
}
