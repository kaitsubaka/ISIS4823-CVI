package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"git.maze.io/go/math32"
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width              = 1080
	height             = 720
	vertexShaderSource = `
		#version 410 core
		layout (location = 0) in vec3 position;
		out vec2 TexCoord;
		uniform mat4 world;
		uniform mat4 camera;
		uniform mat4 project;
		void main()
		{
			gl_Position = project * camera * world * vec4(position, 1.0);
		}`

	fragmentShaderSource = `
		#version 410 core
		out vec4 color;
		uniform vec3 objectColor;
		uniform vec3 lightColor;
		void main()
		{
			color = vec4(objectColor * lightColor, 1.0f);
		}`
)

var (
	treePositions = []mgl32.Vec3{
		{-2.5, 0, -0.5},
		{-2, 0, 1.5},
		{-2, 0, -2.4},
		{-1, 0, -1},
		{-1, 0, 1},
		{1, 0, 1},
		{2, 0, -1.5},
	}
)

func getSphereVertices(r float32, numVertex int, slices int) (sphere []mgl32.Vec3, top []mgl32.Vec3, bottom []mgl32.Vec3) {
	sliceHeight := r / float32(slices)
	auxHeight := r - sliceHeight
	auxR := math32.Sqrt(math32.Pow(r, 2) - math32.Pow(auxHeight, 2))
	auxNextRHeight := auxHeight - sliceHeight
	auxNextR := math32.Sqrt(math32.Pow(r, 2) - math32.Pow(auxNextRHeight, 2))
	bottom = getCircleVertices(0, r-auxHeight, 0, auxR, numVertex)

	bottom[0] = mgl32.Vec3{0, 0, 0}
	top = getCircleVertices(0, 2*r-sliceHeight, 0, auxR, numVertex)
	top[0] = mgl32.Vec3{0, 2 * r, 0}
	for i := 1; i < slices; i++ {
		currentCicle := getCircleVertices(0, r-auxHeight, 0, auxR, numVertex)
		nextCicle := getCircleVertices(0, r-auxNextRHeight, 0, auxNextR, numVertex)
		for i := 1; i < len(currentCicle); i++ {
			sphere = append(sphere, currentCicle[i])
			sphere = append(sphere, nextCicle[i])

		}
		auxHeight = auxNextRHeight
		auxR = auxNextR
		auxNextRHeight = auxHeight - sliceHeight
		auxNextR = math32.Sqrt(math32.Pow(r, 2) - math32.Pow(auxNextRHeight, 2))
	}

	semiSphereLen := len(sphere)
	for i := semiSphereLen - 2; i > 0; i-- {
		sphere = append(sphere, mgl32.Vec3{sphere[i].X(), -sphere[i].Y() + 2*r, sphere[i].Z()})
	}

	return
}

func getCircleVertices(x float32, y float32, z float32, r float32, vertices int) []mgl32.Vec3 {
	var sectorStep = 2 * math32.Pi / float32(vertices)
	var sectorAngle float32 // radian
	var unitCircleVertices = []mgl32.Vec3{{x, y, z}}
	for i := 0; i <= vertices; i++ {
		sectorAngle = float32(i) * sectorStep
		unitCircleVertices = append(unitCircleVertices, mgl32.Vec3{x + r*math32.Cos(sectorAngle), y, z + r*math32.Sin(sectorAngle)})
	}
	return unitCircleVertices
}

func getCylinderVertices(h float32, r float32, vertices int) (side []mgl32.Vec3, top []mgl32.Vec3, bottom []mgl32.Vec3) {
	var slices int
	if slices = int(h / (math32.Sin(2*math32.Pi/float32(vertices)) * r)); slices == 0 {
		slices = 1
	}
	top = getCircleVertices(0, h, 0, r, vertices)
	bottom = getCircleVertices(0, 0, 0, r, vertices)
	for slice := 0; slice < slices; slice++ {
		for i := 1; i < len(bottom); i++ {
			side = append(side, mgl32.Vec3{bottom[i].X(), bottom[i].Y() + float32(slice)*(h/float32(slices)), bottom[i].Z()})
			side = append(side, mgl32.Vec3{bottom[i].X(), bottom[i].Y() + float32(slice+1)*(h/float32(slices)), bottom[i].Z()})
		}
	}
	return side, top, bottom
}

func getPlaneVertices(h int, w int, l int) []mgl32.Vec3 {
	var vertices []mgl32.Vec3
	var i = -1
	for row := -h / 2; row < h/2; row += l {
		i *= -1
		for col := -w / 2; col <= w/2; col += l {
			vertices = append(vertices, mgl32.Vec3{float32(col * i), 0, float32(row)})
			vertices = append(vertices, mgl32.Vec3{float32(col * i), 0, float32(row + l)})
		}
	}
	return vertices
}

func getCubeVertices(x float32) []mgl32.Vec3 {
	var vertices = []mgl32.Vec3{
		{-x / 2, x, x / 2},   // Front-top-left
		{x / 2, x, x / 2},    // Front-top-right
		{-x / 2, 0, x / 2},   // Front-bottom-left
		{x / 2, 0, x / 2},    // Front-bottom-right
		{x / 2, 0, -x / 2},   // Back-bottom-right
		{x / 2, x, x / 2},    // Front-top-right
		{x / 2, x, -x / 2},   // Back-top-right
		{-x / 2, x, x / 2},   // Front-top-left
		{-x / 2, x, -x / 2},  // Back-top-left
		{-x / 2, -0, x / 2},  // Front-bottom-left
		{-x / 2, -0, -x / 2}, // Back-bottom-left
		{x / 2, -0, -x / 2},  // Back-bottom-right
		{-x / 2, x, -x / 2},  // Back-top-left
		{x / 2, x, -x / 2},   // Back-top-right
	}
	return vertices
}
func compileShader(src string, sType uint32) (uint32, error) {
	shader := gl.CreateShader(sType)
	glSrc, freeFn := gl.Strs(src + "\x00")
	defer freeFn()
	gl.ShaderSource(shader, 1, glSrc, nil)
	gl.CompileShader(shader)
	err := getGlError(shader, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog,
		"SHADER::COMPILE_FAILURE::")
	if err != nil {
		return 0, err
	}
	return shader, nil
}

type getObjIv func(uint32, uint32, *int32)
type getObjInfoLog func(uint32, int32, *int32, *uint8)

func getGlError(glHandle uint32, checkTrueParam uint32, getObjIvFn getObjIv,
	getObjInfoLogFn getObjInfoLog, failMsg string) error {

	var success int32
	getObjIvFn(glHandle, checkTrueParam, &success)

	if success == gl.FALSE {
		var logLength int32
		getObjIvFn(glHandle, gl.INFO_LOG_LENGTH, &logLength)

		log := gl.Str(strings.Repeat("\x00", int(logLength)))
		getObjInfoLogFn(glHandle, logLength, nil, log)

		return fmt.Errorf("%s: %s", failMsg, gl.GoStr(log))
	}

	return nil
}

/*
 * Creates the Vertex Array Object for a triangle.
 * indices is leftover from earlier samples and not used here.
 */
func createVAO(vertices []mgl32.Vec3) uint32 {

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
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4*3, gl.Ptr(vertices), gl.STATIC_DRAW)

	// size of one whole vertex (sum of attrib sizes)
	var stride int32 = 3 * 4
	var offset int = 0

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(0)
	offset += 3 * 4

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return VAO
}

func programLoop(window *glfw.Window) error {

	// the linked shader program determines how the data will be rendered
	vertShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertShader)
	gl.AttachShader(program, fragShader)
	gl.LinkProgram(program)
	err = getGlError(program, gl.LINK_STATUS, gl.GetProgramiv, gl.GetProgramInfoLog,
		"PROGRAM::LINKING_FAILURE")
	if err != nil {
		return err
	}
	defer gl.DeleteShader(program)

	// ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.UseProgram(program)

	// creates camara
	camera := mgl32.LookAtV(mgl32.Vec3{3, 2, 5}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("camera\x00")), 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width/height), 0.1, 100.0)
	gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("project\x00")), 1, false, &projectTransform[0])

	// light
	gl.Uniform3f(gl.GetUniformLocation(program, gl.Str("lightColor\x00")), 1, 1, 1)

	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("world\x00"))
	colorModel := gl.GetUniformLocation(program, gl.Str("objectColor\x00"))

	cubeVertices := getCubeVertices(1)
	cubeVAO := createVAO(cubeVertices)

	sideVertices, topVertices, bottomVertices := getCylinderVertices(1, 0.1, 5)
	sideVAO, topVAO, bottomVAO := createVAO(sideVertices), createVAO(topVertices), createVAO(bottomVertices)

	planeVertices := getPlaneVertices(10, 10, 1)
	planeVAO := createVAO(planeVertices)

	sphereVertices, sphereTop, sphereBottom := getSphereVertices(0.3, 16, 8)

	sphereVao, sphereTopVao, sphereBotVao := createVAO(sphereVertices), createVAO(sphereTop), createVAO(sphereBottom)

	noseVertices := getCircleVertices(0, 0, 0, 0.05, 8)
	noseVertices[0] = mgl32.Vec3{0, 0.2, 0}
	noseVao := createVAO(noseVertices)

	for !window.ShouldClose() {

		// update events
		window.SwapBuffers()
		glfw.PollEvents()

		// background color
		gl.ClearColor(0.380, 0.435, 1, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT) // depth buffer needed for DEPTH_TEST

		for _, pos := range treePositions {
			treeTranslate := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())

			//log
			gl.Uniform3f(colorModel, 0.4, 0.2, 0)
			gl.BindVertexArray(sideVAO)
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideVertices)))

			gl.BindVertexArray(topVAO)
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(topVertices)))

			gl.BindVertexArray(bottomVAO)
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(bottomVertices)))

			//leaves

			//big
			gl.Uniform3f(colorModel, 0.196, 0.905, 0.694)
			gl.BindVertexArray(cubeVAO)
			treeTranslate = treeTranslate.Mul4(mgl32.Translate3D(0, 1., 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
			// snow
			gl.Uniform3f(colorModel, 0.713, 0.925, 0.917)
			snowTranslate := treeTranslate.Mul4(mgl32.Scale3D(1, 0.05, 1)).Mul4(mgl32.Translate3D(0, 20, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &snowTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

			//med
			gl.Uniform3f(colorModel, 0.392, 0.929, 0.768)
			treeTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75)).Mul4(mgl32.Translate3D(0, 0.75, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
			// snow
			gl.Uniform3f(colorModel, 0.713, 0.925, 0.917)
			snowTranslate = treeTranslate.Mul4(mgl32.Scale3D(1, 0.05, 1)).Mul4(mgl32.Translate3D(0, 20, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &snowTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

			//smol
			gl.Uniform3f(colorModel, 0.552, 0.886, 0.788)
			treeTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5)).Mul4(mgl32.Translate3D(0, 1.7, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(treeTranslate)))
			// snow
			gl.Uniform3f(colorModel, 0.713, 0.925, 0.917)
			snowTranslate = treeTranslate.Mul4(mgl32.Scale3D(1, 0.05, 1)).Mul4(mgl32.Translate3D(0, 20, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &snowTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))

		}
		snowmanTranslate := mgl32.Translate3D(2, 0, 2)
		gl.Uniform3f(colorModel, 1, 1, 1)
		// fist sphere

		gl.BindVertexArray(sphereTopVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereTop)))

		gl.BindVertexArray(sphereVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereVertices)))

		gl.BindVertexArray(sphereBotVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereBottom)))
		gl.BindVertexArray(0)
		//secodn sphere

		snowmanTranslate = snowmanTranslate.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75)).Mul4(mgl32.Translate3D(0, 0.6, 0))
		gl.BindVertexArray(sphereTopVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereTop)))

		gl.BindVertexArray(sphereVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereVertices)))

		gl.BindVertexArray(sphereBotVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereBottom)))
		gl.BindVertexArray(0)

		// head

		snowmanTranslate = snowmanTranslate.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75)).Mul4(mgl32.Translate3D(0, 0.7, 0))
		gl.BindVertexArray(sphereTopVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereTop)))

		gl.BindVertexArray(sphereVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereVertices)))

		gl.BindVertexArray(sphereBotVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereBottom)))
		gl.BindVertexArray(0)

		// nose
		snowmanNoseTranslate := snowmanTranslate.Mul4(mgl32.Translate3D(0, 0.2, 0.3)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(90)))
		gl.Uniform3f(colorModel, 1, 0.541, 0.380)
		gl.BindVertexArray(noseVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanNoseTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(noseVertices)))
		gl.BindVertexArray(0)

		// hat

		snowmanHatTranslate := snowmanTranslate.Mul4(mgl32.Scale3D(5, 1, 5)).Mul4(mgl32.Translate3D(0, 0.5, 0.))
		gl.Uniform3f(colorModel, 0, 0, 0)
		gl.BindVertexArray(bottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanHatTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(bottomVertices)))

		snowmanHatTranslate = snowmanHatTranslate.Mul4(mgl32.Scale3D(0.55, 0.5, 0.55))
		gl.BindVertexArray(sideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanHatTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideVertices)))
		gl.BindVertexArray(0)
		//floor

		gl.Uniform3f(colorModel, 0.713, 0.925, 0.917)
		gl.BindVertexArray(planeVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(planeVertices)))
		gl.BindVertexArray(0)
	}

	return nil
}

func main() {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to inifitialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 0)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Scene", nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow (go function bindings)
	if err := gl.Init(); err != nil {
		log.Fatalln("failed to initialize gl bindings:", err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	err = programLoop(window)
	if err != nil {
		log.Fatalln(err)
	}
}
