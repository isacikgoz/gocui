// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

// Attribute represents a terminal attribute, like color, font style, etc. They
// can be combined using bitwise OR (|). Note that it is not possible to
// combine multiple color attributes.
type Attribute uint16

// Color attributes.
const (
	ColorDefault Attribute = iota
	ColorBlack
)

// the colors are a little different in tcell
const (
	ColorRed Attribute = iota + 10
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

// Text style attributes.
const (
	AttrBold Attribute = 1 << (9 + iota)
	AttrUnderline
	AttrReverse
)
