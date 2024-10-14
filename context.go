// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package microui

import (
	"image"
)

func (c *Context) drawFrame(rect image.Rectangle, colorid int) {
	c.drawRect(rect, c.style.colors[colorid])
	if colorid == ColorScrollBase ||
		colorid == ColorScrollThumb ||
		colorid == ColorTitleBG {
		return
	}

	// draw border
	if c.style.colors[ColorBorder].A != 0 {
		c.drawBox(rect.Inset(-1), c.style.colors[ColorBorder])
	}
}

func NewContext() *Context {
	return &Context{
		style: &defaultStyle,
	}
}
