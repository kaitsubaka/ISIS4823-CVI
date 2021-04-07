package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/kaitsubaka/glutils/gfx"
	"github.com/kaitsubaka/glutils/win"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width  = 1080
	height = 720
	title  = "Light"
)

var (
	vert = []string{"shaders/phong.vert", "shaders/gouraud.vert", "shaders/flat.vert"}
	frag = []string{"shaders/phong.frag", "shaders/gouraud.frag", "shaders/flat.frag"}
)

func createVAO(vertices []float32, indices []uint32) uint32 {

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)

	var EBO uint32
	gl.GenBuffers(1, &EBO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	// copy vertices data into VBO (it needs to be bound first)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// size of one whole vertex (sum of attrib sizes)
	var stride int32 = 3*4 + 3*4
	var offset int = 0

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(0)
	offset += 3 * 4

	// normal
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(1)
	offset += 3 * 4

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return VAO
}

type AnimationManager struct {
	previousTime         float64
	totalElapsed         float64
	globalAnimationCount int
	angle                float64
	animationTimes       []float64
	animationFunctions   []func(t float32)
}

func (am *AnimationManager) GetAngle() float64 {
	return am.angle
}

func NewAnimationManager() *AnimationManager {
	return &AnimationManager{}
}

func (am *AnimationManager) AddAnimation(animation func(t float32), time float64) {
	am.animationFunctions, am.animationTimes = append(am.animationFunctions, animation), append(am.animationTimes, time)
}

func (am *AnimationManager) Update() {
	time := glfw.GetTime()
	elapsed := time - am.previousTime
	am.totalElapsed += elapsed
	am.previousTime = time
	am.angle += elapsed

	if am.globalAnimationCount < len(am.animationFunctions) {
		if animationTime := am.animationTimes[am.globalAnimationCount]; am.totalElapsed < animationTime {
			t := float32(am.totalElapsed / animationTime)
			am.animationFunctions[am.globalAnimationCount](t)
		} else {
			am.animationFunctions[am.globalAnimationCount](1)
			am.totalElapsed = 0
			am.globalAnimationCount++
		}
	}
}

func (am *AnimationManager) Init() {
	am.previousTime = glfw.GetTime()
}

func programLoop(window *win.Window, shader int) error {

	// Shaders and textures
	vertShader, err := gfx.NewShaderFromFile(vert[shader], gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := gfx.NewShaderFromFile(frag[shader], gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program, err := gfx.NewProgram(vertShader, fragShader)
	if err != nil {
		return err
	}
	defer program.Delete()

	lightFragShader, err := gfx.NewShaderFromFile("shaders/light.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	// special shader program so that lights themselves are not affected by lighting
	lightProgram, err := gfx.NewProgram(vertShader, lightFragShader)
	if err != nil {
		return err
	}

	// Ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Base model
	//model := mgl32.Ident4()

	// Uniform
	modelUniformLocation := program.GetUniformLocation("model")
	viewUniformLocation := program.GetUniformLocation("view")
	projectUniformLocation := program.GetUniformLocation("projection")
	lightColorUniformLocation := program.GetUniformLocation("lightColor")
	objectColorUniformLocation := program.GetUniformLocation("objectColor")
	lightPosUniformLocation := program.GetUniformLocation("lightPos")

	modelLightUniformLocation := lightProgram.GetUniformLocation("model")
	viewLightUniformLocation := lightProgram.GetUniformLocation("view")
	projectLightUniformLocation := lightProgram.GetUniformLocation("projection")

	// creates camara
	eye := mgl32.Vec3{2, 2, 2}
	center := mgl32.Vec3{1, 1, 0}
	camera := mgl32.LookAtV(eye, center, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(viewUniformLocation, 1, false, &camera[0])

	lightPos := mgl32.Vec3{0, 2, 0}
	var lightTransform mgl32.Mat4

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

	// Uncomment to turn on polygon mode
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Scene and animation
	animationCtl := NewAnimationManager()
	VAO := createVAO(cubeVertices, nil)
	lightVAO := VAO

	animationCtl.Init()

	// main loop
	for !window.ShouldClose() {
		window.StartFrame()

		// background color
		gl.ClearColor(0, 0, 0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Scene update
		animationCtl.Update()

		lightTransform = mgl32.Translate3D(lightPos.X(), lightPos.Y(), lightPos.Z()).Mul4(mgl32.Scale3D(0.2, 0.2, 0.2))

		// You shall draw here

		program.Use()
		gl.UniformMatrix4fv(viewUniformLocation, 1, false, &camera[0])
		gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

		gl.BindVertexArray(VAO)

		// obj is colored, light is white
		gl.Uniform3f(objectColorUniformLocation, 1., .0, .2)
		gl.Uniform3f(lightColorUniformLocation, 1.0, 1.0, 1.0)
		gl.Uniform3f(lightPosUniformLocation, lightPos.X(), lightPos.Y(), lightPos.Z())

		worldTranslate := mgl32.Translate3D(0.0, 0.0, 0.0)
		worldTransform := worldTranslate.Mul4(mgl32.HomogRotate3DY(float32(animationCtl.GetAngle())).Mul4(mgl32.HomogRotate3DZ(float32(animationCtl.GetAngle()))))
		gl.UniformMatrix4fv(modelUniformLocation, 1, false, &worldTransform[0])
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		gl.BindVertexArray(0)

		// Draw the light obj after the other boxes using its separate shader program
		// this means that we must re-bind any uniforms
		lightProgram.Use()
		gl.BindVertexArray(lightVAO)
		gl.UniformMatrix4fv(modelLightUniformLocation, 1, false, &lightTransform[0])
		gl.UniformMatrix4fv(viewLightUniformLocation, 1, false, &camera[0])
		gl.UniformMatrix4fv(projectLightUniformLocation, 1, false, &projectTransform[0])
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		gl.BindVertexArray(0)
	}

	return nil
}

func main() {
	fmt.Println("Please select a shader:\n[0] Phong\n[1] Gouraund\n[2] Flat")
	var shader int
	fmt.Scanln(&shader)
	if shader > 2 {
		shader = 0
	}

	runtime.LockOSThread()
	win.InitGlfw(4, 0)
	defer glfw.Terminate()
	window := win.NewWindow(width, height, title)
	gfx.InitGl()

	err := programLoop(window, shader)
	if err != nil {
		log.Fatal(err)
	}
}
