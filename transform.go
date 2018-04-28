// Copyright 2018 by the rasterx Authors. All rights reserved.
//
// created: 2018 by S.R.Wiley
package rasterx

import (
	"golang.org/x/image/math/fixed"
)

type (
	TransAdder struct {
		Adder
		X, Y, ScaleX, ScaleY float32
	}
)

func NewTransAdder(r Adder) *TransAdder {
	t := new(TransAdder)
	t.Reset()
	t.Adder = r
	return t
}

func (t *TransAdder) Reset() {
	t.ScaleX = 1.0
	t.ScaleY = 1.0
	t.X = 0.0
	t.Y = 0.0
}

func (t *TransAdder) Transform(p fixed.Point26_6) fixed.Point26_6 {
	p.X = fixed.Int26_6((t.X + float32(p.X)/64*t.ScaleX) * 64)
	p.Y = fixed.Int26_6((t.Y + float32(p.Y)/64*t.ScaleY) * 64)
	return p
}

func (t *TransAdder) Start(a fixed.Point26_6) {
	t.Adder.Start(t.Transform(a))
}

// Line adds a linear segment to the current curve.
func (t *TransAdder) Line(b fixed.Point26_6) {
	t.Adder.Line(t.Transform(b))
}

// QuadBezier adds a quadratic segment to the current curve.
func (t *TransAdder) QuadBezier(b, c fixed.Point26_6) {
	t.Adder.QuadBezier(t.Transform(b), t.Transform(c))
}

// CubeBezier adds a cubic segment to the current curve.
func (t *TransAdder) CubeBezier(b, c, d fixed.Point26_6) {
	t.Adder.CubeBezier(t.Transform(b), t.Transform(c), t.Transform(d))
}

// SetScale sets the absolute value of the x and y scales.
func (t *TransAdder) SetScale(sx, sy float32) {
	t.ScaleX = sx
	t.ScaleY = sy
}

// TranslateTo sets the absolute value of the x and y translations.
func (t *TransAdder) TranslateTo(x, y float32) {
	t.X = x
	t.Y = y
}
