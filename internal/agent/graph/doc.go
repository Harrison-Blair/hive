// Package graph shows the agents as a live graph of card trees in pods.
//
// The project pod of an agent named <project>-coordinator holds that agent
// and its descendants. The home pod holds each other agent. The model
// refreshes each second, and on r, through the agent list operation. A
// failed refresh keeps the last snapshot on the screen.
//
// The graph opens in tree mode: each parent card is centered over its
// children, and the cards of one depth share a row. g switches to list mode,
// where each child is one column right of its parent, and back. The board
// scrolls down and sideways to keep the selected card on the screen.
//
//   - pods.go groups the entries into pods of card trees.
//   - card.go draws one card on a canvas: glyphs and spinner frames.
//   - checkout.go reads the Git state of an agent's start directory.
//   - layout.go puts the pods in rows that fit the width and draws the
//     indented lists of list mode.
//   - tree.go draws the tidy trees of tree mode and moves the selection
//     across the screen in tree mode.
//   - model.go is the Bubble Tea model: keys, mode, scroll, refresh, clock,
//     spinner, and the time of each agent in its current state.
//   - run.go runs the model on the terminal with live entries.
package graph
