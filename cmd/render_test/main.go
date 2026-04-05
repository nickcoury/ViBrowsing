package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// This is a TEST FILE to verify the rendering pipeline works
// without any network dependency

type Game struct {
	frameCount int
	testImg    *ebiten.Image
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1024, 768
}

func (g *Game) Update() error {
	g.frameCount++
	if g.frameCount == 1 {
		println("FIRST UPDATE")

		// Create a test image from a real PNG file to verify image loading works
		// First, let's create a simple test image programmatically
		rgba := image.NewRGBA(image.Rect(0, 0, 400, 300))

		// Fill with a gradient
		for y := 0; y < 300; y++ {
			for x := 0; x < 400; x++ {
				r := uint8((x * 255) / 400)
				b := uint8((y * 255) / 300)
				rgba.SetRGBA(x, y, color.RGBA{R: r, G: 100, B: b, A: 255})
			}
		}

		// Draw some text markers
		for y := 140; y < 160; y++ {
			for x := 10; x < 200; x++ {
				rgba.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}

		// Convert to ebiten image
		g.testImg = ebiten.NewImageFromImage(rgba)

		// Also save the RGBA to a PNG file for verification
		f, err := os.Create("/tmp/test_canvas_output.png")
		if err == nil {
			png.Encode(f, rgba)
			f.Close()
			println("Saved test image to /tmp/test_canvas_output.png")
		}

		println(fmt.Sprintf("Test image created: %dx%d", g.testImg.Bounds().Dx(), g.testImg.Bounds().Dy()))
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.frameCount == 1 {
		println("FIRST DRAW")
	}

	// Fill background white
	screen.Fill(color.White)

	// Draw nav bar
	navBar := ebiten.NewImage(1024, 40)
	navBar.Fill(color.RGBA{R: 50, G: 50, B: 60, A: 255})
	navGeoM := ebiten.GeoM{}
	navGeoM.Translate(0, 0)
	screen.DrawImage(navBar, &ebiten.DrawImageOptions{GeoM: navGeoM})

	// Draw back button
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

	// Draw text
	ebitenutil.DebugPrintAt(screen, "ViBrowsing Rendering Test", 10, 60)
	ebitenutil.DebugPrintAt(screen, "Nav bar (dark), Back button (blue), URL bar (white), Go button (blue)", 10, 85)

	// Draw the test image that we created from pixels
	if g.testImg != nil {
		testGeoM := ebiten.GeoM{}
		testGeoM.Translate(100, 120)
		screen.DrawImage(g.testImg, &ebiten.DrawImageOptions{GeoM: testGeoM})
		ebitenutil.DebugPrintAt(screen, "TEST IMAGE from pixel data at (100,120)", 105, 125)
	}

	// Draw a direct red rectangle (no image loading)
	redRect := ebiten.NewImage(200, 100)
	redRect.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	redGeoM := ebiten.GeoM{}
	redGeoM.Translate(550, 120)
	screen.DrawImage(redRect, &ebiten.DrawImageOptions{GeoM: redGeoM})
	ebitenutil.DebugPrintAt(screen, "DIRECT RED RECT (no image loading) at (550,120)", 555, 125)

	// Debug info
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, 700)
}

func main() {
	ebiten.SetWindowTitle("ViBrowsing Rendering Test")
	ebiten.SetWindowSize(1024, 768)

	game := &Game{}

	fmt.Println("Starting rendering test...")
	if err := ebiten.RunGame(game); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
