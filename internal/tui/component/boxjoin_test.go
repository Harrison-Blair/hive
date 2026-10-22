package component

import "testing"

// Decisions 66 to 68: a lit arm is heavy. A joint with lit and unlit arms
// uses the mixed glyph, as the Live boards show.
func TestBoxRuneOfArms(t *testing.T) {
	for _, tc := range []struct {
		arms Arms
		want rune
	}{
		{Arms{Down: Light, Right: Light}, '╭'},
		{Arms{Down: Light, Left: Light}, '╮'},
		{Arms{Up: Light, Left: Light}, '╯'},
		{Arms{Down: Heavy, Right: Heavy}, '┏'},
		{Arms{Down: Heavy, Left: Heavy}, '┓'},
		{Arms{Down: Heavy, Left: Light, Right: Light}, '┰'},
		{Arms{Up: Heavy, Left: Light, Right: Light}, '┸'},
		{Arms{Up: Heavy, Left: Light, Right: Heavy}, '┺'},
		{Arms{Up: Heavy, Left: Heavy, Right: Light}, '┹'},
		{Arms{Up: Heavy, Left: Heavy, Right: Heavy}, '┻'},
		{Arms{Up: Heavy, Down: Light, Left: Heavy, Right: Light}, '╃'},
		{Arms{Up: Light, Left: Heavy, Right: Heavy}, '┷'},
		{Arms{Up: Heavy, Down: Heavy, Right: Light}, '┠'},
		{Arms{Up: Light, Down: Heavy}, '╽'},
		{Arms{Up: Light, Down: Light, Left: Light, Right: Light}, '┼'},
	} {
		if got, ok := BoxRune(tc.arms); !ok || got != tc.want {
			t.Errorf("BoxRune(%v) = %q, %t; want %q", tc.arms, got, ok, tc.want)
		}
	}
	if _, ok := BoxRune(Arms{Up: Light}); ok {
		t.Error("BoxRune of one arm is a glyph; want none")
	}
}

func TestBoxArmsOfRune(t *testing.T) {
	for _, tc := range []struct {
		r    rune
		want Arms
	}{
		{'├', Arms{Up: Light, Down: Light, Right: Light}},
		{'┯', Arms{Down: Light, Left: Heavy, Right: Heavy}},
		{'━', Arms{Left: Heavy, Right: Heavy}},
		{'╰', Arms{Up: Light, Right: Light}},
		{'┨', Arms{Up: Heavy, Down: Heavy, Left: Light}},
	} {
		if got, ok := BoxArms(tc.r); !ok || got != tc.want {
			t.Errorf("BoxArms(%q) = %v, %t; want %v", tc.r, got, ok, tc.want)
		}
	}
	if _, ok := BoxArms('a'); ok {
		t.Error("BoxArms('a') is a box glyph; want none")
	}
}
