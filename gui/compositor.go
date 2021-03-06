package gui

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"sort"
	"sync"
)

type Compositor struct {
	Zbuffer map[int][]WindowID
	Windows map[WindowID]Drawable
	win     *pixelgl.Window
	Dnd     DragNDrop
	btnLock *sync.Mutex
	buttonsLocked bool
}

func (c *Compositor) DestroyAllWindows() {
	c.Windows = map[WindowID]Drawable{}
	c.Zbuffer = map[int][]WindowID{}
}

type WindowID int
var currWid WindowID = 0

func (c *Compositor) AddWindow(w Drawable) WindowID {
	z := w.GetZindex()
	wid := currWid

	if _, ok := c.Zbuffer[z]; ok {
		c.Zbuffer[z] = append(c.Zbuffer[z], wid)
	} else {
		c.Zbuffer[z] = []WindowID{wid}
	}

	c.Windows[wid] = w

	currWid++
	return wid
}

func (c *Compositor) DrawAllWindows() {
	zindices := make([]int, len(c.Zbuffer), len(c.Zbuffer))

	i := 0
	for idx := range c.Zbuffer {
		zindices[i] = idx
		i++
	}

	sort.Ints(zindices)

	for _, idx := range zindices {
		for _, wid := range c.Zbuffer[idx] {
			c.Windows[wid].Draw(c.win)
		}
	}

	if c.buttonsLocked {
		bounds := c.win.Bounds()

		imd := imdraw.New(nil)
		imd.Color = colornames.Red

		imd.Push(pixel.Vec{0, 0})
		imd.Push(pixel.Vec{bounds.W(), 0})
		imd.Push(pixel.Vec{bounds.W(), bounds.H()})
		imd.Push(pixel.Vec{0, bounds.H()})

		imd.Rectangle(2)
		imd.Draw(c.win)
	}
}

func (c *Compositor) GetWindowTitleAt(vec pixel.Vec) *RichWindow {
	res := map[int]*RichWindow{}

	for _, window := range c.Windows {
		if rw, ok := window.(*RichWindow); ok {
			if rw.GetTitleRectangle().Contains(vec) {
				res[rw.Zindex] = rw
			}
		}
	}

	keys := make([]int, len(res), len(res))
	i := 0
	for zindex := range res {
		keys[i] = zindex
		i++
	}
	sort.Ints(keys)

	if len(keys) != 0 {
		return res[keys[len(keys) - 1]]
	}

	return nil
}


func NewCompositor(win *pixelgl.Window) Compositor {
	return Compositor {
		Zbuffer: map[int][]WindowID{},
		Windows: map[WindowID]Drawable{},
		win:     win,
		Dnd:     DragNDrop{Initiated: false},
		btnLock: &sync.Mutex{},
		buttonsLocked: false,
	}
}

func (c *Compositor) GetWindowAt(vec pixel.Vec) *RichWindow {
	res := map[int]*RichWindow{}

	for _, window := range c.Windows {
		if rw, ok := window.(*RichWindow); ok {
			if rw.GetBoundaries().Contains(vec) {
				res[rw.Zindex] = rw
			}
		}
	}

	keys := make([]int, len(res), len(res))
	i := 0
	for zindex := range res {
		keys[i] = zindex
		i++
	}
	sort.Ints(keys)

	if len(keys) != 0 {
		return res[keys[len(keys) - 1]]
	}

	return nil
}

func (c *Compositor) LockAllButtons() {
	c.btnLock.Lock()
	defer c.btnLock.Unlock()

	c.buttonsLocked = true
}

func (c *Compositor) UnlockAllButtons() {
	c.btnLock.Lock()
	defer c.btnLock.Unlock()

	c.buttonsLocked = false
}

func (c *Compositor) CheckButtons() {
	if c.buttonsLocked {
		return
	}

	if !c.win.JustReleased(pixelgl.MouseButtonLeft) {
		return
	}

	pos := c.win.MousePosition()
	activeWindow := c.GetWindowAt(pos)

	if activeWindow != nil {
		for _, child := range activeWindow.Children {
			if child.GetBoundaries().Contains(pixel.Vec{X: pos.X - float64(activeWindow.X), Y: pos.Y - float64(activeWindow.Y)}) {
				go child.Click(pos.X, pos.Y)
				return
			}
		}
	}
}