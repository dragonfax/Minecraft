package main

import (
	"math"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	WALKING_SPEED = 5
	FLYING_SPEED  = 15

	GRAVITY           = 20.0
	MAX_JUMP_HEIGHT   = 1.0 // About the height of a block.
	TERMINAL_VELOCITY = 50

	PLAYER_HEIGHT = 2

	// Size of sectors used to ease block loading.
	SECTOR_SIZE = 16
)

var JUMP_SPEED = float32(math.Sqrt(2 * GRAVITY * MAX_JUMP_HEIGHT))

type Window struct {
	glwindow  *glfw.Window
	exclusive bool
	flying    bool
	strafe    struct {
		x int
		y int
	}
	position Vertex
	rotation struct {
		x float32
		y float32
	}
	sector    Vertex
	reticle   CoordList
	dy        float32
	inventory []TextureType
	block     TextureType
	model     *Model
	num_keys  map[glfw.Key]int
}

func NewWindow(glwindow *glfw.Window) *Window {
	self := &Window{}
	self.glwindow = glwindow
	// func __init__(self, *args, **kwargs){
	// super(Window, self).__init__(*args, **kwargs)

	// Whether or not the window exclusively captures the mouse.
	self.exclusive = false

	// When flying gravity has no effect and speed is increased.
	self.flying = false

	// Strafing is moving lateral to the direction you are facing,
	// e.g. moving to the left or right while continuing to face forward.

	// First element is -1 when moving forward, 1 when moving back, and 0
	// otherwise. The second element is -1 when moving left, 1 when moving
	// right, and 0 otherwise.
	// self.strafe = [0, 0]

	// Current (x, y, z) position in the world, specified with floats. Note
	// that, perhaps unlike in math class, the y-axis is the vertical axis.
	self.position = NewVertex(0, 0, 0)

	// First element is rotation of the player in the x-z plane (ground
	// plane) measured from the z-axis down. The second is the rotation
	// angle from the ground plane up. Rotation is in degrees.

	// The vertical plane rotation ranges from -90 (looking straight down) to
	// 90 (looking straight up). The horizontal rotation range is unbounded.
	// self.rotation = (0, 0)

	// Which sector the player is currently in.
	self.sector = nilVertex

	// The crosshairs at the center of the screen.
	self.reticle = CoordList{}

	// Velocity in the y (upward) direction.
	self.dy = 0

	// A list of blocks the player can place. Hit num keys to cycle.
	self.inventory = []TextureType{BRICK, GRASS, SAND}

	// The current block the user can place. Hit num keys to cycle.
	self.block = self.inventory[0]

	// Convenience list of num keys.
	self.num_keys = map[glfw.Key]int{glfw.Key1: 0, glfw.Key2: 1, glfw.Key3: 2, glfw.Key4: 3, glfw.Key5: 4, glfw.Key6: 5, glfw.Key7: 6, glfw.Key8: 7, glfw.Key9: 8, glfw.Key0: 9}

	// Instance of the model that handles the world.
	self.model = NewModel()

	// The label that is displayed in the top left of the canvas.
	// self.label = NewLabel("", font_name="Arial", font_size=18, x=10, y=self.height - 10, anchor_x="left", anchor_y="top", color=(0, 0, 0, 255))

	// This call schedules the `update()` method to be called
	// TICKS_PER_SEC. This is the main game event loop.
	// pyglet.clock.schedule_interval(self.update, 1.0 / TICKS_PER_SEC)
	// schedule_interval(self.update, 1.0/TICKS_PER_SEC)

	// Hide the mouse cursor and prevent the mouse from leaving the window.
	self.set_exclusive_mouse(true)

	gl.ClearColor(0.5, 0.69, 1.0, 1)
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.TEXTURE_2D)
	load_texture(TEXTURE_PATH)

	glwindow.SetCursorPosCallback(func(w *glfw.Window, xpos float64, ypos float64) {
		self.on_mouse_motion(xpos, ypos)
	})
	glwindow.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			self.on_key_press(key, mods)
		} else if action == glfw.Press {
			self.on_key_release(key, mods)
		}
	})
	glwindow.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		if action == glfw.Press {
			x, y := w.GetCursorPos()
			self.on_mouse_press(x, y, button, mod)
		}
	})
	glwindow.SetSizeCallback(func(w *glfw.Window, width int, height int) {
		self.on_resize(width, height)
	})

	setup_fog()

	return self
}

func (self *Window) size() (int, int) {
	return self.glwindow.GetSize()
}

func (self *Window) width() int {
	width, _ := self.size()
	return width
}

func (self *Window) height() int {
	_, height := self.size()
	return height
}

func (self *Window) set_exclusive_mouse(exclusive bool) {
	/* If `exclusive` is true, the game will capture the mouse, if false
	   the game will ignore the mouse.

	*/
	value := glfw.CursorNormal
	if exclusive {
		value = glfw.CursorDisabled
	}
	if self.glwindow == nil {
		panic("no glwindow set yet")
	}
	self.glwindow.SetInputMode(glfw.CursorMode, value)
	self.exclusive = exclusive
}

func (self *Window) get_sight_vector() Vertex {
	// Returns the current line of sight vector indicating the direction the player is looking.

	//
	// y ranges from -90 to 90, or -pi/2 to pi/2, so m ranges from 0 to 1 and
	// is 1 when looking ahead parallel to the ground and 0 when looking
	// straight up or down.
	m := math.Cos(radians(float64(self.rotation.y)))
	// dy ranges from -1 to 1 and is -1 when looking straight down and 1 when
	// looking straight up.
	dy := math.Sin(radians(float64(self.rotation.y)))
	dx := math.Cos(radians(float64(self.rotation.x-90))) * m
	dz := math.Sin(radians(float64(self.rotation.x-90))) * m
	return NewVertex(float32(dx), float32(dy), float32(dz))
}

func (self *Window) get_motion_vector() Vertex {
	/* Returns the current motion vector indicating the velocity of the
	   player.

	   Returns
	   -------
	   vector : tuple of len 3
	       Tuple containing the velocity in x, y, and z respectively.

	*/
	var dy, dx, dz float64
	if self.strafe.x != 0 || self.strafe.y != 0 {
		strafe := degrees(math.Atan2(float64(self.strafe.x), float64(self.strafe.y)))
		y_angle := radians(float64(self.rotation.y))
		x_angle := radians(float64(self.rotation.x) + strafe)
		if self.flying {
			m := math.Cos(y_angle)
			dy = math.Sin(y_angle)
			if self.strafe.y != 0 {
				// Moving left or right.
				dy = 0.0
				m = 1
			}
			if self.strafe.x > 0 {
				// Moving backwards.
				dy *= -1
			}
			// When you are flying up or down, you have less left and right
			// motion.
			dx = math.Cos(x_angle) * m
			dz = math.Sin(x_angle) * m
		} else {
			dy = 0.0
			dx = math.Cos(x_angle)
			dz = math.Sin(x_angle)
		}
	} else {
		dy = 0.0
		dx = 0.0
		dz = 0.0
	}
	return NewVertex(float32(dx), float32(dy), float32(dz))
}

func (self *Window) update(dt float32) {
	/* This method is scheduled to be called repeatedly by the pyglet
	   clock.

	   Parameters
	   ----------
	   dt : float
	       The change in time since the last call.

	*/
	sector := sectorize(self.position)
	if sector != self.sector {
		// TODO self.model.change_sectors(self.sector, sector)
		self.sector = sector
	}
	m := 8
	dt = min(dt, 0.2)
	for _ = range xrange(0, m, 1) {
		self._update(dt / float32(m))
	}
}

func (self *Window) _update(dt float32) {
	/* Private implementation of the `update()` method. This is where most
	   of the motion logic lives, along with gravity and collision detection.

	   Parameters
	   ----------
	   dt : float
	       The change in time since the last call.

	*/
	// walking
	var speed float32
	if self.flying {
		speed = FLYING_SPEED
	} else {
		speed = WALKING_SPEED
	}
	d := dt * speed // distance covered this tick.
	dv := self.get_motion_vector()
	// New position in space, before accounting for gravity.
	dx, dy, dz := dv.x*d, dv.y*d, dv.z*d
	// gravity
	if !self.flying {
		// Update your vertical speed: if you are falling, speed up until you
		// hit terminal velocity; if you are jumping, slow down until you
		// start falling.
		self.dy -= dt * GRAVITY
		self.dy = max(self.dy, -TERMINAL_VELOCITY)
		dy += self.dy * dt
	}
	// collisions
	self.position = self.collide(NewVertex(self.position.x+dx, self.position.y+dy, self.position.z+dz), PLAYER_HEIGHT)
}

const PAD = 0.25

func (self *Window) collide(position Vertex, height int) Vertex {
	/* Checks to see if the player at the given `position` and `height`
	   is colliding with any blocks in the world.

	   Parameters
	   ----------
	   position : tuple of len 3
	       The (x, y, z) position to check for collisions at.
	   height : int or float
	       The height of the player.

	   Returns
	   -------
	   position : tuple of len 3
	       The new position of the player taking into account collisions.

	*/
	// How much overlap with a dimension of a surrounding block you need to
	// have to count as a collision. If 0, touching terrain at all counts as
	// a collision. If .49, you sink into the ground, as if walking through
	// tall grass. If >= .5, you"ll fall through the ground.
	p := position
	np := normalize(position)
	for _, face := range FACES { // check all surrounding blocks
		for _, i := range xrange(0, 3, 1) { // check each dimension independently
			if face.get(i) == 0 {
				continue
			}
			// How uch overlap you have with this dimension.
			d := (p.get(i) - np.get(i)) * face.get(i)
			if d < PAD {
				continue
			}
			for _, dy := range xrange(0, height, 1) { // check each height
				op := np
				op.set(1, op.get(1)-float32(dy))
				op.set(i, op.get(i)+face.get(i))
				if _, ok := self.model.world[op]; !ok {
					continue
				}
				p.set(i, p.get(i)-(d-PAD)*face.get(i))
				if face == NewVertex(0, -1, 0) || face == NewVertex(0, 1, 0) {
					// You are colliding with the ground or ceiling, so stop
					// falling / rising.
					self.dy = 0
				}
				break
			}
		}
	}
	return p
}

func (self *Window) on_mouse_press(x, y float64, button glfw.MouseButton, modifiers glfw.ModifierKey) {
	/* Called when a mouse button is pressed. See pyglet docs for button
	amd modifier mappings.

	Parameters
	----------
	x, y : int
	The coordinates of the mouse click. Always center of the screen if
	the mouse is captured.
	button : int
	Number representing mouse button that was clicked. 1 = left button,
	4 = right button.
	modifiers : int
	Number representing any modifying keys that were pressed when the
	mouse button was clicked.

	*/
	if self.exclusive {
		vector := self.get_sight_vector()
		block, previous := self.model.hit_test(self.position, vector, 8)
		if (button == glfw.MouseButtonRight) || ((button == glfw.MouseButtonLeft) && (modifiers&glfw.ModControl) != 0) {
			// ON OSX, control + left click = right click.
			if !previous.isNil() {
				self.model.add_block(previous, self.block)
			}
		} else if button == glfw.MouseButtonLeft && !block.isNil() {
			texture := self.model.world[block]
			if texture != STONE {
				self.model.remove_block(block)
			}
		}
	} else {
		self.set_exclusive_mouse(true)
	}
}

var last_mouse_x, last_mouse_y float64 = 0, 0

func (self *Window) on_mouse_motion(x, y float64) {
	/* Called when the player moves the mouse.

	   Parameters
	   ----------
	   x, y : int
	       The coordinates of the mouse click. Always center of the screen if
	       the mouse is captured.
	   dx, dy : float
	       The movement of the mouse.

	*/

	dx, dy := x-last_mouse_x, -(y - last_mouse_y)
	last_mouse_x, last_mouse_y = x, y

	if self.exclusive {
		m := float32(0.15)
		x, y := float32(self.rotation.x)+float32(dx)*m, float32(self.rotation.y)+float32(dy)*m
		y = max(-90, min(90, y))
		self.rotation.x = x
		self.rotation.y = y
	}
}

func (self *Window) on_key_press(symbol glfw.Key, modifiers glfw.ModifierKey) {
	/* Called when the player presses a key. See pyglet docs for key
	   mappings.

	   Parameters
	   ----------
	   symbol : int
	       Number representing the key that was pressed.
	   modifiers : int
	       Number representing any modifying keys that were pressed.

	*/

	if symbol == glfw.KeyS {
		self.strafe.x -= 1
	} else if symbol == glfw.KeyW {
		self.strafe.x += 1
	} else if symbol == glfw.KeyD {
		self.strafe.y -= 1
	} else if symbol == glfw.KeyA {
		self.strafe.y += 1
	} else if symbol == glfw.KeySpace {
		if self.dy == 0 {
			self.dy = JUMP_SPEED
		}
	} else if symbol == glfw.KeyEscape {
		self.set_exclusive_mouse(false)
	} else if symbol == glfw.KeyTab {
		self.flying = !self.flying
	} else if _, ok := self.num_keys[symbol]; ok {
		index := self.num_keys[symbol] % len(self.inventory)
		self.block = self.inventory[index]
	}
}

func (self *Window) on_key_release(symbol glfw.Key, modifiers glfw.ModifierKey) {
	/* Called when the player releases a key. See pyglet docs for key
	   mappings.

	   Parameters
	   ----------
	   symbol : int
	       Number representing the key that was pressed.
	   modifiers : int
	       Number representing any modifying keys that were pressed.

	*/
	if symbol == glfw.KeyS {
		self.strafe.x += 1
	} else if symbol == glfw.KeyW {
		self.strafe.x -= 1
	} else if symbol == glfw.KeyD {
		self.strafe.y += 1
	} else if symbol == glfw.KeyA {
		self.strafe.y -= 1
	}
}

func (self *Window) on_resize(width, height int) {
	// Called when the window is resized to a new `width` and `height`.

	// label
	// self.label.y = height - 10
	// reticle
	/*if self.reticle != nil {
		self.reticle.delete()
	}*/
	x, y := self.width()/2, self.height()/2
	n := 10
	self.reticle = NewCoordList([]Coord{{x - n, y}, {x + n, y}, {x, y - n}, {x, y + n}})
}

func (self *Window) set_2d() {
	// Configure OpenGL to draw in 2d.

	//
	gl.Disable(gl.DEPTH_TEST)
	gl.Viewport(0, 0, int32(self.width()), int32(self.height()))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, float64(self.width()), 0, float64(self.height()), -1, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

func (self *Window) set_3d() {
	// Configure OpenGL to draw in 3d.

	//
	gl.Enable(gl.DEPTH_TEST)
	gl.Viewport(0, 0, int32(self.width()*2), int32(self.height()*2))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()

	gluPerspective(45.0, float32(self.width())/float32(self.height()), 0.1, 60.0)

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Rotatef(self.rotation.x, 0, 1, 0)
	gl.Rotatef(-self.rotation.y, float32(math.Cos(radians(float64(self.rotation.x)))), 0, float32(math.Sin(radians(float64(self.rotation.x)))))
	gl.Translatef(-self.position.x, -self.position.y, -self.position.z)
}

func gluPerspective(fovy, aspect, near, far float32) {
	mat := mgl32.Perspective(fovy, aspect, near, far)
	gl.LoadMatrixf(&mat[0])
}

func (self *Window) on_draw() {
	// Called by pyglet to draw the canvas.

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	self.set_3d()
	self.model.batch.draw()
	self.draw_focused_block()
	self.set_2d()
	// self.draw_label()
	self.draw_reticle()
	self.glwindow.SwapBuffers()
}

func (self *Window) draw_focused_block() {
	// Draw black edges around the block that is currently under the crosshairs.

	//
	vector := self.get_sight_vector()
	block, _ := self.model.hit_test(self.position, vector, 8)
	if !block.isNil() {
		vertex_data := cube_vertices(block, 0.51)
		gl.Color3d(0, 0, 0)
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		drawPolygon(gl.QUADS, vertex_data)
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}
}

/*
func (self *Window) draw_label() {
	// Draw the label in the top left of the screen.

	self.label.text = fmt.Sprintf("%02d (%.2f, %.2f, %.2f) %d / %d",
		pyglet.clock.get_fps(), self.position.x, self.position.y, self.position.z,
		len(self.model._shown), len(self.model.world))
	self.label.draw()
}
*/

func (self *Window) draw_reticle() {
	// Draw the crosshairs in the center of the screen.

	//
	gl.Color3d(0, 0, 0)
	self.reticle.draw(gl.LINES)
}
