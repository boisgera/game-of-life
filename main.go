package main

import (
	"cart/w4"
	"math/rand"
	"strconv"
)

const (
	WIDTH  = 160
	HEIGHT = 160
)

type Color = uint8

func print(xs ...interface{}) {
	s := ""
	for _, x := range xs {
		switch v := x.(type) {
		case string:
			s += v + " "
		case int:
			s += strconv.Itoa(v) + " "
		case uint8:
			s += strconv.Itoa(int(v)) + " "
		case uint16:
			s += strconv.Itoa(int(v)) + " "
		case Image3x3: // todo: use Image interface and getPixel, with ellipses?
			s += "["
			for j := 0; j < 3; j++ {
				s += "["
				for i := 0; i < 3; i++ {
					array := [3][3]uint8(v)
					s += strconv.Itoa(int(array[i][j])) + ", "
				}
				s += "]"
			}
			s += "]"
		default:
			panic("can't do")
		}
	}
	w4.Trace(s[:len(s)-1])
}

// TODO: add pixel iterator and width & height functions
type Image interface {
	getPixel(x, y int) (color uint8)
	setPixel(x, y int, color uint8)
}

type Screen [WIDTH * HEIGHT / 4]uint8

func (buffer *Screen) getPixel(x, y int) (color uint8) {
	x %= WIDTH
	if x < 0 {
		x += WIDTH
	}
	y %= HEIGHT
	if y < 0 {
		y += HEIGHT
	}
	index := (x + y*WIDTH) >> 2
	shift := uint8((x & 0b11) << 1)
	mask := uint8(0b11 << shift)
	color = (buffer[index] & mask) >> shift
	return
}

func (buffer *Screen) setPixel(x, y int, color uint8) {
	x %= WIDTH
	if x < 0 {
		x += WIDTH
	}
	y %= HEIGHT
	if y < 0 {
		y += HEIGHT
	}
	index := (x + y*WIDTH) >> 2
	shift := uint8((x & 0b11) << 1)
	mask := uint8(0b11 << shift)
	buffer[index] = (color << shift) | (buffer[index] &^ mask)
}

func (buffer *Screen) clear() {
	for i := range *buffer {
		buffer[i] = 0
	}
}

var buffer = new(Screen)
var screen = (*Screen)(w4.FRAMEBUFFER)

type Image3x3 [3][3]uint8 // Not a multiple of four pixels, can't use
// and don't need to) use the compact layout used for the screen image.

func (image *Image3x3) getPixel(x, y int) (color uint8) {
	x = ((x % 3) + 3) % 3
	y = ((y % 3) + 3) % 3
	return image[x][y]
}

func (image *Image3x3) setPixel(x, y int, color uint8) {
	x = ((x % 3) + 3) % 3
	y = ((y % 3) + 3) % 3
	image[x][y] = color
}

func (screen *Screen) getImage3x3(x, y int) (image3x3 Image3x3) {
	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			color := screen.getPixel(x+i, y+j)
			image3x3.setPixel(i, j, color)
		}
	}
	return
}

func (screen *Screen) setImage3x3(x, y int, image3x3 Image3x3) {
	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			color := image3x3.getPixel(i, j)
			screen.setPixel(x+i, y+j, color)
		}
	}
}

type Rule func(image3x3 Image3x3) (color Color, match bool)

func (image3x3 *Image3x3) colorDistribution() (count map[Color]int) {
	count = make(map[Color]int)
	for _, color := range []Color{0, 1, 2, 3} {
		count[color] = 0
	}
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			color := image3x3.getPixel(x, y)
			count[color] += 1
		}
	}
	return
}

func grassSpreads(image3x3 Image3x3) (color Color, match bool) {
	if image3x3.getPixel(1, 1) == 0 {
		count := image3x3.colorDistribution()
		if count[1] >= 1 {
			match = true
			color = 1
		}
	}
	return
}

func rabbitsEatGrass(image3x3 Image3x3) (color Color, match bool) {
	if image3x3.getPixel(1, 1) == 1 {
		count := image3x3.colorDistribution()
		if count[2] >= 1 {
			match = true
			color = 2
		}
	}
	return
}

func tooManyRabbits(image3x3 Image3x3) (color Color, match bool) {
	if image3x3.getPixel(1, 1) == 3 {
		count := image3x3.colorDistribution()
		if count[2] >= 8 {
			match = true
			color = 0
		}
	}
	return
}

func tooManywolves(image3x3 Image3x3) (color Color, match bool) {
	if image3x3.getPixel(1, 1) == 3 {
		count := image3x3.colorDistribution()
		if count[3] >= 3 {
			match = true
			color = 0
		}
	}
	return
}

func wolvesKillRabbits(image3x3 Image3x3) (color Color, match bool) {
	if image3x3.getPixel(1, 1) == 2 {
		count := image3x3.colorDistribution()
		if count[3] >= 1 {
			match = true
			color = 3
		}
	}
	return
}

var rules = []Rule{
	grassSpreads,
	rabbitsEatGrass,
	tooManyRabbits,
	tooManywolves,
	wolvesKillRabbits,
}

func next() {
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			image3x3 := buffer.getImage3x3(x-1, y-1)
			for _, rule := range rules {
				color, match := rule(image3x3)
				if match {
					screen.setPixel(x, y, color)
					break
				}
			}
		}
	}
	*buffer = *screen
}

func initScreen() {
	for j := 0; j < HEIGHT; j++ {
		for i := 0; i < WIDTH; i++ {
			screen.setPixel(i, j, uint8(rand.Intn(4)))
		}
	}
}

//go:export start
func start() {
	w4.PALETTE[0] = 0xf8f9fa
	w4.PALETTE[1] = 0xc0eb75
	w4.PALETTE[2] = 0xffa94d
	w4.PALETTE[3] = 0x495057
}

var first = true

//go:export update
func update() {

	if first {
		initScreen()
		first = false
	} else {
		if *w4.MOUSE_BUTTONS&0b1 != 0 {
			x, y := int(*w4.MOUSE_X), int(*w4.MOUSE_Y)
			for j := 0; j < 21; j++ {
				for i := 0; i < 21; i++ {
					buffer.setPixel(x-10+i, y-10+j, 2)
				}
			}
		}
		*screen = *buffer
	}

	next()
}
