package math3d

import "math"

// Matrix struct holds a 4x4 matrix
type Matrix struct {
	X Vector
	Y Vector
	Z Vector
	W Vector
}

// ToArray converts the matrix to its array representation
func (m Matrix) ToArray() []float64 {
	ret := append(m.X.ToArray(), m.Y.ToArray()...)
	ret = append(ret, m.Z.ToArray()...)
	return append(ret, m.W.ToArray()...)
}

// FromArray converts an array to a matrix
func (m *Matrix) FromArray(a []float64) {
	m.X.FromArray(a[0:4])
	m.Y.FromArray(a[4:8])
	m.Z.FromArray(a[8:12])
	m.W.FromArray(a[12:16])
}

// LoadIdentity loads the identity matrix
func (m *Matrix) LoadIdentity() {
	*m = IdentityMatrix()
}

// Transpose mirrors the content of the matrix diagonally
func (m *Matrix) Transpose() {
	m.X.Y, m.Y.X = m.Y.X, m.X.Y
	m.X.Z, m.Y.Y, m.Z.X, m.Z.Y = m.Z.X, m.Z.Y, m.X.Z, m.Y.Y
	m.X.W, m.Y.W, m.Z.W, m.W.X, m.W.Y, m.W.Z = m.W.X, m.W.Y, m.W.Z, m.X.W, m.Y.W, m.Z.W
}

// Multiply returns the product of: m * n
func Multiply(m, n Matrix) Matrix {
	return Matrix{
		X: Vector{
			X: m.X.X*n.X.X + m.Y.X*n.X.Y + m.Z.X*n.X.Z + m.W.X*n.X.W,
			Y: m.X.Y*n.X.X + m.Y.Y*n.X.Y + m.Z.Y*n.X.Z + m.W.Y*n.X.W,
			Z: m.X.Z*n.X.X + m.Y.Z*n.X.Y + m.Z.Z*n.X.Z + m.W.Z*n.X.W,
			W: m.X.W*n.X.X + m.Y.W*n.X.Y + m.Z.W*n.X.Z + m.W.W*n.X.W,
		},
		Y: Vector{
			X: m.X.X*n.Y.X + m.Y.X*n.Y.Y + m.Z.X*n.Y.Z + m.W.X*n.Y.W,
			Y: m.X.Y*n.Y.X + m.Y.Y*n.Y.Y + m.Z.Y*n.Y.Z + m.W.Y*n.Y.W,
			Z: m.X.Z*n.Y.X + m.Y.Z*n.Y.Y + m.Z.Z*n.Y.Z + m.W.Z*n.Y.W,
			W: m.X.W*n.Y.X + m.Y.W*n.Y.Y + m.Z.W*n.Y.Z + m.W.W*n.Y.W,
		},
		Z: Vector{
			X: m.X.X*n.Z.X + m.Y.X*n.Z.Y + m.Z.X*n.Z.Z + m.W.X*n.Z.W,
			Y: m.X.Y*n.Z.X + m.Y.Y*n.Z.Y + m.Z.Y*n.Z.Z + m.W.Y*n.Z.W,
			Z: m.X.Z*n.Z.X + m.Y.Z*n.Z.Y + m.Z.Z*n.Z.Z + m.W.Z*n.Z.W,
			W: m.X.W*n.Z.X + m.Y.W*n.Z.Y + m.Z.W*n.Z.Z + m.W.W*n.Z.W,
		},
		W: Vector{
			X: m.X.X*n.W.X + m.Y.X*n.W.Y + m.Z.X*n.W.Z + m.W.X*n.W.W,
			Y: m.X.Y*n.W.X + m.Y.Y*n.W.Y + m.Z.Y*n.W.Z + m.W.Y*n.W.W,
			Z: m.X.Z*n.W.X + m.Y.Z*n.W.Y + m.Z.Z*n.W.Z + m.W.Z*n.W.W,
			W: m.X.W*n.W.X + m.Y.W*n.W.Y + m.Z.W*n.W.Z + m.W.W*n.W.W,
		},
	}
}

// Transform returns a transformed vector, which is m * v
func Transform(m Matrix, v Vector, isNormalVec bool) Vector {
	if !isNormalVec {
		return Vector{
			X: m.X.X*v.X + m.Y.X*v.Y + m.Z.X*v.Z + m.W.X*v.W,
			Y: m.X.Y*v.X + m.Y.Y*v.Y + m.Z.Y*v.Z + m.W.Y*v.W,
			Z: m.X.Z*v.X + m.Y.Z*v.Y + m.Z.Z*v.Z + m.W.Z*v.W,
			W: m.X.W*v.X + m.Y.W*v.Y + m.Z.W*v.Z + m.W.W*v.W,
		}
	}
	return Vector{
		X: m.X.X*v.X + m.Y.X*v.Y + m.Z.X*v.Z,
		Y: m.X.Y*v.X + m.Y.Y*v.Y + m.Z.Y*v.Z,
		Z: m.X.Z*v.X + m.Y.Z*v.Y + m.Z.Z*v.Z,
		W: 0,
	}
}

// Translate returns the matrix translated by x, y, z
func Translate(m Matrix, x, y, z float64) Matrix {
	translationMat := IdentityMatrix()
	translationMat.W = Vector{x, y, z, 1}
	return Multiply(m, translationMat)
}

// Scale returns the matrix scaled by x, y, z
func Scale(m Matrix, x, y, z float64) Matrix {
	translationMat := IdentityMatrix()
	translationMat.X.X = x
	translationMat.Y.Y = y
	translationMat.Z.Z = z
	translationMat.W.W = 1
	return Multiply(m, translationMat)
}

// Rotate rottes the matrix on one axis
func Rotate(m Matrix, radians, x, y, z float64) Matrix {
	s := math.Sin(radians)
	c := math.Cos(radians)
	rotationMat := IdentityMatrix()
	rotationMat.X = Vector{x*x*(1-c) + c, y*x*(1-c) - z*s, z*x*(1-c) + y*s, 0}
	rotationMat.Y = Vector{x*y*(1-c) + z*s, y*y*(1-c) + c, z*y*(1-c) - x*s, 0}
	rotationMat.Z = Vector{x*z*(1-c) - y*s, y*z*(1-c) + x*s, z*z*(1-c) + c, 0}
	return Multiply(m, rotationMat)
}

// IdentityMatrix returns the identity matrix
func IdentityMatrix() Matrix {
	return Matrix{
		X: Vector{1, 0, 0, 0},
		Y: Vector{0, 1, 0, 0},
		Z: Vector{0, 0, 1, 0},
		W: Vector{0, 0, 0, 1},
	}
}

// ProjectionMatrix creates a projection matrix with the given arguments
func ProjectionMatrix(fov, aspectRatio, zNear, zFar float64) Matrix {
	t := math.Tan(fov*math.Pi/360.) * zNear
	r := t * aspectRatio
	return Matrix{
		X: Vector{zNear / r, 0, 0, 0},
		Y: Vector{0, zNear / t, 0, 0},
		Z: Vector{0, 0, (-zFar - zNear) / (zFar - zNear), -1},
		W: Vector{0, 0, -2. * zNear * zFar / (zFar - zNear), 0},
	}
}

// Viewport creates a viewport matrix for a window
func Viewport(x, y, w, h float64) Matrix {
	vp := IdentityMatrix()
	vp = Translate(vp, x, y, 0)
	vp = Scale(vp, w/2., -h/2., 1./2.)
	vp = Translate(vp, 1, -1, 1)
	return vp
}
