package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math/cmplx"
	"math/rand"
	"sync"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	xmin, ymin, xmax, ymax = -2, -2, 2, 2
	width, height          = 1280, 1280
	maxIterations          = 256 * 3
	sizeReduction          = 4.0
)

var (
	size = 3.0
	xc   = -0.5
	yc   = 0.0
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var palette []color.RGBA

func init() {
	rand.Seed(time.Now().UnixNano())
	palette = make([]color.RGBA, maxIterations+1)
	for i := 0; i <= maxIterations; i++ {
		r := 0
		g := 0
		b := 0
		switch {
		case i <= 255:
			{
				r = i
			}
		case i <= 510 && i > 255:
			{
				r = 255
				g = i - 255
			}
		case i <= 765 && i > 510:
			{
				r = 255
				g = 255
				b = i - 510
			}
		}

		palette[i] = color.RGBA{
			uint8(r),
			uint8(g),
			uint8(b),
			255}
	}

	palette[maxIterations] = color.RGBA{0, 0, 0, 255}
}

func mand(c complex128) int {
	z := c
	for i := 0; i < maxIterations; i++ {
		if cmplx.Abs(z) > 2 {
			return i
		}

		z = cmplx.Pow(z, 2) + c
	}

	return maxIterations
}

type imgMsg struct {
	x     int
	y     int
	color color.RGBA
}

func newImg() (*image.RGBA, chan imgMsg) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	imgCh := make(chan imgMsg, 0)

	go func(imgCh chan imgMsg) {
		for {
			select {
			case msg, ok := <-imgCh:
				if !ok {
					break
				}
				img.Set(msg.x, msg.y, msg.color)
			}
		}
	}(imgCh)

	return img, imgCh
}

func saveImg(img image.Image) {
	buf := bytes.NewBuffer([]byte{})
	err := png.Encode(buf, img)
	must(err)
	err = ioutil.WriteFile("/tmp/img.png", buf.Bytes(), 0644)
	must(err)
}

func tranlate(x, inMin, inMax, outMin, outMax float64) float64 {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}

func renderImage(newX, newY, newSize float64) {
	log.Println("Running")

	size = newSize
	pixelRatioW := size / width
	pixelRatioH := size / height
	xCoord := tranlate(newX, 0, width, 0, width*pixelRatioW) * sizeReduction
	yCoord := tranlate(newY, 0, width, 0, width*pixelRatioH) * sizeReduction
	xc += xCoord
	yc += yCoord

	log.Printf("Coordinates input: %f,%f", newX, newY)
	log.Printf("Mapped input: %f,%f", xCoord, yCoord)
	log.Printf("New coords x = %f, y = %f, size = %f", xc, yc, size)

	img, imgCh := newImg()

	var wg sync.WaitGroup

	wg.Add(height)
	for px := 0; px < height; px++ {
		go func(px int) {
			defer wg.Done()
			for py := 0; py < width; py++ {
				x0 := xc - size/2 + float64(px)*pixelRatioW
				y0 := yc - size/2 + float64(py)*pixelRatioH
				coord := complex(x0, y0)
				f := mand(coord)
				clr := palette[int(f)]
				imgCh <- imgMsg{x: px, y: py, color: clr}
			}
		}(px)
	}

	wg.Wait()
	close(imgCh)

	saveImg(img)

	log.Println("Done")
}

func reset() {
	xc = -0.5
	yc = 0.0
	size = 3.0
	renderImage(0.0, 0.0, size)
}

func handleClick(event *sdl.MouseButtonEvent) {
	if event.Button == sdl.BUTTON_LEFT && event.Type == sdl.MOUSEBUTTONUP {
		// size = size / 2
		newX := float64(event.X) - width/2
		newY := float64(event.Y) - height/2
		renderImage(newX, newY, size/sizeReduction)
	}
	if event.Button == sdl.BUTTON_RIGHT && event.Type == sdl.MOUSEBUTTONUP {
		reset()
	}

}

func loadImage(renderer *sdl.Renderer) {
	img, err := img.Load("/tmp/img.png")
	must(err)
	defer img.Free()

	texture, err := renderer.CreateTextureFromSurface(img)
	must(err)
	defer texture.Destroy()

	src := sdl.Rect{0, 0, width, height}
	renderer.Clear()
	renderer.Copy(texture, &src, &src)
	renderer.Present()
}

func main() {
	must(sdl.Init(sdl.INIT_EVERYTHING))
	defer sdl.Quit()

	windowHeight := int32(height)
	windowWidth := int32(width)

	window, renderer, err := sdl.CreateWindowAndRenderer(windowWidth, windowHeight, sdl.WINDOW_SHOWN)
	must(err)
	defer window.Destroy()

	reset()
	loadImage(renderer)

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.MouseButtonEvent:
				e, _ := event.(*sdl.MouseButtonEvent)
				handleClick(e)
				loadImage(renderer)
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}
