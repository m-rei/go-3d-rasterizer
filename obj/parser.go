package obj

import (
	"bufio"
	"go-3d-rasterizer/math3d"
	m "go-3d-rasterizer/math3d"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Model holds the wavefront obj model data
type Model struct {
	vertices    []m.Vector
	normals     []m.Vector
	texCoords   []texCoord
	materials   []material
	materialMap map[string]*material

	triangles []indices
}

type texCoord struct {
	s float64
	t float64
}

type indices struct {
	v0, v1, v2, v3 int
	n0, n1, n2, n3 int
	t0, t1, t2, t3 int
	material       int
	hasNormals     bool
	hasTexture     bool
	hasFour        bool
}

type material struct {
	idx   int
	name  string
	mapKd string
	mapKs string

	ambientColor     m.Vector
	diffuseColor     m.Vector
	specularColor    m.Vector
	specularExponent float64

	mapKdData   []rl.Color
	mapKdWidth  int
	mapKdHeight int

	mapKsData   []rl.Color
	mapKsWidth  int
	mapKsHeight int
}

// ParseFile lodds & parses an .obj file
func ParseFile(filename string) (*Model, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ret := &Model{}
	ret.materialMap = make(map[string]*material)
	matIdx := -1

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(strings.Trim(scanner.Text(), " "), " ")
		if len(parts) == 2 {
			if parts[0] == "mtllib" {
				mats, err := parseMaterial(filepath.Join(filepath.Dir(filename), parts[1]))
				if err != nil {
					return nil, err
				}
				for i, m := range mats {
					m.idx = len(ret.materials) + i
				}
				ret.materials = append(ret.materials, mats...)
				for i, m := range mats {
					ret.materialMap[m.name] = &ret.materials[i]
				}
			} else if parts[0] == "usemtl" {
				matIdx = ret.materialMap[parts[1]].idx
			}
		}
		if len(parts) >= 4 && len(parts) <= 5 && parts[0] == "v" {
			x, _ := strconv.ParseFloat(parts[1], 32)
			y, _ := strconv.ParseFloat(parts[2], 32)
			z, _ := strconv.ParseFloat(parts[3], 32)
			w := 1.
			if len(parts) == 5 {
				w, _ = strconv.ParseFloat(parts[4], 32)
			}
			ret.vertices = append(ret.vertices, m.Vector{X: x, Y: y, Z: z, W: w})
		}
		if len(parts) == 4 && parts[0] == "vn" {
			x, _ := strconv.ParseFloat(parts[1], 32)
			y, _ := strconv.ParseFloat(parts[2], 32)
			z, _ := strconv.ParseFloat(parts[3], 32)
			ret.normals = append(ret.normals, m.Normalize(m.Vector{X: x, Y: y, Z: z, W: 1}))
		}
		if len(parts) >= 3 && len(parts) <= 4 && parts[0] == "vt" { // ignore 3rd parameter
			s, _ := strconv.ParseFloat(parts[1], 32)
			t, _ := strconv.ParseFloat(parts[2], 32)
			ret.texCoords = append(ret.texCoords, texCoord{s: s, t: t})
		}
		if len(parts) >= 4 && len(parts) <= 5 && parts[0] == "f" {
			subParts1 := strings.Split(parts[1], "/")
			subParts2 := strings.Split(parts[2], "/")
			subParts3 := strings.Split(parts[3], "/")
			if len(subParts2) != len(subParts1) || len(subParts3) != len(subParts1) {
				continue // faullty format according to spec, skip
			}
			var subParts4 []string
			hasFour := len(parts) == 5
			if hasFour {
				subParts4 = strings.Split(parts[4], "/")
			}

			// cases
			//		1			v
			//		1/2			v/t
			//		1/2/3		v/t/n
			//		1//3 		v//n
			v0, _ := strconv.Atoi(subParts1[0])
			v1, _ := strconv.Atoi(subParts2[0])
			v2, _ := strconv.Atoi(subParts3[0])
			v3 := 0
			if hasFour {
				v3, _ = strconv.Atoi(subParts4[0])
			}
			t0, t1, t2, t3 := 0, 0, 0, 0
			hasTexture := false
			if len(subParts1) > 1 {
				hasTexture = len(subParts1[1]) != 0
				if hasTexture {
					t0, _ = strconv.Atoi(subParts1[1])
					t1, _ = strconv.Atoi(subParts2[1])
					t2, _ = strconv.Atoi(subParts3[1])
					if hasFour {
						t3, _ = strconv.Atoi(subParts4[1])
					}
				}
			}
			n0, n1, n2, n3 := 0, 0, 0, 0
			hasNormals := false
			if len(subParts1) == 3 {
				hasNormals = true
				n0, _ = strconv.Atoi(subParts1[2])
				n1, _ = strconv.Atoi(subParts2[2])
				n2, _ = strconv.Atoi(subParts3[2])
				if hasFour {
					n3, _ = strconv.Atoi(subParts4[2])
				}
			}

			if v0 < 0 {
				v0 = len(ret.vertices) + v0
			}
			if v1 < 0 {
				v1 = len(ret.vertices) + v1
			}
			if v2 < 0 {
				v2 = len(ret.vertices) + v2
			}
			if hasFour && v3 < 0 {
				v3 = len(ret.vertices) + v3
			}

			if n0 < 0 {
				n0 = len(ret.normals) + n0
			}
			if n1 < 0 {
				n1 = len(ret.normals) + n1
			}
			if n2 < 0 {
				n2 = len(ret.normals) + n2
			}
			if hasFour && n3 < 0 {
				n3 = len(ret.normals) + n3
			}

			if t0 < 0 {
				t0 = len(ret.texCoords) + t0
			}
			if t1 < 0 {
				t1 = len(ret.texCoords) + t1
			}
			if t2 < 0 {
				t2 = len(ret.texCoords) + t2
			}
			if hasFour && t3 < 0 {
				t3 = len(ret.texCoords) + t3
			}

			ret.triangles = append(ret.triangles, indices{
				v0: v0 - 1, v1: v1 - 1, v2: v2 - 1, v3: v3 - 1,
				n0: n0 - 1, n1: n1 - 1, n2: n2 - 1, n3: n3 - 1,
				t0: t0 - 1, t1: t1 - 1, t2: t2 - 1, t3: t3 - 1,
				material:   matIdx,
				hasNormals: hasNormals,
				hasTexture: hasTexture,
				hasFour:    hasFour})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

// CenterVertices centers all the vertices in the model
func (m *Model) CenterVertices() {
	v := math3d.Vector{}
	for _, vertex := range m.vertices {
		v = math3d.Add(v, vertex)
	}
	v = math3d.Mul(v, 1./float64(len(m.vertices)))

	for i := range m.vertices {
		m.vertices[i] = math3d.Sub(m.vertices[i], v)
	}
}

// NormalizeVertices puts the model vertex data into a [-1, 1] interval
func (m *Model) NormalizeVertices(scale float64) {
	bbMin, bbMax := math3d.CalculateBoundingBox(m.vertices...)
	max := math.Abs(bbMin.X)
	max = math.Max(max, math.Abs(bbMax.X))
	max = math.Max(max, math.Abs(bbMin.Y))
	max = math.Max(max, math.Abs(bbMax.Y))
	max = math.Max(max, math.Abs(bbMin.Z))
	max = math.Max(max, math.Abs(bbMax.Z))
	max = scale / max
	for i := range m.vertices {
		n := math3d.Mul(m.vertices[i], max)
		n.W = 1
		m.vertices[i] = n
	}
}

func parseMaterial(filename string) ([]material, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ret []material
	isNotEmpty := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		if len(parts) == 2 {
			if parts[0] == "newmtl" {
				ret = append(ret, material{name: parts[1]})
				isNotEmpty = true
			} else if parts[0] == "map_Kd" && isNotEmpty {
				texFilename := filepath.Join(filepath.Dir(filename), parts[1])
				ret[len(ret)-1].mapKd = texFilename
				img := rl.LoadImage(texFilename)
				defer rl.UnloadImage(img)
				rl.ImageFlipVertical(img)
				ret[len(ret)-1].mapKdData = rl.GetImageData(img)
				ret[len(ret)-1].mapKdWidth = int(img.Width)
				ret[len(ret)-1].mapKdHeight = int(img.Height)
			} else if parts[0] == "map_Ks" && isNotEmpty {
				texFilename := filepath.Join(filepath.Dir(filename), parts[1])
				ret[len(ret)-1].mapKs = texFilename
				img := rl.LoadImage(texFilename)
				defer rl.UnloadImage(img)
				rl.ImageFlipVertical(img)
				ret[len(ret)-1].mapKsData = rl.GetImageData(img)
				ret[len(ret)-1].mapKsWidth = int(img.Width)
				ret[len(ret)-1].mapKsHeight = int(img.Height)

			} else if parts[0] == "Ns" && isNotEmpty {
				v, _ := strconv.ParseFloat(parts[1], 32)
				ret[len(ret)-1].specularExponent = v
			}
		} else if len(parts) == 4 && (parts[0] == "Ka" || parts[0] == "Kd" || parts[0] == "Ks") {
			x, _ := strconv.ParseFloat(parts[1], 32)
			y, _ := strconv.ParseFloat(parts[2], 32)
			z, _ := strconv.ParseFloat(parts[3], 32)
			col := m.Vector{X: x, Y: y, Z: z, W: 1}
			if parts[0] == "Ka" {
				ret[len(ret)-1].ambientColor = col
			} else if parts[0] == "Kd" {
				ret[len(ret)-1].diffuseColor = col
			} else {
				ret[len(ret)-1].specularColor = col
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}
