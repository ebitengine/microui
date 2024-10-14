// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ebitengine/microui"
)

type Game struct {
	ctx *microui.Context

	logBuf       string
	logSubmitBuf string
	logUpdated   bool
	bg           [3]float64
	checks       [3]bool
	num1         float64
	num2         float64
}

func New() *Game {
	return &Game{
		ctx:    microui.NewContext(),
		bg:     [3]float64{90, 95, 100},
		checks: [3]bool{true, false, true},
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	g.ctx.Update(func() {
		g.testWindow()
		g.logWindow()
	})
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.ctx.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1280, 960
}

func main() {
	ebiten.SetWindowTitle("Ebitengine Microui Demo")
	ebiten.SetWindowSize(1280, 960)
	if err := ebiten.RunGame(New()); err != nil {
		log.Fatal("err: ", err)
	}
}
