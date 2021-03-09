package main

import (
	"log"
	"runtime"

	"git.maze.io/go/math32"
	"github.com/StevenTarazona/glcore/ge"
	"github.com/StevenTarazona/glcore/gfx"
	"github.com/StevenTarazona/glcore/win"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width  = 1080
	height = 720
	title  = "Core"
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

func programLoop(window *win.Window) error {

	// Shaders and textures
	vertShader, err := gfx.NewShaderFromFile("shaders/basic.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := gfx.NewShaderFromFile("shaders/basic.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program, err := gfx.NewProgram(vertShader, fragShader)
	if err != nil {
		return err
	}
	defer program.Delete()

	// Ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	program.Use()

	// set texture0 to uniform0 in the fragment shader

	// Base model
	snowManPathModel := mgl32.Translate3D(2, 0, 2)
	model := mgl32.Ident4()

	// Uniform locations
	WorldUniformLocation := program.GetUniformLocation("world")
	colorUniformLocation := program.GetUniformLocation("objectColor")
	lightColorUniformLocation := program.GetUniformLocation("lightColor")
	cameraUniformLocation := program.GetUniformLocation("camera")
	projectUniformLocation := program.GetUniformLocation("project")
	textureUniformLocation := program.GetUniformLocation("texture")

	// creates camara
	camera := mgl32.LookAtV(mgl32.Vec3{3.5, 2.5, 5}, mgl32.Vec3{0, 1, 0}, mgl32.Vec3{0, 1, 0})
	//camera := mgl32.LookAtV(mgl32.Vec3{5, 5, 5}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(cameraUniformLocation, 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

	// creates light
	gl.Uniform3f(lightColorUniformLocation, 1, 1, 1)

	// Uncomment to turn on polygon mode
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Scene and animation
	angle := 0.0
	previousTime := glfw.GetTime()
	totalElapsed := float64(0)
	movementControlCount := 0

	movementTimes := []float64{}
	movementFunctions := []func(t float32){}

	// Textures
	leavesTexture, err := gfx.NewTextureFromFile("images/leaves.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	snowTexture, err := gfx.NewTextureFromFile("images/snow.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	snowTexture2, err := gfx.NewTextureFromFile("images/snowThrees.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	// Get primitive vertices and create VAOs
	var theVoid []mgl32.Vec2
	cubeVertices := ge.GetCubicHexahedronVertices3(1.5, 1, 1.5)
	cubeTextureCoords := ge.GetCubicHexahedronTextureCoords(1, 1, 1)
	cubeVAO := ge.CreateVAO(cubeVertices, cubeTextureCoords)

	sideVertices, topVertices, bottomVertices := ge.GetCylinderVertices3(1, 0.1, 0.1, 5)
	sideVAO, topVAO, bottomVAO := ge.CreateVAO(sideVertices, theVoid), ge.CreateVAO(topVertices, theVoid), ge.CreateVAO(bottomVertices, theVoid)

	planeVertices := ge.GetPlaneVertices3(12, 12, 1)
	planeTextureCoords := ge.GetPlaneTextureCoords(12, 12, 1)
	planeVAO := ge.CreateVAO(planeVertices, planeTextureCoords)

	sphereVertices, sphereTop, sphereBottom := ge.GetSphereVertices3(0.3, 16)
	sphereVao, sphereTopVao, sphereBotVao := ge.CreateVAO(sphereVertices, theVoid), ge.CreateVAO(sphereTop, theVoid), ge.CreateVAO(sphereBottom, theVoid)

	noseVertices := ge.GetCircleVertices3(0.05, 8)
	noseVertices[0] = mgl32.Vec3{0, 0.2, 0}
	noseVao := ge.CreateVAO(noseVertices, theVoid)

	snowCarpetVertices := ge.GetCubicHexahedronVertices3(1.5, 0.1, 1.5)
	snowCarpetTextureCoords := ge.GetCubicHexahedronTextureCoords(1, 1, 1)
	snowCarpetVAO := ge.CreateVAO(snowCarpetVertices, snowCarpetTextureCoords)

	sideVerticesHat, topVerticesHat, bottomVerticesHat := ge.GetCylinderVertices3(0.5, 0.5, 0.5, 5)
	sideHatVAO, topHatVAO, bottomHatVAO := ge.CreateVAO(sideVerticesHat, theVoid), ge.CreateVAO(topVerticesHat, theVoid), ge.CreateVAO(bottomVerticesHat, theVoid)

	for !window.ShouldClose() {
		window.StartFrame()

		// background color
		gl.ClearColor(0, 0.27, 0.7, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		time := glfw.GetTime()
		elapsed := time - previousTime
		totalElapsed += elapsed
		previousTime = time
		angle += elapsed

		// Scene update
		if movementControlCount < len(movementFunctions) {
			if animationTime := movementTimes[movementControlCount]; totalElapsed <= animationTime {
				t := float32(totalElapsed / animationTime)
				movementFunctions[movementControlCount](t)
			} else {
				movementFunctions[movementControlCount](1)
				totalElapsed = 0
				movementControlCount++
			}
		}

		// You shall draw here
		for _, pos := range treePositions {
			treeTranslate := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())
			scale1 := 1 - math32.Abs(math32.Sin(float32(time)))*0.04
			scale2 := 1 - math32.Abs(math32.Cos(float32(time)))*0.04
			//log
			gl.Uniform3f(colorUniformLocation, 0.4, 0.2, 0)
			gl.BindVertexArray(sideVAO)
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideVertices)))

			gl.BindVertexArray(topVAO)
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(topVertices)))

			gl.BindVertexArray(bottomVAO)
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(bottomVertices)))

			//leaves

			//big leavesTexture
			leavesTexture.Bind(gl.TEXTURE0)
			leavesTexture.SetUniform(textureUniformLocation)

			gl.Uniform3f(colorUniformLocation, 1, 1, 1)
			gl.BindVertexArray(cubeVAO)
			treeTranslate = treeTranslate.Mul4(mgl32.Translate3D(0, 1.*scale1, 0))
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
			gl.BindVertexArray(0)
			leavesTexture.UnBind()
			// snow
			snowTexture2.Bind(gl.TEXTURE0)
			snowTexture2.SetUniform(textureUniformLocation)

			gl.Uniform3f(colorUniformLocation, 1, 1, 1)
			gl.BindVertexArray(snowCarpetVAO)
			snowTranslate := treeTranslate.Mul4(mgl32.Translate3D(0, 1, 0))
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(snowCarpetVertices)))
			gl.BindVertexArray(0)
			snowTexture2.UnBind()
			//med
			leavesTexture.Bind(gl.TEXTURE0)
			leavesTexture.SetUniform(textureUniformLocation)
			gl.Uniform3f(colorUniformLocation, 1, 1, 1)
			gl.BindVertexArray(cubeVAO)
			treeTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75)).Mul4(mgl32.Translate3D(0, 0.75*scale2, 0))
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
			gl.BindVertexArray(0)
			leavesTexture.UnBind()
			// snow med
			snowTexture2.Bind(gl.TEXTURE0)
			snowTexture2.SetUniform(textureUniformLocation)
			gl.Uniform3f(colorUniformLocation, 1, 1, 1)
			gl.BindVertexArray(snowCarpetVAO)
			snowTranslate = treeTranslate.Mul4(mgl32.Translate3D(0, 1, 0))
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(snowCarpetVertices)))
			gl.BindVertexArray(0)
			snowTexture2.UnBind()

			//smol
			leavesTexture.Bind(gl.TEXTURE0)
			leavesTexture.SetUniform(textureUniformLocation)

			gl.Uniform3f(colorUniformLocation, 1, 1, 1)
			gl.BindVertexArray(cubeVAO)
			treeTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5)).Mul4(mgl32.Translate3D(0, 1.7*scale1, 0))
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
			gl.BindVertexArray(0)
			leavesTexture.UnBind()
			// snow smol
			snowTexture2.Bind(gl.TEXTURE0)
			snowTexture2.SetUniform(textureUniformLocation)

			gl.Uniform3f(colorUniformLocation, 1, 1, 1)
			gl.BindVertexArray(snowCarpetVAO)
			snowTranslate = treeTranslate.Mul4(mgl32.Translate3D(0, 1, 0))
			gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(snowCarpetVertices)))
			gl.BindVertexArray(0)
			snowTexture2.UnBind()

		}

		snowmanTranslate := snowManPathModel
		gl.Uniform3f(colorUniformLocation, 1, 1, 1)
		// fist sphere

		gl.BindVertexArray(sphereTopVao)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereTop)))

		gl.BindVertexArray(sphereVao)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereVertices)))

		gl.BindVertexArray(sphereBotVao)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereBottom)))
		gl.BindVertexArray(0)
		//secodn sphere

		snowmanTranslate = snowmanTranslate.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75)).Mul4(mgl32.Translate3D(0, 0.6, 0))
		gl.BindVertexArray(sphereTopVao)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereTop)))

		gl.BindVertexArray(sphereVao)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereVertices)))

		gl.BindVertexArray(sphereBotVao)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereBottom)))
		gl.BindVertexArray(0)

		// head

		snowmanTranslate = snowmanTranslate.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75)).Mul4(mgl32.Translate3D(0, 0.65, 0))
		gl.BindVertexArray(sphereTopVao)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereTop)))

		gl.BindVertexArray(sphereVao)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sphereVertices)))

		gl.BindVertexArray(sphereBotVao)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(sphereBottom)))
		gl.BindVertexArray(0)

		// nose
		snowmanNoseTranslate := snowmanTranslate.Mul4(mgl32.Translate3D(0, 0.3, 0.25)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(90)))
		gl.Uniform3f(colorUniformLocation, 1, 0.541, 0.380)
		gl.BindVertexArray(noseVao)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanNoseTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(noseVertices)))
		gl.BindVertexArray(0)

		snowmanHatTranslate := snowmanTranslate.Mul4(mgl32.Translate3D(0, 0.5, 0))
		gl.Uniform3f(colorUniformLocation, 0, 0, 0)
		gl.BindVertexArray(bottomHatVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanHatTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(bottomVerticesHat)))
		gl.BindVertexArray(0)

		snowmanHatTranslate = snowmanHatTranslate.Mul4(mgl32.Scale3D(0.55, 1, 0.55))
		gl.BindVertexArray(sideHatVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanHatTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideVerticesHat)))
		gl.BindVertexArray(0)

		gl.BindVertexArray(topHatVAO)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &snowmanHatTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(topVerticesHat)))
		gl.BindVertexArray(0)

		// plane

		gl.BindVertexArray(planeVAO)
		snowTexture.Bind(gl.TEXTURE0)
		snowTexture.SetUniform(textureUniformLocation)
		gl.Uniform3f(colorUniformLocation, 1, 1, 1)
		gl.UniformMatrix4fv(WorldUniformLocation, 1, false, &model[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(planeVertices)))
		snowTexture.UnBind()

		gl.BindVertexArray(0)
	}

	return nil
}

func main() {
	runtime.LockOSThread()

	win.InitGlfw(4, 1)
	defer glfw.Terminate()
	window := win.NewWindow(width, height, title)
	gfx.InitGl()

	err := programLoop(window)
	if err != nil {
		log.Fatal(err)
	}
}
