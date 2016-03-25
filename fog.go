package main

import "github.com/go-gl/gl/v2.1/gl"

var fog_color = []float32{0.5, 0.69, 1.0, 1}

func setup_fog() {
	// Configure the OpenGL fog properties.

	//
	// Enable fog. Fog "blends a fog color with each rasterized pixel fragment"s
	// post-texturing color."
	gl.Enable(gl.FOG)
	// Set the fog color.
	gl.Fogfv(gl.FOG_COLOR, &fog_color[0])
	// Say we have no preference between rendering speed and quality.
	gl.Hint(gl.FOG_HINT, gl.DONT_CARE)
	// Specify the equation used to compute the blending factor.
	gl.Fogi(gl.FOG_MODE, gl.LINEAR)
	// How close and far away fog starts and ends. The closer the start and end,
	// the denser the fog in the fog range.
	gl.Fogf(gl.FOG_START, 20.0)
	gl.Fogf(gl.FOG_END, 60.0)
}
