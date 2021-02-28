package main

import (
	"log"
	"runtime"

	"git.maze.io/go/math32"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	glut "github.com/kaitsubaka/glae" // libreria propia implementada desde 0, https://github.com/kaitsubaka/glae
)

// Animation is a function that animates the models due a time
type Animation func(t float32)

const (
	width              = 1080
	height             = 720
	windowName         = "Core"
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
	initHatHeight       = 4.0
	fetherControlPoints = []mgl32.Vec3{
		{1, 0, 2.5},
		{1, 0, 5},
		{-5, 0, 5},
	}
)

func programLoop(window *glfw.Window) error {

	// the linked shader program determines how the data will be rendered
	vertShader, err := glut.CompileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := glut.CompileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}
	var shaders = []uint32{vertShader, fragShader}

	program, err := glut.CreateProgram(shaders)
	if err != nil {
		return nil
	}

	// ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.UseProgram(program)

	// creates camara
	cameraModel := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	camera := mgl32.LookAtV(mgl32.Vec3{3, 1, 5}, mgl32.Vec3{0, 1, 0}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])

	// creates perspective
	fov := float32(60.0)
	projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(width)/height, 0.1, 100.0)
	gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("project\x00")), 1, false, &projectTransform[0])

	// light
	gl.Uniform3f(gl.GetUniformLocation(program, gl.Str("lightColor\x00")), 1, 1, 1)
	//------------------------------------------------------------------------------------------------------------
	// Camera and positions
	//------------------------------------------------------------------------------------------------------------
	camPositions := [][2]mgl32.Vec3{
		{{1.7, 0.7, 1.7}, {1, 1, 1}},
		{{1.7, 0.7, 1.7}, {1, 0, 1}},
		{{5, 4, 5}, {0, 0, 0}},
	}
	camPathPoints := [][]mgl32.Vec3{
		{{1, 1, 1}, {1, 0, 1}},
	}
	//------------------------------------------------------------------------------------------------------------
	// models and vaos
	//------------------------------------------------------------------------------------------------------------
	model := mgl32.Ident4() // main model that position everything in 0,0,0
	modelUniform := gl.GetUniformLocation(program, gl.Str("world\x00"))
	colorModel := gl.GetUniformLocation(program, gl.Str("objectColor\x00"))

	cubeVertices := glut.GetCubicHexahedronVertices3(1.5, 1, 1.5)
	cubeVAO := glut.CreateVAO(cubeVertices)

	sideVertices, topVertices, bottomVertices := glut.GetCylinderVertices3(1, 0.1, 0.1, 5)
	sideVAO, topVAO, bottomVAO := glut.CreateVAO(sideVertices), glut.CreateVAO(topVertices), glut.CreateVAO(bottomVertices)

	planeVertices := glut.GetPlaneVertices3(10, 10, 1)
	planeVAO := glut.CreateVAO(planeVertices)

	sphereVertices, sphereTop, sphereBottom := glut.GetSphereVertices3(0.3, 16)

	sphereVao, sphereTopVao, sphereBotVao := glut.CreateVAO(sphereVertices), glut.CreateVAO(sphereTop), glut.CreateVAO(sphereBottom)

	noseVertices := glut.GetCircleVertices3(0.05, 8)
	noseVertices[0] = mgl32.Vec3{0, 0.2, 0}
	noseVao := glut.CreateVAO(noseVertices)

	snowCarpetVertices := glut.GetCubicHexahedronVertices3(1.5, 0.1, 1.5)
	snowCarpetVAO := glut.CreateVAO(snowCarpetVertices)

	sideVerticesHat, topVerticesHat, bottomVerticesHat := glut.GetCylinderVertices3(0.5, 0.5, 0.5, 5)
	sideHatVAO, topHatVAO, bottomHatVAO := glut.CreateVAO(sideVerticesHat), glut.CreateVAO(topVerticesHat), glut.CreateVAO(bottomVerticesHat)

	eyeVertices, eyeTop, eyeBottom := glut.GetSphereVertices3(0.06, 15)

	eyeVao, eyeTopVao, eyeBotVao := glut.CreateVAO(eyeVertices), glut.CreateVAO(eyeTop), glut.CreateVAO(eyeBottom)

	snowManPathModel := mgl32.Translate3D(2, 0, 2)
	hatModel := mgl32.Translate3D(0, float32(initHatHeight), 0)
	rEyeModel := mgl32.Translate3D(-0.08, 0.3, 0)
	lEyeModel := mgl32.Translate3D(0.08, 0.3, 0)
	mouthControlPoints := []mgl32.Vec3{{-0.09, 0.17, 0}, {0, 0.11, 0.02}, {0.09, 0.17, 0}}
	mouthPositions := mgl32.MakeBezierCurve3D(4, mouthControlPoints)
	mouthModels := []mgl32.Mat4{}
	for _, v := range mouthPositions {
		mouthModels = append(mouthModels, mgl32.Translate3D(v.Elem()))
	}
	extremitySideVertices, extremityTopVertices, extremityBottomVertices := glut.GetCapsuleVertices3(0.4, 0.07, 0.05, 8)
	extremitySideVAO, extremityTopVAO, extremityBottomVAO := glut.CreateVAO(extremitySideVertices), glut.CreateVAO(extremityTopVertices), glut.CreateVAO(extremityBottomVertices)
	rightArmModelRef := mgl32.Translate3D(1, 1.4, 1.3)
	rightArmModel := rightArmModelRef
	leftArmModelRef := mgl32.Translate3D(1.3, 1.4, 1.3)
	leftArmModel := leftArmModelRef
	//------------------------------------------------------------------------------------------------------------
	//variables and inits
	//------------------------------------------------------------------------------------------------------------
	var currentDelta float32
	angle := 0.0
	previousTime := glfw.GetTime()
	noFallingSections := 4.0
	relativeFallingSectionHight := initHatHeight / noFallingSections
	amplituCaida := 1
	var fallPathCtlPoints [][]mgl32.Vec3
	for i := 0; i < int(noFallingSections); i++ {
		var tmpHatPath []mgl32.Vec3
		if i%2 == 0 {
			tmpHatPath = append(tmpHatPath, mgl32.Vec3{0, float32(initHatHeight - relativeFallingSectionHight*float64(i)), 0})
			tmpHatPath = append(tmpHatPath, mgl32.Vec3{0.4, float32(initHatHeight - relativeFallingSectionHight*float64(i+1)*1.2), 0})
			tmpHatPath = append(tmpHatPath, mgl32.Vec3{float32(amplituCaida), float32(initHatHeight - relativeFallingSectionHight*float64(i+1)), 0})
		} else {
			tmpHatPath = append(tmpHatPath, mgl32.Vec3{float32(amplituCaida), float32(initHatHeight - relativeFallingSectionHight*float64(i)), 0})
			tmpHatPath = append(tmpHatPath, mgl32.Vec3{float32(amplituCaida) - 0.4, float32(initHatHeight - relativeFallingSectionHight*float64(i+1)*1.2), 0})
			tmpHatPath = append(tmpHatPath, mgl32.Vec3{0, float32(initHatHeight - relativeFallingSectionHight*float64(i+1)), 0})
		}
		fallPathCtlPoints = append(fallPathCtlPoints, tmpHatPath)
	}
	fallPathCtlPoints[int(noFallingSections-1)][1] = fallPathCtlPoints[int(noFallingSections-1)][2]

	var totalElapsed float64
	movementControlCount := 0
	fallingControlCount := 0

	//------------------------------------------------------------------------------------------------------------
	// scenes and animations
	//------------------------------------------------------------------------------------------------------------
	var movementTimes []float64
	var movementFunctions []Animation
	// Test area

	//---------------------------------scene 1 -----------------------------------------------------------
	// hat falling
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		if fallingControlCount < int(noFallingSections) {
			hatModel = mgl32.Translate3D(mgl32.BezierCurve3D(t, fallPathCtlPoints[fallingControlCount]).Elem()).Mul4(mgl32.HomogRotate3DY(float32(angle * 2))).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(10)))

			if t == 1.0 {
				fallingControlCount++
				movementControlCount--
			}
		}
	}), append(movementTimes, 2.0)

	// hat repositioning
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		deltaTime := t - float32(currentDelta)
		hatModel = hatModel.Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(-10) * deltaTime))
		currentDelta = t

	}), append(movementTimes, 0.8)
	// eyes positioning
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		deltaTime := t - float32(currentDelta)
		rEyeModel = rEyeModel.Mul4(mgl32.Translate3D(0, 0, 0.23*deltaTime))
		lEyeModel = lEyeModel.Mul4(mgl32.Translate3D(0, 0, 0.23*deltaTime))
		currentDelta = t
	}), append(movementTimes, 0.8)
	// mouth positioning
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		deltaTime := t - float32(currentDelta)
		for i := range mouthModels {
			mouthModels[i] = mouthModels[i].Mul4(mgl32.Translate3D(0, 0, 0.27*deltaTime))
		}
		currentDelta = t
	}), append(movementTimes, 0.8)
	//pause
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {}), append(movementTimes, 2.0)
	// snowman looks around
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		waves := float32(1)
		rot := math32.Sin(2*math32.Pi*waves*t) * mgl32.DegToRad(80)

		snowManPathModel = snowManPathModel.Mul4(mgl32.HomogRotate3DY(currentDelta - rot))
		currentDelta = rot
	}), append(movementTimes, 1.5)
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		waves := float32(1)
		rot := math32.Sin(2*math32.Pi*waves*t) * mgl32.DegToRad(80)

		snowManPathModel = snowManPathModel.Mul4(mgl32.HomogRotate3DY(currentDelta - rot))
		currentDelta = rot
	}), append(movementTimes, 6)
	//pause
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {}), append(movementTimes, 1.0)
	//arms falling
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {

		rightArmModel = rightArmModelRef.Mul4(mgl32.Translate3D(0, -1.4*t, 0)).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(90))).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(45)))
		leftArmModel = leftArmModelRef.Mul4(mgl32.Translate3D(0, -1.4*t, 0)).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(90))).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(45)))

	}), append(movementTimes, 0.5)
	// snowmans looks around again
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		waves := float32(1)
		rot := math32.Sin(2*math32.Pi*waves*t) * mgl32.DegToRad(80)

		snowManPathModel = snowManPathModel.Mul4(mgl32.HomogRotate3DY(currentDelta - rot))
		currentDelta = rot
	}), append(movementTimes, 5)
	// snowman looks the arms
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		delta := t - currentDelta
		snowManPathModel = snowManPathModel.Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(-100) * delta))
		currentDelta = t
	}), append(movementTimes, 2)

	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {}), append(movementTimes, 1)
	//----------------------------------second scene----------------------------------------------------------------
	//first person camera
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera = mgl32.LookAtV(camPositions[0][0], camPositions[0][1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
	}), append(movementTimes, 1.0)
	//pause
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {}), append(movementTimes, 0.05)
	// snowman looks arms in first person
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera = mgl32.LookAtV(camPositions[0][0], mgl32.BezierCurve3D(t, camPathPoints[0]), mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
	}), append(movementTimes, .5)
	//----------------------------------last scene----------------------------------------------------------------
	//pause
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {}), append(movementTimes, 2.0)
	//Isometric Framing
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		camera = mgl32.LookAtV(camPositions[2][0], camPositions[2][1], mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(cameraModel, 1, false, &camera[0])
		projectTransform := mgl32.Ortho(-5, 5, -5, 5, 1, 100)
		gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("project\x00")), 1, false, &projectTransform[0])

	}), append(movementTimes, 1.0)

	// snowman movement to arms
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		delta := t - currentDelta
		snowManPathModel = snowManPathModel.Mul4(mgl32.Translate3D(0, 0, 1.*delta))
		currentDelta = t
	}), append(movementTimes, .5)
	// arms reposition

	// snowman looks the camera
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		delta := t - currentDelta
		snowManPathModel = snowManPathModel.Mul4(mgl32.HomogRotate3DY(mgl32.DegToRad(140) * delta))
		currentDelta = t
	}), append(movementTimes, .5)

	// snowman comes back
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		delta := t - currentDelta
		snowManPathModel = snowManPathModel.Mul4(mgl32.Translate3D(0, 0, 1.*delta))

		rightArmModel = snowManPathModel.Mul4(mgl32.Translate3D(0.15, 0.7, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(-45))).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(-135)))
		leftArmModel = snowManPathModel.Mul4(mgl32.Translate3D(-0.15, 0.7, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(-45))).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(135)))
		currentDelta = t
	}), append(movementTimes, .5)
	// arm wave
	movementFunctions, movementTimes = append(movementFunctions, func(t float32) {
		waves := float32(1)
		rot := math32.Sin(2*math32.Pi*waves*t) * mgl32.DegToRad(-20)

		rightArmModel = snowManPathModel.Mul4(mgl32.Translate3D(0.15, 0.7, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(-45))).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(-135)))
		leftArmModel = snowManPathModel.Mul4(mgl32.Translate3D(-0.15, 0.7, 0)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(45))).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(45)))
		leftArmModel = leftArmModel.Mul4(mgl32.HomogRotate3DZ(rot))
		if t == 1 {
			movementControlCount--
		}
	}), append(movementTimes, .5)

	//------------------------------------------------------------------------------------------------------------
	//main loop
	//------------------------------------------------------------------------------------------------------------
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	for !window.ShouldClose() {
		time := glfw.GetTime()
		elapsed := time - previousTime
		totalElapsed += elapsed
		previousTime = time
		angle += elapsed

		glut.NewFrame(window, mgl32.Vec4{0.627, 0.596, 0.596, 1})
		// update
		if movementControlCount < len(movementFunctions) {
			if animationTime := movementTimes[movementControlCount]; totalElapsed <= animationTime {
				t := float32(totalElapsed) / float32(animationTime)
				movementFunctions[movementControlCount](t)
			} else {
				movementFunctions[movementControlCount](1)
				totalElapsed = 0
				currentDelta = 0
				movementControlCount++
			}
		}

		// You shall draw here

		for _, pos := range treePositions {
			treeTranslate := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z())
			scale1 := 1 - math32.Abs(math32.Sin(float32(time)))*0.04
			scale2 := 1 - math32.Abs(math32.Cos(float32(time)))*0.04
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
			treeTranslate = treeTranslate.Mul4(mgl32.Translate3D(0, 1.*scale1, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
			gl.BindVertexArray(0)
			// snow

			gl.Uniform3f(colorModel, 0.713, 0.925, 0.917)
			gl.BindVertexArray(snowCarpetVAO)
			snowTranslate := treeTranslate.Mul4(mgl32.Translate3D(0, 1, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &snowTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(snowCarpetVertices)))
			gl.BindVertexArray(0)
			//med

			gl.Uniform3f(colorModel, 0.392, 0.929, 0.768)
			gl.BindVertexArray(cubeVAO)
			treeTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75)).Mul4(mgl32.Translate3D(0, 0.75*scale2, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
			gl.BindVertexArray(0)
			// snow med

			gl.Uniform3f(colorModel, 0.713, 0.925, 0.917)
			gl.BindVertexArray(snowCarpetVAO)
			snowTranslate = treeTranslate.Mul4(mgl32.Translate3D(0, 1, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &snowTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(snowCarpetVertices)))
			gl.BindVertexArray(0)

			//smol

			gl.Uniform3f(colorModel, 0.552, 0.886, 0.788)
			gl.BindVertexArray(cubeVAO)
			treeTranslate = treeTranslate.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5)).Mul4(mgl32.Translate3D(0, 1.7*scale1, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &treeTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(cubeVertices)))
			gl.BindVertexArray(0)
			// snow smol

			gl.Uniform3f(colorModel, 0.713, 0.925, 0.917)
			gl.BindVertexArray(snowCarpetVAO)
			snowTranslate = treeTranslate.Mul4(mgl32.Translate3D(0, 1, 0))
			gl.UniformMatrix4fv(modelUniform, 1, false, &snowTranslate[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(snowCarpetVertices)))
			gl.BindVertexArray(0)

		}

		snowmanTranslate := snowManPathModel
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

		snowmanTranslate = snowmanTranslate.Mul4(mgl32.Scale3D(0.75, 0.75, 0.75)).Mul4(mgl32.Translate3D(0, 0.65, 0))
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
		snowmanNoseTranslate := snowmanTranslate.Mul4(mgl32.Translate3D(0, 0.3, 0.25)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(90)))
		gl.Uniform3f(colorModel, 1, 0.541, 0.380)
		gl.BindVertexArray(noseVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanNoseTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(noseVertices)))
		gl.BindVertexArray(0)

		// eyes

		snowmanREyeTranslate := snowmanTranslate.Mul4(rEyeModel)
		gl.Uniform3f(colorModel, 0, 0, 0)
		gl.BindVertexArray(eyeVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanREyeTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(eyeVertices)))

		gl.BindVertexArray(eyeTopVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanREyeTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(eyeTop)))

		gl.BindVertexArray(eyeBotVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanREyeTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(eyeBottom)))
		gl.BindVertexArray(0)

		snowmanLEyeTranslate := snowmanTranslate.Mul4(lEyeModel)
		gl.Uniform3f(colorModel, 0, 0, 0)
		gl.BindVertexArray(eyeVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanLEyeTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(eyeVertices)))

		gl.BindVertexArray(eyeTopVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanLEyeTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(eyeTop)))

		gl.BindVertexArray(eyeBotVao)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanLEyeTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(eyeBottom)))
		gl.BindVertexArray(0)

		// mouth
		for _, model := range mouthModels {
			mouthModel := snowmanTranslate.Mul4(model).Mul4(mgl32.Scale3D(0.4, 0.4, 0.4))
			gl.Uniform3f(colorModel, 0, 0, 0)
			gl.BindVertexArray(eyeVao)
			gl.UniformMatrix4fv(modelUniform, 1, false, &mouthModel[0])
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(eyeVertices)))

			gl.BindVertexArray(eyeTopVao)
			gl.UniformMatrix4fv(modelUniform, 1, false, &mouthModel[0])
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(eyeTop)))

			gl.BindVertexArray(eyeBotVao)
			gl.UniformMatrix4fv(modelUniform, 1, false, &mouthModel[0])
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(eyeBottom)))
			gl.BindVertexArray(0)
		}

		// hat

		snowmanHatTranslate := snowmanTranslate.Mul4(hatModel).Mul4(mgl32.Translate3D(0, 0.5, 0))
		gl.Uniform3f(colorModel, 0, 0, 0)
		gl.BindVertexArray(bottomHatVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanHatTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(bottomVerticesHat)))
		gl.BindVertexArray(0)

		snowmanHatTranslate = snowmanHatTranslate.Mul4(mgl32.Scale3D(0.55, 1, 0.55))
		gl.BindVertexArray(sideHatVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanHatTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(sideVerticesHat)))
		gl.BindVertexArray(0)

		gl.BindVertexArray(topHatVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &snowmanHatTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(topVerticesHat)))
		gl.BindVertexArray(0)

		// arms
		armsRTranslate := rightArmModel
		gl.Uniform3f(colorModel, 0.4, 0.2, 0)
		gl.BindVertexArray(extremitySideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsRTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(extremitySideVertices)))

		gl.BindVertexArray(extremityTopVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsRTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityTopVertices)))

		gl.BindVertexArray(extremityBottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsRTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityBottomVertices)))

		armsLTranslate := leftArmModel
		gl.BindVertexArray(extremitySideVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsLTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(extremitySideVertices)))

		gl.BindVertexArray(extremityTopVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsLTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityTopVertices)))

		gl.BindVertexArray(extremityBottomVAO)
		gl.UniformMatrix4fv(modelUniform, 1, false, &armsLTranslate[0])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, int32(len(extremityBottomVertices)))

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

	window := glut.InitGlfw(width, height, 4, 1, windowName)
	defer glfw.Terminate()

	glut.InitGl()

	err := programLoop(window)
	if err != nil {
		log.Fatalln(err)
	}
}
