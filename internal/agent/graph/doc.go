// Package graph shows the agents as a live graph of card trees in pods.
//
// The project pod of an agent named <project>-coordinator holds that agent
// and its descendants. The home pod holds each other agent. The model
// refreshes each second, and on r, through the agent list operation. A
// failed refresh keeps the last snapshot on the screen.
//
//   - pods.go groups the entries into pods of card trees.
//   - card.go draws one card on a canvas: glyphs, palette, and spinner frames.
//   - layout.go draws the pods and the tree joins in rows that fit the width.
//   - model.go is the Bubble Tea model: keys, refresh, clock, and spinner.
//   - run.go runs the model on the terminal with live entries.
package graph
