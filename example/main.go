package main

import (
	"fmt"

	"github.com/jdpalmer/glfw"
)

func main() {
	if err := glfw.Init(); err != nil {
		fmt.Println("Failed to initialize GLFW:", err)
		return
	}
	defer glfw.Terminate()

	glfw.DefaultWindowHints()

	window := glfw.CreateWindow(640, 480, "GLFW Test", nil, nil)
	if window == nil {
		fmt.Println("Failed to create window")
		return
	}
	defer window.Destroy()

	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	fmt.Println("Window opened. Close it to exit.")

	for !window.ShouldClose() {
		window.SwapBuffers()
		glfw.PollEvents()
	}

	fmt.Println("Done.")
}
