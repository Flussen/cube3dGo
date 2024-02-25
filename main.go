package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width              = 800
	height             = 600
	vertexShaderSource = `#version 410
        layout (location = 0) in vec3 aPos;
        uniform mat4 model;
        uniform mat4 view;
        uniform mat4 projection;
        void main() {
            gl_Position = projection * view * model * vec4(aPos, 1.0);
        }
    ` + "\x00"
	fragmentShaderSource = `#version 410
        out vec4 FragColor;
        void main() {
            FragColor = vec4(1.0, 1.0, 1.0, 1.0);
        }
    ` + "\x00"
)

func init() {
	// Esto vincula el main a la gorutina principal, necesario para que funcione GLFW.
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "LearnOpenGL", nil, nil)
	if err != nil {
		log.Fatalln("failed to create glfw window:", err)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatalln("failed to initialize gl:", err)
	}

	gl.Enable(gl.DEPTH_TEST)

	program := initOpenGL()
	vao := makeVao(program)

	for !window.ShouldClose() {
		draw(window, program, vao)
	}
}

func initOpenGL() uint32 {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	return program
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

var (
	indices = []uint32{
		0, 1, 2, 2, 3, 0, // Cara trasera
		4, 5, 6, 6, 7, 4, // Cara frontal
		8, 9, 10, 10, 11, 8, // Cara izquierda
		12, 13, 14, 14, 15, 12, // Cara derecha
		16, 17, 18, 18, 19, 16, // Cara inferior
		20, 21, 22, 22, 23, 20, // Cara superior
	}
)

func makeVao(program uint32) uint32 {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	vertices := []float32{
		-0.5, -0.5, -0.5, 0.5, -0.5, -0.5, 0.5, 0.5, -0.5, -0.5, 0.5, -0.5, // Cara trasera
		-0.5, -0.5, 0.5, 0.5, -0.5, 0.5, 0.5, 0.5, 0.5, -0.5, 0.5, 0.5, // Cara frontal
		-0.5, 0.5, 0.5, -0.5, 0.5, -0.5, -0.5, -0.5, -0.5, -0.5, -0.5, 0.5, // Cara izquierda
		0.5, 0.5, 0.5, 0.5, 0.5, -0.5, 0.5, -0.5, -0.5, 0.5, -0.5, 0.5, // Cara derecha
		-0.5, -0.5, -0.5, 0.5, -0.5, -0.5, 0.5, -0.5, 0.5, -0.5, -0.5, 0.5, // Cara inferior
		-0.5, 0.5, -0.5, 0.5, 0.5, -0.5, 0.5, 0.5, 0.5, -0.5, 0.5, 0.5, // Cara superior
	}
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	position := gl.GetAttribLocation(program, gl.Str("aPos\x00"))
	gl.EnableVertexAttribArray(uint32(position))
	gl.VertexAttribPointer(uint32(position), 3, gl.FLOAT, false, 3*4, nil)

	gl.BindVertexArray(0)

	return vao
}

func draw(window *glfw.Window, program uint32, vao uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	// Obtiene el tiempo transcurrido
	time := glfw.GetTime()

	// Crea una matriz de modelo que rota el cubo
	// Rota alrededor del eje Y; el ángulo de rotación depende del tiempo para crear un efecto de movimiento continuo
	model := mgl32.HomogRotate3DY(float32(time))

	view := mgl32.LookAtV(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(width)/float32(height), 0.1, 100.0)

	modelLoc := gl.GetUniformLocation(program, gl.Str("model\x00"))
	viewLoc := gl.GetUniformLocation(program, gl.Str("view\x00"))
	projLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))

	gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])
	gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	gl.BindVertexArray(vao)
	gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, gl.PtrOffset(0))

	window.SwapBuffers()
	glfw.PollEvents()
}
