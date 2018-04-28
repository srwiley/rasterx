// Copyright 2018 by the rasterx Authors. All rights reserved.
//_
// Created 2017 by S.R.Wiley
// This version of Scanner wraps the golang.org/x/image/vector
// rasterizer.

package rasterx

import (
	"image"

	"image/color"
	"image/draw"

	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

type (

	// Rasterizer converts a path to a raster using the grainless algorithm.
	ScannerGV struct {
		r vector.Rasterizer
		//a, first fixed.Point26_6
		Dest   draw.Image
		Targ   image.Rectangle
		Source *image.Uniform
		Offset image.Point
	}
)

func (s *ScannerGV) SetWinding(useNonZeroWinding bool) {
	// no-op as scanner gv does not support even-odd winding
}

func (s *ScannerGV) SetColor(c color.Color) {
	s.Source.C = c
}

// Start starts a new path at the given point.
func (s *ScannerGV) Start(a fixed.Point26_6) {
	s.r.MoveTo(float32(a.X)/64, float32(a.Y)/64)
}

// Line adds a linear segment to the current curve.
func (s *ScannerGV) Line(b fixed.Point26_6) {
	s.r.LineTo(float32(b.X)/64, float32(b.Y)/64)
	//s.a = b
}

func (s *ScannerGV) Draw() {
	s.r.Draw(s.Dest, s.Targ, s.Source, s.Offset)
}

// Clear cancels any previous accumulated scans
func (s *ScannerGV) Clear() {
	p := s.r.Size()
	s.r.Reset(p.X, p.Y)
}

// SetBounds sets the maximum width and height of the rasterized image and
// calls Clear. The width and height are in pixels, not fixed.Int26_6 units.
func (s *ScannerGV) SetBounds(width, height int) {
	s.r.Reset(width, height)
}

// NewScanner creates a new Scanner with the given bounds.
func NewScannerGV(width, height int, dest draw.Image, targ image.Rectangle,
	source *image.Uniform,
	offset image.Point) *ScannerGV {
	s := new(ScannerGV)
	s.SetBounds(width, height)
	s.Dest = dest
	s.Source = source
	s.Targ = targ
	s.Offset = offset
	return s
}
