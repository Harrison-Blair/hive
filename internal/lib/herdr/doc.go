// Package herdr is the typed client of the herdr socket API.
//
// client.go owns the socket path, the request, the reply, the caps, and the
// request IDs. protocol.go owns protocol 22 and the ping gate. errors.go owns
// the server error type and the codes that hive handles. types.go owns the data
// types that hive reads. workspace.go owns the workspace methods and the typed
// worktree metadata. worktree.go owns worktree.create. agent.go owns the agent
// methods and the typed session. pane.go owns the pane, process, and close
// methods.
package herdr
