package main

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"git.maze.io/go/math32"
	"github.com/kaitsubaka/glutils/gfx"
	"github.com/kaitsubaka/glutils/win"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width  = 1080
	height = 720
	title  = "Textured scene"
)

var (
	bezierPoints = []mgl32.Vec3{
		{45, 30, -10},
		{40, 40, 0},
		{-45, 30, 10},
		//{-5, 5, 10},
	}
	pointLightPositions = []mgl32.Vec3{
		{5, 5, 0},
		{-5, 5, 0},
		{-15, 20, -15},
		{-0.2, 12, 0},
		{40, 20, 0},
		//{-5, 5, 0},
	}
	pointLightColors = []mgl32.Vec3{
		{0.933, 0.486, 0.486},
		{0.486, 0.933, 0.929},
		{1, 1, 1},
		{0.792, 0.792, 0.725},
		{1, 1, 1},
	}
	pointLightColorsRef = []mgl32.Vec3{
		{0.839, 0.298, 0.337},
		{0.505, 0.549, 0.854},
		{1, 1, 1},
		{0.792, 0.792, 0.725},
		{1, 1, 1},
	}
)

func turnStar(im *win.InputManager, colorNum int, changeColor bool) (mgl32.Vec3, int, bool) {
	colors := []mgl32.Vec3{
		pointLightColorsRef[3],
		{0.160, 0.160, 0.160},
	}

	if im.IsActive(win.PLAYER_SWITCH) && changeColor {

		colorNum = (colorNum + 1) % 2
		pointLightColors[3] = colors[colorNum]
		return colors[colorNum], colorNum, false

	} else {
		return colors[colorNum], colorNum, true
	}

}

func createVAO(vertices, normals, tCoords []float32, indices []uint32) uint32 {

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.BindVertexArray(VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	var NBO uint32
	gl.GenBuffers(1, &NBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, NBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(normals)*4, gl.Ptr(normals), gl.STATIC_DRAW)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	if len(tCoords) > 0 {
		var TBO uint32
		gl.GenBuffers(1, &TBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, TBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(tCoords)*4, gl.Ptr(tCoords), gl.STATIC_DRAW)
		gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))
		gl.EnableVertexAttribArray(2)
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	}

	var EBO uint32
	gl.GenBuffers(1, &EBO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.BindVertexArray(0)

	return VAO
}

func pointLightsUniformLocations(program *gfx.Program) [][]int32 {
	uniformLocations := [][]int32{}
	for i := 0; i < len(pointLightPositions); i++ {
		uniformLocations = append(uniformLocations,
			[]int32{program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].position")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].ambient")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].diffuse")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].specular")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].constant")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].linear")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].quadratic")),
				program.GetUniformLocation(fmt.Sprint("pointLights[", i, "].lightColor"))})
	}
	return uniformLocations
}

func programLoop(window *win.Window) error {

	// Shaders and textures
	vertShader, err := gfx.NewShaderFromFile("shaders/phong_ml.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	fragShader, err := gfx.NewShaderFromFile("shaders/phong_ml.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program, err := gfx.NewProgram(vertShader, fragShader)
	if err != nil {
		return err
	}
	defer program.Delete()

	sourceVertShader, err := gfx.NewShaderFromFile("shaders/source.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	sourceFragShader, err := gfx.NewShaderFromFile("shaders/source.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	// special shader program so that lights themselves are not affected by lighting
	sourceProgram, err := gfx.NewProgram(sourceVertShader, sourceFragShader)
	if err != nil {
		return err
	}

	// Ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Base model
	model := mgl32.Ident4()

	// Uniform
	modelUniformLocation := program.GetUniformLocation("model")
	viewUniformLocation := program.GetUniformLocation("view")
	projectUniformLocation := program.GetUniformLocation("projection")
	objectColorUniformLocation := program.GetUniformLocation("objectColor")
	viewPosUniformLocation := program.GetUniformLocation("viewPos")
	numLightsUniformLocation := program.GetUniformLocation("numLights")
	textureUniformLocation := program.GetUniformLocation("texSampler")
	texture2UniformLocation := program.GetUniformLocation("texSampler2")

	modelSourceUniformLocation := sourceProgram.GetUniformLocation("model")
	viewSourceUniformLocation := sourceProgram.GetUniformLocation("view")
	projectSourceUniformLocation := sourceProgram.GetUniformLocation("projection")
	objectColorSourceUniformLocation := sourceProgram.GetUniformLocation("objectColor")
	texSampler3SourceUniformLocation := sourceProgram.GetUniformLocation("texSampler3")

	pointLightsUniformLocations := pointLightsUniformLocations(program)

	// creates camara
	eye := mgl32.Vec3{0, 10, 15}
	//center := mgl32.Vec3{0, 2, 0}
	camera := NewFpsCamera(eye, mgl32.Vec3{0, -1, 0}, -90, 0, window.InputManager())

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

	// Textures
	snowTexture, err := gfx.NewTextureFromFile("textures/snow.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	logTexture, err := gfx.NewTextureFromFile("textures/Bark.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	leavesTexture, err := gfx.NewTextureFromFile("textures/leaves2.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	decoratorLeavesTexture, err := gfx.NewTextureFromFile("textures/decorator.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	moonTexture, err := gfx.NewTextureFromFile("textures/moon.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	discoBall, err := gfx.NewTextureFromFile("textures/discoBall.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	starsTexture, err := gfx.NewTextureFromFile("textures/sky.jpg",
		gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	// Colors
	objectColor := mgl32.Vec3{1., 0., 1.}
	backgroundColor := mgl32.Vec3{0.2, 0.2, 0.2}
	lightColor := mgl32.Vec3{1, 0.95, 0.75}

	// Uncomment to turn on polygon mode
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Geometry

	// Generation
	verticesPlane, normalsPlane, tCoordsPlane, indicesPlane := Square(20, 20, 2)
	verticesSpere, normalsSpere, tCoordsSpere, indicesSpere := Sphere(15, 15)
	verticesCone, normalsCone, tCoordsCone, indicesCone := Cone(15, 15, 10)
	verticesCylinder, normalsCylinder, tCoordsCylinder, indicesCylinder := Cylinder(15, 15, 10)

	// models
	logModel := mgl32.Ident4()

	// Buffers
	cylinderVAO := createVAO(verticesCylinder, normalsCylinder, tCoordsCylinder, indicesCylinder)
	planeVAO := createVAO(verticesPlane, normalsPlane, tCoordsPlane, indicesPlane)
	coneVAO := createVAO(verticesCone, normalsCone, tCoordsCone, indicesCone)
	lightVAO := createVAO(verticesSpere, normalsSpere, tCoordsSpere, indicesSpere)
	skyVAO := createVAO(Cube(80, 80, 80))
	// Scene and animation always needs to be after the model and buffers initialization
	animationCtl := gfx.NewAnimationManager()
	lightColor, numColor, changeColor := turnStar(window.InputManager(), 0, true)
	r := float32(5)
	count := 0
	freq := 20
	change := false
	animationCtl.AddContunuousAnimation(func() {

		pointLightPositions[0] = mgl32.Vec3{r * math32.Cos(float32(animationCtl.GetAngle())), 8, r * math32.Sin(float32(animationCtl.GetAngle()))}
		pointLightPositions[1] = mgl32.Vec3{-r * math32.Cos(float32(animationCtl.GetAngle())), 8, -r * math32.Sin(float32(animationCtl.GetAngle()))}

		if count%freq == 0 {
			change = !change
		}
		for index := 0; index < len(pointLightColors)-3; index++ {
			if change {
				pointLightColors[index] = pointLightColorsRef[index]

			} else {
				pointLightColors[index] = mgl32.Vec3{0, 0, 0}
			}

		}
		count++

	})

	animationCtl.AddAnimation(func(t float32) {
		pointLightPositions[4] = mgl32.BezierCurve3D(t, bezierPoints)
	}, 2)

	animationCtl.AddAnimation(func(t float32) {
		if t == 1 {
			animationCtl.GlobalAnimationCount = -1
		}
	}, 2)

	animationCtl.Init() // always needs to be before the main loop in order to get correct times
	// main loop
	for !window.ShouldClose() {
		window.StartFrame()

		// background color
		gl.ClearColor(backgroundColor.X(), backgroundColor.Y(), backgroundColor.Z(), 1.)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Scene update
		animationCtl.Update()
		camera.Update(window.SinceLastFrame())
		camTransform := camera.GetTransform()

		if window.InputManager().IsActive(win.PLAYER_UP) {
			if freq <= 3000 {
				freq += 1
			}
		} else if window.InputManager().IsActive(win.PLAYER_DOWN) {
			if freq > 1 {
				freq -= 1
			}
		}
		lightColor, numColor, changeColor = turnStar(window.InputManager(), numColor, changeColor)

		// You shall draw here
		program.Use()
		gl.UniformMatrix4fv(viewUniformLocation, 1, false, &camTransform[0])
		gl.UniformMatrix4fv(projectUniformLocation, 1, false, &projectTransform[0])

		gl.Uniform3fv(viewPosUniformLocation, 1, &eye[0])
		gl.Uniform3f(objectColorUniformLocation, objectColor.X(), objectColor.Y(), objectColor.Z())
		// gl.Uniform3f(lightColorUniformLocation, lightColor.X(), lightColor.Y(), lightColor.Z())
		gl.Uniform1i(numLightsUniformLocation, int32(len(pointLightPositions)))

		//luces
		for index, pointLightPosition := range pointLightPositions {

			if index == 2 || index == 4 {
				gl.Uniform3fv(pointLightsUniformLocations[index][0], 1, &pointLightPosition[0])
				gl.Uniform3f(pointLightsUniformLocations[index][1], (backgroundColor.X()+0.2)*0.5, (backgroundColor.Y()+0.2)*0.5, (backgroundColor.Z()+0.2)*0.5)
				gl.Uniform3f(pointLightsUniformLocations[index][2], 25, 25, 25)
				gl.Uniform3f(pointLightsUniformLocations[index][3], 1., 1., 1.)
				gl.Uniform1f(pointLightsUniformLocations[index][4], 1.)
				gl.Uniform1f(pointLightsUniformLocations[index][5], 0.09)
				gl.Uniform1f(pointLightsUniformLocations[index][6], 0.032)
				gl.Uniform3f(pointLightsUniformLocations[index][7], pointLightColors[index].X(), pointLightColors[index].Y(), pointLightColors[index].Z())
			} else if index == 3 {
				gl.Uniform3fv(pointLightsUniformLocations[index][0], 1, &pointLightPosition[0])
				gl.Uniform3f(pointLightsUniformLocations[index][1], (backgroundColor.X()+0.2)*0.5, (backgroundColor.Y()+0.2)*0.5, (backgroundColor.Z()+0.2)*0.5)
				gl.Uniform3f(pointLightsUniformLocations[index][2], 10, 10, 10)
				gl.Uniform3f(pointLightsUniformLocations[index][3], 1., 1., 1.)
				gl.Uniform1f(pointLightsUniformLocations[index][4], 1.)
				gl.Uniform1f(pointLightsUniformLocations[index][5], 0.09)
				gl.Uniform1f(pointLightsUniformLocations[index][6], 0.032)
				gl.Uniform3f(pointLightsUniformLocations[index][7], pointLightColors[index].X(), pointLightColors[index].Y(), pointLightColors[index].Z())
			} else {
				gl.Uniform3fv(pointLightsUniformLocations[index][0], 1, &pointLightPosition[0])
				gl.Uniform3f(pointLightsUniformLocations[index][1], (backgroundColor.X()+0.2)*0.5, (backgroundColor.Y()+0.2)*0.5, (backgroundColor.Z()+0.2)*0.5)
				gl.Uniform3f(pointLightsUniformLocations[index][2], 5, 5, 5)
				gl.Uniform3f(pointLightsUniformLocations[index][3], 1., 1., 1.)
				gl.Uniform1f(pointLightsUniformLocations[index][4], 1.)
				gl.Uniform1f(pointLightsUniformLocations[index][5], 0.09)
				gl.Uniform1f(pointLightsUniformLocations[index][6], 0.032)
				gl.Uniform3f(pointLightsUniformLocations[index][7], pointLightColors[index].X(), pointLightColors[index].Y(), pointLightColors[index].Z())
			}

		}

		// render models
		gl.BindVertexArray(planeVAO)
		snowTexture.Bind(gl.TEXTURE0)
		snowTexture.SetUniform(textureUniformLocation)

		boxModel := model

		gl.UniformMatrix4fv(modelUniformLocation, 1, false, &boxModel[0])
		gl.DrawElements(gl.TRIANGLES, int32(len(indicesPlane))*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		snowTexture.UnBind()
		gl.BindVertexArray(0)

		// log
		logModelTransform := logModel.Mul4(mgl32.Scale3D(1, 3, 1))
		gl.BindVertexArray(cylinderVAO)
		logTexture.Bind(gl.TEXTURE0)
		logTexture.SetUniform(textureUniformLocation)
		gl.UniformMatrix4fv(modelUniformLocation, 1, false, &logModelTransform[0])
		gl.DrawElements(gl.TRIANGLES, int32(len(indicesCylinder))*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		logTexture.UnBind()
		gl.BindVertexArray(0)
		// leave 1
		leaveModelTranslate := logModelTransform.Mul4(mgl32.Scale3D(4, 1, 4).Mul4(mgl32.Translate3D(0, 1, 0)))
		leaveOneModel := leaveModelTranslate.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(10), mgl32.Vec3{0, 0, 1})).Mul4(mgl32.Scale3D(1, 1.5, 1))
		gl.BindVertexArray(coneVAO)
		leavesTexture.Bind(gl.TEXTURE0)
		leavesTexture.SetUniform(textureUniformLocation)
		decoratorLeavesTexture.Bind(gl.TEXTURE1)
		decoratorLeavesTexture.SetUniform(texture2UniformLocation)
		gl.UniformMatrix4fv(modelUniformLocation, 1, false, &leaveOneModel[0])
		gl.DrawElements(gl.TRIANGLES, int32(len(indicesCone))*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))

		// leave 2
		leaveModelTranslate = leaveModelTranslate.Mul4(mgl32.Translate3D(0, 0.7, 0)).Mul4(mgl32.Scale3D(0.8, 1, 0.8))
		leaveTwoModel := leaveModelTranslate.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(-6), mgl32.Vec3{0, 0, 1})).Mul4(mgl32.Scale3D(1, 1.3, 1))

		gl.UniformMatrix4fv(modelUniformLocation, 1, false, &leaveTwoModel[0])
		gl.DrawElements(gl.TRIANGLES, int32(len(indicesCone))*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		// leave 3
		leaveModelTranslate = leaveModelTranslate.Mul4(mgl32.Translate3D(0, 0.6, 0)).Mul4(mgl32.Scale3D(0.8, 1, 0.8))
		leaveThreeModel := leaveModelTranslate.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(5), mgl32.Vec3{0, 0, 1}))

		gl.UniformMatrix4fv(modelUniformLocation, 1, false, &leaveThreeModel[0])
		gl.DrawElements(gl.TRIANGLES, int32(len(indicesCone))*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		decoratorLeavesTexture.UnBind()
		leavesTexture.UnBind()
		gl.BindVertexArray(0)

		// obj is colored, light have the same color
		sourceProgram.Use()
		gl.UniformMatrix4fv(projectSourceUniformLocation, 1, false, &projectTransform[0])
		gl.UniformMatrix4fv(viewSourceUniformLocation, 1, false, &camTransform[0])
		moonTexture.Bind(gl.TEXTURE0)
		moonTexture.SetUniform(texSampler3SourceUniformLocation)
		gl.BindVertexArray(lightVAO)

		cubeM := mgl32.Ident4()
		cubeM = cubeM.Mul4(mgl32.Translate3D(pointLightPositions[2].Elem())).Mul4(mgl32.Scale3D(3, 3, 3))
		gl.Uniform3f(objectColorSourceUniformLocation, pointLightColors[2].X(), pointLightColors[2].Y(), pointLightColors[2].Z())
		gl.UniformMatrix4fv(modelSourceUniformLocation, 1, false, &cubeM[0])
		gl.DrawElements(gl.TRIANGLES, int32(len(indicesSpere))*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))

		moonTexture.UnBind()
		gl.BindVertexArray(0)
		discoBall.Bind(gl.TEXTURE0)
		discoBall.SetUniform(texSampler3SourceUniformLocation)
		gl.BindVertexArray(lightVAO)

		gl.Uniform3f(objectColorSourceUniformLocation, lightColor.X(), lightColor.Y(), lightColor.Z())
		starModel := model.Mul4(mgl32.Translate3D(-0.2, 9.9, 0)).Mul4(mgl32.Scale3D(0.5, 0.5, 0.5)).Mul4(mgl32.HomogRotate3DY(float32(animationCtl.GetAngle())))
		gl.UniformMatrix4fv(modelSourceUniformLocation, 1, false, &starModel[0])
		gl.DrawElements(gl.TRIANGLES, int32(len(indicesSpere))*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		discoBall.UnBind()
		gl.BindVertexArray(0)

		gl.BindVertexArray(lightVAO)
		gl.Uniform3f(objectColorSourceUniformLocation, pointLightColorsRef[4].X(), pointLightColorsRef[4].Y(), pointLightColorsRef[4].Z())
		shootingModel := model.Mul4(mgl32.Translate3D(pointLightPositions[4].Elem())).Mul4(mgl32.Scale3D(0.5, 0.5, 0.5)).Mul4(mgl32.HomogRotate3DY(float32(animationCtl.GetAngle())))
		gl.UniformMatrix4fv(modelSourceUniformLocation, 1, false, &shootingModel[0])
		gl.DrawElements(gl.TRIANGLES, int32(len(indicesSpere))*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		gl.BindVertexArray(0)

		//Sky box
		gl.BindVertexArray(skyVAO)
		starsTexture.Bind(gl.TEXTURE0)
		starsTexture.SetUniform(texSampler3SourceUniformLocation)
		gl.Uniform3f(objectColorSourceUniformLocation, backgroundColor.X(), backgroundColor.Y(), backgroundColor.Z())
		skyRotate := model
		gl.UniformMatrix4fv(modelSourceUniformLocation, 1, false, &skyRotate[0])
		gl.DrawElements(gl.TRIANGLES, 6*6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
		starsTexture.UnBind()
		gl.BindVertexArray(0)

	}

	return nil
}

func main() {

	runtime.LockOSThread()
	win.InitGlfw(4, 0)
	defer glfw.Terminate()
	window := win.NewWindow(width, height, title)
	gfx.InitGl()

	err := programLoop(window)
	if err != nil {
		log.Fatal(err)
	}
}
