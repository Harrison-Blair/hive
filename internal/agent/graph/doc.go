// Package graph shows the agents as a live graph of card trees in pods.
//
// The project pod of an agent named <project>-coordinator holds that agent
// and its descendants. The home pod holds each other agent. The model
// refreshes each second, and on r, through the agent list operation. A
// failed refresh keeps the last snapshot on the screen.
//
// The cards and the pods are on a grid of tracks, one card column each, 40
// cells apart. A row of pods holds as many tracks as fit the screen width.
// The graph opens in tree mode: each subtree takes one track for each leaf,
// each parent card is over its first child, and the cards of one depth share
// a row. g switches to list mode, where a pod is one track and each child is
// 3 cells right of its parent and ends in the same column, and back. The
// board scrolls down and sideways to keep the selected card on the screen.
// The footer shows the legend and the keys on one row when both fit, else on
// two rows.
//
//   - pods.go groups the entries into pods of card trees.
//   - card.go draws one card on a canvas: glyphs and spinner frames.
//   - checkout.go reads the Git state of an agent's start directory.
//   - layout.go puts the pods on the tracks of rows that fit the width and
//     draws the stepped lists of list mode.
//   - tree.go draws the tidy trees of tree mode and moves the selection
//     across the screen in tree mode.
//   - model.go is the Bubble Tea model: keys, mode, scroll, refresh, clock,
//     spinner, header, footer, and the time of each agent in its current
//     state.
//   - run.go runs the model on the terminal with live entries.
package graph
