//go:build ignore

package main

import (
    "fmt"
    "image/png"
    "os"

    "github.com/nickcoury/ViBrowsing/internal/css"
    "github.com/nickcoury/ViBrowsing/internal/render"
)

func main() {
    canvas := render.NewCanvas(200, 200)
    canvas.Clear()

    // Draw red rect
    red := css.ParseColor("red")
    r, g, b, a := red.RGBA()
    fmt.Printf("Red css.Color.RGBA() = %d,%d,%d,%d (16-bit)\n", r, g, b, a)
    canvas.FillRect(10, 10, 50, 50, red)

    // Draw blue rect
    blue := css.ParseColor("blue")
    canvas.FillRect(80, 80, 50, 50, blue)

    // Save
    canvas.SavePNG("/tmp/test_draw.png")

    // Read back
    f, _ := os.Open("/tmp/test_draw.png")
    img, _ := png.Decode(f)
    f.Close()

    // Check pixels
    check := func(x, y int) {
        r, g, b, a := img.At(x, y).RGBA()
        fmt.Printf("pixel(%d,%d): r=%d g=%d b=%d a=%d\n", x, y, r>>8, g>>8, b>>8, a>>8)
    }
    check(5, 5)     // white (background)
    check(30, 30)   // red
    check(105, 105) // blue
    check(0, 0)     // white
}
