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

**Note:** This package provides the bindings only. You must also have the 
native GLFW dynamic library installed on your system or bundled with your 
application.

## Platform-Specific Setup

### Windows

Download the pre-compiled binaries from [https://glfw.org/]. For distribution,
include the `glfw3.dll` in the same directory as your application binary.

## Linux (Ubuntu/Debian)

Install via the package manager:

`sudo apt install libglfw3`

You should also ensure your system has the necessary development headers for 
OpenGL, OpenGL ES, or Vulkan depending on your target graphics API.  You can
also install these with `apt`.

## MacOS

While Homebrew provides a `glfw` package, it is often unsuitable for
production applications due to a lack of `rpath` support and inability to
easily link with ANGLE (for OpenGL ES).

**Recommended:** Use the official binary distribution from 
[glfw.org](https://glfw.org). For bundled macOS applications (.app), use the 
following structure:

```text
YourApp.app/
  Contents/
    MacOS/YourApp
    Frameworks/
      libglfw.3.dylib
      libEGL.dylib
      libGLESv2.dylib
```

## Configuring rpath

To ensure your binary finds the dynamic libraries at runtime, add the rpath 
to your executable:

```bash
go build -o YourApp.app/Contents/MacOS/YourApp ./cmd/yourapp
install_name_tool -add_rpath '@executable_path/../Frameworks' \
  YourApp.app/Contents/MacOS/YourApp
```

If you are using CGO/external linking, you can bake this into the build 
process:

```bash
CGO_ENABLED=1 go build \
  -ldflags '-extldflags=-Wl,-rpath,@executable_path/../Frameworks' \
  -o YourApp.app/Contents/MacOS/YourApp ./cmd/yourapp
```

For rapid development, you may simply place the `.dylib` files in the same 
folder as your binary.

## Quick Start Example

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

## Threading Requirements

GLFW has strict threading requirements. Most calls **must** be executed on 
the main thread.

* `glfw.Init` automatically calls `runtime.LockOSThread` on the caller.
* Ensure that all window creation, context management, and event polling 
  occur on this locked main thread.

## Documentation

* API Reference: Use the official GLFW Documentation.
* Go Bindings: Use `go doc github.com/jdpalmer/glfw` for package-specific
  details.
    
## License

MIT — see [LICENSE](LICENSE). GLFW itself is licensed separately (zlib).
