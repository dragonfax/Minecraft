package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const TICKS_PER_SEC = 60

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func enable_cpuprofile() {

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
}

func initGLFW() *glfw.Window {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glwindow, err := glfw.CreateWindow(640, 480, "Gocraft", nil, nil)
	if err != nil {
		panic(err)
	}

	glwindow.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	return glwindow
}

func get_time() float64 {
	// in seconds
	return float64(time.Now().UnixNano()) / float64(time.Second)
}

func main() {

	glwindow := initGLFW()
	defer glfw.Terminate()

	window := NewWindow(glwindow)

	enable_cpuprofile()

	last_time := get_time()
	for !glwindow.ShouldClose() {

		now := get_time()
		dt := now - last_time // in seconds
		if dt > 1/TICKS_PER_SEC {
			last_time = now
			window.update(float32(dt))
			window.on_draw()
		}

		glfw.PollEvents()
	}

}
