package component

import "time"

// WorkingFrameInterval is the frame interval of the Pi 1.4.0 working
// indicator. The animation uses its own timer, so it does not refresh data.
const WorkingFrameInterval = 80 * time.Millisecond

// workingFrames gives the Pi 1.4.0 working indicator, in order. Frame 0 of a
// marker is its static state glyph, so a still frame keeps the approved
// boards.
var workingFrames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// WorkingGlyph gives the working marker at frame. A frame of 0 or less gives
// rest, the glyph of a still screen. Each other frame cycles through the Pi
// frames.
func WorkingGlyph(frame int, rest string) string {
	if frame <= 0 {
		return rest
	}
	return workingFrames[(frame-1)%len(workingFrames)]
}
