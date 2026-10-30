// Package agent holds the agent records and the code that the agent commands
// share. Each agent command has its operation in a child package.
//
// record.go keeps one JSON record for each agent at <home>/agents/<name>.json;
// caller.go finds the caller record and holds the shared helpers.
package agent
