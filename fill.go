// Copyright 2010 The Freetype-Go Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2 (or
// any later version), both of which can be found in the LICENSE file.
//_
// Package raster provides an anti-aliasing 2-D rasterizer.
// It is part of the larger Freetype suite of font-related packages, but the
// raster package is not specific to font rasterization, and can be used
// standalone without any other Freetype package.
// Rasterization is done by the same area/coverage accumulation algorithm as
// the Freetype "smooth" module, and the Anti-Grain Geometry library. A
// description of the area/coverage algorithm is at
// http://projects.tuxee.net/cl-vectors/section-the-cl-aa-algorithm
// _
// _
// 2017: Modifications and refactorizations have been made by the rasterx package
// in order to isolate isolates the scanner struct which adds lines
// to the rasterizer vs the filler which can add bezier curves to the scanner.
// All modifications are subject to the same licenses as the original rights
// declaration.
// The format of the Path data is changed and expaned to include close (see geom.go).
// All modifications are subject to the same licenses as the original rights
// declaration.

package rasterx

import (
	"strconv"

	"golang.org/x/image/math/fixed"
)

// dev is 32-bit, and nsplit++ every time we shift off 2 bits, so maxNsplit
// is 16.
const maxNsplit = 16

type (
	Rasterizer interface {
		Adder
		SetBounds(width, height int)
		Rasterize(p Painter)
		lineF(b fixed.Point26_6)
		joinF()
	}

	// Filler fills a path using the grainless algorithm.
	Filler struct {
		Scanner

		// splitScaleN is the scaling factor used to determine how many times
		// to decompose a quadratic or cubic segment into a linear approximation.
		splitScale2, splitScale3 int

		//Stacks used by add2 and add3 for bezier curve decomposition
		pStack []fixed.Point26_6
		sStack []int
	}
)

// QuadBezier adds a quadratic segment to the current curve.
func (r *Filler) QuadBezier(b, c fixed.Point26_6) {
	r.QuadBezierF(r, b, c)
}

// QuadBezierF adds a quadratic segment to the sgm Rasterizer.
func (r *Filler) QuadBezierF(sgm Rasterizer, b, c fixed.Point26_6) {
	// check for degenerate bezier
	if r.a == b || b == c {
		sgm.Line(c)
		return
	}

	sgm.joinF()

	// Calculate nSplit (the number of recursive decompositions) based on how
	// 'curvy' it is. Specifically, how much the middle point b deviates from
	// (a+c)/2.
	dev := maxAbs(r.a.X-2*b.X+c.X, r.a.Y-2*b.Y+c.Y) / fixed.Int26_6(r.splitScale2)
	nsplit := 0
	for dev > 0 {
		dev /= 4
		nsplit++
	}
	if nsplit > maxNsplit {
		panic("freetype/raster: Add2 nsplit too large: " + strconv.Itoa(nsplit))
	}
	// Recursively decompose the curve nSplit levels deep.
	var i, pPlace, sPlace = 0, len(r.pStack), len(r.sStack)

	r.ExpandStacks(pPlace+2*maxNsplit+3, sPlace+maxNsplit+1)

	r.sStack[sPlace] = nsplit
	r.pStack[pPlace] = c
	r.pStack[pPlace+1] = b
	r.pStack[pPlace+2] = r.a
	for i >= 0 {
		s := r.sStack[i+sPlace]
		p := r.pStack[2*i+pPlace:]
		if s > 0 {
			// Split the quadratic curve p[:3] into an equivalent set of two
			// shorter curves: p[:3] and p[2:5]. The new p[4] is the old p[2],
			// and p[0] is unchanged.
			mx := p[1].X
			p[4].X = p[2].X
			p[3].X = (p[4].X + mx) / 2
			p[1].X = (p[0].X + mx) / 2
			p[2].X = (p[1].X + p[3].X) / 2
			my := p[1].Y
			p[4].Y = p[2].Y
			p[3].Y = (p[4].Y + my) / 2
			p[1].Y = (p[0].Y + my) / 2
			p[2].Y = (p[1].Y + p[3].Y) / 2
			// The two shorter curves have one less split to do.
			r.sStack[i+sPlace] = s - 1
			r.sStack[i+1+sPlace] = s - 1
			i++
		} else {
			// Replace the level-0 quadratic with a two-linear-piece
			// approximation.
			midx := (p[0].X + 2*p[1].X + p[2].X) / 4
			midy := (p[0].Y + 2*p[1].Y + p[2].Y) / 4
			sgm.lineF(fixed.Point26_6{midx, midy})
			sgm.lineF(p[0])
			i--
		}
	}
	r.sStack = r.sStack[:sPlace]
	r.pStack = r.pStack[:pPlace]
}

// CubeBezier adds a cubic bezier to the curve
func (r *Filler) CubeBezier(b, c, d fixed.Point26_6) {
	r.CubeBezierF(r, b, c, d)
}

// joinF is a no-op for a filling rasterizer. This is used in stroking and dashed
// stroking
func (r *Filler) joinF() {

}

// lineF for a filling rasterizer is just the line call in scan
func (r *Filler) lineF(b fixed.Point26_6) {
	r.Line(b)
}

// CubeBezier adds a cubic bezier to the curve. sending the line calls the the
// sgm Rasterizer
func (r *Filler) CubeBezierF(sgm Rasterizer, b, c, d fixed.Point26_6) {
	if (r.a == b && c == d) || (r.a == b && b == c) || (c == b && d == c) {
		sgm.Line(d)
		return
	}
	sgm.joinF()
	// Calculate nSplit (the number of recursive decompositions) based on how
	// 'curvy' it is.
	dev2 := maxAbs(r.a.X-3*(b.X+c.X)+d.X, r.a.Y-3*(b.Y+c.Y)+d.Y) / fixed.Int26_6(r.splitScale2)
	dev3 := maxAbs(r.a.X-2*b.X+d.X, r.a.Y-2*b.Y+d.Y) / fixed.Int26_6(r.splitScale3)
	nsplit := 0
	for dev2 > 0 || dev3 > 0 {
		dev2 /= 8
		dev3 /= 4
		nsplit++
	}

	// devN is 32-bit, and nsplit++ every time we shift off 2 bits, so
	// maxNsplit is 16.
	//const maxNsplit = 16
	if nsplit > maxNsplit {
		panic("freetype/raster: Add3 nsplit too large: " + strconv.Itoa(nsplit))
	}
	// Recursively decompose the curve nSplit levels deep.
	var i, pPlace, sPlace = 0, len(r.pStack), len(r.sStack)
	r.ExpandStacks(pPlace+3*maxNsplit+4, sPlace+maxNsplit+1)
	r.sStack[sPlace] = nsplit
	r.pStack[pPlace] = d
	r.pStack[pPlace+1] = c
	r.pStack[pPlace+2] = b
	r.pStack[pPlace+3] = r.a
	for i >= 0 {
		s := r.sStack[i+sPlace]
		p := r.pStack[3*i+pPlace:]

		if s > 0 {
			// Split the cubic curve p[:4] into an equivalent set of two
			// shorter curves: p[:4] and p[3:7]. The new p[6] is the old p[3],
			// and p[0] is unchanged.
			m01x := (p[0].X + p[1].X) / 2
			m12x := (p[1].X + p[2].X) / 2
			m23x := (p[2].X + p[3].X) / 2
			p[6].X = p[3].X
			p[5].X = m23x
			p[1].X = m01x
			p[2].X = (m01x + m12x) / 2
			p[4].X = (m12x + m23x) / 2
			p[3].X = (p[2].X + p[4].X) / 2
			m01y := (p[0].Y + p[1].Y) / 2
			m12y := (p[1].Y + p[2].Y) / 2
			m23y := (p[2].Y + p[3].Y) / 2
			p[6].Y = p[3].Y
			p[5].Y = m23y
			p[1].Y = m01y
			p[2].Y = (m01y + m12y) / 2
			p[4].Y = (m12y + m23y) / 2
			p[3].Y = (p[2].Y + p[4].Y) / 2
			// The two shorter curves have one less split to do.
			r.sStack[i+sPlace] = s - 1
			r.sStack[i+1+sPlace] = s - 1
			i++
		} else {
			// Replace the level-0 cubic with a two-linear-piece approximation.
			midx := (p[0].X + 3*(p[1].X+p[2].X) + p[3].X) / 8
			midy := (p[0].Y + 3*(p[1].Y+p[2].Y) + p[3].Y) / 8
			sgm.lineF(fixed.Point26_6{midx, midy})
			sgm.lineF(p[0])
			i--
		}
	}
	r.sStack = r.sStack[:sPlace]
	r.pStack = r.pStack[:pPlace]
}

// SetBounds sets the maximum width and height of the rasterized image and
// calls Clear. The width and height are in pixels, not fixed.Int26_6 units.
func (r *Filler) SetBounds(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	// Use the same ssN heuristic as the C Freetype (version 2.4.0)
	// implementation.
	ss2, ss3 := 32, 16
	if width > 24 || height > 24 {
		ss2, ss3 = 2*ss2, 2*ss3
		if width > 120 || height > 120 {
			ss2, ss3 = 2*ss2, 2*ss3
		}
	}
	r.splitScale2 = ss2
	r.splitScale3 = ss3
	r.Scanner.SetBounds(width, height)
}

// NewFiller returns a Filler ptr with default values.
// A Filler in addition to rasterizing lines like a Scann,
// can also rasterize quadratic and cubic bezier curves.
func NewFiller(width, height int) *Filler {
	r := new(Filler)
	r.SetBounds(width, height)
	r.UseNonZeroWinding = true
	return r
}

// ExpandStacks expands the recursion stacks to respective sizes,
// and reallocates slice if nec. It is exposed so that users can pre-expand.
func (r *Filler) ExpandStacks(pSize, sSize int) {
	//Expand pStack if required
	if pSize > cap(r.pStack) {
		newSlice := make([]fixed.Point26_6, pSize, 2*pSize)
		copy(newSlice, r.pStack)
		r.pStack = newSlice
	} else {
		r.pStack = r.pStack[:pSize]
	}
	//Expand sStack if required
	if sSize > cap(r.sStack) {
		newSlice := make([]int, sSize, 2*sSize)
		copy(newSlice, r.sStack)
		r.sStack = newSlice
	} else {
		r.sStack = r.sStack[:sSize]
	}
}
