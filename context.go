// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package microui

import (
	"image"
)

func (c *Context) drawFrame(rect image.Rectangle, colorid int) {
	c.drawRect(rect, c.Style.Colors[colorid])
	if colorid == ColorScrollBase ||
		colorid == ColorScrollThumb ||
		colorid == ColorTitleBG {
		return
	}

	// draw border
	if c.Style.Colors[ColorBorder].A != 0 {
		c.drawBox(rect.Inset(-1), c.Style.Colors[ColorBorder])
	}
}

func NewContext() *Context {
	return &Context{
		Style: &defaultStyle,
	}
}
