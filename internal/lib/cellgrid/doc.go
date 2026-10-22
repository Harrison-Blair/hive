// Package cellgrid models a grid of terminal cells for the TUI v2 golden and
// live checks: a rune, a palette token, and a bold flag for each cell.
//
// palette.go holds the one table of the palette tokens.
// grid.go holds the grid and reads and writes the grid file format.
// convert.go turns ANSI text into a grid.
// compare.go compares two grids with masks and reads the mask file.
package cellgrid
