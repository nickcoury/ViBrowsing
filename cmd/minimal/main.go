package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	frameCount int
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1024, 768
}

func (g *Game) Update() error {
	g.frameCount++
	if g.frameCount == 1 {
		println("FIRST UPDATE - game loop started")
	}
	if g.frameCount%60 == 0 {
		println(fmt.Sprintf("Update frame %d", g.frameCount))
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.frameCount == 1 {
		println("FIRST DRAW")
	}

	// Fill background white
	screen.Fill(color.White)

	// Draw nav bar background
	navBar := ebiten.NewImage(1024, 40)
	navBar.Fill(color.RGBA{R: 50, G: 50, B: 60, A: 255})
	screen.DrawImage(navBar, &ebiten.DrawImageOptions{})

	// Draw back button (blue rect)
	backBtn := ebiten.NewImage(40, 24)
	backBtn.Fill(color.RGBA{R: 70, G: 120, B: 200, A: 255})
	backGeoM := ebiten.GeoM{}
	backGeoM.Translate(10, 8)
	screen.DrawImage(backBtn, &ebiten.DrawImageOptions{GeoM: backGeoM})

	// Draw forward button
	fwdBtn := ebiten.NewImage(40, 24)
	fwdBtn.Fill(color.RGBA{R: 70, G: 120, B: 200, A: 255})
	fwdGeoM := ebiten.GeoM{}
	fwdGeoM.Translate(55, 8)
	screen.DrawImage(fwdBtn, &ebiten.DrawImageOptions{GeoM: fwdGeoM})

	// Draw URL bar
	urlBar := ebiten.NewImage(900, 24)
	urlBar.Fill(color.White)
	urlGeoM := ebiten.GeoM{}
	urlGeoM.Translate(110, 8)
	screen.DrawImage(urlBar, &ebiten.DrawImageOptions{GeoM: urlGeoM})

	// Draw Go button
	goBtn := ebiten.NewImage(40, 24)
	goBtn.Fill(color.RGBA{R: 70, G: 120, B: 200, A: 255})
	goGeoM := ebiten.GeoM{}
	goGeoM.Translate(980, 8)
	screen.DrawImage(goBtn, &ebiten.DrawImageOptions{GeoM: goGeoM})

	// Draw "Go" text
	ebitenutil.DebugPrintAt(screen, "Go", 990, 16)

	// Draw welcome text in content area
	ebitenutil.DebugPrintAt(screen, "ViBrowsing Test - If you see colored buttons above, rendering works!", 10, 60)
	ebitenutil.DebugPrintAt(screen, "Nav bar (dark), Back button (blue), URL bar (white), Go button (blue)", 10, 85)

	// Draw a red rectangle in the content area to verify pixel drawing
	contentRect := ebiten.NewImage(200, 100)
	contentRect.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	redGeoM := ebiten.GeoM{}
	redGeoM.Translate(100, 120)
	screen.DrawImage(contentRect, &ebiten.DrawImageOptions{GeoM: redGeoM})
	ebitenutil.DebugPrintAt(screen, "RED RECTANGLE (200x100) should be visible at (100,120)", 105, 125)

	// Draw a green rectangle
	greenRect := ebiten.NewImage(150, 80)
	greenRect.Fill(color.RGBA{R: 0, G: 200, B: 0, A: 255})
	greenGeoM := ebiten.GeoM{}
	greenGeoM.Translate(400, 120)
	screen.DrawImage(greenRect, &ebiten.DrawImageOptions{GeoM: greenGeoM})
	ebitenutil.DebugPrintAt(screen, "GREEN RECT (150x80) at (400,120)", 405, 125)
}

func main() {
	ebiten.SetWindowTitle("ViBrowsing Minimal Test")
	ebiten.SetWindowSize(1024, 768)

	game := &Game{}

	fmt.Println("Starting minimal test...")
	if err := ebiten.RunGame(game); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
