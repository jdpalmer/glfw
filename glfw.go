// glfw is a purego binding of GLFW 3.4.
//
// # Threading
//
// Most GLFW calls must run on the main thread. Init locks the calling OS thread
// via runtime.LockOSThread; keep event polling and window creation on that thread.
//
// # Handles and GC
//
// Window, Monitor, and Cursor are Go-owned wrappers around opaque C handles.
// Callbacks receive the same Go *Window / *Monitor values registered at create
// time. User data from SetUserPointer is stored on the Go wrapper as any; it is
// not written into GLFW's C user-pointer slot, so the GC can track it safely.
//
// Callbacks use one stable C trampoline per callback kind. Set*Callback replaces
// the Go func and returns the previous one. Destroy and Terminate clear retained
// callbacks and handle maps.
package glfw

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Version constants.
const (
	VersionMajor    = 3
	VersionMinor    = 4
	VersionRevision = 0
)

// Action corresponds to a key or button action.
type Action int32

const (
	Release Action = 0
	Press   Action = 1
	Repeat  Action = 2
)

// JoystickHatState corresponds to joystick hat states.
type JoystickHatState int32

const (
	HatCentered  JoystickHatState = 0
	HatUp        JoystickHatState = 1
	HatRight     JoystickHatState = 2
	HatDown      JoystickHatState = 4
	HatLeft      JoystickHatState = 8
	HatRightUp   JoystickHatState = HatRight | HatUp
	HatRightDown JoystickHatState = HatRight | HatDown
	HatLeftUp    JoystickHatState = HatLeft | HatUp
	HatLeftDown  JoystickHatState = HatLeft | HatDown
)

// Key corresponds to a keyboard key.
type Key int32

const (
	KeyUnknown Key = -1

	/* Printable keys */
	KeySpace        Key = 32
	KeyApostrophe   Key = 39 // '
	KeyComma        Key = 44 // ,
	KeyMinus        Key = 45 // -
	KeyPeriod       Key = 46 // .
	KeySlash        Key = 47 // /
	Key0            Key = 48
	Key1            Key = 49
	Key2            Key = 50
	Key3            Key = 51
	Key4            Key = 52
	Key5            Key = 53
	Key6            Key = 54
	Key7            Key = 55
	Key8            Key = 56
	Key9            Key = 57
	KeySemicolon    Key = 59 // ;
	KeyEqual        Key = 61 // =
	KeyA            Key = 65
	KeyB            Key = 66
	KeyC            Key = 67
	KeyD            Key = 68
	KeyE            Key = 69
	KeyF            Key = 70
	KeyG            Key = 71
	KeyH            Key = 72
	KeyI            Key = 73
	KeyJ            Key = 74
	KeyK            Key = 75
	KeyL            Key = 76
	KeyM            Key = 77
	KeyN            Key = 78
	KeyO            Key = 79
	KeyP            Key = 80
	KeyQ            Key = 81
	KeyR            Key = 82
	KeyS            Key = 83
	KeyT            Key = 84
	KeyU            Key = 85
	KeyV            Key = 86
	KeyW            Key = 87
	KeyX            Key = 88
	KeyY            Key = 89
	KeyZ            Key = 90
	KeyLeftBracket  Key = 91  // [
	KeyBackslash    Key = 92  // \
	KeyRightBracket Key = 93  // ]
	KeyGraveAccent  Key = 96  // `
	KeyWorld_1      Key = 161 // non-US #1
	KeyWorld_2      Key = 162 // non-US #2

	/* Function keys */
	KeyEscape        Key = 256
	KeyEnter         Key = 257
	KeyTab           Key = 258
	KeyBackspace     Key = 259
	KeyInsert        Key = 260
	KeyDelete        Key = 261
	KeyRight         Key = 262
	KeyLeft          Key = 263
	KeyDown          Key = 264
	KeyUp            Key = 265
	KeyPageUp        Key = 266
	KeyPageDown      Key = 267
	KeyHome          Key = 268
	KeyEnd           Key = 269
	KeyCapsLock      Key = 280
	KeyScrollLock    Key = 281
	KeyNumLock       Key = 282
	KeyPrintScreen   Key = 283
	KeyPause         Key = 284
	KeyF1            Key = 290
	KeyF2            Key = 291
	KeyF3            Key = 292
	KeyF4            Key = 293
	KeyF5            Key = 294
	KeyF6            Key = 295
	KeyF7            Key = 296
	KeyF8            Key = 297
	KeyF9            Key = 298
	KeyF10           Key = 299
	KeyF11           Key = 300
	KeyF12           Key = 301
	KeyF13           Key = 302
	KeyF14           Key = 303
	KeyF15           Key = 304
	KeyF16           Key = 305
	KeyF17           Key = 306
	KeyF18           Key = 307
	KeyF19           Key = 308
	KeyF20           Key = 309
	KeyF21           Key = 310
	KeyF22           Key = 311
	KeyF23           Key = 312
	KeyF24           Key = 313
	KeyF25           Key = 314
	KeyKP_0          Key = 320
	KeyKP_1          Key = 321
	KeyKP_2          Key = 322
	KeyKP_3          Key = 323
	KeyKP_4          Key = 324
	KeyKP_5          Key = 325
	KeyKP_6          Key = 326
	KeyKP_7          Key = 327
	KeyKP_8          Key = 328
	KeyKP_9          Key = 329
	KeyKP_Decimal    Key = 330
	KeyKP_Divide     Key = 331
	KeyKP_Multiply   Key = 332
	KeyKP_Subtract   Key = 333
	KeyKP_Add        Key = 334
	KeyKP_Enter      Key = 335
	KeyKP_Equal      Key = 336
	KeyLeft_Shift    Key = 340
	KeyLeft_Control  Key = 341
	KeyLeft_Alt      Key = 342
	KeyLeft_Super    Key = 343
	KeyRight_Shift   Key = 344
	KeyRight_Control Key = 345
	KeyRight_Alt     Key = 346
	KeyRight_Super   Key = 347
	KeyMenu          Key = 348
	KeyLast          Key = KeyMenu
)

// ModifierKey corresponds to a modifier key.
type ModifierKey int32

const (
	ModShift    ModifierKey = 0x0001
	ModControl  ModifierKey = 0x0002
	ModAlt      ModifierKey = 0x0004
	ModSuper    ModifierKey = 0x0008
	ModCapsLock ModifierKey = 0x0010
	ModNumLock  ModifierKey = 0x0020
)

// MouseButton corresponds to a mouse button.
type MouseButton int32

const (
	MouseButton1      MouseButton = 0
	MouseButton2      MouseButton = 1
	MouseButton3      MouseButton = 2
	MouseButton4      MouseButton = 3
	MouseButton5      MouseButton = 4
	MouseButton6      MouseButton = 5
	MouseButton7      MouseButton = 6
	MouseButton8      MouseButton = 7
	MouseButtonLast   MouseButton = MouseButton8
	MouseButtonLeft   MouseButton = MouseButton1
	MouseButtonRight  MouseButton = MouseButton2
	MouseButtonMiddle MouseButton = MouseButton3
)

// Joystick corresponds to a joystick ID.
type Joystick int32

const (
	Joystick1    Joystick = 0
	Joystick2    Joystick = 1
	Joystick3    Joystick = 2
	Joystick4    Joystick = 3
	Joystick5    Joystick = 4
	Joystick6    Joystick = 5
	Joystick7    Joystick = 6
	Joystick8    Joystick = 7
	Joystick9    Joystick = 8
	Joystick10   Joystick = 9
	Joystick11   Joystick = 10
	Joystick12   Joystick = 11
	Joystick13   Joystick = 12
	Joystick14   Joystick = 13
	Joystick15   Joystick = 14
	Joystick16   Joystick = 15
	JoystickLast Joystick = Joystick16
)

// GamepadButton corresponds to a gamepad button.
type GamepadButton int32

const (
	GamepadButtonA           GamepadButton = 0
	GamepadButtonB           GamepadButton = 1
	GamepadButtonX           GamepadButton = 2
	GamepadButtonY           GamepadButton = 3
	GamepadButtonLeftBumper  GamepadButton = 4
	GamepadButtonRightBumper GamepadButton = 5
	GamepadButtonBack        GamepadButton = 6
	GamepadButtonStart       GamepadButton = 7
	GamepadButtonGuide       GamepadButton = 8
	GamepadButtonLeftThumb   GamepadButton = 9
	GamepadButtonRightThumb  GamepadButton = 10
	GamepadButtonDpadUp      GamepadButton = 11
	GamepadButtonDpadRight   GamepadButton = 12
	GamepadButtonDpadDown    GamepadButton = 13
	GamepadButtonDpadLeft    GamepadButton = 14
	GamepadButtonLast        GamepadButton = GamepadButtonDpadLeft
	GamepadButtonCross       GamepadButton = GamepadButtonA
	GamepadButtonCircle      GamepadButton = GamepadButtonB
	GamepadButtonSquare      GamepadButton = GamepadButtonX
	GamepadButtonTriangle    GamepadButton = GamepadButtonY
)

// GamepadAxis corresponds to a gamepad axis.
type GamepadAxis int32

const (
	GamepadAxisLeftX        GamepadAxis = 0
	GamepadAxisLeftY        GamepadAxis = 1
	GamepadAxisRightX       GamepadAxis = 2
	GamepadAxisRightY       GamepadAxis = 3
	GamepadAxisLeftTrigger  GamepadAxis = 4
	GamepadAxisRightTrigger GamepadAxis = 5
	GamepadAxisLast         GamepadAxis = GamepadAxisRightTrigger
)

// ErrorCode corresponds to an error code.
type ErrorCode int32

const (
	NoError              ErrorCode = 0
	NotInitialized       ErrorCode = 0x00010001
	NoCurrentContext     ErrorCode = 0x00010002
	InvalidEnum          ErrorCode = 0x00010003
	InvalidValue         ErrorCode = 0x00010004
	OutOfMemory          ErrorCode = 0x00010005
	APIUnavailable       ErrorCode = 0x00010006
	VersionUnavailable   ErrorCode = 0x00010007
	PlatformError        ErrorCode = 0x00010008
	FormatUnavailable    ErrorCode = 0x00010009
	NoWindowContext      ErrorCode = 0x0001000A
	CursorUnavailable    ErrorCode = 0x0001000B
	FeatureUnavailable   ErrorCode = 0x0001000C
	FeatureUnimplemented ErrorCode = 0x0001000D
	PlatformUnavailable  ErrorCode = 0x0001000E
)

type Hint int32

const (
	Focused                Hint = 0x00020001
	Iconified              Hint = 0x00020002
	Resizable              Hint = 0x00020003
	Visible                Hint = 0x00020004
	Decorated              Hint = 0x00020005
	AutoIconify            Hint = 0x00020006
	Floating               Hint = 0x00020007
	Maximized              Hint = 0x00020008
	CenterCursor           Hint = 0x00020009
	TransparentFramebuffer Hint = 0x0002000A
	Hovered                Hint = 0x0002000B
	FocusOnShow            Hint = 0x0002000C
	MousePassthrough       Hint = 0x0002000D
	PositionX              Hint = 0x0002000E
	PositionY              Hint = 0x0002000F

	RedBits        Hint = 0x00021001
	GreenBits      Hint = 0x00021002
	BlueBits       Hint = 0x00021003
	AlphaBits      Hint = 0x00021004
	DepthBits      Hint = 0x00021005
	StencilBits    Hint = 0x00021006
	AccumRedBits   Hint = 0x00021007
	AccumGreenBits Hint = 0x00021008
	AccumBlueBits  Hint = 0x00021009
	AccumAlphaBits Hint = 0x0002100A
	AuxBuffers     Hint = 0x0002100B
	Stereo         Hint = 0x0002100C
	Samples        Hint = 0x0002100D
	SRGBCapable    Hint = 0x0002100E
	RefreshRate    Hint = 0x0002100F
	Doublebuffer   Hint = 0x00021010

	ClientAPI              Hint = 0x00022001
	ContextVersionMajor    Hint = 0x00022002
	ContextVersionMinor    Hint = 0x00022003
	ContextRevision        Hint = 0x00022004
	ContextRobustness      Hint = 0x00022005
	OpenGLForwardCompat    Hint = 0x00022006
	ContextDebug           Hint = 0x00022007
	OpenGLDebugContext     Hint = ContextDebug
	OpenGLProfile          Hint = 0x00022008
	ContextReleaseBehavior Hint = 0x00022009
	ContextNoError         Hint = 0x0002200A
	ContextCreationAPI     Hint = 0x0002200B
	ScaleToMonitor         Hint = 0x0002200C
	ScaleFramebuffer       Hint = 0x0002200D
	CocoaRetinaFramebuffer Hint = 0x00023001
	CocoaFrameName         Hint = 0x00023002
	CocoaGraphicsSwitching Hint = 0x00023003
	X11ClassName           Hint = 0x00024001
	X11InstanceName        Hint = 0x00024002
	Win32KeyboardMenu      Hint = 0x00025001
	Win32ShowDefault       Hint = 0x00025002
	WaylandAppID           Hint = 0x00026001

	JoystickHatButtons  Hint = 0x00050001
	AnglePlatformType   Hint = 0x00050002
	Platform            Hint = 0x00050003
	CocoaChdirResources Hint = 0x00051001
	CocoaMenubar        Hint = 0x00051002
	X11XcbVulkanSurface Hint = 0x00052001
	WaylandLibdecor     Hint = 0x00053001
)

// InputMode corresponds to an input mode.
type InputMode int32

const (
	CursorMode         InputMode = 0x00033001
	StickyKeys         InputMode = 0x00033002
	StickyMouseButtons InputMode = 0x00033003
	LockKeyMods        InputMode = 0x00033004
	RawMouseMotion     InputMode = 0x00033005
)

// Cursor mode values (use with InputMode CursorMode).
const (
	CursorNormal   int32 = 0x00034001
	CursorHidden   int32 = 0x00034002
	CursorDisabled int32 = 0x00034003
	CursorCaptured int32 = 0x00034004
)

// StandardCursor corresponds to a standard cursor shape.
type StandardCursor int32

const (
	ArrowCursor        StandardCursor = 0x00036001
	IbeamCursor        StandardCursor = 0x00036002
	CrosshairCursor    StandardCursor = 0x00036003
	PointingHandCursor StandardCursor = 0x00036004
	ResizeEWCursor     StandardCursor = 0x00036005
	ResizeNSCursor     StandardCursor = 0x00036006
	ResizeNWSECursor   StandardCursor = 0x00036007
	ResizeNESWCursor   StandardCursor = 0x00036008
	ResizeAllCursor    StandardCursor = 0x00036009
	NotAllowedCursor   StandardCursor = 0x0003600A
	HResizeCursor      StandardCursor = ResizeEWCursor
	VResizeCursor      StandardCursor = ResizeNSCursor
	HandCursor         StandardCursor = PointingHandCursor
)

// PeripheralEvent corresponds to a monitor or joystick connection event.
type PeripheralEvent int32

const (
	Connected    PeripheralEvent = 0x00040001
	Disconnected PeripheralEvent = 0x00040002
)

// PlatformID corresponds to a platform returned by GetPlatform.
type PlatformID int32

const (
	AnyPlatform     PlatformID = 0x00060000
	PlatformWin32   PlatformID = 0x00060001
	PlatformCocoa   PlatformID = 0x00060002
	PlatformWayland PlatformID = 0x00060003
	PlatformX11     PlatformID = 0x00060004
	PlatformNull    PlatformID = 0x00060005
)

const (
	True  int32 = 1
	False int32 = 0

	NoAPI       int32 = 0
	OpenGLAPI   int32 = 0x00030001
	OpenGLESAPI int32 = 0x00030002

	NoRobustness        int32 = 0
	NoResetNotification int32 = 0x00031001
	LoseContextOnReset  int32 = 0x00031002

	OpenGLAnyProfile    int32 = 0
	OpenGLCoreProfile   int32 = 0x00032001
	OpenGLCompatProfile int32 = 0x00032002

	AnyReleaseBehavior   int32 = 0
	ReleaseBehaviorFlush int32 = 0x00035001
	ReleaseBehaviorNone  int32 = 0x00035002

	NativeContextAPI int32 = 0x00036001
	EGLContextAPI    int32 = 0x00036002
	OSMesaContextAPI int32 = 0x00036003

	AnglePlatformTypeNone     int32 = 0x00037001
	AnglePlatformTypeOpenGL   int32 = 0x00037002
	AnglePlatformTypeOpenGLES int32 = 0x00037003
	AnglePlatformTypeD3D9     int32 = 0x00037004
	AnglePlatformTypeD3D11    int32 = 0x00037005
	AnglePlatformTypeVulkan   int32 = 0x00037007
	AnglePlatformTypeMetal    int32 = 0x00037008

	WaylandPreferLibdecor  int32 = 0x00038001
	WaylandDisableLibdecor int32 = 0x00038002

	AnyPosition int32 = -0x80000000
	DontCare    int32 = -1
)

type ErrorFunc func(errorCode ErrorCode, description string)
type WindowPosFunc func(window *Window, xpos int32, ypos int32)
type WindowSizeFunc func(window *Window, width int32, height int32)
type WindowCloseFunc func(window *Window)
type WindowRefreshFunc func(window *Window)
type WindowFocusFunc func(window *Window, focused int32)
type WindowIconifyFunc func(window *Window, iconified int32)
type WindowMaximizeFunc func(window *Window, maximized int32)
type FramebufferSizeFunc func(window *Window, width int32, height int32)
type WindowContentScaleFunc func(window *Window, xscale float32, yscale float32)
type MouseButtonFunc func(window *Window, button MouseButton, action Action, mods ModifierKey)
type CursorPosFunc func(window *Window, xpos float64, ypos float64)
type CursorEnterFunc func(window *Window, entered int32)
type ScrollFunc func(window *Window, xoffset float64, yoffset float64)
type KeyFunc func(window *Window, key Key, scancode int32, action Action, mods ModifierKey)
type CharFunc func(window *Window, codepoint uint32)
type CharModsFunc func(window *Window, codepoint uint32, mods ModifierKey)
type DropFunc func(window *Window, paths []string)
type MonitorFunc func(monitor *Monitor, event PeripheralEvent)
type JoystickFunc func(jid Joystick, event PeripheralEvent)

type VidMode struct {
	Width       int32
	Height      int32
	RedBits     int32
	GreenBits   int32
	BlueBits    int32
	RefreshRate int32
}

type GammaRamp struct {
	Red   []uint16
	Green []uint16
	Blue  []uint16
	Size  uint32
}

type gammaRampC struct {
	Red   *uint16
	Green *uint16
	Blue  *uint16
	Size  uint32
}

type Image struct {
	Width  int32
	Height int32
	Pixels []byte
}

type imageC struct {
	Width  int32
	Height int32
	Pixels *byte
}

type GamepadState struct {
	Buttons [15]byte
	Axes    [6]float32
}

// Monitor is a Go-owned wrapper around a GLFWmonitor*.
// Use SetUserPointer to attach Go values; they are not stored in C.
type Monitor struct {
	ptr  unsafe.Pointer
	user any
}

// Window is a Go-owned wrapper around a GLFWwindow*.
// Callbacks installed with Set*Callback receive this same *Window.
// Use SetUserPointer to attach Go values; they are not stored in C.
type Window struct {
	ptr  unsafe.Pointer
	user any

	posCb             WindowPosFunc
	sizeCb            WindowSizeFunc
	closeCb           WindowCloseFunc
	refreshCb         WindowRefreshFunc
	focusCb           WindowFocusFunc
	iconifyCb         WindowIconifyFunc
	maximizeCb        WindowMaximizeFunc
	framebufferSizeCb FramebufferSizeFunc
	contentScaleCb    WindowContentScaleFunc
	keyCb             KeyFunc
	charCb            CharFunc
	charModsCb        CharModsFunc
	mouseButtonCb     MouseButtonFunc
	cursorPosCb       CursorPosFunc
	cursorEnterCb     CursorEnterFunc
	scrollCb          ScrollFunc
	dropCb            DropFunc
}

// Cursor is a Go-owned wrapper around a GLFWcursor*.
type Cursor struct {
	ptr unsafe.Pointer
}

// Init loads the platform GLFW shared library, registers trampolines, locks the
// OS thread, and initializes GLFW. Call Terminate before exit.
func Init() error {
	//var lib uintptr
	//var err error

	lib, err := loadGLFW()

	if err != nil {
		return err
	}

	purego.RegisterLibFunc(&glfwTerminate, lib, "glfwTerminate")
	purego.RegisterLibFunc(&glfwInitHint, lib, "glfwInitHint")
	purego.RegisterLibFunc(&glfwGetVersion, lib, "glfwGetVersion")
	purego.RegisterLibFunc(&glfwGetVersionString, lib, "glfwGetVersionString")
	purego.RegisterLibFunc(&glfwGetError, lib, "glfwGetError")
	purego.RegisterLibFunc(&glfwSetErrorCallback, lib, "glfwSetErrorCallback")
	purego.RegisterLibFunc(&glfwGetPlatform, lib, "glfwGetPlatform")
	purego.RegisterLibFunc(&glfwPlatformSupported, lib, "glfwPlatformSupported")
	purego.RegisterLibFunc(&glfwGetMonitors, lib, "glfwGetMonitors")
	purego.RegisterLibFunc(&glfwGetPrimaryMonitor, lib, "glfwGetPrimaryMonitor")
	purego.RegisterLibFunc(&glfwGetMonitorPos, lib, "glfwGetMonitorPos")
	purego.RegisterLibFunc(&glfwGetMonitorWorkarea, lib, "glfwGetMonitorWorkarea")
	purego.RegisterLibFunc(&glfwGetMonitorPhysicalSize, lib, "glfwGetMonitorPhysicalSize")
	purego.RegisterLibFunc(&glfwGetMonitorContentScale, lib, "glfwGetMonitorContentScale")
	purego.RegisterLibFunc(&glfwGetMonitorName, lib, "glfwGetMonitorName")
	purego.RegisterLibFunc(&glfwSetMonitorCallback, lib, "glfwSetMonitorCallback")
	purego.RegisterLibFunc(&glfwGetVideoModes, lib, "glfwGetVideoModes")
	purego.RegisterLibFunc(&glfwGetVideoMode, lib, "glfwGetVideoMode")
	purego.RegisterLibFunc(&glfwSetGamma, lib, "glfwSetGamma")
	purego.RegisterLibFunc(&glfwGetGammaRampC, lib, "glfwGetGammaRamp")
	purego.RegisterLibFunc(&glfwSetGammaRamp, lib, "glfwSetGammaRamp")
	purego.RegisterLibFunc(&glfwDefaultWindowHints, lib, "glfwDefaultWindowHints")
	purego.RegisterLibFunc(&glfwWindowHint, lib, "glfwWindowHint")
	purego.RegisterLibFunc(&glfwWindowHintString, lib, "glfwWindowHintString")
	purego.RegisterLibFunc(&glfwCreateWindow, lib, "glfwCreateWindow")
	purego.RegisterLibFunc(&glfwDestroyWindow, lib, "glfwDestroyWindow")
	purego.RegisterLibFunc(&glfwWindowShouldClose, lib, "glfwWindowShouldClose")
	purego.RegisterLibFunc(&glfwSetWindowShouldClose, lib, "glfwSetWindowShouldClose")
	purego.RegisterLibFunc(&glfwGetWindowTitle, lib, "glfwGetWindowTitle")
	purego.RegisterLibFunc(&glfwSetWindowTitle, lib, "glfwSetWindowTitle")
	purego.RegisterLibFunc(&glfwSetWindowIcon, lib, "glfwSetWindowIcon")
	purego.RegisterLibFunc(&glfwGetWindowPos, lib, "glfwGetWindowPos")
	purego.RegisterLibFunc(&glfwSetWindowPos, lib, "glfwSetWindowPos")
	purego.RegisterLibFunc(&glfwGetWindowSize, lib, "glfwGetWindowSize")
	purego.RegisterLibFunc(&glfwSetWindowSizeLimits, lib, "glfwSetWindowSizeLimits")
	purego.RegisterLibFunc(&glfwSetWindowAspectRatio, lib, "glfwSetWindowAspectRatio")
	purego.RegisterLibFunc(&glfwSetWindowSize, lib, "glfwSetWindowSize")
	purego.RegisterLibFunc(&glfwGetFramebufferSize, lib, "glfwGetFramebufferSize")
	purego.RegisterLibFunc(&glfwGetWindowFrameSize, lib, "glfwGetWindowFrameSize")
	purego.RegisterLibFunc(&glfwGetWindowContentScale, lib, "glfwGetWindowContentScale")
	purego.RegisterLibFunc(&glfwGetWindowOpacity, lib, "glfwGetWindowOpacity")
	purego.RegisterLibFunc(&glfwSetWindowOpacity, lib, "glfwSetWindowOpacity")
	purego.RegisterLibFunc(&glfwIconifyWindow, lib, "glfwIconifyWindow")
	purego.RegisterLibFunc(&glfwRestoreWindow, lib, "glfwRestoreWindow")
	purego.RegisterLibFunc(&glfwMaximizeWindow, lib, "glfwMaximizeWindow")
	purego.RegisterLibFunc(&glfwShowWindow, lib, "glfwShowWindow")
	purego.RegisterLibFunc(&glfwHideWindow, lib, "glfwHideWindow")
	purego.RegisterLibFunc(&glfwFocusWindow, lib, "glfwFocusWindow")
	purego.RegisterLibFunc(&glfwRequestWindowAttention, lib, "glfwRequestWindowAttention")
	purego.RegisterLibFunc(&glfwGetWindowMonitor, lib, "glfwGetWindowMonitor")
	purego.RegisterLibFunc(&glfwSetWindowMonitor, lib, "glfwSetWindowMonitor")
	purego.RegisterLibFunc(&glfwGetWindowAttrib, lib, "glfwGetWindowAttrib")
	purego.RegisterLibFunc(&glfwSetWindowAttrib, lib, "glfwSetWindowAttrib")
	purego.RegisterLibFunc(&glfwSetWindowPosCallback, lib, "glfwSetWindowPosCallback")
	purego.RegisterLibFunc(&glfwSetWindowSizeCallback, lib, "glfwSetWindowSizeCallback")
	purego.RegisterLibFunc(&glfwSetWindowCloseCallback, lib, "glfwSetWindowCloseCallback")
	purego.RegisterLibFunc(&glfwSetWindowRefreshCallback, lib, "glfwSetWindowRefreshCallback")
	purego.RegisterLibFunc(&glfwSetWindowFocusCallback, lib, "glfwSetWindowFocusCallback")
	purego.RegisterLibFunc(&glfwSetWindowIconifyCallback, lib, "glfwSetWindowIconifyCallback")
	purego.RegisterLibFunc(&glfwSetWindowMaximizeCallback, lib, "glfwSetWindowMaximizeCallback")
	purego.RegisterLibFunc(&glfwSetFramebufferSizeCallback, lib, "glfwSetFramebufferSizeCallback")
	purego.RegisterLibFunc(&glfwSetWindowContentScaleCallback, lib, "glfwSetWindowContentScaleCallback")
	purego.RegisterLibFunc(&glfwPollEvents, lib, "glfwPollEvents")
	purego.RegisterLibFunc(&glfwWaitEvents, lib, "glfwWaitEvents")
	purego.RegisterLibFunc(&glfwWaitEventsTimeout, lib, "glfwWaitEventsTimeout")
	purego.RegisterLibFunc(&glfwPostEmptyEvent, lib, "glfwPostEmptyEvent")
	purego.RegisterLibFunc(&glfwGetInputMode, lib, "glfwGetInputMode")
	purego.RegisterLibFunc(&glfwSetInputMode, lib, "glfwSetInputMode")
	purego.RegisterLibFunc(&glfwRawMouseMotionSupported, lib, "glfwRawMouseMotionSupported")
	purego.RegisterLibFunc(&glfwGetKeyName, lib, "glfwGetKeyName")
	purego.RegisterLibFunc(&glfwGetKeyScancode, lib, "glfwGetKeyScancode")
	purego.RegisterLibFunc(&glfwGetKey, lib, "glfwGetKey")
	purego.RegisterLibFunc(&glfwGetMouseButton, lib, "glfwGetMouseButton")
	purego.RegisterLibFunc(&glfwGetCursorPos, lib, "glfwGetCursorPos")
	purego.RegisterLibFunc(&glfwSetCursorPos, lib, "glfwSetCursorPos")
	purego.RegisterLibFunc(&glfwCreateCursor, lib, "glfwCreateCursor")
	purego.RegisterLibFunc(&glfwCreateStandardCursor, lib, "glfwCreateStandardCursor")
	purego.RegisterLibFunc(&glfwDestroyCursor, lib, "glfwDestroyCursor")
	purego.RegisterLibFunc(&glfwSetCursor, lib, "glfwSetCursor")
	purego.RegisterLibFunc(&glfwSetKeyCallback, lib, "glfwSetKeyCallback")
	purego.RegisterLibFunc(&glfwSetCharCallback, lib, "glfwSetCharCallback")
	purego.RegisterLibFunc(&glfwSetCharModsCallback, lib, "glfwSetCharModsCallback")
	purego.RegisterLibFunc(&glfwSetMouseButtonCallback, lib, "glfwSetMouseButtonCallback")
	purego.RegisterLibFunc(&glfwSetCursorPosCallback, lib, "glfwSetCursorPosCallback")
	purego.RegisterLibFunc(&glfwSetCursorEnterCallback, lib, "glfwSetCursorEnterCallback")
	purego.RegisterLibFunc(&glfwSetScrollCallback, lib, "glfwSetScrollCallback")
	purego.RegisterLibFunc(&glfwSetDropCallback, lib, "glfwSetDropCallback")
	purego.RegisterLibFunc(&glfwJoystickPresent, lib, "glfwJoystickPresent")
	purego.RegisterLibFunc(&glfwGetJoystickAxes, lib, "glfwGetJoystickAxes")
	purego.RegisterLibFunc(&glfwGetJoystickButtons, lib, "glfwGetJoystickButtons")
	purego.RegisterLibFunc(&glfwGetJoystickHats, lib, "glfwGetJoystickHats")
	purego.RegisterLibFunc(&glfwGetJoystickName, lib, "glfwGetJoystickName")
	purego.RegisterLibFunc(&glfwGetJoystickGUID, lib, "glfwGetJoystickGUID")
	purego.RegisterLibFunc(&glfwJoystickIsGamepad, lib, "glfwJoystickIsGamepad")
	purego.RegisterLibFunc(&glfwSetJoystickCallback, lib, "glfwSetJoystickCallback")
	purego.RegisterLibFunc(&glfwUpdateGamepadMappings, lib, "glfwUpdateGamepadMappings")
	purego.RegisterLibFunc(&glfwGetGamepadName, lib, "glfwGetGamepadName")
	purego.RegisterLibFunc(&glfwGetGamepadState, lib, "glfwGetGamepadState")
	purego.RegisterLibFunc(&glfwSetClipboardString, lib, "glfwSetClipboardString")
	purego.RegisterLibFunc(&glfwGetClipboardString, lib, "glfwGetClipboardString")
	purego.RegisterLibFunc(&glfwGetTime, lib, "glfwGetTime")
	purego.RegisterLibFunc(&glfwSetTime, lib, "glfwSetTime")
	purego.RegisterLibFunc(&glfwGetTimerValue, lib, "glfwGetTimerValue")
	purego.RegisterLibFunc(&glfwGetTimerFrequency, lib, "glfwGetTimerFrequency")
	purego.RegisterLibFunc(&glfwMakeContextCurrent, lib, "glfwMakeContextCurrent")
	purego.RegisterLibFunc(&glfwGetCurrentContext, lib, "glfwGetCurrentContext")
	purego.RegisterLibFunc(&glfwSwapBuffers, lib, "glfwSwapBuffers")
	purego.RegisterLibFunc(&glfwSwapInterval, lib, "glfwSwapInterval")
	purego.RegisterLibFunc(&glfwExtensionSupported, lib, "glfwExtensionSupported")
	purego.RegisterLibFunc(&glfwGetProcAddress, lib, "glfwGetProcAddress")
	purego.RegisterLibFunc(&glfwVulkanSupported, lib, "glfwVulkanSupported")
	purego.RegisterLibFunc(&glfwGetRequiredInstanceExtensions, lib, "glfwGetRequiredInstanceExtensions")
	purego.RegisterLibFunc(&glfwInit, lib, "glfwInit")

	installTrampolines()

	runtime.LockOSThread()
	if glfwInit() != True {
		return fmt.Errorf("glfwInit failed")
	}
	return nil
}

// Terminate shuts down GLFW and clears handle maps, user data, and global callbacks.
func Terminate() {
	glfwTerminate()
	windowMap = sync.Map{}
	monitorMap = sync.Map{}
	cursorMap = sync.Map{}
	joyUserMap = sync.Map{}
	errorCallback = nil
	monitorCallback = nil
	joystickCallback = nil
}

// SetErrorCallback sets the global error callback and returns the previous one.
// Pass nil to remove. The callback may be invoked on any thread GLFW uses.
func SetErrorCallback(cb ErrorFunc) ErrorFunc {
	prev := errorCallback
	errorCallback = cb
	if cb != nil {
		glfwSetErrorCallback(trampError)
	} else {
		glfwSetErrorCallback(0)
	}
	return prev
}

// GetVersion returns the compile-time major, minor, and revision of the GLFW library.
func GetVersion() (major, minor, rev int32) {
	var m, mn, r int32
	glfwGetVersion(&m, &mn, &r)
	return m, mn, r
}

// GetVersionString returns a compile-time version string for the GLFW library.
func GetVersionString() string {
	return glfwGetVersionString()
}

// GetError returns and clears the last error code and its UTF-8 description.
func GetError() (code ErrorCode, description string) {
	var desc *byte
	code = ErrorCode(glfwGetError(&desc))
	return code, goString(desc)
}

// GetPlatform returns the currently selected platform.
func GetPlatform() PlatformID {
	return PlatformID(glfwGetPlatform())
}

// PlatformSupported reports whether the specified platform is supported on this machine.
func PlatformSupported(platform PlatformID) bool {
	return glfwPlatformSupported(int32(platform)) != 0
}

// GetMonitors returns the currently connected monitors.
func GetMonitors() ([]*Monitor, error) {
	var count int32
	ptr := glfwGetMonitors(&count)
	if ptr == nil || count == 0 {
		return nil, nil
	}
	arr := unsafe.Slice(ptr, int(count))
	result := make([]*Monitor, int(count))
	for i := range arr {
		result[i] = wrapMonitor(arr[i])
	}
	return result, nil
}

// GetPrimaryMonitor returns the primary monitor.
func GetPrimaryMonitor() *Monitor {
	return wrapMonitor(glfwGetPrimaryMonitor())
}

// SetMonitorCallback sets the global monitor connection callback and returns the previous one.
func SetMonitorCallback(cb MonitorFunc) MonitorFunc {
	prev := monitorCallback
	monitorCallback = cb
	if cb != nil {
		glfwSetMonitorCallback(trampMonitor)
	} else {
		glfwSetMonitorCallback(0)
	}
	return prev
}

// DefaultWindowHints resets all window hints to their default values.
func DefaultWindowHints() {
	glfwDefaultWindowHints()
}

// WindowHint sets a window creation hint. hint is a Hint name; value is typically
// True/False, an API enumerant (OpenGLAPI, OpenGLCoreProfile, …), DontCare, or a size.
func WindowHint(hint Hint, value int32) {
	glfwWindowHint(int32(hint), value)
}

// WindowHintString sets a string-valued window hint (for example CocoaFrameName).
func WindowHintString(hint Hint, value string) {
	glfwWindowHintString(int32(hint), value)
}

// CreateWindow creates a window and OpenGL/Vulkan-capable context (per hints).
// The returned *Window is a Go wrapper registered for callbacks; call Destroy when done.
// Pass nil monitor for windowed mode and nil share for no context sharing.
func CreateWindow(width, height int32, title string, monitor *Monitor, share *Window) *Window {
	var mptr, sptr unsafe.Pointer
	if monitor != nil {
		mptr = monitor.ptr
	}
	if share != nil {
		sptr = share.ptr
	}
	return wrapWindow(glfwCreateWindow(width, height, title, mptr, sptr))
}

// PollEvents processes pending events without blocking.
func PollEvents() {
	glfwPollEvents()
}

// WaitEvents waits until events are available and processes them.
func WaitEvents() {
	glfwWaitEvents()
}

// WaitEventsTimeout waits with timeout for events and then processes them.
func WaitEventsTimeout(timeout float64) {
	glfwWaitEventsTimeout(timeout)
}

// PostEmptyEvent posts an empty event from another thread to wake WaitEvents.
func PostEmptyEvent() {
	glfwPostEmptyEvent()
}

// RawMouseMotionSupported reports whether raw mouse motion is supported.
func RawMouseMotionSupported() bool {
	return glfwRawMouseMotionSupported() != 0
}

// GetKeyName returns the printable name of a key.
func GetKeyName(key Key, scancode int32) string {
	return glfwGetKeyName(int32(key), scancode)
}

// GetKeyScancode returns the platform-specific scancode of a key.
func GetKeyScancode(key Key) int32 {
	return glfwGetKeyScancode(int32(key))
}

// CreateStandardCursor creates a cursor from a standard shape.
func CreateStandardCursor(shape StandardCursor) *Cursor {
	return wrapCursor(glfwCreateStandardCursor(int32(shape)))
}

// CreateCursor creates a cursor from an RGBA image. image pixels must remain
// reachable until the call returns (GLFW copies them).
func CreateCursor(image *Image, xhot, yhot int32) *Cursor {
	if image == nil {
		return nil
	}
	cImg := imageC{Width: image.Width, Height: image.Height, Pixels: bytePtr(image.Pixels)}
	ptr := glfwCreateCursor(&cImg, xhot, yhot)
	runtime.KeepAlive(image)
	runtime.KeepAlive(cImg)
	return wrapCursor(ptr)
}

// JoystickPresent reports whether the specified joystick is present.
func JoystickPresent(jid Joystick) bool {
	return glfwJoystickPresent(int32(jid)) != 0
}

// GetJoystickAxes returns the values of all axes of the specified joystick.
func GetJoystickAxes(jid Joystick) ([]float32, error) {
	return joystickSlice(jid, glfwGetJoystickAxes)
}

// GetJoystickButtons returns the state of all buttons of the specified joystick.
func GetJoystickButtons(jid Joystick) ([]byte, error) {
	return joystickSlice(jid, glfwGetJoystickButtons)
}

// GetJoystickHats returns the state of all hats of the specified joystick.
func GetJoystickHats(jid Joystick) ([]byte, error) {
	return joystickSlice(jid, glfwGetJoystickHats)
}

// JoystickName returns the name of the specified joystick.
func JoystickName(jid Joystick) string {
	return glfwGetJoystickName(int32(jid))
}

// JoystickGUID returns the SDL-compatible GUID of the specified joystick.
func JoystickGUID(jid Joystick) string {
	return glfwGetJoystickGUID(int32(jid))
}

// SetJoystickUserPointer stores an arbitrary Go value for a joystick (Go-side only).
func SetJoystickUserPointer(jid Joystick, ptr any) {
	if ptr == nil {
		joyUserMap.Delete(jid)
		return
	}
	joyUserMap.Store(jid, ptr)
}

// JoystickUserPointer returns the value previously passed to SetJoystickUserPointer.
func JoystickUserPointer(jid Joystick) any {
	if v, ok := joyUserMap.Load(jid); ok {
		return v
	}
	return nil
}

// JoystickIsGamepad reports whether the joystick has a gamepad mapping.
func JoystickIsGamepad(jid Joystick) bool { return glfwJoystickIsGamepad(int32(jid)) != 0 }

// SetJoystickCallback sets the global joystick connection callback and returns the previous one.
func SetJoystickCallback(cb JoystickFunc) JoystickFunc {
	prev := joystickCallback
	joystickCallback = cb
	if cb != nil {
		glfwSetJoystickCallback(trampJoystick)
	} else {
		glfwSetJoystickCallback(0)
	}
	return prev
}

// UpdateGamepadMappings adds or updates gamepad mappings from an ASCII string.
func UpdateGamepadMappings(str string) bool {
	return glfwUpdateGamepadMappings(str) != 0
}

// GamepadName returns the human-readable name of the gamepad mapped to the joystick.
func GamepadName(jid Joystick) string {
	return glfwGetGamepadName(int32(jid))
}

// GetGamepadState returns the gamepad input state for the specified joystick.
func GetGamepadState(jid Joystick) (*GamepadState, int32) {
	var s GamepadState
	return &s, glfwGetGamepadState(int32(jid), &s)
}

// GetTime returns the value of the GLFW timer in seconds.
func GetTime() float64 {
	return glfwGetTime()
}

// SetTime sets the value of the GLFW timer in seconds.
func SetTime(time float64) {
	glfwSetTime(time)
}

// GetTimerValue returns the current value of the raw timer.
func GetTimerValue() uint64 {
	return glfwGetTimerValue()
}

// GetTimerFrequency returns the frequency of the raw timer in Hz.
func GetTimerFrequency() uint64 {
	return glfwGetTimerFrequency()
}

// CurrentContext returns the window whose OpenGL or OpenGL ES context is current.
func CurrentContext() *Window { return wrapWindow(glfwGetCurrentContext()) }

// SwapInterval sets the swap interval for the current OpenGL or OpenGL ES context.
func SwapInterval(interval int32) {
	glfwSwapInterval(interval)
}

// ExtensionSupported reports whether the specified OpenGL or OpenGL ES extension is supported.
func ExtensionSupported(extension string) bool {
	return glfwExtensionSupported(extension) != 0
}

// GetProcAddress returns a raw OpenGL/OpenGL ES function pointer as uintptr
// (not a Go func value). Cast or pass it to your GL loader as needed.
func GetProcAddress(procname string) uintptr {
	return glfwGetProcAddress(procname)
}

// VulkanSupported reports whether the Vulkan loader and an ICD have been found.
func VulkanSupported() bool {
	return glfwVulkanSupported() != 0
}

// GetRequiredInstanceExtensions returns the Vulkan instance extensions required by GLFW.
func GetRequiredInstanceExtensions() []string {
	var count uint32
	ptr := glfwGetRequiredInstanceExtensions(&count)
	return goStrings(ptr, int(count))
}

func (w *Window) clearCallbacks() {
	w.posCb = nil
	w.sizeCb = nil
	w.closeCb = nil
	w.refreshCb = nil
	w.focusCb = nil
	w.iconifyCb = nil
	w.maximizeCb = nil
	w.framebufferSizeCb = nil
	w.contentScaleCb = nil
	w.keyCb = nil
	w.charCb = nil
	w.charModsCb = nil
	w.mouseButtonCb = nil
	w.cursorPosCb = nil
	w.cursorEnterCb = nil
	w.scrollCb = nil
	w.dropCb = nil
}

// ShouldClose reports whether the window has been requested to close.
func (w *Window) ShouldClose() bool {
	return glfwWindowShouldClose(w.ptr) != 0
}

// SetShouldClose sets whether the window should be closed.
func (w *Window) SetShouldClose(v bool) {
	glfwSetWindowShouldClose(w.ptr, boolToI32(v))
}

// Title returns the window title.
func (w *Window) Title() string {
	return glfwGetWindowTitle(w.ptr)
}

// SetTitle sets the window title.
func (w *Window) SetTitle(title string) {
	glfwSetWindowTitle(w.ptr, title)
}

// SetIcon sets the window icon from one or more images.
func (w *Window) SetIcon(images []Image) {
	if len(images) == 0 {
		glfwSetWindowIcon(w.ptr, 0, nil)
		return
	}
	cImages := make([]imageC, len(images))
	for i, img := range images {
		cImages[i] = imageC{
			Width:  img.Width,
			Height: img.Height,
			Pixels: bytePtr(img.Pixels),
		}
	}
	glfwSetWindowIcon(w.ptr, int32(len(cImages)), &cImages[0])
	runtime.KeepAlive(images)
	runtime.KeepAlive(cImages)
}

// Pos returns the position of the window's content area.
func (w *Window) Pos() (x, y int32) {
	glfwGetWindowPos(w.ptr, &x, &y)
	return
}

// SetPos sets the position of the window's content area.
func (w *Window) SetPos(x, y int32) {
	glfwSetWindowPos(w.ptr, x, y)
}

// Size returns the size of the window's content area.
func (w *Window) Size() (w2, h int32) {
	glfwGetWindowSize(w.ptr, &w2, &h)
	return
}

// SetSizeLimits sets the minimum and maximum size limits of the content area.
func (w *Window) SetSizeLimits(minW, minH, maxW, maxH int32) {
	glfwSetWindowSizeLimits(w.ptr, minW, minH, maxW, maxH)
}

// SetAspectRatio sets the required aspect ratio of the content area.
func (w *Window) SetAspectRatio(numer, denom int32) {
	glfwSetWindowAspectRatio(w.ptr, numer, denom)
}

// SetSize sets the size of the window's content area.
func (w *Window) SetSize(width, height int32) {
	glfwSetWindowSize(w.ptr, width, height)
}

// FramebufferSize returns the size of the framebuffer.
func (w *Window) FramebufferSize() (w2, h int32) {
	glfwGetFramebufferSize(w.ptr, &w2, &h)
	return
}

// FrameSize returns the size of the window frame edges.
func (w *Window) FrameSize() (left, top, right, bottom int32) {
	glfwGetWindowFrameSize(w.ptr, &left, &top, &right, &bottom)
	return
}

// ContentScale returns the content scale of the window.
func (w *Window) ContentScale() (xscale, yscale float32) {
	glfwGetWindowContentScale(w.ptr, &xscale, &yscale)
	return
}

// Opacity returns the opacity of the window.
func (w *Window) Opacity() float32 {
	return glfwGetWindowOpacity(w.ptr)
}

// SetOpacity sets the opacity of the window.
func (w *Window) SetOpacity(opacity float32) {
	glfwSetWindowOpacity(w.ptr, opacity)
}

// Iconify iconifies the window.
func (w *Window) Iconify() {
	glfwIconifyWindow(w.ptr)
}

// Restore restores the window if it was iconified or maximized.
func (w *Window) Restore() {
	glfwRestoreWindow(w.ptr)
}

// Maximize maximizes the window.
func (w *Window) Maximize() {
	glfwMaximizeWindow(w.ptr)
}

// Show makes the window visible.
func (w *Window) Show() {
	glfwShowWindow(w.ptr)
}

// Hide hides the window.
func (w *Window) Hide() {
	glfwHideWindow(w.ptr)
}

// Focus brings the window to front and sets input focus.
func (w *Window) Focus() {
	glfwFocusWindow(w.ptr)
}

// RequestAttention requests user attention to the window.
func (w *Window) RequestAttention() {
	glfwRequestWindowAttention(w.ptr)
}

// Monitor returns the monitor the window is fullscreen on, or nil.
func (w *Window) Monitor() *Monitor {
	return wrapMonitor(glfwGetWindowMonitor(w.ptr))
}

// SetMonitor sets the monitor and video mode for fullscreen or windowed mode.
// Pass nil monitor to switch back to windowed mode.
func (w *Window) SetMonitor(monitor *Monitor, xpos, ypos, width, height, refreshRate int32) {
	var mptr unsafe.Pointer
	if monitor != nil {
		mptr = monitor.ptr
	}
	glfwSetWindowMonitor(w.ptr, mptr, xpos, ypos, width, height, refreshRate)
}

// GetAttrib returns the value of a window attribute.
func (w *Window) GetAttrib(attrib Hint) int32 {
	return glfwGetWindowAttrib(w.ptr, int32(attrib))
}

// SetAttrib sets the value of a window attribute.
func (w *Window) SetAttrib(attrib Hint, value int32) {
	glfwSetWindowAttrib(w.ptr, int32(attrib), value)
}

// SetUserPointer stores an arbitrary Go value on this window.
// Unlike glfwSetWindowUserPointer, the value is kept only on the Go wrapper
// so the garbage collector can track it. Pass nil to clear.
func (w *Window) SetUserPointer(ptr any) { w.user = ptr }

// UserPointer returns the value previously passed to SetUserPointer.
func (w *Window) UserPointer() any { return w.user }

// SetPosCallback sets the window position callback and returns the previous one.
// Pass nil to remove. All Set*Callback methods on Window share this replace semantics.
func (w *Window) SetPosCallback(cb WindowPosFunc) WindowPosFunc {
	prev := w.posCb
	w.posCb = cb
	if cb != nil {
		glfwSetWindowPosCallback(w.ptr, trampWindowPos)
	} else {
		glfwSetWindowPosCallback(w.ptr, 0)
	}
	return prev
}

// SetSizeCallback sets the window size callback and returns the previous one.
func (w *Window) SetSizeCallback(cb WindowSizeFunc) WindowSizeFunc {
	prev := w.sizeCb
	w.sizeCb = cb
	if cb != nil {
		glfwSetWindowSizeCallback(w.ptr, trampWindowSize)
	} else {
		glfwSetWindowSizeCallback(w.ptr, 0)
	}
	return prev
}

// SetCloseCallback sets the close callback and returns the previous one.
func (w *Window) SetCloseCallback(cb WindowCloseFunc) WindowCloseFunc {
	prev := w.closeCb
	w.closeCb = cb
	if cb != nil {
		glfwSetWindowCloseCallback(w.ptr, trampWindowClose)
	} else {
		glfwSetWindowCloseCallback(w.ptr, 0)
	}
	return prev
}

// SetRefreshCallback sets the refresh callback and returns the previous one.
func (w *Window) SetRefreshCallback(cb WindowRefreshFunc) WindowRefreshFunc {
	prev := w.refreshCb
	w.refreshCb = cb
	if cb != nil {
		glfwSetWindowRefreshCallback(w.ptr, trampWindowRefresh)
	} else {
		glfwSetWindowRefreshCallback(w.ptr, 0)
	}
	return prev
}

// SetFocusCallback sets the focus callback and returns the previous one.
func (w *Window) SetFocusCallback(cb WindowFocusFunc) WindowFocusFunc {
	prev := w.focusCb
	w.focusCb = cb
	if cb != nil {
		glfwSetWindowFocusCallback(w.ptr, trampWindowFocus)
	} else {
		glfwSetWindowFocusCallback(w.ptr, 0)
	}
	return prev
}

// SetIconifyCallback sets the iconify callback and returns the previous one.
func (w *Window) SetIconifyCallback(cb WindowIconifyFunc) WindowIconifyFunc {
	prev := w.iconifyCb
	w.iconifyCb = cb
	if cb != nil {
		glfwSetWindowIconifyCallback(w.ptr, trampWindowIconify)
	} else {
		glfwSetWindowIconifyCallback(w.ptr, 0)
	}
	return prev
}

// SetMaximizeCallback sets the maximize callback and returns the previous one.
func (w *Window) SetMaximizeCallback(cb WindowMaximizeFunc) WindowMaximizeFunc {
	prev := w.maximizeCb
	w.maximizeCb = cb
	if cb != nil {
		glfwSetWindowMaximizeCallback(w.ptr, trampWindowMaximize)
	} else {
		glfwSetWindowMaximizeCallback(w.ptr, 0)
	}
	return prev
}

// SetFramebufferSizeCallback sets the framebuffer size callback and returns the previous one.
func (w *Window) SetFramebufferSizeCallback(cb FramebufferSizeFunc) FramebufferSizeFunc {
	prev := w.framebufferSizeCb
	w.framebufferSizeCb = cb
	if cb != nil {
		glfwSetFramebufferSizeCallback(w.ptr, trampFramebufferSize)
	} else {
		glfwSetFramebufferSizeCallback(w.ptr, 0)
	}
	return prev
}

// SetContentScaleCallback sets the content scale callback and returns the previous one.
func (w *Window) SetContentScaleCallback(cb WindowContentScaleFunc) WindowContentScaleFunc {
	prev := w.contentScaleCb
	w.contentScaleCb = cb
	if cb != nil {
		glfwSetWindowContentScaleCallback(w.ptr, trampWindowContentScale)
	} else {
		glfwSetWindowContentScaleCallback(w.ptr, 0)
	}
	return prev
}

// GetInputMode returns the value of an input option for the window.
func (w *Window) GetInputMode(mode InputMode) int32 {
	return glfwGetInputMode(w.ptr, int32(mode))
}

// SetInputMode sets an input option for the window.
func (w *Window) SetInputMode(mode InputMode, value int32) {
	glfwSetInputMode(w.ptr, int32(mode), value)
}

// Key returns the last reported state of a keyboard key.
func (w *Window) Key(key Key) Action {
	return Action(glfwGetKey(w.ptr, int32(key)))
}

// MouseButton returns the last reported state of a mouse button.
func (w *Window) MouseButton(button MouseButton) Action {
	return Action(glfwGetMouseButton(w.ptr, int32(button)))
}

// CursorPos returns the position of the cursor relative to the content area.
func (w *Window) CursorPos() (x, y float64) {
	glfwGetCursorPos(w.ptr, &x, &y)
	return
}

// SetCursorPos sets the position of the cursor relative to the content area.
func (w *Window) SetCursorPos(x, y float64) {
	glfwSetCursorPos(w.ptr, x, y)
}

// SetCursor sets the cursor image used when the cursor is over the content area.
// Pass nil to restore the default arrow cursor.
func (w *Window) SetCursor(cursor *Cursor) {
	var cptr unsafe.Pointer
	if cursor != nil {
		cptr = cursor.ptr
	}
	glfwSetCursor(w.ptr, cptr)
}

// SetKeyCallback sets the key callback and returns the previous one.
func (w *Window) SetKeyCallback(cb KeyFunc) KeyFunc {
	prev := w.keyCb
	w.keyCb = cb
	if cb != nil {
		glfwSetKeyCallback(w.ptr, trampKey)
	} else {
		glfwSetKeyCallback(w.ptr, 0)
	}
	return prev
}

// SetCharCallback sets the Unicode character callback and returns the previous one.
func (w *Window) SetCharCallback(cb CharFunc) CharFunc {
	prev := w.charCb
	w.charCb = cb
	if cb != nil {
		glfwSetCharCallback(w.ptr, trampChar)
	} else {
		glfwSetCharCallback(w.ptr, 0)
	}
	return prev
}

// SetCharModsCallback sets the character-with-modifiers callback and returns the previous one.
func (w *Window) SetCharModsCallback(cb CharModsFunc) CharModsFunc {
	prev := w.charModsCb
	w.charModsCb = cb
	if cb != nil {
		glfwSetCharModsCallback(w.ptr, trampCharMods)
	} else {
		glfwSetCharModsCallback(w.ptr, 0)
	}
	return prev
}

// SetMouseButtonCallback sets the mouse button callback and returns the previous one.
func (w *Window) SetMouseButtonCallback(cb MouseButtonFunc) MouseButtonFunc {
	prev := w.mouseButtonCb
	w.mouseButtonCb = cb
	if cb != nil {
		glfwSetMouseButtonCallback(w.ptr, trampMouseButton)
	} else {
		glfwSetMouseButtonCallback(w.ptr, 0)
	}
	return prev
}

// SetCursorPosCallback sets the cursor position callback and returns the previous one.
func (w *Window) SetCursorPosCallback(cb CursorPosFunc) CursorPosFunc {
	prev := w.cursorPosCb
	w.cursorPosCb = cb
	if cb != nil {
		glfwSetCursorPosCallback(w.ptr, trampCursorPos)
	} else {
		glfwSetCursorPosCallback(w.ptr, 0)
	}
	return prev
}

// SetCursorEnterCallback sets the cursor enter/leave callback and returns the previous one.
func (w *Window) SetCursorEnterCallback(cb CursorEnterFunc) CursorEnterFunc {
	prev := w.cursorEnterCb
	w.cursorEnterCb = cb
	if cb != nil {
		glfwSetCursorEnterCallback(w.ptr, trampCursorEnter)
	} else {
		glfwSetCursorEnterCallback(w.ptr, 0)
	}
	return prev
}

// SetScrollCallback sets the scroll callback and returns the previous one.
func (w *Window) SetScrollCallback(cb ScrollFunc) ScrollFunc {
	prev := w.scrollCb
	w.scrollCb = cb
	if cb != nil {
		glfwSetScrollCallback(w.ptr, trampScroll)
	} else {
		glfwSetScrollCallback(w.ptr, 0)
	}
	return prev
}

// SetDropCallback sets the path drop callback and returns the previous one.
func (w *Window) SetDropCallback(cb DropFunc) DropFunc {
	prev := w.dropCb
	w.dropCb = cb
	if cb != nil {
		glfwSetDropCallback(w.ptr, trampDrop)
	} else {
		glfwSetDropCallback(w.ptr, 0)
	}
	return prev
}

// ClipboardString returns the contents of the system clipboard.
func (w *Window) ClipboardString() string {
	return glfwGetClipboardString(w.ptr)
}

// SetClipboardString sets the system clipboard to the specified string.
func (w *Window) SetClipboardString(str string) {
	glfwSetClipboardString(w.ptr, str)
}

// SwapBuffers swaps the front and back buffers of the window.
func (w *Window) SwapBuffers() {
	glfwSwapBuffers(w.ptr)
}

// MakeContextCurrent makes the window's OpenGL or OpenGL ES context current.
func (w *Window) MakeContextCurrent() {
	glfwMakeContextCurrent(w.ptr)
}

// Destroy destroys the window, clears its callbacks and user data, and
// unregisters it from the handle map. The Window must not be used afterward.
func (w *Window) Destroy() {
	if w == nil || w.ptr == nil {
		return
	}
	ptr := w.ptr
	w.clearCallbacks()
	w.user = nil
	windowMap.Delete(ptr)
	glfwDestroyWindow(ptr)
	w.ptr = nil
}

// Pos returns the position of the monitor's viewport on the virtual desktop.
func (m *Monitor) Pos() (x, y int32) {
	glfwGetMonitorPos(m.ptr, &x, &y)
	return
}

// Workarea returns the work area of the monitor.
func (m *Monitor) Workarea() (x, y, w2, h int32) {
	glfwGetMonitorWorkarea(m.ptr, &x, &y, &w2, &h)
	return
}

// PhysicalSize returns the physical size of the monitor in millimetres.
func (m *Monitor) PhysicalSize() (wMM, hMM int32) {
	glfwGetMonitorPhysicalSize(m.ptr, &wMM, &hMM)
	return
}

// ContentScale returns the content scale of the monitor.
func (m *Monitor) ContentScale() (xscale, yscale float32) {
	glfwGetMonitorContentScale(m.ptr, &xscale, &yscale)
	return
}

// Name returns a human-readable name of the monitor.
func (m *Monitor) Name() string {
	return glfwGetMonitorName(m.ptr)
}

// SetUserPointer stores an arbitrary Go value on this monitor (Go-side only).
func (m *Monitor) SetUserPointer(ptr any) {
	m.user = ptr
}

// UserPointer returns the value previously passed to SetUserPointer.
func (m *Monitor) UserPointer() any {
	return m.user
}

// VideoModes returns all video modes supported by the monitor.
func (m *Monitor) VideoModes() []VidMode {
	var count int32
	ptr := glfwGetVideoModes(m.ptr, &count)
	if ptr == nil || count == 0 {
		return nil
	}
	arr := unsafe.Slice(ptr, int(count))
	result := make([]VidMode, int(count))
	copy(result, arr)
	return result
}

// VideoMode returns the current video mode of the monitor.
func (m *Monitor) VideoMode() VidMode {
	p := glfwGetVideoMode(m.ptr)
	if p == nil {
		return VidMode{}
	}
	return *p
}

// SetGamma generates a gamma ramp from the specified exponent and applies it.
func (m *Monitor) SetGamma(gamma float32) {
	glfwSetGamma(m.ptr, gamma)
}

// GammaRamp returns the current gamma ramp of the monitor.
func (m *Monitor) GammaRamp() GammaRamp {
	p := glfwGetGammaRampC(m.ptr)
	if p == nil || p.Size == 0 {
		return GammaRamp{}
	}
	n := int(p.Size)
	return GammaRamp{
		Red:   unsafe.Slice(p.Red, n),
		Green: unsafe.Slice(p.Green, n),
		Blue:  unsafe.Slice(p.Blue, n),
		Size:  p.Size,
	}
}

// SetGammaRamp sets the current gamma ramp of the monitor.
func (m *Monitor) SetGammaRamp(ramp *GammaRamp) {
	if ramp == nil || ramp.Size == 0 {
		glfwSetGammaRamp(m.ptr, nil)
		return
	}
	cRamp := &gammaRampC{
		Red:   uint16Ptr(ramp.Red),
		Green: uint16Ptr(ramp.Green),
		Blue:  uint16Ptr(ramp.Blue),
		Size:  ramp.Size,
	}
	glfwSetGammaRamp(m.ptr, cRamp)
	runtime.KeepAlive(ramp)
	runtime.KeepAlive(cRamp)
}

// Destroy destroys the cursor and unregisters it from the handle map.
func (c *Cursor) Destroy() {
	if c == nil || c.ptr == nil {
		return
	}
	ptr := c.ptr
	cursorMap.Delete(ptr)
	glfwDestroyCursor(ptr)
	c.ptr = nil
}

func boolToI32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func goString(p *byte) string {
	if p == nil {
		return ""
	}
	var n uintptr
	for *(*byte)(unsafe.Add(unsafe.Pointer(p), n)) != 0 {
		n++
	}
	return string(unsafe.Slice(p, n))
}

func goStrings(pp **byte, count int) []string {
	if pp == nil || count == 0 {
		return nil
	}
	ptrs := unsafe.Slice(pp, count)
	out := make([]string, count)
	for i, p := range ptrs {
		out[i] = goString(p)
	}
	return out
}

func joystickSlice[T any](jid Joystick, get func(int32, *int32) *T) ([]T, error) {
	var count int32
	ptr := get(int32(jid), &count)
	if ptr == nil {
		return nil, nil
	}
	return unsafe.Slice(ptr, int(count)), nil
}

func bytePtr(b []byte) *byte {
	if len(b) == 0 {
		return nil
	}
	return &b[0]
}

func uint16Ptr(u []uint16) *uint16 {
	if len(u) == 0 {
		return nil
	}
	return &u[0]
}

func wrapWindow(ptr unsafe.Pointer) *Window {
	if ptr == nil {
		return nil
	}
	if v, ok := windowMap.Load(ptr); ok {
		return v.(*Window)
	}
	w := &Window{ptr: ptr}
	actual, _ := windowMap.LoadOrStore(ptr, w)
	return actual.(*Window)
}

func getWindow(ptr unsafe.Pointer) *Window {
	if ptr == nil {
		return nil
	}
	if v, ok := windowMap.Load(ptr); ok {
		return v.(*Window)
	}
	return nil
}

func wrapMonitor(ptr unsafe.Pointer) *Monitor {
	if ptr == nil {
		return nil
	}
	if v, ok := monitorMap.Load(ptr); ok {
		return v.(*Monitor)
	}
	m := &Monitor{ptr: ptr}
	actual, _ := monitorMap.LoadOrStore(ptr, m)
	return actual.(*Monitor)
}

func wrapCursor(ptr unsafe.Pointer) *Cursor {
	if ptr == nil {
		return nil
	}
	if v, ok := cursorMap.Load(ptr); ok {
		return v.(*Cursor)
	}
	c := &Cursor{ptr: ptr}
	actual, _ := cursorMap.LoadOrStore(ptr, c)
	return actual.(*Cursor)
}

var (
	windowMap  sync.Map // unsafe.Pointer -> *Window
	monitorMap sync.Map // unsafe.Pointer -> *Monitor
	cursorMap  sync.Map // unsafe.Pointer -> *Cursor
	joyUserMap sync.Map // Joystick -> any

	errorCallback    ErrorFunc
	monitorCallback  MonitorFunc
	joystickCallback JoystickFunc

	// Stable C-callable trampolines (one NewCallback each, created in Init).
	trampWindowPos          uintptr
	trampWindowSize         uintptr
	trampWindowClose        uintptr
	trampWindowRefresh      uintptr
	trampWindowFocus        uintptr
	trampWindowIconify      uintptr
	trampWindowMaximize     uintptr
	trampFramebufferSize    uintptr
	trampWindowContentScale uintptr
	trampKey                uintptr
	trampChar               uintptr
	trampCharMods           uintptr
	trampMouseButton        uintptr
	trampCursorPos          uintptr
	trampCursorEnter        uintptr
	trampScroll             uintptr
	trampDrop               uintptr
	trampError              uintptr
	trampMonitor            uintptr
	trampJoystick           uintptr
)

var glfwInit func() int32
var glfwTerminate func()
var glfwInitHint func(hint int32, value int32)
var glfwGetVersion func(major *int32, minor *int32, rev *int32)
var glfwGetVersionString func() string
var glfwGetError func(description **byte) int32
var glfwSetErrorCallback func(callback uintptr) uintptr
var glfwGetPlatform func() int32
var glfwPlatformSupported func(platform int32) int32
var glfwGetMonitors func(count *int32) *unsafe.Pointer
var glfwGetPrimaryMonitor func() unsafe.Pointer
var glfwGetMonitorPos func(monitor unsafe.Pointer, xpos *int32, ypos *int32)
var glfwGetMonitorWorkarea func(monitor unsafe.Pointer, xpos *int32, ypos *int32, width *int32, height *int32)
var glfwGetMonitorPhysicalSize func(monitor unsafe.Pointer, widthMM *int32, heightMM *int32)
var glfwGetMonitorContentScale func(monitor unsafe.Pointer, xscale *float32, yscale *float32)
var glfwGetMonitorName func(monitor unsafe.Pointer) string
var glfwSetMonitorCallback func(callback uintptr) uintptr
var glfwGetVideoModes func(monitor unsafe.Pointer, count *int32) *VidMode
var glfwGetVideoMode func(monitor unsafe.Pointer) *VidMode
var glfwSetGamma func(monitor unsafe.Pointer, gamma float32)
var glfwGetGammaRampC func(monitor unsafe.Pointer) *gammaRampC
var glfwSetGammaRamp func(monitor unsafe.Pointer, ramp *gammaRampC)
var glfwDefaultWindowHints func()
var glfwWindowHint func(hint int32, value int32)
var glfwWindowHintString func(hint int32, value string)
var glfwCreateWindow func(width int32, height int32, title string, monitor unsafe.Pointer, share unsafe.Pointer) unsafe.Pointer
var glfwDestroyWindow func(window unsafe.Pointer)
var glfwWindowShouldClose func(window unsafe.Pointer) int32
var glfwSetWindowShouldClose func(window unsafe.Pointer, value int32)
var glfwGetWindowTitle func(window unsafe.Pointer) string
var glfwSetWindowTitle func(window unsafe.Pointer, title string)
var glfwSetWindowIcon func(window unsafe.Pointer, count int32, images *imageC)
var glfwGetWindowPos func(window unsafe.Pointer, xpos *int32, ypos *int32)
var glfwSetWindowPos func(window unsafe.Pointer, xpos int32, ypos int32)
var glfwGetWindowSize func(window unsafe.Pointer, width *int32, height *int32)
var glfwSetWindowSizeLimits func(window unsafe.Pointer, minwidth int32, minheight int32, maxwidth int32, maxheight int32)
var glfwSetWindowAspectRatio func(window unsafe.Pointer, numer int32, denom int32)
var glfwSetWindowSize func(window unsafe.Pointer, width int32, height int32)
var glfwGetFramebufferSize func(window unsafe.Pointer, width *int32, height *int32)
var glfwGetWindowFrameSize func(window unsafe.Pointer, left *int32, top *int32, right *int32, bottom *int32)
var glfwGetWindowContentScale func(window unsafe.Pointer, xscale *float32, yscale *float32)
var glfwGetWindowOpacity func(window unsafe.Pointer) float32
var glfwSetWindowOpacity func(window unsafe.Pointer, opacity float32)
var glfwIconifyWindow func(window unsafe.Pointer)
var glfwRestoreWindow func(window unsafe.Pointer)
var glfwMaximizeWindow func(window unsafe.Pointer)
var glfwShowWindow func(window unsafe.Pointer)
var glfwHideWindow func(window unsafe.Pointer)
var glfwFocusWindow func(window unsafe.Pointer)
var glfwRequestWindowAttention func(window unsafe.Pointer)
var glfwGetWindowMonitor func(window unsafe.Pointer) unsafe.Pointer
var glfwSetWindowMonitor func(window unsafe.Pointer, monitor unsafe.Pointer, xpos int32, ypos int32, width int32, height int32, refreshRate int32)
var glfwGetWindowAttrib func(window unsafe.Pointer, attrib int32) int32
var glfwSetWindowAttrib func(window unsafe.Pointer, attrib int32, value int32)
var glfwSetWindowPosCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetWindowSizeCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetWindowCloseCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetWindowRefreshCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetWindowFocusCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetWindowIconifyCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetWindowMaximizeCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetFramebufferSizeCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetWindowContentScaleCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwPollEvents func()
var glfwWaitEvents func()
var glfwWaitEventsTimeout func(timeout float64)
var glfwPostEmptyEvent func()
var glfwGetInputMode func(window unsafe.Pointer, mode int32) int32
var glfwSetInputMode func(window unsafe.Pointer, mode int32, value int32)
var glfwRawMouseMotionSupported func() int32
var glfwGetKeyName func(key int32, scancode int32) string
var glfwGetKeyScancode func(key int32) int32
var glfwGetKey func(window unsafe.Pointer, key int32) int32
var glfwGetMouseButton func(window unsafe.Pointer, button int32) int32
var glfwGetCursorPos func(window unsafe.Pointer, xpos *float64, ypos *float64)
var glfwSetCursorPos func(window unsafe.Pointer, xpos float64, ypos float64)
var glfwCreateCursor func(image *imageC, xhot int32, yhot int32) unsafe.Pointer
var glfwCreateStandardCursor func(shape int32) unsafe.Pointer
var glfwDestroyCursor func(cursor unsafe.Pointer)
var glfwSetCursor func(window unsafe.Pointer, cursor unsafe.Pointer)
var glfwSetKeyCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetCharCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetCharModsCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetMouseButtonCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetCursorPosCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetCursorEnterCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetScrollCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwSetDropCallback func(window unsafe.Pointer, callback uintptr) uintptr
var glfwJoystickPresent func(jid int32) int32
var glfwGetJoystickAxes func(jid int32, count *int32) *float32
var glfwGetJoystickButtons func(jid int32, count *int32) *byte
var glfwGetJoystickHats func(jid int32, count *int32) *byte
var glfwGetJoystickName func(jid int32) string
var glfwGetJoystickGUID func(jid int32) string
var glfwJoystickIsGamepad func(jid int32) int32
var glfwSetJoystickCallback func(callback uintptr) uintptr
var glfwUpdateGamepadMappings func(str string) int32
var glfwGetGamepadName func(jid int32) string
var glfwGetGamepadState func(jid int32, state *GamepadState) int32
var glfwSetClipboardString func(window unsafe.Pointer, str string)
var glfwGetClipboardString func(window unsafe.Pointer) string
var glfwGetTime func() float64
var glfwSetTime func(time float64)
var glfwGetTimerValue func() uint64
var glfwGetTimerFrequency func() uint64
var glfwMakeContextCurrent func(window unsafe.Pointer)
var glfwGetCurrentContext func() unsafe.Pointer
var glfwSwapBuffers func(window unsafe.Pointer)
var glfwSwapInterval func(interval int32)
var glfwExtensionSupported func(extension string) int32
var glfwGetProcAddress func(procname string) uintptr
var glfwVulkanSupported func() int32
var glfwGetRequiredInstanceExtensions func(count *uint32) **byte

func installTrampolines() {
	trampWindowPos = purego.NewCallback(func(wptr unsafe.Pointer, xpos, ypos int32) {
		if w := getWindow(wptr); w != nil && w.posCb != nil {
			w.posCb(w, xpos, ypos)
		}
	})
	trampWindowSize = purego.NewCallback(func(wptr unsafe.Pointer, width, height int32) {
		if w := getWindow(wptr); w != nil && w.sizeCb != nil {
			w.sizeCb(w, width, height)
		}
	})
	trampWindowClose = purego.NewCallback(func(wptr unsafe.Pointer) {
		if w := getWindow(wptr); w != nil && w.closeCb != nil {
			w.closeCb(w)
		}
	})
	trampWindowRefresh = purego.NewCallback(func(wptr unsafe.Pointer) {
		if w := getWindow(wptr); w != nil && w.refreshCb != nil {
			w.refreshCb(w)
		}
	})
	trampWindowFocus = purego.NewCallback(func(wptr unsafe.Pointer, focused int32) {
		if w := getWindow(wptr); w != nil && w.focusCb != nil {
			w.focusCb(w, focused)
		}
	})
	trampWindowIconify = purego.NewCallback(func(wptr unsafe.Pointer, iconified int32) {
		if w := getWindow(wptr); w != nil && w.iconifyCb != nil {
			w.iconifyCb(w, iconified)
		}
	})
	trampWindowMaximize = purego.NewCallback(func(wptr unsafe.Pointer, maximized int32) {
		if w := getWindow(wptr); w != nil && w.maximizeCb != nil {
			w.maximizeCb(w, maximized)
		}
	})
	trampFramebufferSize = purego.NewCallback(func(wptr unsafe.Pointer, width, height int32) {
		if w := getWindow(wptr); w != nil && w.framebufferSizeCb != nil {
			w.framebufferSizeCb(w, width, height)
		}
	})
	trampWindowContentScale = purego.NewCallback(func(wptr unsafe.Pointer, xscale, yscale float32) {
		if w := getWindow(wptr); w != nil && w.contentScaleCb != nil {
			w.contentScaleCb(w, xscale, yscale)
		}
	})
	trampKey = purego.NewCallback(func(wptr unsafe.Pointer, key, scancode, action, mods int32) {
		if w := getWindow(wptr); w != nil && w.keyCb != nil {
			w.keyCb(w, Key(key), scancode, Action(action), ModifierKey(mods))
		}
	})
	trampChar = purego.NewCallback(func(wptr unsafe.Pointer, codepoint uint32) {
		if w := getWindow(wptr); w != nil && w.charCb != nil {
			w.charCb(w, codepoint)
		}
	})
	trampCharMods = purego.NewCallback(func(wptr unsafe.Pointer, codepoint uint32, mods int32) {
		if w := getWindow(wptr); w != nil && w.charModsCb != nil {
			w.charModsCb(w, codepoint, ModifierKey(mods))
		}
	})
	trampMouseButton = purego.NewCallback(func(wptr unsafe.Pointer, button, action, mods int32) {
		if w := getWindow(wptr); w != nil && w.mouseButtonCb != nil {
			w.mouseButtonCb(w, MouseButton(button), Action(action), ModifierKey(mods))
		}
	})
	trampCursorPos = purego.NewCallback(func(wptr unsafe.Pointer, xpos, ypos float64) {
		if w := getWindow(wptr); w != nil && w.cursorPosCb != nil {
			w.cursorPosCb(w, xpos, ypos)
		}
	})
	trampCursorEnter = purego.NewCallback(func(wptr unsafe.Pointer, entered int32) {
		if w := getWindow(wptr); w != nil && w.cursorEnterCb != nil {
			w.cursorEnterCb(w, entered)
		}
	})
	trampScroll = purego.NewCallback(func(wptr unsafe.Pointer, xoffset, yoffset float64) {
		if w := getWindow(wptr); w != nil && w.scrollCb != nil {
			w.scrollCb(w, xoffset, yoffset)
		}
	})
	trampDrop = purego.NewCallback(func(wptr unsafe.Pointer, pathCount int32, paths **byte) {
		if w := getWindow(wptr); w != nil && w.dropCb != nil {
			w.dropCb(w, goStrings(paths, int(pathCount)))
		}
	})
	trampError = purego.NewCallback(func(errorCode int32, description *byte) {
		if errorCallback != nil {
			errorCallback(ErrorCode(errorCode), goString(description))
		}
	})
	trampMonitor = purego.NewCallback(func(mptr unsafe.Pointer, event int32) {
		if monitorCallback != nil {
			monitorCallback(wrapMonitor(mptr), PeripheralEvent(event))
		}
	})
	trampJoystick = purego.NewCallback(func(jid, event int32) {
		if joystickCallback != nil {
			joystickCallback(Joystick(jid), PeripheralEvent(event))
		}
	})
}

func loadGLFW() (uintptr, error) {
	var names []string
	var searchPaths []string

	// 1. Determine OS-specific file names
	switch runtime.GOOS {
	case "darwin":
		names = []string{"libglfw.3.dylib", "libglfw.dylib"}
		searchPaths = []string{
			"", // "" forces a raw system check first (handles standard paths & global locations)
			"/opt/homebrew/lib",
			"/usr/local/lib",
		}
		// Method 1: Add macOS App Bundle context if we can find the executable
		if execPath, err := os.Executable(); err == nil {
			searchPaths = append(searchPaths, filepath.Join(filepath.Dir(execPath), "..", "Frameworks"))
			searchPaths = append(searchPaths, filepath.Dir(execPath)) // Sibling dir fallback
		}
	case "linux":
		names = []string{"libglfw.so.3", "libglfw.so"}
		searchPaths = []string{""} // System paths
		if execPath, err := os.Executable(); err == nil {
			searchPaths = append(searchPaths, filepath.Dir(execPath)) // Sibling dir fallback
		}
	case "windows":
		names = []string{"glfw3.dll"}
		searchPaths = []string{""} // System paths & Executable directory auto-search
	default:
		return 0, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// 2. Single unified search loop
	for _, path := range searchPaths {
		for _, name := range names {
			fullPath := name
			if path != "" {
				fullPath = filepath.Join(path, name)
			}

			// Attempt to open the resolved target path
			if handle, err := purego.Dlopen(fullPath, purego.RTLD_NOW|purego.RTLD_GLOBAL); err == nil {
				return handle, nil // Success!
			}
		}
	}

	return 0, fmt.Errorf("failed to load GLFW on %s: all paths exhausted", runtime.GOOS)
}
