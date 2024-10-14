// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package microui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type ID uint64

type poolItem struct {
	id         ID
	lastUpdate int
}

type baseCommand struct {
	typ int
}

type jumpCommand struct {
	dstIdx int
}

type clipCommand struct {
	rect image.Rectangle
}

type rectCommand struct {
	rect  image.Rectangle
	color color.Color
}

type textCommand struct {
	pos   image.Point
	color color.Color
	str   string
}

type iconCommand struct {
	rect  image.Rectangle
	icon  icon
	color color.Color
}

type drawCommand struct {
	f func(screen *ebiten.Image)
}

type layout struct {
	body      image.Rectangle
	position  image.Point
	height    int
	max       image.Point
	widths    []int
	itemIndex int
	nextRow   int
	indent    int
}

type command struct {
	typ  int
	idx  int
	base baseCommand // type 0 (TODO)
	jump jumpCommand // type 1
	clip clipCommand // type 2
	rect rectCommand // type 3
	text textCommand // type 4
	icon iconCommand // type 5
	draw drawCommand // type 6
}

type Container struct {
	headIdx     int
	tailIdx     int
	Rect        image.Rectangle
	Body        image.Rectangle
	ContentSize image.Point
	Scroll      image.Point
	zIndex      int
	open        bool
}

type Style struct {
	Size          image.Point
	Padding       int
	Spacing       int
	Indent        int
	TitleHeight   int
	ScrollbarSize int
	ThumbSize     int
	Colors        [ColorMax + 1]color.RGBA
}

type Context struct {
	// core state

	Style         *Style
	hover         ID
	focus         ID
	LastID        ID
	lastRect      image.Rectangle
	lastZIndex    int
	keepFocus     bool
	tick          int
	hoverRoot     *Container
	nextHoverRoot *Container
	scrollTarget  *Container
	numberEditBuf string
	numberEdit    ID

	// stacks

	commandList    []*command
	rootList       []*Container
	containerStack []*Container
	clipStack      []image.Rectangle
	idStack        []ID
	layoutStack    []layout

	// retained state pools

	containerPool [containerPoolSize]poolItem
	containers    [containerPoolSize]Container
	treeNodePool  [treeNodePoolSize]poolItem

	// input state

	mousePos     image.Point
	lastMousePos image.Point
	mouseDelta   image.Point
	scrollDelta  image.Point
	mouseDown    int
	mousePressed int
	keyDown      int
	keyPressed   int
	textInput    []rune
}
