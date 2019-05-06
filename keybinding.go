// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"github.com/gdamore/tcell"
)

// Keybidings are used to link a given key-press event with a handler.
type keybinding struct {
	viewName string
	key      Key
	ch       rune
	mod      Modifier
	handler  func(*Gui, *View) error
}

// newKeybinding returns a new Keybinding object.
func newKeybinding(viewname string, key Key, ch rune, mod Modifier, handler func(*Gui, *View) error) (kb *keybinding) {
	kb = &keybinding{
		viewName: viewname,
		key:      key,
		ch:       ch,
		mod:      mod,
		handler:  handler,
	}
	return kb
}

// matchKeypress returns if the keybinding matches the keypress.
func (kb *keybinding) matchKeypress(key Key, ch rune, mod Modifier) bool {
	return kb.key == key && kb.ch == ch && kb.mod == mod
}

// matchView returns if the keybinding matches the current view.
func (kb *keybinding) matchView(v *View) bool {
	if kb.viewName == "" {
		return true
	}
	return v != nil && kb.viewName == v.name
}

// Key represents special keys or keys combinations.
type Key tcell.Key

// Special keys.
const (
	KeyF1         Key = Key(tcell.KeyF1)
	KeyF2             = Key(tcell.KeyF2)
	KeyF3             = Key(tcell.KeyF3)
	KeyF4             = Key(tcell.KeyF4)
	KeyF5             = Key(tcell.KeyF5)
	KeyF6             = Key(tcell.KeyF6)
	KeyF7             = Key(tcell.KeyF7)
	KeyF8             = Key(tcell.KeyF8)
	KeyF9             = Key(tcell.KeyF9)
	KeyF10            = Key(tcell.KeyF10)
	KeyF11            = Key(tcell.KeyF11)
	KeyF12            = Key(tcell.KeyF12)
	KeyInsert         = Key(tcell.KeyInsert)
	KeyDelete         = Key(tcell.KeyDelete)
	KeyHome           = Key(tcell.KeyHome)
	KeyEnd            = Key(tcell.KeyEnd)
	KeyPgup           = Key(tcell.KeyPgUp)
	KeyPgdn           = Key(tcell.KeyPgDn)
	KeyArrowUp        = Key(tcell.KeyUp)
	KeyArrowDown      = Key(tcell.KeyDown)
	KeyArrowLeft      = Key(tcell.KeyLeft)
	KeyArrowRight     = Key(tcell.KeyRight)

	// MouseLeft      = Key(tcell.MouseLeft)
	// MouseMiddle    = Key(tcell.MouseMiddle)
	// MouseRight     = Key(tcell.MouseRight)
	// MouseRelease   = Key(tcell.MouseRelease)
	// MouseWheelUp   = Key(tcell.MouseWheelUp)
	// MouseWheelDown = Key(tcell.MouseWheelDown)
)

// Keys combinations.
const (
	KeyCtrlSpace      Key = Key(tcell.KeyCtrlSpace)
	KeyCtrlA              = Key(tcell.KeyCtrlA)
	KeyCtrlB              = Key(tcell.KeyCtrlB)
	KeyCtrlC              = Key(tcell.KeyCtrlC)
	KeyCtrlD              = Key(tcell.KeyCtrlD)
	KeyCtrlE              = Key(tcell.KeyCtrlE)
	KeyCtrlF              = Key(tcell.KeyCtrlF)
	KeyCtrlG              = Key(tcell.KeyCtrlG)
	KeyBackspace          = Key(tcell.KeyBackspace)
	KeyCtrlH              = Key(tcell.KeyCtrlH)
	KeyTab                = Key(tcell.KeyTab)
	KeyCtrlI              = Key(tcell.KeyCtrlI)
	KeyCtrlJ              = Key(tcell.KeyCtrlJ)
	KeyCtrlK              = Key(tcell.KeyCtrlK)
	KeyCtrlL              = Key(tcell.KeyCtrlL)
	KeyEnter              = Key(tcell.KeyEnter)
	KeyCtrlM              = Key(tcell.KeyCtrlM)
	KeyCtrlN              = Key(tcell.KeyCtrlN)
	KeyCtrlO              = Key(tcell.KeyCtrlO)
	KeyCtrlP              = Key(tcell.KeyCtrlP)
	KeyCtrlQ              = Key(tcell.KeyCtrlQ)
	KeyCtrlR              = Key(tcell.KeyCtrlR)
	KeyCtrlS              = Key(tcell.KeyCtrlS)
	KeyCtrlT              = Key(tcell.KeyCtrlT)
	KeyCtrlU              = Key(tcell.KeyCtrlU)
	KeyCtrlV              = Key(tcell.KeyCtrlV)
	KeyCtrlW              = Key(tcell.KeyCtrlW)
	KeyCtrlX              = Key(tcell.KeyCtrlX)
	KeyCtrlY              = Key(tcell.KeyCtrlY)
	KeyCtrlZ              = Key(tcell.KeyCtrlZ)
	KeyEsc                = Key(tcell.KeyEsc)
	KeyCtrlBackslash      = Key(tcell.KeyCtrlBackslash)
	KeyCtrlUnderscore     = Key(tcell.KeyCtrlUnderscore)
	KeyBackspace2         = Key(tcell.KeyBackspace2)
	KeySpace              = Key(0x20)
)

// Modifier allows to define special keys combinations. They can be used
// in combination with Keys or Runes when a new keybinding is defined.
type Modifier tcell.ModMask

// Modifiers.
const (
	ModNone Modifier = Modifier(0)
	ModAlt           = Modifier(tcell.ModAlt)
)
