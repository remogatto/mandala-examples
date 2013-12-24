package cubelib

type Reshaper interface {
	Resize(width, height int)
	Width() int
	Height() int
}
