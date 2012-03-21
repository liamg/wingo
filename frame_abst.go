package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xrect"
)

type abstFrame struct {
    parent *frameParent
    clientPos clientPos
    moving moveState
    resizing resizeState
}

func newFrameAbst(p *frameParent, c Client, cp clientPos) *abstFrame {
    f := &abstFrame {
        parent: p,
        clientPos: cp,
        moving: moveState{},
        resizing: resizeState{},
    }

    return f
}

func (f *abstFrame) Destroy() {
    f.parent.window.destroy()
}

func (f *abstFrame) Map() {
    f.parent.window.map_()
}

func (f *abstFrame) Unmap() {
    f.parent.window.unmap()
}

func (f *abstFrame) Client() Client {
    return f.parent.client
}

func (f *abstFrame) Reset() {
    geom := f.Client().Geom()
    f.Moveresize(DoW | DoH, 0, 0, geom.Width(), geom.Height(), false)
}

func (f *abstFrame) Moveresize(flags uint16, x, y int16, w, h uint16,
                               ignoreHints bool) {
    f.ConfigureClient(flags, x, y, w, h, xgb.Id(0), 0, ignoreHints)
}

// Configure is from the perspective of the client.
// Namely, the width and height specified here will be precisely the width
// and height that the client itself ends up with, assuming it passes
// validation. (Therefore, the actual window itself will be bigger, because
// of decorations.)
// Moreover, the x and y coordinates are gravitized. Yuck.
func (f *abstFrame) ConfigureClient(flags uint16, x, y int16, w, h uint16,
                                    sibling xgb.Id, stackMode byte,
                                    ignoreHints bool) {
    // Defy gravity!
    if DoX & flags > 0 {
        x = f.Client().GravitizeX(x)
    }
    if DoY & flags > 0 {
        y = f.Client().GravitizeY(y)
    }

    // This will change with other frames
    if DoW & flags > 0 {
        w += f.clientPos.w
    }
    if DoH & flags > 0 {
        h += f.clientPos.h
    }

    f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints)
}

// ConfigureFrame is from the perspective of the frame.
// The fw and fh specify the width of the entire window, so that the client
// will end up slightly smaller than the width/height specified here.
// Also, the fx and fy coordinates are interpreted plainly as root window
// coordinates. (No gravitization.)
func (f *abstFrame) ConfigureFrame(flags uint16, fx, fy int16, fw, fh uint16,
                                   sibling xgb.Id, stackMode byte,
                                   ignoreHints bool) {
    cw, ch := fw, fh
    framex, framey, _, _ := xrect.RectPieces(f.Geom())
    _, _, clientw, clienth := xrect.RectPieces(f.Client().Geom())

    if DoX & flags > 0 {
        framex = fx
    }
    if DoY & flags > 0 {
        framey = fy
    }
    if DoW & flags > 0 {
        cw -= f.clientPos.w
        if !ignoreHints {
            cw = f.Client().ValidateWidth(cw)
            fw = cw + f.clientPos.w
        }
        clientw = cw
    }
    if DoH & flags > 0 {
        ch -= f.clientPos.h
        if !ignoreHints {
            ch = f.Client().ValidateHeight(ch)
            fh = ch + f.clientPos.h
        }
        clienth = ch
    }

    configNotify := xevent.NewConfigureNotify(f.Client().Id(), f.Client().Id(),
                                              0, framex, framey,
                                              clientw, clienth, 0, false)
    X.Conn().SendEvent(false, f.Client().Id(), xgb.EventMaskStructureNotify,
                       configNotify.Bytes())

    f.Client().Win().moveresize(flags | DoX | DoY,
                                f.clientPos.x, f.clientPos.y, cw, ch)
    f.Parent().Win().configure(flags, fx, fy, fw, fh, sibling, stackMode)
}

func (f *abstFrame) Geom() xrect.Rect {
    return f.parent.window.geom
}

func (f *abstFrame) Moving() bool {
    return f.moving.moving
}

func (f *abstFrame) Parent() *frameParent {
    return f.parent
}

func (f *abstFrame) ParentId() xgb.Id {
    return f.parent.window.id
}

func (f *abstFrame) ParentWin() *window {
    return f.parent.window
}

func (f *abstFrame) Resizing() bool {
    return f.resizing.resizing
}

// ValidateHeight validates a height of a *frame*, which is equivalent
// to validating the height of a client.
func (f *abstFrame) ValidateHeight(height uint16) uint16 {
    return f.Client().ValidateHeight(height - f.clientPos.h) + f.clientPos.h
}

// ValidateWidth validates a width of a *frame*, which is equivalent
// to validating the width of a client.
func (f *abstFrame) ValidateWidth(width uint16) uint16 {
    return f.Client().ValidateWidth(width - f.clientPos.w) + f.clientPos.w
}
