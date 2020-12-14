package math3d

import "math"

// Vector struct holds the X, Y, Z and W coordinates
type Vector struct {
	X float64
	Y float64
	Z float64
	W float64
}

// ToArray converts a vector to an array
func (v Vector) ToArray() []float64 {
	return []float64{v.X, v.Y, v.Z, v.W}
}

// FromArray converts an array to a vector
func (v *Vector) FromArray(a []float64) {
	*v = Vector{a[0], a[1], a[2], a[3]}
}

// VectorOf constructs a vector of a given value
func VectorOf(value float64) Vector {
	return Vector{X: value, Y: value, Z: value, W: value}
}

// ClampValue clamps a vector to a min max range
func ClampValue(v Vector, min, max float64) Vector {
	return Vector{X: math.Min(max, math.Max(min, v.X)),
		Y: math.Min(max, math.Max(min, v.Y)),
		Z: math.Min(max, math.Max(min, v.Z)),
		W: math.Min(max, math.Max(min, v.W))}
}

// Add w to v
func Add(v, w Vector) Vector {
	return Vector{v.X + w.X, v.Y + w.Y, v.Z + w.Z, 1}
}

// Sub w to v
func Sub(v, w Vector) Vector {
	return Vector{v.X - w.X, v.Y - w.Y, v.Z - w.Z, 1}
}

// Mul multiplies v by f
func Mul(v Vector, f float64) Vector {
	return Vector{v.X * f, v.Y * f, v.Z * f, 1}
}

// MulComponentWise multiplies the components of two vectors
func MulComponentWise(v, w Vector) Vector {
	return Vector{v.X * w.X, v.Y * w.Y, v.Z * w.Z, v.W * w.W}
}

// Dot calculates the dot product of v * w
func Dot(v, w Vector) float64 {
	return v.X*w.X + v.Y*w.Y + v.Z*w.Z
}

// Magnitude determines the size of the vector
func Magnitude(v Vector) float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Normalize sets hthe size of the vector to 1
func Normalize(v Vector) Vector {
	mag := Magnitude(v)
	v.X /= mag
	v.Y /= mag
	v.Z /= mag
	return v
}

// Cross calculates the crossproduct of v x w
func Cross(v, w Vector) Vector {
	return Vector{
		X: v.Y*w.Z - v.Z*w.Y,
		Y: v.Z*w.X - v.X*w.Z,
		Z: v.X*w.Y - v.Y*w.X,
		W: 1,
	}
}

// Reflect reflects the (direction) vector v towards the (surface) normal
func Reflect(v, normal Vector) Vector {
	return Add(Mul(v, -1.), Mul(normal, Dot(v, normal)*2))
}

// Lerp linearly interpolates between v and w using t [0, 1]
func Lerp(v, w Vector, t float64) Vector {
	return Add(Mul(v, 1.-t), Mul(w, t))
}

// LerpTri linearly interpolates between v, w and u using s, t € [0, 1]
func LerpTri(v, w, u Vector, s, t float64) Vector {
	return Add(Add(Mul(v, 1.-s-t), Mul(w, s)), Mul(u, t))
}

// CalculateBoundingBox calculates the bounding box for given vectors
func CalculateBoundingBox(vectors ...Vector) (Vector, Vector) {
	var min, max Vector

	firstIteration := true
	for _, v := range vectors {
		if firstIteration {
			min = v
			max = v
			firstIteration = false
			continue
		}
		if v.X < min.X {
			min.X = v.X
		} else if v.X > max.X {
			max.X = v.X
		}
		if v.Y < min.Y {
			min.Y = v.Y
		} else if v.Y > max.Y {
			max.Y = v.Y
		}
	}

	return min, max
}
