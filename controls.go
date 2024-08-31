// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package microui

import (
	"fmt"
	"image"
	"strconv"
	"unsafe"
)

func (c *Context) inHoverRoot() bool {
	for i := len(c.containerStack) - 1; i >= 0; i-- {
		if c.containerStack[i] == c.HoverRoot {
			return true
		}
		// only root containers have their `head` field set; stop searching if we've
		// reached the current root container
		if c.containerStack[i].HeadIdx >= 0 {
			break
		}
	}
	return false
}

func (c *Context) DrawControlFrame(id ID, rect image.Rectangle, colorid int, opt Option) {
	if (opt & OptNoFrame) != 0 {
		return
	}
	if c.Focus == id {
		colorid += 2
	} else if c.Hover == id {
		colorid++
	}
	c.drawFrame(rect, colorid)
}

func (c *Context) DrawControlText(str string, rect image.Rectangle, colorid int, opt Option) {
	var pos image.Point
	tw := textWidth(str)
	c.PushClipRect(rect)
	pos.Y = rect.Min.Y + (rect.Dy()-textHeight())/2
	if (opt & OptAlignCenter) != 0 {
		pos.X = rect.Min.X + (rect.Dx()-tw)/2
	} else if (opt & OptAlignRight) != 0 {
		pos.X = rect.Min.X + rect.Dx() - tw - c.Style.Padding
	} else {
		pos.X = rect.Min.X + c.Style.Padding
	}
	c.DrawText(str, pos, c.Style.Colors[colorid])
	c.PopClipRect()
}

func (c *Context) mouseOver(rect image.Rectangle) bool {
	return c.mousePos.In(rect) && c.mousePos.In(c.ClipRect()) && c.inHoverRoot()
}

func (c *Context) UpdateControl(id ID, rect image.Rectangle, opt Option) {
	mouseover := c.mouseOver(rect)

	if c.Focus == id {
		c.UpdatedFocus = true
	}
	if (opt & OptNoInteract) != 0 {
		return
	}
	if mouseover && c.mouseDown == 0 {
		c.Hover = id
	}

	if c.Focus == id {
		if c.mousePressed != 0 && !mouseover {
			c.SetFocus(0)
		}
		if c.mouseDown == 0 && (^opt&OptHoldFocus) != 0 {
			c.SetFocus(0)
		}
	}

	if c.Hover == id {
		if c.mousePressed != 0 {
			c.SetFocus(id)
		} else if !mouseover {
			c.Hover = 0
		}
	}
}

func (c *Context) Text(text string) {
	var start_idx, end_idx, p int
	color := c.Style.Colors[ColorText]
	c.layoutBeginColumn()
	c.LayoutRow(1, []int{-1}, textHeight())
	for end_idx < len(text) {
		r := c.LayoutNext()
		w := 0
		end_idx = p
		start_idx = end_idx
		for end_idx < len(text) && text[end_idx] != '\n' {
			word := p
			for p < len(text) && text[p] != ' ' && text[p] != '\n' {
				p++
			}
			w += textWidth(text[word:p])
			if w > r.Dx() && end_idx != start_idx {
				break
			}
			if p < len(text) {
				w += textWidth(string(text[p]))
			}
			end_idx = p
			p++
		}
		c.DrawText(text[start_idx:end_idx], r.Min, color)
		p = end_idx + 1
	}
	c.layoutEndColumn()
}

func (c *Context) Label(text string) {
	c.DrawControlText(text, c.LayoutNext(), ColorText, 0)
}

func (c *Context) ButtonEx(label string, icon Icon, opt Option) Res {
	var res Res
	var id ID
	if len(label) > 0 {
		id = c.id([]byte(label))
	} else {
		iconPtr := &icon
		// TODO: investigate if this okay, if icon represents an icon ID we might need
		// to refer to the value instead of a pointer, like commented below:
		// unsafe.Slice((*byte)(unsafe.Pointer(&icon)), unsafe.Sizeof(icon)))
		id = c.id(ptrToBytes(unsafe.Pointer(iconPtr)))
	}
	r := c.LayoutNext()
	c.UpdateControl(id, r, opt)
	// handle click
	if c.mousePressed == mouseLeft && c.Focus == id {
		res |= ResSubmit
	}
	// draw
	c.DrawControlFrame(id, r, ColorButton, opt)
	if len(label) > 0 {
		c.DrawControlText(label, r, ColorText, opt)
	}
	if icon != 0 {
		c.DrawIcon(icon, r, c.Style.Colors[ColorText])
	}
	return res
}

func (c *Context) Checkbox(label string, state *bool) Res {
	var res Res
	id := c.id(ptrToBytes(unsafe.Pointer(state)))
	r := c.LayoutNext()
	box := image.Rect(r.Min.X, r.Min.Y, r.Min.X+r.Dy(), r.Max.Y)
	c.UpdateControl(id, r, 0)
	// handle click
	if c.mousePressed == mouseLeft && c.Focus == id {
		res |= ResChange
		*state = !*state
	}
	// draw
	c.DrawControlFrame(id, box, ColorBase, 0)
	if *state {
		c.DrawIcon(IconCheck, box, c.Style.Colors[ColorText])
	}
	r = image.Rect(r.Min.X+box.Dx(), r.Min.Y, r.Max.X, r.Max.Y)
	c.DrawControlText(label, r, ColorText, 0)
	return res
}

func (c *Context) TextBoxRaw(buf *string, id ID, r image.Rectangle, opt Option) Res {
	var res Res
	c.UpdateControl(id, r, opt|OptHoldFocus)
	buflen := len(*buf)

	if c.Focus == id {
		// handle text input
		if len(c.textInput) > 0 {
			*buf += string(c.textInput)
			res |= ResChange
		}
		// handle backspace
		if (c.keyPressed&keyBackspace) != 0 && buflen > 0 {
			*buf = (*buf)[:buflen-1]
			res |= ResChange
		}
		// handle return
		if (c.keyPressed & keyReturn) != 0 {
			c.SetFocus(0)
			res |= ResSubmit
		}
	}

	// draw
	c.DrawControlFrame(id, r, ColorBase, opt)
	if c.Focus == id {
		color := c.Style.Colors[ColorText]
		textw := textWidth(*buf)
		texth := textHeight()
		ofx := r.Dx() - c.Style.Padding - textw - 1
		textx := r.Min.X + min(ofx, c.Style.Padding)
		texty := r.Min.Y + (r.Dy()-texth)/2
		c.PushClipRect(r)
		c.DrawText(*buf, image.Pt(textx, texty), color)
		c.DrawRect(image.Rect(textx+textw, texty, textx+textw+1, texty+texth), color)
		c.PopClipRect()
	} else {
		c.DrawControlText(*buf, r, ColorText, opt)
	}

	return res
}

func (c *Context) numberTextBox(value *float64, r image.Rectangle, id ID) bool {
	if c.mousePressed == mouseLeft && (c.keyDown&keyShift) != 0 &&
		c.Hover == id {
		c.NumberEdit = id
		c.NumberEditBuf = fmt.Sprintf(realFmt, *value)
	}
	if c.NumberEdit == id {
		res := c.TextBoxRaw(&c.NumberEditBuf, id, r, 0)
		if (res&ResSubmit) != 0 || c.Focus != id {
			nval, err := strconv.ParseFloat(c.NumberEditBuf, 32)
			if err != nil {
				nval = 0
			}
			*value = float64(nval)
			c.NumberEdit = 0
		} else {
			return true
		}
	}
	return false
}

func (c *Context) TextBoxEx(buf *string, opt Option) Res {
	id := c.id(ptrToBytes(unsafe.Pointer(buf)))
	r := c.LayoutNext()
	return c.TextBoxRaw(buf, id, r, opt)
}

func (c *Context) SliderEx(value *float64, low, high, step float64, format string, opt Option) Res {
	var res Res
	last := *value
	v := last
	id := c.id(ptrToBytes(unsafe.Pointer(value)))
	base := c.LayoutNext()

	// handle text input mode
	if c.numberTextBox(&v, base, id) {
		return res
	}

	// handle normal mode
	c.UpdateControl(id, base, opt)

	// handle input
	if c.Focus == id && (c.mouseDown|c.mousePressed) == mouseLeft {
		v = low + float64(c.mousePos.X-base.Min.X)*(high-low)/float64(base.Dx())
		if step != 0 {
			v = ((v + step/2) / step) * step
		}
	}
	// clamp and store value, update res
	*value = clampF(v, low, high)
	v = *value
	if last != v {
		res |= ResChange
	}

	// draw base
	c.DrawControlFrame(id, base, ColorBase, opt)
	// draw thumb
	w := c.Style.ThumbSize
	x := int((v - low) * float64(base.Dx()-w) / (high - low))
	thumb := image.Rect(base.Min.X+x, base.Min.Y, base.Min.X+x+w, base.Max.Y)
	c.DrawControlFrame(id, thumb, ColorButton, opt)
	// draw text
	text := fmt.Sprintf(format, v)
	c.DrawControlText(text, base, ColorText, opt)

	return res
}

func (c *Context) NumberEx(value *float64, step float64, format string, opt Option) Res {
	var res Res
	id := c.id(ptrToBytes(unsafe.Pointer(&value)))
	base := c.LayoutNext()
	last := *value

	// handle text input mode
	if c.numberTextBox(value, base, id) {
		return res
	}

	// handle normal mode
	c.UpdateControl(id, base, opt)

	// handle input
	if c.Focus == id && c.mouseDown == mouseLeft {
		*value += float64(c.mouseDelta.X) * step
	}
	// set flag if value changed
	if *value != last {
		res |= ResChange
	}

	// draw base
	c.DrawControlFrame(id, base, ColorBase, opt)
	// draw text
	text := fmt.Sprintf(format, *value)
	c.DrawControlText(text, base, ColorText, opt)

	return res
}

func (c *Context) header(label string, istreenode bool, opt Option) Res {
	id := c.id([]byte(label))
	idx := c.poolGet(c.treeNodePool[:], id)
	c.LayoutRow(1, []int{-1}, 0)

	active := idx >= 0
	var expanded bool
	if (opt & OptExpanded) != 0 {
		expanded = !active
	} else {
		expanded = active
	}
	r := c.LayoutNext()
	c.UpdateControl(id, r, 0)

	// handle click (TODO (port): check if this is correct)
	clicked := c.mousePressed == mouseLeft && c.Focus == id
	v1, v2 := 0, 0
	if active {
		v1 = 1
	}
	if clicked {
		v2 = 1
	}
	active = (v1 ^ v2) == 1

	// update pool ref
	if idx >= 0 {
		if active {
			c.poolUpdate(c.treeNodePool[:], idx)
		} else {
			c.treeNodePool[idx] = poolItem{}
		}
	} else if active {
		c.poolInit(c.treeNodePool[:], id)
	}

	// draw
	if istreenode {
		if c.Hover == id {
			c.drawFrame(r, ColorButtonHover)
		}
	} else {
		c.DrawControlFrame(id, r, ColorButton, 0)
	}
	var icon Icon
	if expanded {
		icon = IconExpanded
	} else {
		icon = IconCollapsed
	}
	c.DrawIcon(
		icon,
		image.Rect(r.Min.X, r.Min.Y, r.Min.X+r.Dy(), r.Max.Y),
		c.Style.Colors[ColorText],
	)
	r.Min.X += r.Dy() - c.Style.Padding
	c.DrawControlText(label, r, ColorText, 0)

	if expanded {
		return ResActive
	}
	return 0
}

func (c *Context) HeaderEx(label string, opt Option) Res {
	return c.header(label, false, opt)
}

func (c *Context) TreeNodeEx(label string, opt Option, f func(res Res)) {
	if res := c.beginTreeNodeEx(label, opt); res != 0 {
		defer c.endTreeNode()
		f(res)
	}
}

func (c *Context) beginTreeNodeEx(label string, opt Option) Res {
	res := c.header(label, true, opt)
	if (res & ResActive) != 0 {
		c.layout().indent += c.Style.Indent
		// push()
		c.idStack = append(c.idStack, c.LastID)
	}
	return res
}

func (c *Context) endTreeNode() {
	c.layout().indent -= c.Style.Indent
	c.popID()
}

// x = x, y = y, w = w, h = h
func (c *Context) scrollbarVertical(cnt *Container, b image.Rectangle, cs image.Point) {
	maxscroll := cs.Y - b.Dy()
	if maxscroll > 0 && b.Dy() > 0 {
		id := c.id([]byte("!scrollbar" + "y"))

		// get sizing / positioning
		base := b
		base.Min.X = b.Max.X
		base.Max.X = base.Min.X + c.Style.ScrollbarSize

		// handle input
		c.UpdateControl(id, base, 0)
		if c.Focus == id && c.mouseDown == mouseLeft {
			cnt.Scroll.Y += c.mouseDelta.Y * cs.Y / base.Dy()
		}
		// clamp scroll to limits
		cnt.Scroll.Y = clamp(cnt.Scroll.Y, 0, maxscroll)

		// draw base and thumb
		c.drawFrame(base, ColorScrollBase)
		thumb := base
		thumb.Max.Y = thumb.Min.Y + max(c.Style.ThumbSize, base.Dy()*b.Dy()/cs.Y)
		thumb = thumb.Add(image.Pt(0, cnt.Scroll.Y*(base.Dy()-thumb.Dy())/maxscroll))
		c.drawFrame(thumb, ColorScrollThumb)

		// set this as the scroll_target (will get scrolled on mousewheel)
		// if the mouse is over it
		if c.mouseOver(b) {
			c.ScrollTarget = cnt
		}
	} else {
		cnt.Scroll.Y = 0
	}
}

// x = y, y = x, w = h, h = w
func (c *Context) scrollbarHorizontal(cnt *Container, b image.Rectangle, cs image.Point) {
	maxscroll := cs.X - b.Dx()
	if maxscroll > 0 && b.Dx() > 0 {
		id := c.id([]byte("!scrollbar" + "x"))

		// get sizing / positioning
		base := b
		base.Min.Y = b.Max.Y
		base.Max.Y = base.Min.Y + c.Style.ScrollbarSize

		// handle input
		c.UpdateControl(id, base, 0)
		if c.Focus == id && c.mouseDown == mouseLeft {
			cnt.Scroll.X += c.mouseDelta.X * cs.X / base.Dx()
		}
		// clamp scroll to limits
		cnt.Scroll.X = clamp(cnt.Scroll.X, 0, maxscroll)

		// draw base and thumb
		c.drawFrame(base, ColorScrollBase)
		thumb := base
		thumb.Max.X = thumb.Min.X + max(c.Style.ThumbSize, base.Dx()*b.Dx()/cs.X)
		thumb = thumb.Add(image.Pt(cnt.Scroll.X*(base.Dx()-thumb.Dx())/maxscroll, 0))
		c.drawFrame(thumb, ColorScrollThumb)

		// set this as the scroll_target (will get scrolled on mousewheel)
		// if the mouse is over it
		if c.mouseOver(b) {
			c.ScrollTarget = cnt
		}
	} else {
		cnt.Scroll.X = 0
	}
}

// if `swap` is true, X = Y, Y = X, W = H, H = W
func (c *Context) scrollbar(cnt *Container, b image.Rectangle, cs image.Point, swap bool) {
	if swap {
		c.scrollbarHorizontal(cnt, b, cs)
	} else {
		c.scrollbarVertical(cnt, b, cs)
	}
}

func (c *Context) Scrollbars(cnt *Container, body *image.Rectangle) {
	sz := c.Style.ScrollbarSize
	cs := cnt.ContentSize
	cs.X += c.Style.Padding * 2
	cs.Y += c.Style.Padding * 2
	c.PushClipRect(*body)
	// resize body to make room for scrollbars
	if cs.Y > cnt.Body.Dy() {
		body.Max.X -= sz
	}
	if cs.X > cnt.Body.Dx() {
		body.Max.Y -= sz
	}
	// to create a horizontal or vertical scrollbar almost-identical code is
	// used; only the references to `x|y` `w|h` need to be switched
	c.scrollbar(cnt, *body, cs, false)
	c.scrollbar(cnt, *body, cs, true)
	c.PopClipRect()
}

func (c *Context) pushContainerBody(cnt *Container, body image.Rectangle, opt Option) {
	if (^opt & OptNoScroll) != 0 {
		c.Scrollbars(cnt, &body)
	}
	c.pushLayout(body.Inset(c.Style.Padding), cnt.Scroll)
	cnt.Body = body
}

func (c *Context) beginRootContainer(cnt *Container) {
	// push()
	c.containerStack = append(c.containerStack, cnt)
	// push container to roots list and push head command
	// push()
	c.rootList = append(c.rootList, cnt)
	cnt.HeadIdx = c.pushJump(-1)
	// set as hover root if the mouse is overlapping this container and it has a
	// higher zindex than the current hover root
	if c.mousePos.In(cnt.Rect) && (c.NextHoverRoot == nil || cnt.Zindex > c.NextHoverRoot.Zindex) {
		c.NextHoverRoot = cnt
	}
	// clipping is reset here in case a root-container is made within
	// another root-containers's begin/end block; this prevents the inner
	// root-container being clipped to the outer
	// push()
	c.clipStack = append(c.clipStack, unclippedRect)
}

func (c *Context) endRootContainer() {
	// push tail 'goto' jump command and set head 'skip' command. the final steps
	// on initing these are done in End
	cnt := c.CurrentContainer()
	cnt.TailIdx = c.pushJump(-1)
	c.commandList[cnt.HeadIdx].jump.dstIdx = len(c.commandList) //- 1
	// pop base clip rect and container
	c.PopClipRect()
	c.popContainer()
}

func (c *Context) WindowEx(title string, rect image.Rectangle, opt Option, f func(res Res)) {
	if res := c.beginWindowEx(title, rect, opt); res != 0 {
		defer c.endWindow()
		f(res)
	}
}

func (c *Context) beginWindowEx(title string, rect image.Rectangle, opt Option) Res {
	id := c.id([]byte(title))
	cnt := c.container(id, opt)
	if cnt == nil || !cnt.Open {
		return 0
	}
	// push()
	c.idStack = append(c.idStack, id)

	if cnt.Rect.Dx() == 0 {
		cnt.Rect = rect
	}
	c.beginRootContainer(cnt)
	body := cnt.Rect
	rect = body

	// draw frame
	if (^opt & OptNoFrame) != 0 {
		c.drawFrame(rect, ColorWindowBG)
	}

	// do title bar
	if (^opt & OptNoTitle) != 0 {
		tr := rect
		tr.Max.Y = tr.Min.Y + c.Style.TitleHeight
		c.drawFrame(tr, ColorTitleBG)

		// do title text
		if (^opt & OptNoTitle) != 0 {
			id := c.id([]byte("!title"))
			c.UpdateControl(id, tr, opt)
			c.DrawControlText(title, tr, ColorTitleText, opt)
			if id == c.Focus && c.mouseDown == mouseLeft {
				cnt.Rect = cnt.Rect.Add(c.mouseDelta)
			}
			body.Min.Y += tr.Dy()
		}

		// do `close` button
		if (^opt & OptNoClose) != 0 {
			id := c.id([]byte("!close"))
			r := image.Rect(tr.Max.X-tr.Dy(), tr.Min.Y, tr.Max.X, tr.Max.Y)
			tr.Max.X -= r.Dx()
			c.DrawIcon(IconClose, r, c.Style.Colors[ColorTitleText])
			c.UpdateControl(id, r, opt)
			if c.mousePressed == mouseLeft && id == c.Focus {
				cnt.Open = false
			}
		}
	}

	c.pushContainerBody(cnt, body, opt)

	// do `resize` handle
	if (^opt & OptNoResize) != 0 {
		sz := c.Style.TitleHeight
		id := c.id([]byte("!resize"))
		r := image.Rect(rect.Max.X-sz, rect.Max.Y-sz, rect.Max.X, rect.Max.Y)
		c.UpdateControl(id, r, opt)
		if id == c.Focus && c.mouseDown == mouseLeft {
			cnt.Rect.Max.X = cnt.Rect.Min.X + max(96, cnt.Rect.Dx()+c.mouseDelta.X)
			cnt.Rect.Max.Y = cnt.Rect.Min.Y + max(64, cnt.Rect.Dy()+c.mouseDelta.Y)
		}
	}

	// resize to content size
	if (opt & OptAutoSize) != 0 {
		r := c.layout().body
		cnt.Rect.Max.X = cnt.Rect.Min.X + cnt.ContentSize.X + (cnt.Rect.Dx() - r.Dx())
		cnt.Rect.Max.Y = cnt.Rect.Min.Y + cnt.ContentSize.Y + (cnt.Rect.Dy() - r.Dy())
	}

	// close if this is a popup window and elsewhere was clicked
	if (opt&OptPopup) != 0 && c.mousePressed != 0 && c.HoverRoot != cnt {
		cnt.Open = false
	}

	c.PushClipRect(cnt.Body)
	return ResActive
}

func (c *Context) endWindow() {
	c.PopClipRect()
	c.endRootContainer()
}

func (c *Context) OpenPopup(name string) {
	cnt := c.Container(name)
	// set as hover root so popup isn't closed in begin_window_ex()
	c.NextHoverRoot = cnt
	c.HoverRoot = c.NextHoverRoot
	// position at mouse cursor, open and bring-to-front
	cnt.Rect = image.Rect(c.mousePos.X, c.mousePos.Y, c.mousePos.X+1, c.mousePos.Y+1)
	cnt.Open = true
	c.BringToFront(cnt)
}

func (c *Context) Popup(name string, f func(res Res)) {
	if res := c.beginPopup(name); res != 0 {
		defer c.endPopup()
		f(res)
	}
}

func (c *Context) beginPopup(name string) Res {
	opt := OptPopup | OptAutoSize | OptNoResize | OptNoScroll | OptNoTitle | OptClosed
	return c.beginWindowEx(name, image.Rectangle{}, opt)
}

func (c *Context) endPopup() {
	c.endWindow()
}

func (c *Context) PanelEx(name string, opt Option, f func()) {
	c.beginPanelEx(name, opt)
	defer c.endPanel()
	f()
}

func (c *Context) beginPanelEx(name string, opt Option) {
	var cnt *Container
	c.pushID([]byte(name))
	cnt = c.container(c.LastID, opt)
	cnt.Rect = c.LayoutNext()
	if (^opt & OptNoFrame) != 0 {
		c.drawFrame(cnt.Rect, ColorPanelBG)
	}
	// push()
	c.containerStack = append(c.containerStack, cnt)
	c.pushContainerBody(cnt, cnt.Rect, opt)
	c.PushClipRect(cnt.Body)
}

func (c *Context) endPanel() {
	c.PopClipRect()
	c.popContainer()
}
