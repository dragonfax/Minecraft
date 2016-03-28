package main

import "github.com/go-gl/gl/v2.1/gl"

var FACES = []Vertex{
	NewVertex(0, 1, 0),
	NewVertex(0, -1, 0),
	NewVertex(-1, 0, 0),
	NewVertex(1, 0, 0),
	NewVertex(0, 0, 1),
	NewVertex(0, 0, -1),
}

var grass = tex_coords(1, 0, 0, 1, 0, 0)
var sand = tex_coords(1, 1, 1, 1, 1, 1)
var brick = tex_coords(2, 0, 2, 0, 2, 0)
var stone = tex_coords(2, 1, 2, 1, 2, 1)

var textures = map[TextureType][]Point2f{GRASS: grass, SAND: sand, BRICK: brick, STONE: stone}

type TextureType int

const (
	GRASS TextureType = iota
	SAND
	BRICK
	STONE
)

type Model struct {
	world   map[Vertex]TextureType
	shown   map[Vertex]TextureType
	_shown  map[Vertex]CallList
	sectors map[Vertex][]Vertex

	// texture *Texture
	batch *Batch
}

func NewModel() *Model {
	self := &Model{}

	// A Batch is a collection of vertex lists for batched rendering.
	self.batch = NewBatch() /*pyglet.graphics.Batch() */

	// A mapping from position to the texture of the block at that position.
	// This defines all the blocks that are currently in the world.
	self.world = make(map[Vertex]TextureType)

	// Same mapping as `world` but only contains blocks that are shown.
	self.shown = make(map[Vertex]TextureType)

	// Mapping from position to a VertextList for all shown blocks.
	self._shown = make(map[Vertex]CallList)

	// Mapping from sector to a list of positions inside that sector.
	self.sectors = make(map[Vertex][]Vertex)

	self.build_world()

	return self
}

func random_choice(types []TextureType) TextureType {
	n := random_randint(0, len(types))
	return types[n]
}

func (self *Model) build_world() {
	// Initialize the world by placing all the blocks.

	n := 80 // 1/2 width and height of world
	s := 1  // step size
	y := 0  // initial y height
	for _, x := range xrange(-n, n+1, s) {
		for _, z := range xrange(-n, n+1, s) {
			// create a layer stone an grass everywhere.
			self.add_block(NewVertexInt(x, y-2, z), GRASS)
			self.add_block(NewVertexInt(x, y-3, z), STONE)
			if x == -n || x == n || z == -n || z == n {
				// create outer walls.
				for _, dy := range xrange(-2, 3, 1) {
					self.add_block(NewVertexInt(x, y+dy, z), STONE)
				}
			}
		}
	}

	// generate the hills randomly
	o := n - 10
	for _ = range xrange(0, 120, 1) {
		a := random_randint(-o, o) // x position of the hill
		b := random_randint(-o, o) // z position of the hill
		c := -1                    // base of the hill
		h := random_randint(1, 6)  // height of the hill
		s := random_randint(4, 8)  // 2 * s is the side length of the hill
		d := 1                     // how quickly to taper off the hills
		t := random_choice([]TextureType{GRASS, SAND, BRICK})
		for _, y := range xrange(c, c+h, 1) {
			for _, x := range xrange(a-s, a+s+1, 1) {
				for _, z := range xrange(b-s, b+s+1, 1) {
					if PowInt((x-a), 2)+PowInt((z-b), 2) > PowInt((s+1), 2) {
						continue
					}
					if PowInt((x-0), 2)+PowInt((z-0), 2) < PowInt(5, 2) {
						continue
					}
					self.add_block(NewVertexInt(x, y, z), t)
				}
			}
			s -= d // decrement side lenth so hills taper off
		}

	}
}

func (self *Model) hit_test(position Vertex, vector Vertex, max_distance int /*=8*/) (Vertex, Vertex) {
	/* Line of sight search from current position. If a block is
	   intersected it is returned, along with the block previously in the line
	   of sight. If no block is found, return None, None.

	   Parameters
	   ----------
	   position : tuple of len 3
	       The (x, y, z) position to check visibility from.
	   vector : tuple of len 3
	       The line of sight vector.
	   max_distance : int
	       How many blocks away to search for a hit.

	*/
	m := 8
	previous := nilVertex
	x, y, z := position.x, position.y, position.z
	for _ = range xrange(0, max_distance*m, 1) {
		key := normalize(NewVertex(x, y, z))
		if _, ok := self.world[key]; key != previous && ok {
			return key, previous
		}
		previous = key
		x, y, z = x+vector.x/float32(m), y+vector.y/float32(m), z+vector.z/float32(m)
	}
	return nilVertex, nilVertex
}

func (self *Model) exposed(position Vertex) bool {
	/* Returns false is given `position` is surrounded on all 6 sides by
	   blocks, true otherwise.

	*/
	for _, d := range FACES {
		if _, ok := self.world[NewVertex(position.x+d.x, position.y+d.y, position.z+d.z)]; !ok {
			return true
		}
	}
	return false
}

func (self *Model) add_block(position Vertex, texture TextureType) {
	/* Add a block with the given `texture` and `position` to the world.

	   Parameters
	   ----------
	   position : tuple of len 3
	       The (x, y, z) position of the block to add.
	   texture : list of len 3
	       The coordinates of the texture squares. Use `tex_coords()` to
	       generate.
	   immediate : bool
	       Whether or not to draw the block immediately.

	*/
	if _, ok := self.world[position]; ok {
		self.remove_block(position)
	}
	self.world[position] = texture

	s := sectorize(position)
	if _, ok := self.sectors[s]; !ok {
		self.sectors[s] = []Vertex{position}
	} else {
		self.sectors[s] = append(self.sectors[s], position)
	}

	if self.exposed(position) {
		self.show_block(position)
	}
	self.check_neighbors(position)
}

func (self *Model) remove_block(position Vertex) {
	/* Remove the block at the given `position`.

	   Parameters
	   ----------
	   position : tuple of len 3
	       The (x, y, z) position of the block to remove.
	   immediate : bool
	       Whether or not to immediately remove block from canvas.

	*/
	delete(self.world, position)
	sector_id := sectorize(position)
	sector_data := self.sectors[sector_id]

	// remove the vector from the []vector
	for i, v := range sector_data {
		if v == position {
			self.sectors[sector_id] = append(sector_data[:i], sector_data[i+1:]...)
			break
		}
	}

	if _, ok := self.shown[position]; ok {
		self.hide_block(position)
	}
	self.check_neighbors(position)
}

func (self *Model) check_neighbors(position Vertex) {
	/* Check all blocks surrounding `position` and ensure their visual
	   state is current. This means hiding blocks that are not exposed and
	   ensuring that all exposed blocks are shown. Usually used after a block
	   is added or removed.

	*/
	for _, d := range FACES {
		key := NewVertex(position.x+d.x, position.y+d.y, position.z+d.z)
		if _, ok := self.world[key]; !ok {
			continue
		}
		if self.exposed(key) {
			if _, ok := self.shown[key]; !ok {
				self.show_block(key)
			}
		} else {
			if _, ok := self.shown[key]; ok {
				self.hide_block(key)
			}
		}
	}
}

func (self *Model) show_block(position Vertex) {
	/* Show the block at the given `position`. This method assumes the
	   block has already been added with add_block()

	   Parameters
	   ----------
	   position : tuple of len 3
	       The (x, y, z) position of the block to show.
	   immediate : bool
	       Whether or not to show the block immediately.

	*/
	texture := self.world[position]
	self.shown[position] = texture
	self._show_block(position, texture)
}

func (self *Model) _show_block(position Vertex, texture TextureType) {
	/* Private implementation of the `show_block()` method.

	   Parameters
	   ----------
	   position : tuple of len 3
	       The (x, y, z) position of the block to show.
	   texture : list of len 3
	       The coordinates of the texture squares. Use `tex_coords()` to
	       generate.

	*/
	vertex_data := cube_vertices(position, 0.5)
	texture_data := textures[texture]
	// create vertex list
	self._shown[position] = self.batch.add(gl.QUADS, vertex_data, texture_data)
}

func (self *Model) hide_block(position Vertex) {
	/* Hide the block at the given `position`. Hiding does not remove the
	   block from the world.

	   Parameters
	   ----------
	   position : tuple of len 3
	       The (x, y, z) position of the block to hide.
	   immediate : bool
	       Whether or not to immediately remove the block from the canvas.

	*/
	delete(self.shown, position)
	self._hide_block(position)
}

func (self *Model) _hide_block(position Vertex) {
	// Private implementation of the "hide_block()` method.

	self._shown[position].delete()
	delete(self._shown, position)
}

func (self *Model) show_sector(sector Vertex) {
	// Ensure all blocks in the given sector that should be shown are drawn to the canvas.

	//
	for _, position := range self.sectors[sector] {
		if _, ok := self.shown[position]; !ok && self.exposed(position) {
			self.show_block(position)
		}
	}
}

func (self *Model) hide_sector(sector Vertex) {
	// Ensure all blocks in the given sector that should be hidden are removed from the canvas.

	//
	for _, position := range self.sectors[sector] {
		if _, ok := self.shown[position]; ok {
			self.hide_block(position)
		}
	}
}

func (self *Model) change_sectors(before Vertex, after Vertex) {
	/* Move from sector `before` to sector `after`. A sector is a
	   contiguous x, y sub-region of world. Sectors are used to speed up
	   world rendering.

	*/
	before_set := NewVertexSet()
	after_set := NewVertexSet()
	pad := 4
	for _, dx := range xrange(-pad, pad+1, 1) {
		dy := 0 // xrange(-pad, pad + 1){
		for _, dz := range xrange(-pad, pad+1, 1) {
			if PowInt(dx, 2)+PowInt(dy, 2)+PowInt(dz, 2) > PowInt((pad+1), 2) {
				continue
			}
			if !before.isNil() {
				before_set.add(NewVertex(before.x+float32(dx), before.y+float32(dy), before.z+float32(dz)))
			}
			if !after.isNil() {
				after_set.add(NewVertex(after.x+float32(dx), after.y+float32(dy), after.z+float32(dz)))
			}
		}
	}
	show := after_set.Remove(before_set)
	hide := before_set.Remove(after_set)
	for sector := range show {
		self.show_sector(sector)
	}
	for sector := range hide {
		self.hide_sector(sector)
	}
}
