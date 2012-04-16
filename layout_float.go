package main

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

const layoutFloatMaxOverlapClients = 20

type floating struct {
	workspace *workspace
}

func newFloating(wrk *workspace) *floating {
	return &floating{
		workspace: wrk,
	}
}

func (ly *floating) floating() bool {
	return true
}
func (ly *floating) place()           {}
func (ly *floating) unplace()         {}
func (ly *floating) add(c *client)    {}
func (ly *floating) remove(c *client) {}

func (ly *floating) maximizable() bool {
	return true
}

func (ly *floating) move(c *client, x, y int) {
	c.move(x, y)
}

func (ly *floating) resize(c *client, w, h int) {
	c.resize(w, h)
}

func (ly *floating) moveresize(c *client, x, y, w, h int) {
	c.moveresize(x, y, w, h)
}

// xy_no_overlap positions c in a free rectangle on the screen.
// The algorithm is very close to that described here: http://goo.gl/kwPFr
// XXX: xy_no_overlap currently only works when there are fewer than
// layoutFloatMaxOverlapClients visible clients on the active workspace.
// It becomes a bit too slow otherwise.
func (ly *floating) xy_no_overlap(c *client) {
	// Function to determine whether a client's geometry should be
	// factored into the free rectangles calculation.
	// Ignore: clients not on c's workspace, c itself, unmapped clients
	// and clients that aren't floating (i.e., are tiling).
	useClient := func(c2 *client) bool {
		return c.workspace.id == c2.workspace.id && c.Id() != c2.Id() &&
			c2.Mapped() && c2.layout().floating()
	}

	// The geometry of the active head serves as a starting point of our
	// free rectangle algorithm, and also as a fallback if there are no
	// suitable free rectangles.
	headGeom := WM.headActive()

	// If there are too many clients visible on this workspace, give up.
	// (It's too slow otherwise.)
	clientCount := 0
	for _, c2 := range WM.clients {
		if useClient(c2) {
			clientCount++
		}
	}
	if clientCount > layoutFloatMaxOverlapClients {
		c.move(headGeom.X(), headGeom.Y())
		return
	}

	// Constructs a slice of rects representing all empty rectangles on
	// the active head.
	emptyRects := func() []xrect.Rect {
		// start the "empty" area with the current monitor
		empties := []xrect.Rect{headGeom}

		for _, c2 := range WM.clients {
			if !useClient(c2) {
				continue
			}

			temp := make([]xrect.Rect, len(empties))
			copy(temp, empties)
			for _, rect := range temp {
				// Find rect in 'empties' and remove it.
				for i, rect2 := range empties {
					if rect2.X() == rect.X() && rect2.Y() == rect.Y() {
						empties = append(empties[:i], empties[i+1:]...)
						break
					}
				}
				empties = append(empties,
					xrect.Subtract(rect, c2.Frame().Geom())...)
			}
		}

		return empties
	}

	cg := c.Frame().Geom()
	empties := emptyRects()
	var choose xrect.Rect = nil
	for _, rect := range empties {
		if rect.Width() < cg.Width() || rect.Height() < cg.Height() {
			continue
		}
		if choose == nil || rect.Y() < choose.Y() ||
			(rect.Y() == choose.Y() && rect.X() < choose.X()) {

			choose = rect
		}
	}

	if choose == nil {
		c.move(headGeom.X(), headGeom.Y())
	} else {
		c.move(choose.X(), choose.Y())
	}
}