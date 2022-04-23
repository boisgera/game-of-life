package main

import (
	"cart/w4"
	"strconv"
)

const (
	WIDTH  = 160
	HEIGHT = 160
)

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

// TODO. Redefine Mask as 3x3 buffer and make corresponding method getPixel,
// setPixels. Define an interface(?). And make some getSubScreen, setSubScreen
// stuff?

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

func next() {
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			image3x3 := buffer.getImage3x3(x-1, y-1)
			color := image3x3.getPixel(1, 1)
			if color > 0 {
				for j := 0; j < 3; j++ {
					for i := 0; i < 3; i++ {
						image3x3.setPixel(i, j, color)
					}
				}
				screen.setImage3x3(x-1, y-1, image3x3)
			}
		}
	}
	*buffer = *screen
}

//go:export update
func update() {
	if *w4.MOUSE_BUTTONS&0b1 != 0 {
		x, y := int(*w4.MOUSE_X), int(*w4.MOUSE_Y)
		buffer.setPixel(x, y, 3)
	}
	next()
}
