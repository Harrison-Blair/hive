// Package agenttest gives the agent command tests a fake herdr server and log
// helpers.
//
// herdr.go owns the fake herdr server, a test Git repository, a test home,
// and project registration. logs.go reads the command log events of a test
// home and captures the events of one logger.
package agenttest
