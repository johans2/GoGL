package gui

import (
	"math"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/inkyblackness/imgui-go"
)

var glfwButtonIndexByID = map[glfw.MouseButton]int{
	glfw.MouseButton1: 0,
	glfw.MouseButton2: 1,
	glfw.MouseButton3: 2,
}

// ImguiInput is the state holder for the imgui framework
type ImguiInput struct {
	io               *imgui.IO
	time             float64
	mouseJustPressed [3]bool
}

// ImguiMouseState is provided to NewFrame(...), containing the mouse state
type ImguiMouseState struct {
	MousePosX  float32
	MousePosY  float32
	MousePress [3]bool
}

var imguiIO imgui.IO
var inputState ImguiInput

// NewImgui initializes a new imgui context and a input object
func NewImgui() (*imgui.Context, ImguiInput) {
	context := imgui.CreateContext(nil)
	imguiIO = imgui.CurrentIO()
	inputState = ImguiInput{io: &imguiIO, time: 0}
	inputState.setKeyMapping()
	return context, inputState
}

// NewFrame : Initiates a new frame for the input package
func (input *ImguiInput) NewFrame(displaySizeX float32, displaySizeY float32, time float64, isFocused bool, mouseState ImguiMouseState) {
	// Setup display size (every frame to accommodate for window resizing)
	input.io.SetDisplaySize(imgui.Vec2{X: displaySizeX, Y: displaySizeY})

	// Setup time step
	currentTime := time
	if input.time > 0 {
		input.io.SetDeltaTime(float32(currentTime - input.time))
	}
	input.time = currentTime

	// Setup inputs
	if isFocused {
		input.io.SetMousePosition(imgui.Vec2{X: mouseState.MousePosX, Y: mouseState.MousePosY})
	} else {
		input.io.SetMousePosition(imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32})
	}

	for i := 0; i < len(input.mouseJustPressed); i++ {
		down := input.mouseJustPressed[i] || mouseState.MousePress[0] == true
		input.io.SetMouseButtonDown(i, down)
		input.mouseJustPressed[i] = false
	}

	imgui.NewFrame()
}

func (input *ImguiInput) setKeyMapping() {
	// Keyboard mapping. ImGui will use those indices to peek into the input.io.KeysDown[] array.
	input.io.KeyMap(imgui.KeyTab, int(glfw.KeyTab))
	input.io.KeyMap(imgui.KeyLeftArrow, int(glfw.KeyLeft))
	input.io.KeyMap(imgui.KeyRightArrow, int(glfw.KeyRight))
	input.io.KeyMap(imgui.KeyUpArrow, int(glfw.KeyUp))
	input.io.KeyMap(imgui.KeyDownArrow, int(glfw.KeyDown))
	input.io.KeyMap(imgui.KeyPageUp, int(glfw.KeyPageUp))
	input.io.KeyMap(imgui.KeyPageDown, int(glfw.KeyPageDown))
	input.io.KeyMap(imgui.KeyHome, int(glfw.KeyHome))
	input.io.KeyMap(imgui.KeyEnd, int(glfw.KeyEnd))
	input.io.KeyMap(imgui.KeyInsert, int(glfw.KeyInsert))
	input.io.KeyMap(imgui.KeyDelete, int(glfw.KeyDelete))
	input.io.KeyMap(imgui.KeyBackspace, int(glfw.KeyBackspace))
	input.io.KeyMap(imgui.KeySpace, int(glfw.KeySpace))
	input.io.KeyMap(imgui.KeyEnter, int(glfw.KeyEnter))
	input.io.KeyMap(imgui.KeyEscape, int(glfw.KeyEscape))
	input.io.KeyMap(imgui.KeyA, int(glfw.KeyA))
	input.io.KeyMap(imgui.KeyC, int(glfw.KeyC))
	input.io.KeyMap(imgui.KeyV, int(glfw.KeyV))
	input.io.KeyMap(imgui.KeyX, int(glfw.KeyX))
	input.io.KeyMap(imgui.KeyY, int(glfw.KeyY))
	input.io.KeyMap(imgui.KeyZ, int(glfw.KeyZ))
}

// MouseButtonChange passes mouse events to the imgui framework
func (input *ImguiInput) MouseButtonChange(window *glfw.Window, rawButton glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	buttonIndex, known := glfwButtonIndexByID[rawButton]

	if known && (action == glfw.Press) {
		input.mouseJustPressed[buttonIndex] = true
	}
}

// MouseScrollChange passes mouse scrolling to the imgui framework
func (input *ImguiInput) MouseScrollChange(window *glfw.Window, x, y float64) {
	input.io.AddMouseWheelDelta(float32(x), float32(y))
}

// KeyChange passes key events to the imgui framework
func (input *ImguiInput) KeyChange(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		input.io.KeyPress(int(key))
	}
	if action == glfw.Release {
		input.io.KeyRelease(int(key))
	}

	// Modifiers are not reliable across systems
	input.io.KeyCtrl(int(glfw.KeyLeftControl), int(glfw.KeyRightControl))
	input.io.KeyShift(int(glfw.KeyLeftShift), int(glfw.KeyRightShift))
	input.io.KeyAlt(int(glfw.KeyLeftAlt), int(glfw.KeyRightAlt))
	input.io.KeySuper(int(glfw.KeyLeftSuper), int(glfw.KeyRightSuper))
}

// CharChange passes char changes to the imgui framework
func (input *ImguiInput) CharChange(window *glfw.Window, char rune) {
	input.io.AddInputCharacters(string(char))
}
