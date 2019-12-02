package gui

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/inkyblackness/imgui-go"
)

var InputString string = "Lets move input to this package!"

var glfwButtonIndexByID_2 = map[glfw.MouseButton]int{
	glfw.MouseButton1: 0,
	glfw.MouseButton2: 1,
	glfw.MouseButton3: 2,
}

var glfwButtonIDByIndex_2 = map[int]glfw.MouseButton{
	0: glfw.MouseButton1,
	1: glfw.MouseButton2,
	2: glfw.MouseButton3,
}

type ImguiInput struct {
	io *imgui.IO
}

func (input *ImguiInput) SetKeyMapping() {
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
