package rasterizer

import (
	m "go-3d-rasterizer/math3d"
	"math"
)

// Scene contains all the matrices needed to transform a vertex into screen coordinates
type Scene struct {
	ModelViewMatrix  m.Matrix
	ProjectionMatrix m.Matrix
	ViewportMatrix   m.Matrix
	Buffers          buffers

	width  int
	height int
	wh     int
}

type buffers struct {
	FrameBuffer []m.Vector
	DepthBuffer []float64
}

// LightingCalcCb is a callback function type for lighting calculation
type LightingCalcCb func(w, u, t float64, frameBufferIdx int)

// NewScene creates a new scene struct
func NewScene(winWidth, winHeight, fov, zNear, zFar float64) *Scene {
	return &Scene{
		ModelViewMatrix:  m.IdentityMatrix(),
		ProjectionMatrix: m.ProjectionMatrix(90.0, winWidth/winHeight, 1, 1000),
		ViewportMatrix:   m.Viewport(0, 0, winWidth, winHeight),
		Buffers: buffers{
			FrameBuffer: make([]m.Vector, int(winWidth*winHeight)),
			DepthBuffer: make([]float64, int(winWidth*winHeight)),
		},
		width:  int(winWidth),
		height: int(winHeight),
		wh:     int(winWidth * winHeight),
	}
}

// ClearBuffers clears the buffers
func (s *Scene) ClearBuffers(clearColor m.Vector) {
	i := 0
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			s.Buffers.FrameBuffer[i] = clearColor
			s.Buffers.DepthBuffer[i] = 1
			i++
		}
	}
}

// VectorToScreencoords convertex a vector to screen coordinates
func (s *Scene) VectorToScreencoords(v m.Vector) m.Vector {
	v = m.Transform(s.ModelViewMatrix, v, false)
	v = m.Transform(s.ProjectionMatrix, v, false)
	v = m.Mul(v, 1./v.W)
	v = m.Transform(s.ViewportMatrix, v, false)
	return v
}

// RasterizeLine draws a line from a to b with the given color, using the bresenham's line algorithm found here
// https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm#All_cases
func (s *Scene) RasterizeLine(a, b, colorA, colorB m.Vector) {
	a = s.VectorToScreencoords(a)
	b = s.VectorToScreencoords(b)

	x0, y0 := int(a.X), int(a.Y)
	x1, y1 := int(b.X), int(b.Y)

	dx := int(math.Abs(float64(x1 - x0)))
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	dy := int(-math.Abs(float64(y1 - y0)))
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx + dy
	for true {
		var t float64
		if b.X-a.X == 0 {
			t = (float64(y0) - a.Y) / (b.Y - a.Y)
		} else {
			t = (float64(x0) - a.X) / (b.X - a.X)
		}
		depth := a.Z*(1-t) + b.Z*t
		if y0 >= 0 && y0 < s.height && x0 >= 0 && x0 < s.width && depth >= 0 && depth <= 1.0 {
			idx := s.width*y0 + x0
			s.Buffers.FrameBuffer[idx] = lerpTriColor(colorA, colorB, m.Vector{}, (1 - t), t, 0)
			s.Buffers.DepthBuffer[idx] = depth
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func sf(x, y int, a, b, c m.Vector) float64 {
	return ((a.Y-c.Y)*float64(x) + (c.X-a.X)*float64(y) + a.X*c.Y - c.X*a.Y) / ((a.Y-c.Y)*b.X + (c.X-a.X)*b.Y + a.X*c.Y - c.X*a.Y)
}

func tf(x, y int, a, b, c m.Vector) float64 {
	return ((a.Y-b.Y)*float64(x) + (b.X-a.X)*float64(y) + a.X*b.Y - b.X*a.Y) / ((a.Y-b.Y)*c.X + (b.X-a.X)*c.Y + a.X*b.Y - b.X*a.Y)
}

// RasterizeTriangle draws a triangle with the three vectors a, b and c and the given color
// optimization: incremental barycentric coordinate calculation (u,t)
// source: http://gamma.cs.unc.edu/graphicscourse/09_rasterization.pdf (page 32,33)
func (s *Scene) RasterizeTriangle(a, b, c, colorA, colorB, colorC m.Vector, lightCalcCb LightingCalcCb) {
	a = s.VectorToScreencoords(a)
	b = s.VectorToScreencoords(b)
	c = s.VectorToScreencoords(c)

	bbMinX := int(math.Ceil(math.Min(math.Min(a.X, b.X), c.X)))
	bbMinY := int(math.Ceil(math.Min(math.Min(a.Y, b.Y), c.Y)))
	bbMaxX := int(math.Ceil(math.Max(math.Max(a.X, b.X), c.X)))
	bbMaxY := int(math.Ceil(math.Max(math.Max(a.Y, b.Y), c.Y)))

	u := sf(bbMinX, bbMinY, a, b, c)
	ux := sf(bbMinX+1., bbMinY, a, b, c) - u
	uy := sf(bbMinX, bbMinY+1., a, b, c) - u

	t := tf(bbMinX, bbMinY, a, b, c)
	tx := tf(bbMinX+1., bbMinY, a, b, c) - t
	ty := tf(bbMinX, bbMinY+1., a, b, c) - t

	n := float64(bbMaxX - bbMinX + 1)

	for y := bbMinY; y <= bbMaxY; y++ {
		idxOffset := s.width*y + bbMinX
		for x := bbMinX; x <= bbMaxX; x++ {
			insideViewport := x >= 0 && x < s.width && y >= 0 && y < s.height
			if t >= 0 && u >= 0 && t+u <= 1 && insideViewport {
				w := 1. - u - t
				depth := a.Z*w + b.Z*u + c.Z*t
				if depth >= 0. && depth <= s.Buffers.DepthBuffer[idxOffset] {
					s.Buffers.FrameBuffer[idxOffset] = lerpTriColor(colorA, colorB, colorC, w, u, t)
					s.Buffers.DepthBuffer[idxOffset] = depth
					if lightCalcCb != nil {
						lightCalcCb(w, u, t, idxOffset)
					}
				}
			}
			idxOffset++
			u += ux
			t += tx
		}
		u += uy - n*ux
		t += ty - n*tx
	}
}

// DrawTriangleWireframe draws a triangle with lines
func (s *Scene) DrawTriangleWireframe(a, b, c, colorA, colorB, colorC m.Vector) {
	s.RasterizeLine(a, b, colorA, colorB)
	s.RasterizeLine(b, c, colorB, colorC)
	s.RasterizeLine(c, a, colorC, colorA)
}

// DrawAxisLines draws an axis
func (s *Scene) DrawAxisLines(size float64) {
	xAxisA := m.Vector{X: -size, Y: 0, Z: 0, W: 1}
	xAxisB := m.Vector{X: size, Y: 0, Z: 0, W: 1}
	yAxisA := m.Vector{X: 0, Y: -size, Z: 0, W: 1}
	yAxisB := m.Vector{X: 0, Y: size, Z: 0, W: 1}
	zAxisA := m.Vector{X: 0, Y: 0, Z: -size, W: 1}
	zAxisB := m.Vector{X: 0, Y: 0, Z: size, W: 1}

	red := m.Vector{X: 1, Y: 0, Z: 0, W: 1}
	green := m.Vector{X: 0, Y: 1, Z: 0, W: 1}
	blue := m.Vector{X: 0, Y: 0, Z: 1, W: 1}

	s.RasterizeLine(xAxisA, xAxisB, red, red)
	s.RasterizeLine(yAxisA, yAxisB, green, green)
	s.RasterizeLine(zAxisA, zAxisB, blue, blue)
}

// DrawQuad renders 2 triangles
func (s *Scene) DrawQuad(a, b, c, d m.Vector, ca, cb, cc, cd m.Vector) {
	s.RasterizeTriangle(a, b, c, ca, cb, cc, nil)
	s.RasterizeTriangle(a, c, d, ca, cc, cd, nil)
}

// DrawCube renders a cube
func (s *Scene) DrawCube(center, color m.Vector, size float64) {
	size /= 2
	v := []m.Vector{
		{X: -size, Y: -size, Z: -size, W: 1},
		{X: -size, Y: +size, Z: -size, W: 1},
		{X: +size, Y: +size, Z: -size, W: 1},
		{X: +size, Y: -size, Z: -size, W: 1},
		{X: -size, Y: -size, Z: +size, W: 1},
		{X: -size, Y: +size, Z: +size, W: 1},
		{X: +size, Y: +size, Z: +size, W: 1},
		{X: +size, Y: -size, Z: +size, W: 1},
	}
	for i := range v {
		v[i] = m.Add(v[i], center)
	}
	s.DrawQuad(v[0], v[1], v[2], v[3], color, color, color, color)
	s.DrawQuad(v[4], v[5], v[6], v[7], color, color, color, color)

	s.DrawQuad(v[0], v[4], v[7], v[3], color, color, color, color)
	s.DrawQuad(v[1], v[5], v[6], v[2], color, color, color, color)

	s.DrawQuad(v[0], v[1], v[5], v[4], color, color, color, color)
	s.DrawQuad(v[2], v[3], v[7], v[6], color, color, color, color)
}

func lerpTriColor(c1, c2, c3 m.Vector, s, t, u float64) m.Vector {
	return m.Add(m.Add(m.Mul(c1, s), m.Mul(c2, t)), m.Mul(c3, u))
}
