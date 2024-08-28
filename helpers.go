// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package microui

import (
	"image"
	"sort"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func expect(x bool) {
	if !x {
		panic("expect() failed")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func clamp(x, a, b int) int {
	return min(b, max(a, x))
}

func clampF(x, a, b float32) float32 {
	return minF(b, maxF(a, x))
}

func hash(hash *ID, data []byte) {
	for i := 0; i < len(data); i++ {
		*hash = (*hash ^ ID(data[i])) * 16777619
	}
}

func PtrToBytes(ptr unsafe.Pointer) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&ptr)), unsafe.Sizeof(ptr))
}

// GetID returns a hash value based on the data and the last ID on the stack.
func (ctx *Context) GetID(data []byte) ID {
	const (
		hashInitial = 2166136261 // 32bit fnv-1a hash
	)

	idx := len(ctx.IDStack)
	var res ID
	if idx > 0 {
		res = ctx.IDStack[len(ctx.IDStack)-1]
	} else {
		res = hashInitial
	}
	hash(&res, data)
	ctx.LastID = res
	return res
}

func (ctx *Context) PushID(data []byte) {
	// push()
	ctx.IDStack = append(ctx.IDStack, ctx.GetID(data))
}

func (ctx *Context) PopID() {
	expect(len(ctx.IDStack) > 0)
	ctx.IDStack = ctx.IDStack[:len(ctx.IDStack)-1]
}

func (ctx *Context) PushClipRect(rect image.Rectangle) {
	last := ctx.GetClipRect()
	// push()
	ctx.ClipStack = append(ctx.ClipStack, rect.Intersect(last))
}

func (ctx *Context) PopClipRect() {
	expect(len(ctx.ClipStack) > 0)
	ctx.ClipStack = ctx.ClipStack[:len(ctx.ClipStack)-1]
}

func (ctx *Context) GetClipRect() image.Rectangle {
	expect(len(ctx.ClipStack) > 0)
	return ctx.ClipStack[len(ctx.ClipStack)-1]
}

func (ctx *Context) CheckClip(r image.Rectangle) int {
	cr := ctx.GetClipRect()
	if !r.Overlaps(cr) {
		return ClipAll
	}
	if r.In(cr) {
		return 0
	}
	return ClipPart
}

func (ctx *Context) GetLayout() *Layout {
	expect(len(ctx.LayoutStack) > 0)
	return &ctx.LayoutStack[len(ctx.LayoutStack)-1]
}

func (ctx *Context) PopContainer() {
	cnt := ctx.GetCurrentContainer()
	layout := ctx.GetLayout()
	cnt.ContentSize.X = layout.Max.X - layout.Body.Min.X
	cnt.ContentSize.Y = layout.Max.Y - layout.Body.Min.Y
	// pop container, layout and id
	// pop()
	expect(len(ctx.ContainerStack) > 0) // TODO: no expect in original impl
	ctx.ContainerStack = ctx.ContainerStack[:len(ctx.ContainerStack)-1]
	// pop()
	expect(len(ctx.LayoutStack) > 0) // TODO: no expect in original impl
	ctx.LayoutStack = ctx.LayoutStack[:len(ctx.LayoutStack)-1]
	ctx.PopID()
}

func (ctx *Context) GetCurrentContainer() *Container {
	expect(len(ctx.ContainerStack) > 0)
	return ctx.ContainerStack[len(ctx.ContainerStack)-1]
}

func (ctx *Context) getContainer(id ID, opt Option) *Container {
	// try to get existing container from pool
	idx := ctx.poolGet(ctx.containerPool[:], id)
	if idx >= 0 {
		if ctx.containers[idx].Open || (^opt&OptClosed) != 0 {
			ctx.poolUpdate(ctx.containerPool[:], idx)
		}
		return &ctx.containers[idx]
	}
	if (opt & OptClosed) != 0 {
		return nil
	}
	// container not found in pool: init new container
	idx = ctx.poolInit(ctx.containerPool[:], id)
	cnt := &ctx.containers[idx]
	*cnt = Container{}
	cnt.HeadIdx = -1
	cnt.TailIdx = -1
	cnt.Open = true
	ctx.BringToFront(cnt)
	return cnt
}

func (ctx *Context) GetContainer(name string) *Container {
	id := ctx.GetID([]byte(name))
	return ctx.getContainer(id, 0)
}

func (ctx *Context) BringToFront(cnt *Container) {
	ctx.LastZindex++
	cnt.Zindex = ctx.LastZindex
}

func (ctx *Context) SetFocus(id ID) {
	ctx.Focus = id
	ctx.UpdatedFocus = true
}

func (ctx *Context) Begin() {
	ctx.updateInput()

	if ctx.TextWidth == nil {
		ctx.TextWidth = func(font Font, str string) int {
			return int(text.Advance(str, face))
		}
	}
	if ctx.TextHeight == nil {
		ctx.TextHeight = func(font Font) int {
			return 14
		}
	}

	ctx.commandList = ctx.commandList[:0]
	ctx.RootList = ctx.RootList[:0]
	ctx.ScrollTarget = nil
	ctx.HoverRoot = ctx.NextHoverRoot
	ctx.NextHoverRoot = nil
	ctx.mouseDelta.X = ctx.mousePos.X - ctx.lastMousePos.X
	ctx.mouseDelta.Y = ctx.mousePos.Y - ctx.lastMousePos.Y
	ctx.tick++
}

func (ctx *Context) End() {
	// check stacks
	expect(len(ctx.ContainerStack) == 0)
	expect(len(ctx.ClipStack) == 0)
	expect(len(ctx.IDStack) == 0)
	expect(len(ctx.LayoutStack) == 0)

	// handle scroll input
	if ctx.ScrollTarget != nil {
		ctx.ScrollTarget.Scroll.X += ctx.scrollDelta.X
		ctx.ScrollTarget.Scroll.Y += ctx.scrollDelta.Y
	}

	// unset focus if focus id was not touched this frame
	if !ctx.UpdatedFocus {
		ctx.Focus = 0
	}
	ctx.UpdatedFocus = false

	// bring hover root to front if mouse was pressed
	if ctx.mousePressed != 0 && ctx.NextHoverRoot != nil &&
		ctx.NextHoverRoot.Zindex < ctx.LastZindex &&
		ctx.NextHoverRoot.Zindex >= 0 {
		ctx.BringToFront(ctx.NextHoverRoot)
	}

	// reset input state
	ctx.keyPressed = 0
	ctx.textInput = nil
	ctx.mousePressed = 0
	ctx.scrollDelta = image.Pt(0, 0)
	ctx.lastMousePos = ctx.mousePos

	// sort root containers by zindex
	sort.SliceStable(ctx.RootList, func(i, j int) bool {
		return ctx.RootList[i].Zindex < ctx.RootList[j].Zindex
	})

	// set root container jump commands
	for i := 0; i < len(ctx.RootList); i++ {
		cnt := ctx.RootList[i]
		// if this is the first container then make the first command jump to it.
		// otherwise set the previous container's tail to jump to this one
		if i == 0 {
			cmd := ctx.commandList[0]
			expect(cmd.typ == commandJump)
			cmd.jump.dstIdx = cnt.HeadIdx + 1
			expect(cmd.jump.dstIdx < commandListSize)
		} else {
			prev := ctx.RootList[i-1]
			ctx.commandList[prev.TailIdx].jump.dstIdx = cnt.HeadIdx + 1
		}
		// make the last container's tail jump to the end of command list
		if i == len(ctx.RootList)-1 {
			ctx.commandList[cnt.TailIdx].jump.dstIdx = len(ctx.commandList)
		}
	}
}
