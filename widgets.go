// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package microui

import "image"

func (c *Context) Button(label string) Res {
	return c.ButtonEx(label, OptAlignCenter)
}

func (c *Context) TextBox(buf *string) Res {
	return c.TextBoxEx(buf, 0)
}

func (c *Context) Slider(value *float64, lo, hi float64) Res {
	return c.SliderEx(value, lo, hi, 0, sliderFmt, OptAlignCenter)
}

func (c *Context) Number(value *float64, step float64) Res {
	return c.NumberEx(value, step, sliderFmt, OptAlignCenter)
}

func (c *Context) Header(label string) Res {
	return c.HeaderEx(label, 0)
}

func (c *Context) TreeNode(label string, f func(res Res)) {
	c.TreeNodeEx(label, 0, f)
}

func (c *Context) Window(title string, rect image.Rectangle, f func(res Res)) {
	c.WindowEx(title, rect, 0, f)
}

func (c *Context) Panel(name string, f func()) {
	c.PanelEx(name, 0, f)
}
