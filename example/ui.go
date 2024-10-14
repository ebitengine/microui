// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ebitengine/microui"
)

func (g *Game) writeLog(text string) {
	if len(g.logBuf) > 0 {
		g.logBuf += "\n"
	}
	g.logBuf += text
	g.logUpdated = true
}

func (g *Game) testWindow() {
	g.ctx.Window("Demo Window", image.Rect(40, 40, 340, 490), func(res microui.Response, layout microui.Layout) {
		// window info
		if g.ctx.Header("Window Info") != 0 {
			g.ctx.SetLayoutRow([]int{54, -1}, 0)
			g.ctx.Label("Position:")
			g.ctx.Label(fmt.Sprintf("%d, %d", layout.Rect.Min.X, layout.Rect.Min.Y))
			g.ctx.Label("Size:")
			g.ctx.Label(fmt.Sprintf("%d, %d", layout.Rect.Dx(), layout.Rect.Dy()))
		}

		// labels + buttons
		if g.ctx.HeaderEx("Test Buttons", microui.OptExpanded) != 0 {
			g.ctx.SetLayoutRow([]int{100, -110, -1}, 0)
			g.ctx.Label("Test buttons 1:")
			if g.ctx.Button("Button 1") != 0 {
				g.writeLog("Pressed button 1")
			}
			if g.ctx.Button("Button 2") != 0 {
				g.writeLog("Pressed button 2")
			}
			g.ctx.Label("Test buttons 2:")
			if g.ctx.Button("Button 3") != 0 {
				g.writeLog("Pressed button 3")
			}
			if g.ctx.Button("Popup") != 0 {
				g.ctx.OpenPopup("Test Popup")
			}
			g.ctx.Popup("Test Popup", func(res microui.Response, layout microui.Layout) {
				g.ctx.Button("Hello")
				g.ctx.Button("World")
			})
		}

		// tree
		if g.ctx.HeaderEx("Tree and Text", microui.OptExpanded) != 0 {
			g.ctx.SetLayoutRow([]int{140, -1}, 0)
			g.ctx.LayoutColumn(func() {
				g.ctx.TreeNode("Test 1", func(res microui.Response) {
					g.ctx.TreeNode("Test 1a", func(res microui.Response) {
						g.ctx.Label("Hello")
						g.ctx.Label("World")
					})
					g.ctx.TreeNode("Test 1b", func(res microui.Response) {
						if g.ctx.Button("Button 1") != 0 {
							g.writeLog("Pressed button 1")
						}
						if g.ctx.Button("Button 2") != 0 {
							g.writeLog("Pressed button 2")
						}
					})
				})
				g.ctx.TreeNode("Test 2", func(res microui.Response) {
					g.ctx.SetLayoutRow([]int{54, 54}, 0)
					if g.ctx.Button("Button 3") != 0 {
						g.writeLog("Pressed button 3")
					}
					if g.ctx.Button("Button 4") != 0 {
						g.writeLog("Pressed button 4")
					}
					if g.ctx.Button("Button 5") != 0 {
						g.writeLog("Pressed button 5")
					}
					if g.ctx.Button("Button 6") != 0 {
						g.writeLog("Pressed button 6")
					}
				})
				g.ctx.TreeNode("Test 3", func(res microui.Response) {
					g.ctx.Checkbox("Checkbox 1", &g.checks[0])
					g.ctx.Checkbox("Checkbox 2", &g.checks[1])
					g.ctx.Checkbox("Checkbox 3", &g.checks[2])
				})
			})

			g.ctx.Text("Lorem ipsum dolor sit amet, consectetur adipiscing " +
				"elit. Maecenas lacinia, sem eu lacinia molestie, mi risus faucibus " +
				"ipsum, eu varius magna felis a nulla.")
		}

		// background color sliders
		if g.ctx.HeaderEx("Background Color", microui.OptExpanded) != 0 {
			g.ctx.SetLayoutRow([]int{-78, -1}, 74)
			// sliders
			g.ctx.LayoutColumn(func() {
				g.ctx.SetLayoutRow([]int{46, -1}, 0)
				g.ctx.Label("Red:")
				g.ctx.Slider(&g.bg[0], 0, 255, 1, 0)
				g.ctx.Label("Green:")
				g.ctx.Slider(&g.bg[1], 0, 255, 1, 0)
				g.ctx.Label("Blue:")
				g.ctx.Slider(&g.bg[2], 0, 255, 1, 0)
			})
			// color preview
			g.ctx.Control(0, 0, func(r image.Rectangle) microui.Response {
				g.ctx.DrawControl(func(screen *ebiten.Image) {
					vector.DrawFilledRect(
						screen,
						float32(r.Min.X),
						float32(r.Min.Y),
						float32(r.Dx()),
						float32(r.Dy()),
						color.RGBA{byte(g.bg[0]), byte(g.bg[1]), byte(g.bg[2]), 255},
						false)
					txt := fmt.Sprintf("#%02X%02X%02X", int(g.bg[0]), int(g.bg[1]), int(g.bg[2]))
					op := &text.DrawOptions{}
					op.GeoM.Translate(float64((r.Min.X+r.Max.X)/2), float64((r.Min.Y+r.Max.Y)/2))
					op.PrimaryAlign = text.AlignCenter
					op.SecondaryAlign = text.AlignCenter
					microui.DrawText(screen, txt, op)
				})
				return 0
			})
		}

		// Number
		if g.ctx.HeaderEx("Number", microui.OptExpanded) != 0 {
			g.ctx.SetLayoutRow([]int{-1}, 0)
			g.ctx.Number(&g.num1, 0.1, 2)
			g.ctx.Slider(&g.num2, 0, 10, 0.1, 2)
		}
	})
}

func (g *Game) logWindow() {
	g.ctx.Window("Log Window", image.Rect(350, 40, 650, 490), func(res microui.Response, layout microui.Layout) {
		// output text panel
		g.ctx.SetLayoutRow([]int{-1}, -25)
		g.ctx.Panel("Log Output", func(layout microui.Layout) {
			g.ctx.SetLayoutRow([]int{-1}, -1)
			g.ctx.Text(g.logBuf)
			if g.logUpdated {
				g.ctx.SetScroll(image.Pt(layout.Scroll.X, layout.ContentSize.Y))
				g.logUpdated = false
			}
		})

		// input textbox + submit button
		var submitted bool
		g.ctx.SetLayoutRow([]int{-70, -1}, 0)
		if g.ctx.TextBox(&g.logSubmitBuf)&microui.ResponseSubmit != 0 {
			g.ctx.SetFocus(g.ctx.LastID)
			submitted = true
		}
		if g.ctx.Button("Submit") != 0 {
			submitted = true
		}
		if submitted {
			g.writeLog(g.logSubmitBuf)
			g.logSubmitBuf = ""
		}
	})
}
