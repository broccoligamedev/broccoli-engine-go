package main

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
	_ "image/png"
	"log"
	//"math"
	"time"
)

func check(e error) {
	if e != nil {
		log.Println(e.Error())
		panic(e)
	}
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func main() {
	// TODO(ryan): why does sdl's init sometimes start hanging????
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	} else {
		log.Println("SDL initialized")
		defer sdl.Quit()
	}
	// Make things not look like garbage
	// sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)
	// sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 16)
	window, err := sdl.CreateWindow(
		"OpenGL Test",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		600, 600,
		sdl.WINDOW_OPENGL)
	//sdl.GLSetSwapInterval(0)
	//vsync, _ := sdl.GLGetSwapInterval()
	//log.Println("Vsync", vsync)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	_, err = window.GLCreateContext()
	if err != nil {
		panic(err)
	}
	err = gl.Init()
	if err != nil {
		panic(err)
	}
	InitGraphics()
	currentFrame := 0
	for {
		beginTime := time.Now()
		for {
			event := sdl.PollEvent()
			if event == nil {
				break
			}
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		ClearScreen(0.5, 0.5, 0.8)
		limit := 16
		var xOffset float32 = 32
		var yOffset float32 = 32
		var ySeparator float32 = 8
		for i := 0; i < limit; i++ {
			SetDrawColor(
				0.0+float32(i)/float32(limit),
				1.0-float32(i)/float32(limit),
				0.0+float32(i)/float32(limit),
				1.0-float32(i)/float32(limit))
			DrawTriangle(
				xOffset+float32(i*32), yOffset+ySeparator,
				xOffset+float32(i*32)+16, yOffset+ySeparator-24,
				xOffset+float32(i*32)+32, yOffset+ySeparator)
			DrawRectangle(
				xOffset+float32(i*32),
				yOffset+2*ySeparator,
				32, 32)
			DrawPolygon(
				100, 100,
				120, 120,
				100, 140,
				80, 120,
			)
		}
		stopTime := time.Since(beginTime).Seconds() * 1000
		window.GLSwap()
		currentFrame++
		targetRate := float64(1000 / 60)
		if stopTime < targetRate {
			sleepTime := uint32(targetRate - stopTime)
			sdl.Delay(sleepTime)
		}
	}
}
