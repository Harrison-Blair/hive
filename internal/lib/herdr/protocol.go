package herdr

import "context"

// Protocol is the herdr wire protocol that hive supports.
const Protocol = 22

// Ping checks the protocol of the server.
func Ping(ctx context.Context, opts ...Option) (int, error) {
	return 0, nil
}
