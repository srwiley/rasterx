// Copyright 2018 by the rasterx Authors. All rights reserved.
//_
// Created 2017 by S.R.Wiley
// This version of Scanner wraps the golang.org/x/image/vector
// rasterizer.

package rasterx

import (
	"image"
	"math"

	"image/color"

	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

type (
	ColorFuncImage struct {
		image.Uniform
		colorFunc ColorFunc
	}

	// Rasterizer converts a path to a raster using the grainless algorithm.
	ScannerGV struct {
		r vector.Rasterizer
		//a, first fixed.Point26_6
		Dest                   *image.RGBA
		Targ                   image.Rectangle
		colFImage              *ColorFuncImage
		Source                 image.Image
		Offset                 image.Point
		minX, minY, maxX, maxY fixed.Int26_6 // keep track of bounds
	}
)

func (s *ScannerGV) GetPathExtent() fixed.Rectangle26_6 {
	return fixed.Rectangle26_6{fixed.Point26_6{s.minX, s.minY}, fixed.Point26_6{s.maxX, s.maxY}}
}

func (c *ColorFuncImage) At(x, y int) color.Color {
	return c.colorFunc(x, y)
}

func (s *ScannerGV) SetWinding(useNonZeroWinding bool) {
	// no-op as scanner gv does not support even-odd winding
}

func (s *ScannerGV) SetColor(clr interface{}) {
	switch c := clr.(type) {
	case color.Color:
		s.Source = &s.colFImage.Uniform
		s.colFImage.Uniform.C = c
	case ColorFunc:
		s.Source = s.colFImage
		s.colFImage.colorFunc = c
	}
}

func (s *ScannerGV) set(a fixed.Point26_6) {
	if s.maxX < a.X {
		s.maxX = a.X
	}
	if s.maxY < a.Y {
		s.maxY = a.Y
	}
	if s.minX > a.X {
		s.minX = a.X
	}
	if s.minY > a.Y {
		s.minY = a.Y
	}
}

// Start starts a new path at the given point.
func (s *ScannerGV) Start(a fixed.Point26_6) {
	s.set(a)
	s.r.MoveTo(float32(a.X)/64, float32(a.Y)/64)
}

// Line adds a linear segment to the current curve.
func (s *ScannerGV) Line(b fixed.Point26_6) {
	s.set(b)
	s.r.LineTo(float32(b.X)/64, float32(b.Y)/64)
}

func (s *ScannerGV) Draw() {
	// This draws the entire bounds of the image, because
	// at this point the alpha mask does not shift with the
	// placement of the target rectangle in the vector rasterizer
	s.r.Draw(s.Dest, s.Dest.Bounds(), s.Source, s.Offset)

	// Remove the line above and uncomment the lines below if you
	// are using a version of the vector rasterizer that shifts the alpha
	// mask with the placement of the target

	//	s.Targ.Min.X = int(s.minX >> 6)
	//	s.Targ.Min.Y = int(s.minY >> 6)
	//	s.Targ.Max.X = int(s.maxX>>6) + 1
	//	s.Targ.Max.Y = int(s.maxY>>6) + 1
	//	s.Targ = s.Targ.Intersect(s.Dest.Bounds())  // This check should be done by the rasterizer?
	//	s.r.Draw(s.Dest, s.Targ, s.Source, s.Offset)
}

// Clear cancels any previous accumulated scans
func (s *ScannerGV) Clear() {
	p := s.r.Size()
	s.r.Reset(p.X, p.Y)
	const mxfi = fixed.Int26_6(math.MaxInt32)
	s.minX, s.minY, s.maxX, s.maxY = mxfi, mxfi, -mxfi, -mxfi
}

// SetBounds sets the maximum width and height of the rasterized image and
// calls Clear. The width and height are in pixels, not fixed.Int26_6 units.
func (s *ScannerGV) SetBounds(width, height int) {
	s.r.Reset(width, height)
}

// NewScanner creates a new Scanner with the given bounds.
func NewScannerGV(width, height int, dest *image.RGBA,
	targ image.Rectangle) *ScannerGV {
	s := new(ScannerGV)
	s.SetBounds(width, height)
	s.Dest = dest
	s.Targ = targ
	s.colFImage = &ColorFuncImage{}
	s.colFImage.Uniform.C = &color.RGBA{255, 0, 0, 255}
	s.Source = &s.colFImage.Uniform
	s.Offset = image.Point{0, 0}
	return s
}
