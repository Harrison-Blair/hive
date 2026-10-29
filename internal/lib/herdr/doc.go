// Package herdr is the typed client of the herdr socket API.
//
// client.go owns the socket path, the request, the reply, the caps, and the
// request IDs. protocol.go owns protocol 22 and the ping gate. errors.go owns
// the server error type and the codes that hive checks. types.go owns the data
// types that hive reads. workspace.go owns workspace.create, workspace.close,
// and the typed worktree metadata. tab.go owns tab.rename; tab_create.go owns
// tab.create. worktree.go owns worktree.list and worktree.create. agent.go owns
// agent.start, agent.prompt, agent.get, and the typed session. pane.go owns
// pane.get and pane.close.
package herdr
