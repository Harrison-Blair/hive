package component

// Weight is the weight of one arm of a box glyph: none, light, or heavy.
type Weight uint8

// The weights of an arm.
const (
	None Weight = iota
	Light
	Heavy
)

// Arms are the four arms of a box glyph. A lit line makes its arms heavy
// (decisions 66 to 68).
type Arms struct {
	Up, Down, Left, Right Weight
}

// boxArms gives the arms of each box glyph with two or more arms. The light
// corners are the rounded ones.
var boxArms = map[rune]Arms{
	'─': {Left: Light, Right: Light},
	'━': {Left: Heavy, Right: Heavy},
	'│': {Up: Light, Down: Light},
	'┃': {Up: Heavy, Down: Heavy},
	'┍': {Down: Light, Right: Heavy},
	'┎': {Down: Heavy, Right: Light},
	'┏': {Down: Heavy, Right: Heavy},
	'┑': {Down: Light, Left: Heavy},
	'┒': {Down: Heavy, Left: Light},
	'┓': {Down: Heavy, Left: Heavy},
	'┕': {Up: Light, Right: Heavy},
	'┖': {Up: Heavy, Right: Light},
	'┗': {Up: Heavy, Right: Heavy},
	'┙': {Up: Light, Left: Heavy},
	'┚': {Up: Heavy, Left: Light},
	'┛': {Up: Heavy, Left: Heavy},
	'├': {Up: Light, Down: Light, Right: Light},
	'┝': {Up: Light, Down: Light, Right: Heavy},
	'┞': {Up: Heavy, Down: Light, Right: Light},
	'┟': {Up: Light, Down: Heavy, Right: Light},
	'┠': {Up: Heavy, Down: Heavy, Right: Light},
	'┡': {Up: Heavy, Down: Light, Right: Heavy},
	'┢': {Up: Light, Down: Heavy, Right: Heavy},
	'┣': {Up: Heavy, Down: Heavy, Right: Heavy},
	'┤': {Up: Light, Down: Light, Left: Light},
	'┥': {Up: Light, Down: Light, Left: Heavy},
	'┦': {Up: Heavy, Down: Light, Left: Light},
	'┧': {Up: Light, Down: Heavy, Left: Light},
	'┨': {Up: Heavy, Down: Heavy, Left: Light},
	'┩': {Up: Heavy, Down: Light, Left: Heavy},
	'┪': {Up: Light, Down: Heavy, Left: Heavy},
	'┫': {Up: Heavy, Down: Heavy, Left: Heavy},
	'┬': {Down: Light, Left: Light, Right: Light},
	'┭': {Down: Light, Left: Heavy, Right: Light},
	'┮': {Down: Light, Left: Light, Right: Heavy},
	'┯': {Down: Light, Left: Heavy, Right: Heavy},
	'┰': {Down: Heavy, Left: Light, Right: Light},
	'┱': {Down: Heavy, Left: Heavy, Right: Light},
	'┲': {Down: Heavy, Left: Light, Right: Heavy},
	'┳': {Down: Heavy, Left: Heavy, Right: Heavy},
	'┴': {Up: Light, Left: Light, Right: Light},
	'┵': {Up: Light, Left: Heavy, Right: Light},
	'┶': {Up: Light, Left: Light, Right: Heavy},
	'┷': {Up: Light, Left: Heavy, Right: Heavy},
	'┸': {Up: Heavy, Left: Light, Right: Light},
	'┹': {Up: Heavy, Left: Heavy, Right: Light},
	'┺': {Up: Heavy, Left: Light, Right: Heavy},
	'┻': {Up: Heavy, Left: Heavy, Right: Heavy},
	'┼': {Up: Light, Down: Light, Left: Light, Right: Light},
	'┽': {Up: Light, Down: Light, Left: Heavy, Right: Light},
	'┾': {Up: Light, Down: Light, Left: Light, Right: Heavy},
	'┿': {Up: Light, Down: Light, Left: Heavy, Right: Heavy},
	'╀': {Up: Heavy, Down: Light, Left: Light, Right: Light},
	'╁': {Up: Light, Down: Heavy, Left: Light, Right: Light},
	'╂': {Up: Heavy, Down: Heavy, Left: Light, Right: Light},
	'╃': {Up: Heavy, Down: Light, Left: Heavy, Right: Light},
	'╄': {Up: Heavy, Down: Light, Left: Light, Right: Heavy},
	'╅': {Up: Light, Down: Heavy, Left: Heavy, Right: Light},
	'╆': {Up: Light, Down: Heavy, Left: Light, Right: Heavy},
	'╇': {Up: Heavy, Down: Light, Left: Heavy, Right: Heavy},
	'╈': {Up: Light, Down: Heavy, Left: Heavy, Right: Heavy},
	'╉': {Up: Heavy, Down: Heavy, Left: Heavy, Right: Light},
	'╊': {Up: Heavy, Down: Heavy, Left: Light, Right: Heavy},
	'╋': {Up: Heavy, Down: Heavy, Left: Heavy, Right: Heavy},
	'╭': {Down: Light, Right: Light},
	'╮': {Down: Light, Left: Light},
	'╯': {Up: Light, Left: Light},
	'╰': {Up: Light, Right: Light},
	'╼': {Left: Light, Right: Heavy},
	'╽': {Up: Light, Down: Heavy},
	'╾': {Left: Heavy, Right: Light},
	'╿': {Up: Heavy, Down: Light},
}

// boxRunes is the inverse of boxArms. The square light corners are not in
// it: a light corner is rounded.
var boxRunes = func() map[Arms]rune {
	out := make(map[Arms]rune, len(boxArms))
	for r, arms := range boxArms {
		out[arms] = r
	}
	return out
}()

// BoxArms gives the arms of a box glyph.
func BoxArms(r rune) (Arms, bool) {
	arms, ok := boxArms[r]
	return arms, ok
}

// BoxRune gives the box glyph of arms. Arms with fewer than two arms have no
// glyph.
func BoxRune(arms Arms) (rune, bool) {
	r, ok := boxRunes[arms]
	return r, ok
}
