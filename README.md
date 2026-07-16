# glfw

This package provides a **very** thin 
[purego](https://github.com/ebitengine/purego) Go binding for the
[GLFW 3.4](https://www.glfw.org/) library. GLFW provides a simple API for
creating windows, contexts and surfaces, receiving input, and events, with
support for Windows, macOS, Wayland and X11.

## Install

```bash
go get github.com/jdpalmer/glfw
```

Requires **Go 1.22+**.

## System Dependencies

You will need to install the GLFW 3.4 shared library such that your
Operating System can load it (i.e., `libglfw.3.dylib`,
`libglfw.so.3`, or `glfw3.dll`):

| Platform | Typical install |
|----------|-----------------|
| macOS | `brew install glfw` |
| Ubuntu/Debian | `sudo apt install libglfw3` |
| Windows | Provide `glfw3.dll` on `PATH` or beside your binary |

## Example

```go
package main

import (
	"fmt"

	"github.com/jdpalmer/glfw"
)

func main() {
	if err := glfw.Init(); err != nil {
		fmt.Println(err)
		return
	}
	defer glfw.Terminate()

	window := glfw.CreateWindow(640, 480, "GLFW Test", nil, nil)
	if window == nil {
		fmt.Println("Failed to create window")
		return
	}
	defer window.Destroy()

	window.MakeContextCurrent()
	for !window.ShouldClose() {
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
```

## Documentation

Please use the very thorough 
[GLFW Documentation](https://www.glfw.org/documentation) as a reference. 
Documentation specific to the Go binding is also available via `go doc`.

## Threading

Most GLFW calls must run on the main thread. `glfw.Init` calls
`runtime.LockOSThread` on the caller.  You must keep window creation and
event polling on that thread.

## License

MIT — see [LICENSE](LICENSE). GLFW itself is licensed separately (zlib).
