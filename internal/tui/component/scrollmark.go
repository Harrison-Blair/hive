package component

// The scroll mark symbols of decision 5.
const (
	markV = "┃"
	markH = "━"
)

// ScrollMark is the thumb of a scroll mark: Size cells from cell Start of the
// track. The track of a vertical mark is the body rows of a panel. The track
// of a horizontal mark is the inner cells of the bottom border. A Size of 0
// draws no mark.
type ScrollMark struct {
	Start int
	Size  int
}

func (m ScrollMark) covers(i int) bool { return i >= m.Start && i < m.Start+m.Size }

// Thumb gives the thumb of a track of track cells for a view of visible
// items from offset of total items. The size is in proportion to
// visible/total, with a minimum of one cell. Thumb gives no mark when all
// items are visible.
func Thumb(track, total, visible, offset int) ScrollMark {
	if total <= visible || track < 1 {
		return ScrollMark{}
	}
	size := max(1, track*visible/total)
	offset = min(max(offset, 0), total-visible)
	return ScrollMark{Start: offset * (track - size) / (total - visible), Size: size}
}
