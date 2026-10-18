// Package herdr is the typed client of the herdr socket API.
//
// client.go owns the socket path, the request, the reply, the caps, and the
// request IDs. protocol.go owns protocol 22 and the ping gate. errors.go owns
// the server error type and the codes that hive handles. types.go owns the data
// types that hive reads. workspace.go owns the workspace methods, including
// workspace.create and workspace.close, and the typed worktree metadata.
// tab.go owns tab.list; tab_create.go owns tab.create. worktree.go owns
// worktree.create and worktree.open. agent.go owns the
// agent methods and the typed session. pane.go owns the pane, pane.list, the
// process, and the close methods.
package herdr
