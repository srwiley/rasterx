// Copyright 2010 The Freetype-Go Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2 (or
// any later version), both of which can be found in the LICENSE file.
// _
// 2017: Modification have been made by the rasterx to eliminate redundancy
// in the path by placing the command type id number only at the beginning
// of the command and not at both ends. It also adds the Close command
// to the path commands and the adder interface so that appropriate action
// can be taken to close the path depending on the path rendering operation.
// Some unused vector functions have been removed.
// All modifications are subject to the same licenses as the original rights
// declaration.

package rasterx

import (
	"fmt"
	"math"

	"golang.org/x/image/math/fixed"
)

// maxAbs returns the maximum of abs(a) and abs(b).
func maxAbs(a, b fixed.Int26_6) fixed.Int26_6 {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	if a < b {
		return b
	}
	return a
}

// pNeg returns the vector -p, or equivalently p rotated by 180 degrees.
func pNeg(p fixed.Point26_6) fixed.Point26_6 {
	return fixed.Point26_6{-p.X, -p.Y}
}

// pDot returns the dot product p·q.
func pDot(p fixed.Point26_6, q fixed.Point26_6) fixed.Int52_12 {
	px, py := int64(p.X), int64(p.Y)
	qx, qy := int64(q.X), int64(q.Y)
	return fixed.Int52_12(px*qx + py*qy)
}

// pLen returns the length of the vector p.
func pLen(p fixed.Point26_6) fixed.Int26_6 {
	// TODO(nigeltao): use fixed point math.
	x := float64(p.X)
	y := float64(p.Y)
	return fixed.Int26_6(math.Sqrt(x*x + y*y))
}

// pNorm returns the vector p normalized to the given length, or zero if p is
// degenerate.
func pNorm(p fixed.Point26_6, length fixed.Int26_6) fixed.Point26_6 {
	d := pLen(p)
	if d == 0 {
		return fixed.Point26_6{}
	}
	s, t := int64(length), int64(d)
	x := int64(p.X) * s / t
	y := int64(p.Y) * s / t
	return fixed.Point26_6{fixed.Int26_6(x), fixed.Int26_6(y)}
}

// pRot90CW returns the vector p rotated clockwise by 90 degrees.
// Note that the Y-axis grows downwards, so {1, 0}.Rot90CW is {0, 1}.
func pRot90CW(p fixed.Point26_6) fixed.Point26_6 {
	return fixed.Point26_6{-p.Y, p.X}
}

// pRot90CCW returns the vector p rotated counter-clockwise by 90 degrees.
// Note that the Y-axis grows downwards, so {1, 0}.Rot90CCW is {0, -1}.
func pRot90CCW(p fixed.Point26_6) fixed.Point26_6 {
	return fixed.Point26_6{p.Y, -p.X}
}

// Human readable path constants
const (
	PathMoveTo fixed.Int26_6 = iota
	PathLineTo
	PathQuadTo
	PathCubicTo
	PathClose
)

// An Adder accumulates points on a path.
type Adder interface {
	// Start starts a new curve at the given point.
	Start(a fixed.Point26_6)
	// Add1 adds a linear segment to the current curve.
	Line(b fixed.Point26_6)
	// Add2 adds a quadratic segment to the current curve.
	QuadBezier(b, c fixed.Point26_6)
	// Add3 adds a cubic segment to the current curve.
	CubeBezier(b, c, d fixed.Point26_6)
	// Closes the path to the start
	Stop(closeLoop bool)
	// Wipes out the path
	Clear()
}

// A Path is a sequence of curves, and a curve is a start point followed by a
// sequence of linear, quadratic or cubic segments.
type Path []fixed.Int26_6

func (p Path) ToSVGPath() string {
	s := ""
	for i := 0; i < len(p); {
		if i != 0 {
			s += " "
		}
		switch p[i] {
		case PathMoveTo:
			s += fmt.Sprintf("M%4.3f,%4.3f", float32(p[i+1])/64, float32(p[i+2])/64)
			i += 3
		case PathLineTo:
			s += fmt.Sprintf("L%4.3f,%4.3f", float32(p[i+1])/64, float32(p[i+2])/64)
			i += 3
		case PathQuadTo:
			s += fmt.Sprintf("Q%4.3f,%4.3f,%4.3f,%4.3f", float32(p[i+1])/64, float32(p[i+2])/64,
				float32(p[i+3])/64, float32(p[i+4])/64)
			i += 5
		case PathCubicTo:
			s += "C" + fmt.Sprintf("C%4.3f,%4.3f,%4.3f,%4.3f,%4.3f,%4.3f", float32(p[i+1])/64, float32(p[i+2])/64,
				float32(p[i+3])/64, float32(p[i+4])/64, float32(p[i+5])/64, float32(p[i+6])/64)
			i += 7
		case PathClose:
			s += "Z"
			i += 1
		default:
			panic("freetype/rasterx: bad pather")
		}
	}
	return s
}

// String returns a human-readable representation of a Path.
func (p Path) String() string {
	return p.ToSVGPath()
}

// Clears zeros the path slice
func (p *Path) Clear() {
	*p = (*p)[:0]
}

// Start starts a new curve at the given point.
func (p *Path) Start(a fixed.Point26_6) {
	*p = append(*p, PathMoveTo, a.X, a.Y)
}

// Add1 adds a linear segment to the current curve.
func (p *Path) Line(b fixed.Point26_6) {
	*p = append(*p, PathLineTo, b.X, b.Y)
}

// Add2 adds a quadratic segment to the current curve.
func (p *Path) QuadBezier(b, c fixed.Point26_6) {
	*p = append(*p, PathQuadTo, b.X, b.Y, c.X, c.Y)
}

// Add3 adds a cubic segment to the current curve.
func (p *Path) CubeBezier(b, c, d fixed.Point26_6) {
	*p = append(*p, PathCubicTo, b.X, b.Y, c.X, c.Y, d.X, d.Y)
}

// Close joins the ends of the path
func (p *Path) Stop(closeLoop bool) {
	if closeLoop {
		*p = append(*p, PathClose)
	}
}

// AddPath adds the Path p to q. This bridges the path and adder interface.
func (p Path) AddTo(q Adder) {
	for i := 0; i < len(p); {
		switch p[i] {
		case PathMoveTo:
			q.Start(fixed.Point26_6{p[i+1], p[i+2]})
			i += 3
		case PathLineTo:
			q.Line(fixed.Point26_6{p[i+1], p[i+2]})
			i += 3
		case PathQuadTo:
			q.QuadBezier(fixed.Point26_6{p[i+1], p[i+2]}, fixed.Point26_6{p[i+3], p[i+4]})
			i += 5
		case PathCubicTo:
			q.CubeBezier(fixed.Point26_6{p[i+1], p[i+2]},
				fixed.Point26_6{p[i+3], p[i+4]}, fixed.Point26_6{p[i+5], p[i+6]})
			i += 7
		case PathClose:
			q.Stop(true)
			i += 1
		default:
			panic("adder geom: bad path")
		}
	}
	q.Stop(false)
}

// AddPath adds the Path q to p.
func (p *Path) AddPath(q Path) {
	*p = append(*p, q...)
}
