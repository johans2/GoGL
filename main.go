package main

import (
	"fmt"
	"go/build"
	_ "image/png"
	"log"
	"os"
	"runtime"

	"GoGL/gui"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/inkyblackness/imgui-go"
)

const windowWidth = 1600
const windowHeight = 900

type mouseState bool

const (
	press   mouseState = true
	release mouseState = false
)

const (
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

type state struct {
	//glContext        *nk.Context
	//bgColor          nk.Color
	rotationSpeed    float32
	scale            float32
	clearColor       mgl32.Vec4
	bufferVertSource []byte
	bufferFragSource []byte
	activeMaterial   material
	activeModel      []float32
	modelRenderer    renderer
	shaderError      error
}

type data struct {
	sphereVerts []float32
	boxVerts    []float32
	torusVerts  []float32
	planeVerts  []float32
	coneVerts   []float32
}

var reApplyUniformsa = false

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()

	fmt.Println(gui.TestString)

	// Set the working directory to the root of Go package, so that its assets can be accessed.
	dir, err := importPathToDir("GoGL")
	if err != nil {
		log.Fatalln("Unable to find Go package in your GOPATH, it's needed to load assets:", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		log.Panicln("os.Chdir:", err)
	}
}

func main() {
	/*
		window := initGLFW()
		defer glfw.Terminate()
	*/
	context := imgui.CreateContext(nil)
	defer context.Destroy()
	io := imgui.CurrentIO()

	platform, err := gui.NewGLFW(io, gui.GLFWClientAPIOpenGL3)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer platform.Dispose()

	imguiRenderer, err := gui.NewOpenGL3(io)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	defer imguiRenderer.Dispose()

	currentMouseState := release
	/*
		window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
			if button == glfw.MouseButton1 && action == glfw.Press {
				currentMouseState = press
			}

			if button == glfw.MouseButton1 && action == glfw.Release {
				currentMouseState = release
			}
		})*/

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Create state and data holders
	state := new(state)
	data := new(data)

	var shaderGreen shader
	shaderGreen.loadFromFile("Assets/simpleGreen.vert", "Assets/simpleGreen.frag")

	// Set up projection matrix for shader
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 10.0)

	// Set up view matrix for shader
	cameraPos := mgl32.Vec3{0, 2, 3}
	view := mgl32.LookAtV(cameraPos, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})

	// Set up model martix for shader
	model := mgl32.Ident4()

	// Load the model from the obj file
	sphereModel, _ := readOBJ("Assets/sphere.obj")
	boxModel, _ := readOBJ("Assets/box.obj")
	torusModel, _ := readOBJ("Assets/torus.obj")
	planeModel, _ := readOBJ("Assets/plane.obj")
	coneModel, _ := readOBJ("Assets/cone.obj")

	angle := 0.0
	previousTime := glfw.GetTime()

	log.Printf("Finished setup. Now rendering..")

	// Setup data struct
	data.sphereVerts = sphereModel.ToArrayXYZUVN1N2N3()
	data.boxVerts = boxModel.ToArrayXYZUVN1N2N3()
	data.torusVerts = torusModel.ToArrayXYZUVN1N2N3()
	data.planeVerts = planeModel.ToArrayXYZUVN1N2N3()
	data.coneVerts = coneModel.ToArrayXYZUVN1N2N3()

	// Setup initial state
	state.activeMaterial.init(shaderGreen)
	state.activeModel = data.boxVerts
	state.modelRenderer.setData(state.activeModel, state.activeMaterial)
	state.bufferVertSource = make([]byte, 1024)
	state.bufferFragSource = make([]byte, 1024)
	//state.bgColor.SetRGBAi(255, 255, 255, 255)
	state.clearColor = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
	state.rotationSpeed = float32(0.5)
	state.scale = float32(1.0)

	var mouseXPrev float64
	var mouseYPrev float64
	var mouseY float64
	var mouseX float64

	for !platform.ShouldStop() {
		platform.ProcessEvents()

		// Signal start of a new frame
		platform.NewFrame()
		imgui.NewFrame()

		// 2. Show a simple window that we create ourselves. We use a Begin/End pair to created a named window.
		{
			imgui.Begin("Hello, world!") // Create a window called "Hello, world!" and append into it.

			imgui.Text("This is some useful text.") // Display some text

			if imgui.Button("Button") { // Buttons return true when clicked (most widgets return true when edited/activated)
				fmt.Println("Button clicked!")
			}
			imgui.SameLine()
			imgui.Text(fmt.Sprintf("counter = %d", 1))

			imgui.End()
		}

		// Rendering
		imgui.Render() // This call only creates the draw data list. Actual rendering to framebuffer is done below.
		clearColor := [4]float32{0.0, 0.0, 0.0, 1.0}
		imguiRenderer.PreRender(clearColor)

		// Need to reanable these things since Nuklear sets its own gl states when rendering.
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.Enable(gl.DEPTH_TEST)
		gl.Enable(gl.CULL_FACE)
		gl.DepthFunc(gl.LESS)
		gl.ClearColor(state.clearColor.X(),
			state.clearColor.Y(),
			state.clearColor.Z(),
			state.clearColor.W())

		// Update time
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		// Update mouse rotation
		if currentMouseState == press {
			x, y := 0.0, 0.0 //window.GetCursorPos()
			mouseX = float64(x)
			mouseY = y
			deltaX := mouseX - mouseXPrev
			deltaY := mouseY - mouseYPrev
			angle += (deltaX/100 + deltaY/100)
			mouseXPrev = mouseX
			mouseYPrev = mouseY
		} else {
			angle += (elapsed * float64(state.rotationSpeed))
		}

		if reApplyUniformsa {
			state.modelRenderer.material.applyUniforms()
		}

		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})
		model = model.Mul4(mgl32.Scale3D(state.scale, state.scale, state.scale))

		// Render
		state.modelRenderer.issueDrawCall(model, view, projection, cameraPos, float32(time))

		// Maintenance
		imguiRenderer.Render(platform.DisplaySize(), platform.FramebufferSize(), imgui.RenderedDrawData())
		platform.PostRender()
		//platform.PostRender() //window.SwapBuffers()
		//platform.ProcessEvents()
		//glfw.PollEvents()
	}

	//glfw.Terminate()
}

func initGLFW() *glfw.Window {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "GoGL", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// importPathToDir resolves the absolute path from importPath.
// There doesn't need to be a valid Go package inside that import path,
// but the directory must exist.
func importPathToDir(importPath string) (string, error) {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}
