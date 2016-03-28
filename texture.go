package main

import (
	"image"
	"image/draw"
	"log"
	"os"

	"github.com/go-gl/gl/v2.1/gl"
)

const (
	TEXTURE_PATH = "texture.png"
)

func tex_coord(x, y float32) [4]Point2f {
	// Return the bounding vertices of the texture square.

	var n float32 = 4 // original 4,  1/4 is the size of a texture in the image
	m := 1.0 / n
	dx := x * m
	dy := y * m
	return [4]Point2f{{dx, dy}, {dx + m, dy}, {dx + m, dy + m}, {dx, dy + m}}
}

func tex_coords(topx, topy, bottomx, bottomy, sidex, sidey float32) []Point2f {
	// Return a list of the texture squares for the top, bottom and side.

	top := tex_coord(topx, topy)
	bottom := tex_coord(bottomx, bottomy)
	side := tex_coord(sidex, sidey)
	result := make([]Point2f, 24, 24)
	copy(result[0:4], top[:])
	copy(result[4:8], bottom[:])
	copy(result[8:12], side[:])
	copy(result[12:16], side[:])
	copy(result[16:20], side[:])
	copy(result[20:24], side[:])
	return result
}

func load_texture(file string) {

	imgFile, err := os.Open(file)
	if err != nil {
		log.Fatalf("texture %q not found on disk: %v\n", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		panic(err)
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		panic("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	// flip the image on the y axis.
	b := img.Bounds()
	rgba_flipped := image.NewRGBA(b)
	for y := b.Max.Y - 1; y >= b.Min.Y; y-- {
		for x := b.Max.X - 1; x >= b.Min.X; x-- {
			c := rgba.At(x, y)
			rgba_flipped.Set(x, b.Max.Y-y, c)
		}
	}

	var texture_id uint32
	gl.GenTextures(1, &texture_id)
	gl.BindTexture(gl.TEXTURE_2D, texture_id)

	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba_flipped.Pix))

}
