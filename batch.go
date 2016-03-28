package main

import (
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"

	_ "image/png"
)

func cube_vertices(block Vertex, n float32) []Vertex {
	// Return the vertices of the cube at position x, y, z with size 2*n.

	return []Vertex{
		NewVertex(block.x-n, block.y+n, block.z-n), NewVertex(block.x-n, block.y+n, block.z+n), NewVertex(block.x+n, block.y+n, block.z+n), NewVertex(block.x+n, block.y+n, block.z-n), // top
		NewVertex(block.x-n, block.y-n, block.z-n), NewVertex(block.x+n, block.y-n, block.z-n), NewVertex(block.x+n, block.y-n, block.z+n), NewVertex(block.x-n, block.y-n, block.z+n), // bottom
		NewVertex(block.x-n, block.y-n, block.z-n), NewVertex(block.x-n, block.y-n, block.z+n), NewVertex(block.x-n, block.y+n, block.z+n), NewVertex(block.x-n, block.y+n, block.z-n), // left
		NewVertex(block.x+n, block.y-n, block.z+n), NewVertex(block.x+n, block.y-n, block.z-n), NewVertex(block.x+n, block.y+n, block.z-n), NewVertex(block.x+n, block.y+n, block.z+n), // right
		NewVertex(block.x-n, block.y-n, block.z+n), NewVertex(block.x+n, block.y-n, block.z+n), NewVertex(block.x+n, block.y+n, block.z+n), NewVertex(block.x-n, block.y+n, block.z+n), // front
		NewVertex(block.x+n, block.y-n, block.z-n), NewVertex(block.x-n, block.y-n, block.z-n), NewVertex(block.x-n, block.y+n, block.z-n), NewVertex(block.x+n, block.y+n, block.z-n), // back
	}
}

func drawPolygon(gl_mode uint32, vertex_data []Vertex) {
	gl.Begin(gl_mode)
	gl.Color3f(0.5, 0.5, 0.5)
	for _, v := range vertex_data {
		gl.Vertex3f(v.x, v.y, v.z)
	}
	gl.End()
}

type Point2i struct {
	x, y int
}

type Point2f struct {
	x, y float32
}

func drawPoint2i(gl_mode uint32, cl []Point2i) {
	gl.Begin(gl_mode)
	gl.Color3f(0.5, 0.5, 0.5)
	for _, c := range cl {
		gl.Vertex2i(int32(c.x), int32(c.y))
	}
	gl.End()
}

type Batch struct {
	lists []CallList
}

func NewBatch() *Batch {
	return &Batch{lists: make([]CallList, 0, 10)}
}

func (b *Batch) add(gl_mode uint32, vertex_data []Vertex, texture_data []Point2f) CallList {
	bvl := NewCallList(b, gl_mode, vertex_data, texture_data)
	b.lists = append(b.lists, bvl)
	return bvl
}

func (b *Batch) draw() {

	num := int32(len(b.lists))

	id_list := make([]uint32, num, num)
	for i, list := range b.lists {
		id_list[i] = list.list_index
	}

	gl.CallLists(num, gl.UNSIGNED_INT, unsafe.Pointer(&id_list[0]))
}

type CallList struct {
	parent     *Batch
	list_index uint32
}

var last_bvl_id = 0

func NewCallList(b *Batch, gl_mode uint32, vertex_data []Vertex, texture_data []Point2f) CallList {
	bvl := CallList{parent: b}

	list_index := gl.GenLists(1)
	gl.NewList(list_index, gl.COMPILE)
	bvl.draw(gl_mode, vertex_data, texture_data)
	gl.EndList()

	bvl.list_index = list_index

	return bvl
}

func (bvl CallList) delete() {
	// free the structure, if necessary
	// removes itself from the batch its part of as well.
	b := bvl.parent
	for i, bvl1 := range b.lists {
		if bvl.list_index == bvl1.list_index {
			b.lists = append(b.lists[:i], b.lists[i+1:]...)
			return
		}
	}
	gl.DeleteLists(bvl.list_index, 1)
}

func (bvl CallList) draw(gl_mode uint32, vl []Vertex, texture_data []Point2f) {
	gl.Begin(gl_mode)
	for i, v := range vl {
		t := texture_data[i]
		gl.TexCoord2f(t.x, t.y)
		gl.Vertex3f(v.x, v.y, v.z)
	}
	gl.End()
}
