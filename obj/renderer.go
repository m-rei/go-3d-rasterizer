package obj

import (
	m "go-3d-rasterizer/math3d"
	"go-3d-rasterizer/rasterizer"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// RenderWireframe renders the model in wireframe mode
func (o *Model) RenderWireframe(scene *rasterizer.Scene) {
	for _, t := range o.triangles {
		col1, col2, col3, col4 := m.Vector{X: 0, Y: 0, Z: 0, W: 1}, m.Vector{X: 0, Y: 0, Z: 0, W: 1}, m.Vector{X: 0, Y: 0, Z: 0, W: 1}, m.Vector{X: 0, Y: 0, Z: 0, W: 1}

		if t.hasTexture && len(o.materials) > 0 {
			st0, st1, st2 := o.texCoords[t.t0], o.texCoords[t.t1], o.texCoords[t.t2]
			col1 = pixelFromMaterial(o.materials[t.material], st0)
			col2 = pixelFromMaterial(o.materials[t.material], st1)
			col3 = pixelFromMaterial(o.materials[t.material], st2)
			if t.hasFour {
				st3 := o.texCoords[t.t3]
				col4 = pixelFromMaterial(o.materials[t.material], st3)
			}
		}
		if t.hasFour {
			scene.DrawTriangleWireframe(o.vertices[t.v0], o.vertices[t.v1], o.vertices[t.v2], col1, col2, col3)
			scene.DrawTriangleWireframe(o.vertices[t.v0], o.vertices[t.v2], o.vertices[t.v3], col1, col3, col4)
		} else {
			scene.DrawTriangleWireframe(o.vertices[t.v0], o.vertices[t.v1], o.vertices[t.v2], col1, col2, col3)
		}
	}
}

// RenderNormals renders the normals ontop of each vertex
func (o *Model) RenderNormals(scene *rasterizer.Scene) {
	white := m.Vector{X: 1, Y: 1, Z: 1, W: 1}
	for _, t := range o.triangles {
		if !t.hasNormals {
			continue
		}

		scene.RasterizeLine(o.vertices[t.v0], m.Add(o.vertices[t.v0], o.normals[t.n0]), white, white)
		scene.RasterizeLine(o.vertices[t.v1], m.Add(o.vertices[t.v1], o.normals[t.n1]), white, white)
		scene.RasterizeLine(o.vertices[t.v2], m.Add(o.vertices[t.v2], o.normals[t.n2]), white, white)
		if !t.hasFour {
			continue
		}
		scene.RasterizeLine(o.vertices[t.v3], m.Add(o.vertices[t.v3], o.normals[t.n3]), white, white)
	}
}

func lightingCalculator(a, b, c,
	normalA, normalB, normalC m.Vector,
	st0, st1, st2 texCoord,
	lightDir m.Vector, mat *material, s *rasterizer.Scene) rasterizer.LightingCalcCb {

	lightDir = m.Mul(lightDir, -1)
	return func(w, u, t float64, frameBufferIdx int) {
		color := &s.Buffers.FrameBuffer[frameBufferIdx]
		normalAvg := m.Normalize(m.Add(m.Add(m.Mul(normalA, w), m.Mul(normalB, u)), m.Mul(normalC, t)))
		posAvg := m.Add(m.Add(m.Mul(a, w), m.Mul(b, u)), m.Mul(c, t))
		dot := m.Dot(normalAvg, lightDir)
		if dot >= 0 {
			if mat != nil {
				// ambient
				lightCol := mat.ambientColor
				// diffuse
				lightCol = m.Add(lightCol, m.Mul(mat.diffuseColor, dot))
				// specular
				eye := m.Normalize(m.Mul(posAvg, -1))
				half := m.Normalize(m.Sub(eye, lightDir))
				halfNormal := math.Max(0, m.Dot(half, eye))
				shininess := math.Pow(halfNormal, mat.specularExponent)
				specularColor := mat.specularColor
				if st0.s != -1 && mat.mapKsWidth > 0 {
					x, y := stToXy(st0, mat.mapKsWidth, mat.mapKsHeight)
					specCol1 := mat.mapKsData[mat.mapKsWidth*y+x]
					x, y = stToXy(st1, mat.mapKsWidth, mat.mapKsHeight)
					specCol2 := mat.mapKsData[mat.mapKsWidth*y+x]
					x, y = stToXy(st2, mat.mapKsWidth, mat.mapKsHeight)
					specCol3 := mat.mapKsData[mat.mapKsWidth*y+x]
					specularColor = m.Add(m.Add(m.Mul(colorToVector(specCol1), w), m.Mul(colorToVector(specCol2), u)), m.Mul(colorToVector(specCol3), t))
				}
				lightCol = m.Add(lightCol, m.Mul(specularColor, shininess))
				// final mixture:
				// fragment color = fragment color * (ambient + diffuse + specular), clamped
				*color = m.MulComponentWise(*color, m.ClampValue(lightCol, 0, 1))
			}
		} else {
			*color = m.VectorOf(0)
			color.W = 1
		}
	}
}

// Render renders the obj model
func (o *Model) Render(scene *rasterizer.Scene, useLighting bool, lightDirection m.Vector) {
	for _, t := range o.triangles {
		col1 := m.Vector{X: 1, Y: 1, Z: 1, W: 1}
		col2, col3, col4 := col1, col1, col1

		st0 := texCoord{s: -1, t: -1}
		st1, st2, st3 := st0, st0, st0

		if t.hasTexture && len(o.materials) > 0 {
			st0, st1, st2 = o.texCoords[t.t0], o.texCoords[t.t1], o.texCoords[t.t2]
			col1 = pixelFromMaterial(o.materials[t.material], st0)
			col2 = pixelFromMaterial(o.materials[t.material], st1)
			col3 = pixelFromMaterial(o.materials[t.material], st2)
			if t.hasFour {
				st3 = o.texCoords[t.t3]
				col4 = pixelFromMaterial(o.materials[t.material], st3)
			}
		}

		var mat *material = nil
		if t.material != -1 {
			mat = &o.materials[t.material]
		}

		if t.hasFour {
			if t.hasNormals {
				var lightingCalc1 rasterizer.LightingCalcCb = nil
				var lightingCalc2 rasterizer.LightingCalcCb = nil
				if useLighting {
					lightingCalc1 = lightingCalculator(o.vertices[t.v0], o.vertices[t.v1], o.vertices[t.v2],
						o.normals[t.n0], o.normals[t.n1], o.normals[t.n2],
						st0, st1, st2,
						lightDirection, mat, scene)
					lightingCalc2 = lightingCalculator(o.vertices[t.v0], o.vertices[t.v1], o.vertices[t.v2],
						o.normals[t.n0], o.normals[t.n2], o.normals[t.n3],
						st0, st2, st3,
						lightDirection, mat, scene)
				}

				scene.RasterizeTriangle(o.vertices[t.v0], o.vertices[t.v1], o.vertices[t.v2], col1, col2, col3, lightingCalc1)
				scene.RasterizeTriangle(o.vertices[t.v0], o.vertices[t.v2], o.vertices[t.v3], col1, col3, col4, lightingCalc2)
			} else {
				scene.DrawQuad(o.vertices[t.v0], o.vertices[t.v1], o.vertices[t.v2], o.vertices[t.v3], col1, col2, col3, col4)
			}
		} else {
			if t.hasNormals {
				var lightingCalc rasterizer.LightingCalcCb = nil
				if useLighting {
					lightingCalc = lightingCalculator(o.vertices[t.v0], o.vertices[t.v1], o.vertices[t.v2],
						o.normals[t.n0], o.normals[t.n1], o.normals[t.n2],
						st0, st1, st2,
						lightDirection, mat, scene)
				}
				scene.RasterizeTriangle(o.vertices[t.v0], o.vertices[t.v1], o.vertices[t.v2], col1, col2, col3, lightingCalc)
			} else {
				scene.RasterizeTriangle(o.vertices[t.v0], o.vertices[t.v1], o.vertices[t.v2], col1, col2, col3, nil)
			}
		}
	}
}

func pixelFromMaterial(mat material, st texCoord) m.Vector {
	if mat.mapKdWidth == 0 {
		return m.Vector{X: 1, Y: 1, Z: 1, W: 1}
	}
	x, y := stToXy(st, mat.mapKdWidth, mat.mapKdHeight)
	return colorToVector(mat.mapKdData[mat.mapKdWidth*y+x])
}

func stToXy(st texCoord, width, height int) (int, int) {
	return int(st.s * float64(width-1)), int(st.t * float64(height-1))
}

func colorToVector(c rl.Color) m.Vector {
	return m.Vector{X: float64(c.R) / 255., Y: float64(c.G) / 255., Z: float64(c.B) / 255., W: float64(c.A) / 255}
}
