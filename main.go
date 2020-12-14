package main

import (
	m "go-3d-rasterizer/math3d"
	"go-3d-rasterizer/obj"
	r "go-3d-rasterizer/rasterizer"
	_ "image/png"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type modelFile struct {
	fn    string
	scale float64
}

const (
	title  = "3D Rasterizer"
	width  = 1200
	height = 800
)

var (
	scene             *r.Scene   = r.NewScene(width, height, 90, 1, 1000)
	raylibFramebuffer []rl.Color = make([]rl.Color, width*height)
	startTime         time.Time  = time.Now()
	renderNormals     bool       = false
	selectedModel     int        = 0
	models            []*obj.Model

	zoom        float64 = -3
	mode        int     = 1
	autoRotate  bool    = true
	useLighting bool    = false

	modelFiles = []modelFile{
		{"./assets/teapot.obj", 2.},
		{"./assets/spaceship/Spaceship.obj", 1.5},
		{"./assets/leggings/leggins.obj", 2},
	}
)

func loadModels() {
	models = []*obj.Model{}
	for _, mf := range modelFiles {
		model, err := obj.ParseFile(mf.fn)
		if err != nil {
			continue
		}
		model.CenterVertices()
		model.NormalizeVertices(mf.scale)
		models = append(models, model)
	}
}

func clamp(value, min, max float64) float64 {
	return math.Max(math.Min(value, max), min)
}

func handleInput() {
	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		mode = (mode + 1) % 2
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		selectedModel = (selectedModel + 1) % len(models)
	}
	if rl.IsKeyPressed(rl.KeyN) {
		renderNormals = !renderNormals
	}
	if rl.IsKeyPressed(rl.KeyA) {
		autoRotate = !autoRotate
	}
	if rl.IsKeyPressed(rl.KeyL) {
		useLighting = !useLighting
	}
	mw := rl.GetMouseWheelMove()
	if mw > 0 {
		zoom += 0.5
	} else if mw < 0 {
		zoom -= 0.5
	}
	zoom = clamp(zoom, -8, -2)
}

func render() {
	dt := time.Now().Sub(startTime).Seconds()

	scene.ModelViewMatrix = m.IdentityMatrix()
	scene.ModelViewMatrix = m.Translate(scene.ModelViewMatrix, 0, 0, zoom)
	scene.ModelViewMatrix = m.Rotate(scene.ModelViewMatrix, -10.*math.Pi/180., 1, 0, 0)
	scene.ModelViewMatrix = m.Rotate(scene.ModelViewMatrix, dt, 0, 1, 0)

	scene.ClearBuffers(m.Vector{X: 0.5, Y: 0.5, Z: 0.5, W: 1})
	scene.DrawAxisLines(1)

	rot := 2. * math.Pi * float64(rl.GetMousePosition().X) / float64(width)
	if autoRotate {
		rot = dt*3 + math.Pi
	}

	// render light source
	lightDirection := m.Vector{X: math.Sin(rot), Y: 0, Z: math.Cos(rot), W: 1}
	if useLighting {
		scene.DrawCube(m.Mul(lightDirection, -1.5), m.Vector{X: 1, Y: 1, Z: 1, W: 1}, 0.025)
	}

	if renderNormals {
		models[selectedModel].RenderNormals(scene)
	}
	switch mode {
	case 0:
		models[selectedModel].Render(scene, useLighting, lightDirection)
	case 1:
		models[selectedModel].RenderWireframe(scene)
	}
}

func createFrameBuffer(width, height int) rl.Texture2D {
	img := rl.GenImageColor(width, height, rl.White)
	defer rl.UnloadImage(img)
	return rl.LoadTextureFromImage(img)
}

func updateFrameBuffer() {
	idx := 0
	for y := 0; y < height; y++ {
		for x := 1; x < width; x++ {
			c := scene.Buffers.FrameBuffer[idx]
			raylibFramebuffer[idx] = rl.Color{R: uint8(c.X * 255.), G: uint8(c.Y * 255.), B: uint8(c.Z * 255.), A: uint8(c.W * 255.)}
			idx++
		}
	}
}

func main() {
	loadModels()
	rl.InitWindow(width, height, title)
	rl.SetTargetFPS(120)
	frameBuffer := createFrameBuffer(width, height)
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		handleInput()
		render()
		updateFrameBuffer()
		rl.UpdateTexture(frameBuffer, raylibFramebuffer)
		rl.DrawTexture(frameBuffer, 0, 0, rl.White)
		rl.DrawText("Left mouse button - change model", 5, 30, 20, rl.Black)
		rl.DrawText("Right mouse button - toggle wireframe mode", 5, 60, 20, rl.Black)
		rl.DrawText("Mouse wheel - zoom", 5, 90, 20, rl.Black)
		rl.DrawText("L - toggle light", 5, 120, 20, rl.Black)
		rl.DrawText("N - toggle normals", 5, 150, 20, rl.Black)
		rl.DrawText("A - toggle auto rotation", 5, 180, 20, rl.Black)
		rl.DrawFPS(5, 5)
		rl.EndDrawing()
	}
	rl.UnloadTexture(frameBuffer)
	rl.CloseWindow()
}
