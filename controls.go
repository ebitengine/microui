// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package microui

import (
	"fmt"
	"image"
	"strconv"
	"unsafe"
)

/*============================================================================
** controls
**============================================================================*/

func (ctx *Context) InHoverRoot() bool {
	for i := len(ctx.ContainerStack) - 1; i >= 0; i-- {
		if ctx.ContainerStack[i] == ctx.HoverRoot {
			return true
		}
		// only root containers have their `head` field set; stop searching if we've
		// reached the current root container
		if ctx.ContainerStack[i].HeadIdx >= 0 {
			break
		}
	}
	return false
}

func (ctx *Context) DrawControlFrame(id ID, rect image.Rectangle, colorid int, opt Option) {
	if (opt & OptNoFrame) != 0 {
		return
	}
	if ctx.Focus == id {
		colorid += 2
	} else if ctx.Hover == id {
		colorid++
	}
	ctx.DrawFrame(ctx, rect, colorid)
}

func (ctx *Context) DrawControlText(str string, rect image.Rectangle, colorid int, opt Option) {
	var pos image.Point
	font := ctx.Style.Font
	tw := ctx.TextWidth(font, str)
	ctx.PushClipRect(rect)
	pos.Y = rect.Min.Y + (rect.Dy()-ctx.TextHeight(font))/2
	if (opt & OptAlignCenter) != 0 {
		pos.X = rect.Min.X + (rect.Dx()-tw)/2
	} else if (opt & OptAlignRight) != 0 {
		pos.X = rect.Min.X + rect.Dx() - tw - ctx.Style.Padding
	} else {
		pos.X = rect.Min.X + ctx.Style.Padding
	}
	ctx.DrawText(font, str, pos, ctx.Style.Colors[colorid])
	ctx.PopClipRect()
}

func (ctx *Context) mouseOver(rect image.Rectangle) bool {
	return ctx.mousePos.In(rect) && ctx.mousePos.In(ctx.GetClipRect()) && ctx.InHoverRoot()
}

func (ctx *Context) updateControl(id ID, rect image.Rectangle, opt Option) {
	mouseover := ctx.mouseOver(rect)

	if ctx.Focus == id {
		ctx.UpdatedFocus = true
	}
	if (opt & OptNoInteract) != 0 {
		return
	}
	if mouseover && ctx.mouseDown == 0 {
		ctx.Hover = id
	}

	if ctx.Focus == id {
		if ctx.mousePressed != 0 && !mouseover {
			ctx.SetFocus(0)
		}
		if ctx.mouseDown == 0 && (^opt&OptHoldFocus) != 0 {
			ctx.SetFocus(0)
		}
	}

	if ctx.Hover == id {
		if ctx.mousePressed != 0 {
			ctx.SetFocus(id)
		} else if !mouseover {
			ctx.Hover = 0
		}
	}
}

func (ctx *Context) Text(text string) {
	var start_idx, end_idx, p int
	font := ctx.Style.Font
	color := ctx.Style.Colors[ColorText]
	ctx.LayoutBeginColumn()
	ctx.LayoutRow(1, []int{-1}, ctx.TextHeight(font))
	for end_idx < len(text) {
		r := ctx.LayoutNext()
		w := 0
		end_idx = p
		start_idx = end_idx
		for end_idx < len(text) && text[end_idx] != '\n' {
			word := p
			for p < len(text) && text[p] != ' ' && text[p] != '\n' {
				p++
			}
			w += ctx.TextWidth(font, text[word:p])
			if w > r.Dx() && end_idx != start_idx {
				break
			}
			if p < len(text) {
				w += ctx.TextWidth(font, string(text[p]))
			}
			end_idx = p
			p++
		}
		ctx.DrawText(font, text[start_idx:end_idx], r.Min, color)
		p = end_idx + 1
	}
	ctx.LayoutEndColumn()
}

func (ctx *Context) Label(text string) {
	ctx.DrawControlText(text, ctx.LayoutNext(), ColorText, 0)
}

func (ctx *Context) ButtonEx(label string, icon Icon, opt Option) int {
	var res int = 0
	var id ID
	if len(label) > 0 {
		id = ctx.GetID([]byte(label))
	} else {
		iconPtr := &icon
		// TODO: investigate if this okay, if icon represents an icon ID we might need
		// to refer to the value instead of a pointer, like commented below:
		// unsafe.Slice((*byte)(unsafe.Pointer(&icon)), unsafe.Sizeof(icon)))
		id = ctx.GetID(PtrToBytes(unsafe.Pointer(iconPtr)))
	}
	r := ctx.LayoutNext()
	ctx.updateControl(id, r, opt)
	// handle click
	if ctx.mousePressed == mouseLeft && ctx.Focus == id {
		res |= ResSubmit
	}
	// draw
	ctx.DrawControlFrame(id, r, ColorButton, opt)
	if len(label) > 0 {
		ctx.DrawControlText(label, r, ColorText, opt)
	}
	if icon != 0 {
		ctx.DrawIcon(icon, r, ctx.Style.Colors[ColorText])
	}
	return res
}

func (ctx *Context) Checkbox(label string, state *bool) int {
	var res int = 0
	id := ctx.GetID(PtrToBytes(unsafe.Pointer(state)))
	r := ctx.LayoutNext()
	box := image.Rect(r.Min.X, r.Min.Y, r.Min.X+r.Dy(), r.Max.Y)
	ctx.updateControl(id, r, 0)
	// handle click
	if ctx.mousePressed == mouseLeft && ctx.Focus == id {
		res |= ResChange
		*state = !*state
	}
	// draw
	ctx.DrawControlFrame(id, box, ColorBase, 0)
	if *state {
		ctx.DrawIcon(IconCheck, box, ctx.Style.Colors[ColorText])
	}
	r = image.Rect(r.Min.X+box.Dx(), r.Min.Y, r.Max.X, r.Max.Y)
	ctx.DrawControlText(label, r, ColorText, 0)
	return res
}

func (ctx *Context) TextboxRaw(buf *string, id ID, r image.Rectangle, opt Option) int {
	var res int = 0
	ctx.updateControl(id, r, opt|OptHoldFocus)
	buflen := len(*buf)

	if ctx.Focus == id {
		// handle text input
		if len(ctx.textInput) > 0 {
			*buf += string(ctx.textInput)
			res |= ResChange
		}
		// handle backspace
		if (ctx.keyPressed&keyBackspace) != 0 && buflen > 0 {
			*buf = (*buf)[:buflen-1]
			res |= ResChange
		}
		// handle return
		if (ctx.keyPressed & keyReturn) != 0 {
			ctx.SetFocus(0)
			res |= ResSubmit
		}
	}

	// draw
	ctx.DrawControlFrame(id, r, ColorBase, opt)
	if ctx.Focus == id {
		color := ctx.Style.Colors[ColorText]
		font := ctx.Style.Font
		textw := ctx.TextWidth(font, *buf)
		texth := ctx.TextHeight(font)
		ofx := r.Dx() - ctx.Style.Padding - textw - 1
		textx := r.Min.X + min(ofx, ctx.Style.Padding)
		texty := r.Min.Y + (r.Dy()-texth)/2
		ctx.PushClipRect(r)
		ctx.DrawText(font, *buf, image.Pt(textx, texty), color)
		ctx.DrawRect(image.Rect(textx+textw, texty, textx+textw+1, texty+texth), color)
		ctx.PopClipRect()
	} else {
		ctx.DrawControlText(*buf, r, ColorText, opt)
	}

	return res
}

func (ctx *Context) NumberTextBox(value *float32, r image.Rectangle, id ID) bool {
	if ctx.mousePressed == mouseLeft && (ctx.keyDown&keyShift) != 0 &&
		ctx.Hover == id {
		ctx.NumberEdit = id
		ctx.NumberEditBuf = fmt.Sprintf(realFmt, *value)
	}
	if ctx.NumberEdit == id {
		res := ctx.TextboxRaw(&ctx.NumberEditBuf, id, r, 0)
		if (res&ResSubmit) != 0 || ctx.Focus != id {
			nval, err := strconv.ParseFloat(ctx.NumberEditBuf, 32)
			if err != nil {
				nval = 0
			}
			*value = float32(nval)
			ctx.NumberEdit = 0
		} else {
			return true
		}
	}
	return false
}

func (ctx *Context) TextBoxEx(buf *string, opt Option) int {
	id := ctx.GetID(PtrToBytes(unsafe.Pointer(buf)))
	r := ctx.LayoutNext()
	return ctx.TextboxRaw(buf, id, r, opt)
}

func (ctx *Context) SliderEx(value *float32, low, high, step float32, format string, opt Option) int {
	var x, w, res int = 0, 0, 0
	last := *value
	v := last
	id := ctx.GetID(PtrToBytes(unsafe.Pointer(value)))
	base := ctx.LayoutNext()

	// handle text input mode
	if ctx.NumberTextBox(&v, base, id) {
		return res
	}

	// handle normal mode
	ctx.updateControl(id, base, opt)

	// handle input
	if ctx.Focus == id && (ctx.mouseDown|ctx.mousePressed) == mouseLeft {
		v = low + float32(ctx.mousePos.X-base.Min.X)*(high-low)/float32(base.Dx())
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
	ctx.DrawControlFrame(id, base, ColorBase, opt)
	// draw thumb
	w = ctx.Style.ThumbSize
	x = int((v - low) * float32(base.Dx()-w) / (high - low))
	thumb := image.Rect(base.Min.X+x, base.Min.Y, base.Min.X+x+w, base.Max.Y)
	ctx.DrawControlFrame(id, thumb, ColorButton, opt)
	// draw text
	text := fmt.Sprintf(format, v)
	ctx.DrawControlText(text, base, ColorText, opt)

	return res
}

func (ctx *Context) NumberEx(value *float32, step float32, format string, opt Option) int {
	var res int = 0
	id := ctx.GetID(PtrToBytes(unsafe.Pointer(&value)))
	base := ctx.LayoutNext()
	last := *value

	// handle text input mode
	if ctx.NumberTextBox(value, base, id) {
		return res
	}

	// handle normal mode
	ctx.updateControl(id, base, opt)

	// handle input
	if ctx.Focus == id && ctx.mouseDown == mouseLeft {
		*value += float32(ctx.mouseDelta.X) * step
	}
	// set flag if value changed
	if *value != last {
		res |= ResChange
	}

	// draw base
	ctx.DrawControlFrame(id, base, ColorBase, opt)
	// draw text
	text := fmt.Sprintf(format, *value)
	ctx.DrawControlText(text, base, ColorText, opt)

	return res
}

func (ctx *Context) header(label string, istreenode bool, opt Option) int {
	id := ctx.GetID([]byte(label))
	idx := ctx.poolGet(ctx.treeNodePool[:], id)
	ctx.LayoutRow(1, []int{-1}, 0)

	active := idx >= 0
	var expanded bool
	if (opt & OptExpanded) != 0 {
		expanded = !active
	} else {
		expanded = active
	}
	r := ctx.LayoutNext()
	ctx.updateControl(id, r, 0)

	// handle click (TODO (port): check if this is correct)
	clicked := ctx.mousePressed == mouseLeft && ctx.Focus == id
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
			ctx.poolUpdate(ctx.treeNodePool[:], idx)
		} else {
			ctx.treeNodePool[idx] = poolItem{}
		}
	} else if active {
		ctx.poolInit(ctx.treeNodePool[:], id)
	}

	// draw
	if istreenode {
		if ctx.Hover == id {
			ctx.DrawFrame(ctx, r, ColorButtonHover)
		}
	} else {
		ctx.DrawControlFrame(id, r, ColorButton, 0)
	}
	var icon Icon
	if expanded {
		icon = IconExpanded
	} else {
		icon = IconCollapsed
	}
	ctx.DrawIcon(
		icon,
		image.Rect(r.Min.X, r.Min.Y, r.Min.X+r.Dy(), r.Max.Y),
		ctx.Style.Colors[ColorText],
	)
	r.Min.X += r.Dy() - ctx.Style.Padding
	ctx.DrawControlText(label, r, ColorText, 0)

	if expanded {
		return ResActive
	}
	return 0
}

func (ctx *Context) HeaderEx(label string, opt Option) int {
	return ctx.header(label, false, opt)
}

func (ctx *Context) BeginTreeNodeEx(label string, opt Option) int {
	res := ctx.header(label, true, opt)
	if (res & ResActive) != 0 {
		ctx.GetLayout().Indent += ctx.Style.Indent
		// push()
		ctx.IDStack = append(ctx.IDStack, ctx.LastID)
	}
	return res
}

func (ctx *Context) EndTreeNode() {
	ctx.GetLayout().Indent -= ctx.Style.Indent
	ctx.PopID()
}

// x = x, y = y, w = w, h = h
func (ctx *Context) scrollbarVertical(cnt *Container, b image.Rectangle, cs image.Point) {
	maxscroll := cs.Y - b.Dy()
	if maxscroll > 0 && b.Dy() > 0 {
		id := ctx.GetID([]byte("!scrollbar" + "y"))

		// get sizing / positioning
		base := b
		base.Min.X = b.Max.X
		base.Max.X = base.Min.X + ctx.Style.ScrollbarSize

		// handle input
		ctx.updateControl(id, base, 0)
		if ctx.Focus == id && ctx.mouseDown == mouseLeft {
			cnt.Scroll.Y += ctx.mouseDelta.Y * cs.Y / base.Dy()
		}
		// clamp scroll to limits
		cnt.Scroll.Y = clamp(cnt.Scroll.Y, 0, maxscroll)

		// draw base and thumb
		ctx.DrawFrame(ctx, base, ColorScrollBase)
		thumb := base
		thumb.Max.Y = thumb.Min.Y + max(ctx.Style.ThumbSize, base.Dy()*b.Dy()/cs.Y)
		thumb = thumb.Add(image.Pt(0, cnt.Scroll.Y*(base.Dy()-thumb.Dy())/maxscroll))
		ctx.DrawFrame(ctx, thumb, ColorScrollThumb)

		// set this as the scroll_target (will get scrolled on mousewheel)
		// if the mouse is over it
		if ctx.mouseOver(b) {
			ctx.ScrollTarget = cnt
		}
	} else {
		cnt.Scroll.Y = 0
	}
}

// x = y, y = x, w = h, h = w
func (ctx *Context) scrollbarHorizontal(cnt *Container, b image.Rectangle, cs image.Point) {
	maxscroll := cs.X - b.Dx()
	if maxscroll > 0 && b.Dx() > 0 {
		id := ctx.GetID([]byte("!scrollbar" + "x"))

		// get sizing / positioning
		base := b
		base.Min.Y = b.Max.Y
		base.Max.Y = base.Min.Y + ctx.Style.ScrollbarSize

		// handle input
		ctx.updateControl(id, base, 0)
		if ctx.Focus == id && ctx.mouseDown == mouseLeft {
			cnt.Scroll.X += ctx.mouseDelta.X * cs.X / base.Dx()
		}
		// clamp scroll to limits
		cnt.Scroll.X = clamp(cnt.Scroll.X, 0, maxscroll)

		// draw base and thumb
		ctx.DrawFrame(ctx, base, ColorScrollBase)
		thumb := base
		thumb.Max.X = thumb.Min.X + max(ctx.Style.ThumbSize, base.Dx()*b.Dx()/cs.X)
		thumb = thumb.Add(image.Pt(cnt.Scroll.X*(base.Dx()-thumb.Dx())/maxscroll, 0))
		ctx.DrawFrame(ctx, thumb, ColorScrollThumb)

		// set this as the scroll_target (will get scrolled on mousewheel)
		// if the mouse is over it
		if ctx.mouseOver(b) {
			ctx.ScrollTarget = cnt
		}
	} else {
		cnt.Scroll.X = 0
	}
}

// if `swap` is true, X = Y, Y = X, W = H, H = W
func (ctx *Context) AddScrollbar(cnt *Container, b image.Rectangle, cs image.Point, swap bool) {
	if swap {
		ctx.scrollbarHorizontal(cnt, b, cs)
	} else {
		ctx.scrollbarVertical(cnt, b, cs)
	}
}

func (ctx *Context) Scrollbars(cnt *Container, body *image.Rectangle) {
	sz := ctx.Style.ScrollbarSize
	cs := cnt.ContentSize
	cs.X += ctx.Style.Padding * 2
	cs.Y += ctx.Style.Padding * 2
	ctx.PushClipRect(*body)
	// resize body to make room for scrollbars
	if cs.Y > cnt.Body.Dy() {
		body.Max.X -= sz
	}
	if cs.X > cnt.Body.Dx() {
		body.Max.Y -= sz
	}
	// to create a horizontal or vertical scrollbar almost-identical code is
	// used; only the references to `x|y` `w|h` need to be switched
	ctx.AddScrollbar(cnt, *body, cs, false)
	ctx.AddScrollbar(cnt, *body, cs, true)
	ctx.PopClipRect()
}

func (ctx *Context) PushContainerBody(cnt *Container, body image.Rectangle, opt Option) {
	if (^opt & OptNoScroll) != 0 {
		ctx.Scrollbars(cnt, &body)
	}
	ctx.PushLayout(body.Inset(ctx.Style.Padding), cnt.Scroll)
	cnt.Body = body
}

func (ctx *Context) BeginRootContainer(cnt *Container) {
	// push()
	ctx.ContainerStack = append(ctx.ContainerStack, cnt)
	// push container to roots list and push head command
	// push()
	ctx.RootList = append(ctx.RootList, cnt)
	cnt.HeadIdx = ctx.pushJump(-1)
	// set as hover root if the mouse is overlapping this container and it has a
	// higher zindex than the current hover root
	if ctx.mousePos.In(cnt.Rect) && (ctx.NextHoverRoot == nil || cnt.Zindex > ctx.NextHoverRoot.Zindex) {
		ctx.NextHoverRoot = cnt
	}
	// clipping is reset here in case a root-container is made within
	// another root-containers's begin/end block; this prevents the inner
	// root-container being clipped to the outer
	// push()
	ctx.ClipStack = append(ctx.ClipStack, unclippedRect)
}

func (ctx *Context) EndRootContainer() {
	// push tail 'goto' jump command and set head 'skip' command. the final steps
	// on initing these are done in mu_end()
	cnt := ctx.GetCurrentContainer()
	cnt.TailIdx = ctx.pushJump(-1)
	ctx.commandList[cnt.HeadIdx].jump.dstIdx = len(ctx.commandList) //- 1
	// pop base clip rect and container
	ctx.PopClipRect()
	ctx.PopContainer()
}

func (ctx *Context) BeginWindowEx(title string, rect image.Rectangle, opt Option) int {
	id := ctx.GetID([]byte(title))
	cnt := ctx.getContainer(id, opt)
	if cnt == nil || !cnt.Open {
		return 0
	}
	// push()
	ctx.IDStack = append(ctx.IDStack, id)

	if cnt.Rect.Dx() == 0 {
		cnt.Rect = rect
	}
	ctx.BeginRootContainer(cnt)
	body := cnt.Rect
	rect = body

	// draw frame
	if (^opt & OptNoFrame) != 0 {
		ctx.DrawFrame(ctx, rect, ColorWindowBG)
	}

	// do title bar
	if (^opt & OptNoTitle) != 0 {
		tr := rect
		tr.Max.Y = tr.Min.Y + ctx.Style.TitleHeight
		ctx.DrawFrame(ctx, tr, ColorTitleBG)

		// do title text
		if (^opt & OptNoTitle) != 0 {
			id := ctx.GetID([]byte("!title"))
			ctx.updateControl(id, tr, opt)
			ctx.DrawControlText(title, tr, ColorTitleText, opt)
			if id == ctx.Focus && ctx.mouseDown == mouseLeft {
				cnt.Rect = cnt.Rect.Add(ctx.mouseDelta)
			}
			body.Min.Y += tr.Dy()
		}

		// do `close` button
		if (^opt & OptNoClose) != 0 {
			id := ctx.GetID([]byte("!close"))
			r := image.Rect(tr.Max.X-tr.Dy(), tr.Min.Y, tr.Max.X, tr.Max.Y)
			tr.Max.X -= r.Dx()
			ctx.DrawIcon(IconClose, r, ctx.Style.Colors[ColorTitleText])
			ctx.updateControl(id, r, opt)
			if ctx.mousePressed == mouseLeft && id == ctx.Focus {
				cnt.Open = false
			}
		}
	}

	ctx.PushContainerBody(cnt, body, opt)

	// do `resize` handle
	if (^opt & OptNoResize) != 0 {
		sz := ctx.Style.TitleHeight
		id := ctx.GetID([]byte("!resize"))
		r := image.Rect(rect.Max.X-sz, rect.Max.Y-sz, rect.Max.X, rect.Max.Y)
		ctx.updateControl(id, r, opt)
		if id == ctx.Focus && ctx.mouseDown == mouseLeft {
			cnt.Rect.Max.X = cnt.Rect.Min.X + max(96, cnt.Rect.Dx()+ctx.mouseDelta.X)
			cnt.Rect.Max.Y = cnt.Rect.Min.Y + max(64, cnt.Rect.Dy()+ctx.mouseDelta.Y)
		}
	}

	// resize to content size
	if (opt & OptAutoSize) != 0 {
		r := ctx.GetLayout().Body
		cnt.Rect.Max.X = cnt.Rect.Min.X + cnt.ContentSize.X + (cnt.Rect.Dx() - r.Dx())
		cnt.Rect.Max.Y = cnt.Rect.Min.Y + cnt.ContentSize.Y + (cnt.Rect.Dy() - r.Dy())
	}

	// close if this is a popup window and elsewhere was clicked
	if (opt&OptPopup) != 0 && ctx.mousePressed != 0 && ctx.HoverRoot != cnt {
		cnt.Open = false
	}

	ctx.PushClipRect(cnt.Body)
	return ResActive
}

func (ctx *Context) EndWindow() {
	ctx.PopClipRect()
	ctx.EndRootContainer()
}

func (ctx *Context) OpenPopup(name string) {
	cnt := ctx.GetContainer(name)
	// set as hover root so popup isn't closed in begin_window_ex()
	ctx.NextHoverRoot = cnt
	ctx.HoverRoot = ctx.NextHoverRoot
	// position at mouse cursor, open and bring-to-front
	cnt.Rect = image.Rect(ctx.mousePos.X, ctx.mousePos.Y, ctx.mousePos.X+1, ctx.mousePos.Y+1)
	cnt.Open = true
	ctx.BringToFront(cnt)
}

func (ctx *Context) BeginPopup(name string) int {
	opt := OptPopup | OptAutoSize | OptNoResize |
		OptNoScroll | OptNoTitle | OptClosed
	return ctx.BeginWindowEx(name, image.Rectangle{}, opt)
}

func (ctx *Context) EndPopup() {
	ctx.EndWindow()
}

func (ctx *Context) BeginPanelEx(name string, opt Option) {
	var cnt *Container
	ctx.PushID([]byte(name))
	cnt = ctx.getContainer(ctx.LastID, opt)
	cnt.Rect = ctx.LayoutNext()
	if (^opt & OptNoFrame) != 0 {
		ctx.DrawFrame(ctx, cnt.Rect, ColorPanelBG)
	}
	// push()
	ctx.ContainerStack = append(ctx.ContainerStack, cnt)
	ctx.PushContainerBody(cnt, cnt.Rect, opt)
	ctx.PushClipRect(cnt.Body)
}

func (ctx *Context) EndPanel() {
	ctx.PopClipRect()
	ctx.PopContainer()
}
