// Package agenttest gives the agent command tests a fake herdr server, a fake
// harness message socket, and log helpers.
//
// herdr.go owns the fake herdr server, a test Git repository, a test home,
// and project registration. socket.go owns the fake inbox sockets of Claude
// Code and of the Pi extension. logs.go reads the command log events of a test home and captures the
// events of one logger.
package agenttest
