package main

import (
	"fmt"
	"math"
	"math/rand"
)

const PI = math.Pi / 180

func radians(d float64) float64 { return d * PI }
func degrees(r float64) float64 { return r / PI }

// Vertex is used for 3d position, vertex, direction, or other various 3float tuples.
type Vertex struct {
	x        float32
	y        float32
	z        float32
	isNotNil bool
}

func NewVertex(x, y, z float32) Vertex {
	return Vertex{x, y, z, true}
}

func NewVertexInt(x, y, z int) Vertex {
	return NewVertex(float32(x), float32(y), float32(z))
}

var nilVertex = Vertex{}

func (v Vertex) isNil() bool {
	return !v.isNotNil
}

func (v Vertex) get(i int) float32 {
	if i == 0 {
		return v.x
	} else if i == 1 {
		return v.y
	} else if i == 2 {
		return v.z
	} else {
		panic(fmt.Sprintf("unknown index %d in a vertex\n", i))
	}
}

func (v *Vertex) set(i int, f float32) {
	if i == 0 {
		v.x = f
	} else if i == 1 {
		v.y = f
	} else if i == 2 {
		v.z = f
	} else {
		panic("trying to set a bad vertex index")
	}
}

func normalize(position Vertex) Vertex {
	/* Accepts `position` of arbitrary precision and returns the block
	   containing that position.

	   Parameters
	   ----------
	   position : tuple of len 3

	   Returns
	   -------
	   block_position : tuple of ints of len 3

	*/
	x, y, z := round(position.x), round(position.y), round(position.z)
	return NewVertex(x, y, z)
}

func round(x float32) float32 {
	if x < 0 {
		return float32(math.Ceil(float64(x) - 0.5))
	}
	return float32(math.Floor(float64(x) + 0.5))
}

func sectorize(position Vertex) Vertex {
	/* Returns a tuple representing the sector for the given `position`.

	   Parameters
	   ----------
	   position : tuple of len 3

	   Returns
	   -------
	   sector : tuple of len 3

	*/
	normal := normalize(position)
	x, z := normal.x/SECTOR_SIZE, normal.z/SECTOR_SIZE
	return NewVertex(x, 0, z)
}

func xrange(start, end, step int) []int {
	r := make([]int, 0, ((end-start)/step)+1)

	for x := start; x < end; x += step {
		r = append(r, x)
	}

	return r
}

func random_randint(start, end int) int {
	return rand.Intn(end-start) + start
}

func PowInt(x, y int) int {
	return int(math.Pow(float64(x), float64(y)))
}

type VertexSet map[Vertex]bool

func NewVertexSet() VertexSet {
	return VertexSet(make(map[Vertex]bool))
}

func (vs1 VertexSet) Remove(vs2 VertexSet) VertexSet {

	vs3 := make(map[Vertex]bool)

	for v := range vs1 {
		if _, ok := vs2[v]; !ok {
			vs3[v] = true
		}
	}
	return vs3
}

func (vs VertexSet) add(v Vertex) {
	vs[v] = true
}

func min(a, b float32) float32 {

	if a < b {
		return a
	} else {
		return b
	}
}

func max(a, b float32) float32 {
	if a > b {
		return a
	} else {
		return b
	}
}
