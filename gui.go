// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"errors"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
)

var (
	// ErrQuit is used to decide if the MainLoop finished successfully.
	ErrQuit = errors.New("quit")

	// ErrUnknownView allows to assert if a View must be initialized.
	ErrUnknownView = errors.New("unknown view")
)

// OutputMode represents the terminal's output mode (8 or 256 colors).
type OutputMode int

const (
	OutputCurrent OutputMode = iota
	OutputNormal
	Output256
	Output216
	OutputGrayscale
)

// EventType represents the type of event.
type EventType uint8

// Event represents an event like a key press, mouse action, or window resize.
type Event struct {
	Type   EventType
	Mod    Modifier
	Key    Key
	Ch     rune
	Width  int
	Height int
	Err    error
	MouseX int
	MouseY int
	N      int
}

// Event types.
const (
	EventNone EventType = iota
	EventKey
	EventResize
	EventMouse
	EventInterrupt
	EventError
	EventRaw
)

// Gui represents the whole User Interface, including the views, layouts
// and keybindings.
type Gui struct {
	tbEvents    chan Event
	userEvents  chan userEvent
	screen      tcell.Screen
	views       []*View
	currentView *View
	managers    []Manager
	keybindings []*keybinding
	maxX, maxY  int
	outputMode  OutputMode

	// BgColor and FgColor allow to configure the background and foreground
	// colors of the GUI.
	BgColor, FgColor Attribute

	// SelBgColor and SelFgColor allow to configure the background and
	// foreground colors of the frame of the current view.
	SelBgColor, SelFgColor Attribute

	// If Highlight is true, Sel{Bg,Fg}Colors will be used to draw the
	// frame of the current view.
	Highlight bool

	// If Cursor is true then the cursor is enabled.
	Cursor bool

	// If Mouse is true then mouse events will be enabled.
	Mouse bool

	// If InputEsc is true, when ESC sequence is in the buffer and it doesn't
	// match any known sequence, ESC means KeyEsc.
	InputEsc bool

	// If ASCII is true then use ASCII instead of unicode to draw the
	// interface. Using ASCII is more portable.
	ASCII bool
}

// NewGui returns a new Gui object with a given output mode.
func NewGui(mode OutputMode) (*Gui, error) {
	g := &Gui{}
	encoding.Register()
	if s, e := tcell.NewScreen(); e != nil {
		return nil, e
	} else if e = s.Init(); e != nil {
		return nil, e
	} else {
		g.screen = s
	}
	g.outputMode = g.setOutputMode(mode)
	g.tbEvents = make(chan Event, 20)
	g.userEvents = make(chan userEvent, 20)

	g.maxX, g.maxY = g.screen.Size()

	g.BgColor, g.FgColor = ColorDefault, ColorDefault
	g.SelBgColor, g.SelFgColor = ColorDefault, ColorDefault

	return g, nil
}

// Close finalizes the library. It should be called after a successful
// initialization and when gocui is not needed anymore.
func (g *Gui) Close() {
	g.screen.Fini()
}

// Size returns the terminal's size.
func (g *Gui) Size() (x, y int) {
	return g.maxX, g.maxY
}

// setOutputMode is used to set the color palette used.
func (g *Gui) setOutputMode(mode OutputMode) OutputMode {
	if g.screen.Colors() < 256 {
		mode = OutputNormal
	}
	outMode := OutputNormal
	switch mode {
	case OutputCurrent:
		return outMode
	case OutputNormal, Output256, Output216, OutputGrayscale:
		outMode = mode
		return mode
	default:
		return outMode
	}
}

func fixColor(outputMode OutputMode, c tcell.Color) tcell.Color {
	if c == tcell.ColorDefault {
		return c
	}
	switch outputMode {
	case OutputNormal:
		c %= tcell.Color(16)
	case Output256:
		c %= tcell.Color(256)
	case Output216:
		c %= tcell.Color(216)
		c += tcell.Color(16)
	case OutputGrayscale:
		c %= tcell.Color(24)
		c += tcell.Color(232)
	default:
		c = tcell.ColorDefault
	}
	return c
}

func mkStyle(outputMode OutputMode, fg, bg Attribute) tcell.Style {
	st := tcell.StyleDefault

	f := tcell.Color(int(fg)&0x1ff) - 1
	b := tcell.Color(int(bg)&0x1ff) - 1

	f = fixColor(outputMode, f)
	b = fixColor(outputMode, b)
	st = st.Foreground(f).Background(b)
	if (fg|bg)&AttrBold != 0 {
		st = st.Bold(true)
	}
	if (fg|bg)&AttrUnderline != 0 {
		st = st.Underline(true)
	}
	if (fg|bg)&AttrReverse != 0 {
		st = st.Reverse(true)
	}
	return st
}

func makeEvent(tev tcell.Event) Event {
	switch tev := tev.(type) {
	case *tcell.EventInterrupt:
		return Event{Type: EventInterrupt}
	case *tcell.EventResize:
		w, h := tev.Size()
		return Event{Type: EventResize, Width: w, Height: h}
	case *tcell.EventKey:
		k := tev.Key()
		ch := rune(0)
		if k == tcell.KeyRune {
			ch = tev.Rune()
			if ch == ' ' {
				k = tcell.Key(' ')
			}
		}
		mod := tev.Modifiers()
		return Event{
			Type: EventKey,
			Key:  Key(k),
			Ch:   ch,
			Mod:  Modifier(mod),
		}
	default:
		return Event{Type: EventNone}
	}
}

// SetRune writes a rune at the given point, relative to the top-left
// corner of the terminal. It checks if the position is valid and applies
// the given colors.
func (g *Gui) SetRune(x, y int, ch rune, fgColor, bgColor Attribute) error {
	if x < 0 || y < 0 || x >= g.maxX || y >= g.maxY {
		return errors.New("invalid point")
	}
	st := mkStyle(g.outputMode, fgColor, bgColor)
	g.screen.SetContent(x, y, ch, nil, st)
	return nil
}

// Rune returns the rune contained in the cell at the given position.
// It checks if the position is valid.
func (g *Gui) Rune(x, y int) (rune, error) {
	if x < 0 || y < 0 || x >= g.maxX || y >= g.maxY {
		return ' ', errors.New("invalid point")
	}
	c, _, _, _ := g.screen.GetContent(x, y)
	return c, nil
}

// SetView creates a new view with its top-left corner at (x0, y0)
// and the bottom-right one at (x1, y1). If a view with the same name
// already exists, its dimensions are updated; otherwise, the error
// ErrUnknownView is returned, which allows to assert if the View must
// be initialized. It checks if the position is valid.
func (g *Gui) SetView(name string, x0, y0, x1, y1 int) (*View, error) {
	if x0 >= x1 || y0 >= y1 {
		return nil, errors.New("invalid dimensions")
	}
	if name == "" {
		return nil, errors.New("invalid name")
	}

	if v, err := g.View(name); err == nil {
		v.x0 = x0
		v.y0 = y0
		v.x1 = x1
		v.y1 = y1
		v.tainted = true
		return v, nil
	}

	v := newView(name, g.screen, g.outputMode, x0, y0, x1, y1, g.outputMode)
	v.BgColor, v.FgColor = g.BgColor, g.FgColor
	v.SelBgColor, v.SelFgColor = g.SelBgColor, g.SelFgColor
	g.views = append(g.views, v)
	return v, ErrUnknownView
}

// SetViewOnTop sets the given view on top of the existing ones.
func (g *Gui) SetViewOnTop(name string) (*View, error) {
	for i, v := range g.views {
		if v.name == name {
			s := append(g.views[:i], g.views[i+1:]...)
			g.views = append(s, v)
			return v, nil
		}
	}
	return nil, ErrUnknownView
}

// SetViewOnBottom sets the given view on bottom of the existing ones.
func (g *Gui) SetViewOnBottom(name string) (*View, error) {
	for i, v := range g.views {
		if v.name == name {
			s := append(g.views[:i], g.views[i+1:]...)
			g.views = append([]*View{v}, s...)
			return v, nil
		}
	}
	return nil, ErrUnknownView
}

// Views returns all the views in the GUI.
func (g *Gui) Views() []*View {
	return g.views
}

// View returns a pointer to the view with the given name, or error
// ErrUnknownView if a view with that name does not exist.
func (g *Gui) View(name string) (*View, error) {
	for _, v := range g.views {
		if v.name == name {
			return v, nil
		}
	}
	return nil, ErrUnknownView
}

// ViewByPosition returns a pointer to a view matching the given position, or
// error ErrUnknownView if a view in that position does not exist.
func (g *Gui) ViewByPosition(x, y int) (*View, error) {
	// traverse views in reverse order checking top views first
	for i := len(g.views); i > 0; i-- {
		v := g.views[i-1]
		if x > v.x0 && x < v.x1 && y > v.y0 && y < v.y1 {
			return v, nil
		}
	}
	return nil, ErrUnknownView
}

// ViewPosition returns the coordinates of the view with the given name, or
// error ErrUnknownView if a view with that name does not exist.
func (g *Gui) ViewPosition(name string) (x0, y0, x1, y1 int, err error) {
	for _, v := range g.views {
		if v.name == name {
			return v.x0, v.y0, v.x1, v.y1, nil
		}
	}
	return 0, 0, 0, 0, ErrUnknownView
}

// DeleteView deletes a view by name.
func (g *Gui) DeleteView(name string) error {
	for i, v := range g.views {
		if v.name == name {
			g.views = append(g.views[:i], g.views[i+1:]...)
			return nil
		}
	}
	return ErrUnknownView
}

// SetCurrentView gives the focus to a given view.
func (g *Gui) SetCurrentView(name string) (*View, error) {
	for _, v := range g.views {
		if v.name == name {
			g.currentView = v
			return v, nil
		}
	}
	return nil, ErrUnknownView
}

// CurrentView returns the currently focused view, or nil if no view
// owns the focus.
func (g *Gui) CurrentView() *View {
	return g.currentView
}

// SetKeybinding creates a new keybinding. If viewname equals to ""
// (empty string) then the keybinding will apply to all views. key must
// be a rune or a Key.
func (g *Gui) SetKeybinding(viewname string, key interface{}, mod Modifier, handler func(*Gui, *View) error) error {

	k, ch, err := getKey(key)
	if k == KeySpace {
		ch = ' '
	}
	if err != nil {
		return err
	}
	g.keybindings = append(g.keybindings, &keybinding{
		viewName: viewname,
		key:      k,
		ch:       ch,
		mod:      mod,
		handler:  handler,
	})
	return nil
}

// DeleteKeybinding deletes a keybinding.
func (g *Gui) DeleteKeybinding(viewname string, key interface{}, mod Modifier) error {
	k, ch, err := getKey(key)
	if err != nil {
		return err
	}

	for i, kb := range g.keybindings {
		if kb.viewName == viewname && kb.ch == ch && kb.key == k && kb.mod == mod {
			g.keybindings = append(g.keybindings[:i], g.keybindings[i+1:]...)
			return nil
		}
	}
	return errors.New("keybinding not found")
}

// DeleteKeybindings deletes all keybindings of view.
func (g *Gui) DeleteKeybindings(viewname string) {
	var s []*keybinding
	for _, kb := range g.keybindings {
		if kb.viewName != viewname {
			s = append(s, kb)
		}
	}
	g.keybindings = s
}

// getKey takes an empty interface with a key and returns the corresponding
// typed Key or rune.
func getKey(key interface{}) (Key, rune, error) {
	switch t := key.(type) {
	case Key:
		return t, 0, nil
	case rune:
		return 256, t, nil
	default:
		return 0, 0, errors.New("unknown type")
	}
}

// userEvent represents an event triggered by the user.
type userEvent struct {
	f func(*Gui) error
}

// Update executes the passed function. This method can be called safely from a
// goroutine in order to update the GUI. It is important to note that the
// passed function won't be executed immediately, instead it will be added to
// the user events queue. Given that Update spawns a goroutine, the order in
// which the user events will be handled is not guaranteed.
func (g *Gui) Update(f func(*Gui) error) {
	go func() { g.userEvents <- userEvent{f: f} }()
}

// A Manager is in charge of GUI's layout and can be used to build widgets.
type Manager interface {
	// Layout is called every time the GUI is redrawn, it must contain the
	// base views and its initializations.
	Layout(*Gui) error
}

// The ManagerFunc type is an adapter to allow the use of ordinary functions as
// Managers. If f is a function with the appropriate signature, ManagerFunc(f)
// is an Manager object that calls f.
type ManagerFunc func(*Gui) error

// Layout calls f(g)
func (f ManagerFunc) Layout(g *Gui) error {
	return f(g)
}

// SetManager sets the given GUI managers. It deletes all views and
// keybindings.
func (g *Gui) SetManager(managers ...Manager) {
	g.managers = managers
	g.currentView = nil
	g.views = nil
	g.keybindings = nil

	go func() {
		w, h := g.screen.Size()
		g.tbEvents <- Event{Type: EventResize, Width: w, Height: h}
	}()
}

// SetManagerFunc sets the given manager function. It deletes all views and
// keybindings.
func (g *Gui) SetManagerFunc(manager func(*Gui) error) {
	g.SetManager(ManagerFunc(manager))
}

// MainLoop runs the main loop until an error is returned. A successful
// finish should return ErrQuit.
func (g *Gui) MainLoop() error {
	go func() {
		for {
			ev := g.screen.PollEvent()
			g.tbEvents <- makeEvent(ev)
		}
	}()

	if err := g.flush(); err != nil {
		return err
	}
	for {
		select {
		case ev := <-g.tbEvents:
			if err := g.handleEvent(&ev); err != nil {
				return err
			}
		case ev := <-g.userEvents:
			if err := ev.f(g); err != nil {
				return err
			}
		}
		if err := g.consumeevents(); err != nil {
			return err
		}
		if err := g.flush(); err != nil {
			return err
		}
	}
}

// consumeevents handles the remaining events in the events pool.
func (g *Gui) consumeevents() error {
	for {
		select {
		case ev := <-g.tbEvents:
			if err := g.handleEvent(&ev); err != nil {
				return err
			}
		case ev := <-g.userEvents:
			if err := ev.f(g); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

// handleEvent handles an event, based on its type (key-press, error,
// etc.)
func (g *Gui) handleEvent(ev *Event) error {
	switch ev.Type {
	case EventKey, EventMouse:
		return g.onKey(ev)
	case EventError:
		return ev.Err
	default:
		return nil
	}
}

// flush updates the gui, re-drawing frames and buffers.
func (g *Gui) flush() error {
	st := mkStyle(g.outputMode, g.FgColor, g.BgColor)
	w, h := g.screen.Size()
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			g.screen.SetContent(col, row, ' ', nil, st)
		}
	}

	maxX, maxY := g.screen.Size()
	// if GUI's size has changed, we need to redraw all views
	if maxX != g.maxX || maxY != g.maxY {
		for _, v := range g.views {
			v.tainted = true
		}
	}
	g.maxX, g.maxY = maxX, maxY

	for _, m := range g.managers {
		if err := m.Layout(g); err != nil {
			return err
		}
	}
	for _, v := range g.views {
		if v.Frame {
			var fgColor, bgColor Attribute
			if g.Highlight && v == g.currentView {
				fgColor = g.SelFgColor
				bgColor = g.SelBgColor
			} else {
				fgColor = g.FgColor
				bgColor = g.BgColor
			}

			if err := g.drawFrameEdges(v, fgColor, bgColor); err != nil {
				return err
			}
			if err := g.drawFrameCorners(v, fgColor, bgColor); err != nil {
				return err
			}
			if v.Title != "" {
				if err := g.drawTitle(v, fgColor, bgColor); err != nil {
					return err
				}
			}
		}
		if err := g.draw(v); err != nil {
			return err
		}
	}
	g.screen.Show()
	return nil
}

// drawFrameEdges draws the horizontal and vertical edges of a view.
func (g *Gui) drawFrameEdges(v *View, fgColor, bgColor Attribute) error {
	runeH, runeV := '─', '│'
	if g.ASCII {
		runeH, runeV = '-', '|'
	}

	for x := v.x0 + 1; x < v.x1 && x < g.maxX; x++ {
		if x < 0 {
			continue
		}
		if v.y0 > -1 && v.y0 < g.maxY {
			if err := g.SetRune(x, v.y0, runeH, fgColor, bgColor); err != nil {
				return err
			}
		}
		if v.y1 > -1 && v.y1 < g.maxY {
			if err := g.SetRune(x, v.y1, runeH, fgColor, bgColor); err != nil {
				return err
			}
		}
	}
	for y := v.y0 + 1; y < v.y1 && y < g.maxY; y++ {
		if y < 0 {
			continue
		}
		if v.x0 > -1 && v.x0 < g.maxX {
			if err := g.SetRune(v.x0, y, runeV, fgColor, bgColor); err != nil {
				return err
			}
		}
		if v.x1 > -1 && v.x1 < g.maxX {
			if err := g.SetRune(v.x1, y, runeV, fgColor, bgColor); err != nil {
				return err
			}
		}
	}
	return nil
}

// drawFrameCorners draws the corners of the view.
func (g *Gui) drawFrameCorners(v *View, fgColor, bgColor Attribute) error {
	runeTL, runeTR, runeBL, runeBR := '┌', '┐', '└', '┘'
	if g.ASCII {
		runeTL, runeTR, runeBL, runeBR = '+', '+', '+', '+'
	}

	corners := []struct {
		x, y int
		ch   rune
	}{{v.x0, v.y0, runeTL}, {v.x1, v.y0, runeTR}, {v.x0, v.y1, runeBL}, {v.x1, v.y1, runeBR}}

	for _, c := range corners {
		if c.x >= 0 && c.y >= 0 && c.x < g.maxX && c.y < g.maxY {
			if err := g.SetRune(c.x, c.y, c.ch, fgColor, bgColor); err != nil {
				return err
			}
		}
	}
	return nil
}

// drawTitle draws the title of the view.
func (g *Gui) drawTitle(v *View, fgColor, bgColor Attribute) error {
	if v.y0 < 0 || v.y0 >= g.maxY {
		return nil
	}

	for i, ch := range v.Title {
		x := v.x0 + i + 2
		if x < 0 {
			continue
		} else if x > v.x1-2 || x >= g.maxX {
			break
		}
		if err := g.SetRune(x, v.y0, ch, fgColor, bgColor); err != nil {
			return err
		}
	}
	return nil
}

// draw manages the cursor and calls the draw function of a view.
func (g *Gui) draw(v *View) error {
	if g.Cursor {
		if curview := g.currentView; curview != nil {
			vMaxX, vMaxY := curview.Size()
			if curview.cx < 0 {
				curview.cx = 0
			} else if curview.cx >= vMaxX {
				curview.cx = vMaxX - 1
			}
			if curview.cy < 0 {
				curview.cy = 0
			} else if curview.cy >= vMaxY {
				curview.cy = vMaxY - 1
			}

			gMaxX, gMaxY := g.Size()
			cx, cy := curview.x0+curview.cx+1, curview.y0+curview.cy+1
			if cx >= 0 && cx < gMaxX && cy >= 0 && cy < gMaxY {
				g.screen.ShowCursor(cx, cy)
			} else {
				g.screen.ShowCursor(-1, -1)
			}
		}
	} else {
		g.screen.ShowCursor(-1, -1)
	}

	v.clearRunes()
	if err := v.draw(); err != nil {
		return err
	}
	return nil
}

// onKey manages key-press events. A keybinding handler is called when
// a key-press or mouse event satisfies a configured keybinding. Furthermore,
// currentView's internal buffer is modified if currentView.Editable is true.
func (g *Gui) onKey(ev *Event) error {
	switch ev.Type {
	case EventKey:
		matched, err := g.execKeybindings(g.currentView, ev)
		if err != nil {
			return err
		}
		if matched {
			break
		}
		if g.currentView != nil && g.currentView.Editable && g.currentView.Editor != nil {
			g.currentView.Editor.Edit(g.currentView, Key(ev.Key), ev.Ch, Modifier(ev.Mod))
		}
	case EventMouse:
		mx, my := ev.MouseX, ev.MouseY
		v, err := g.ViewByPosition(mx, my)
		if err != nil {
			break
		}
		if err := v.SetCursor(mx-v.x0-1, my-v.y0-1); err != nil {
			return err
		}
		if _, err := g.execKeybindings(v, ev); err != nil {
			return err
		}
	}

	return nil
}

// execKeybindings executes the keybinding handlers that match the passed view
// and event. The value of matched is true if there is a match and no errors.
func (g *Gui) execKeybindings(v *View, ev *Event) (matched bool, err error) {
	matched = false

	for _, kb := range g.keybindings {
		if kb.handler == nil {
			continue
		}
		//log.Fatalf("match: %t, found: %t", kb.matchView(v), kb.matchKeypress(Key(ev.Key), ev.Ch, Modifier(ev.Mod)))
		if kb.matchKeypress(Key(ev.Key), ev.Ch, Modifier(ev.Mod)) && kb.matchView(v) {
			if err := kb.handler(g, v); err != nil {
				return false, err
			}
			matched = true
		}
	}
	return matched, nil
}
