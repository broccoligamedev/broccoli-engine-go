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
	// todo(ryan): why does sdl's init sometimes start hanging????
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	} else {
		log.Println("SDL initialized")
		defer sdl.Quit()
	}
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 16)
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
		for i := 0; i < 10; i++ {
			// TODO: implement alpha blending
			SetDrawColor(1.0-float32(i)*0.1, 0.0+float32(i)*0.1, 0.0+float32(i)*0.1, 1.0-float32(i)*0.1)
			DrawRectangle(100+float32(i*32), 100, 32, 32)
			DrawTriangle(100+float32(i*32), 200, 116+float32(i*32), 168, 132+float32(i*32), 200)
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
