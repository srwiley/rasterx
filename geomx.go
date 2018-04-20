// Copyright 2017 by the rasterx Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2 (or
// any later version), both of which can be found in the LICENSE file.
//_
// created: 2/12/2017 by S.R.Wiley
// geomx adds some additional geometry functions needed by rasterx

package rasterx

import (
	"math"

	"golang.org/x/image/math/fixed"
)

// ClosestPortside returns the closest of p1 or p2 on the port side of the
// line from the bow to the stern. (port means left side of the direction you are heading)
// isIntersecting is just convienice to reduce code, and if false returns false, because p1 and p2 are not valid
func ClosestPortside(bow, stern, p1, p2 fixed.Point26_6, isIntersecting bool) (xt fixed.Point26_6, intersects bool) {
	if isIntersecting == false {
		return
	}
	dir := bow.Sub(stern)
	dp1 := p1.Sub(stern)
	dp2 := p2.Sub(stern)
	cp1 := dir.X*dp1.Y - dp1.X*dir.Y
	cp2 := dir.X*dp2.Y - dp2.X*dir.Y
	switch {
	case cp1 < 0 && cp2 < 0:
		return
	case cp1 < 0 && cp2 >= 0:
		return p2, true
	case cp1 >= 0 && cp2 < 0:
		return p1, true
	default: // both points on port side
		dirdot := pDot(dir, dir)
		// calculate vector rejections of dp1 and dp2 onto dir
		h1 := dp1.Sub(dir.Mul(fixed.Int26_6((pDot(dp1, dir) << 6) / dirdot)))
		h2 := dp2.Sub(dir.Mul(fixed.Int26_6((pDot(dp2, dir) << 6) / dirdot)))
		// return point with smallest vector rejection; i.e. closest to dir line
		if (h1.X*h1.X + h1.Y*h1.Y) > (h2.X*h2.X + h2.Y*h2.Y) {
			return p2, true
		}
		return p1, true
	}
}

// RadCurvature returns the curvature of a Bezier curve end point,
// given an end point, the two adjacent control points and the degree.
// The sign of the value indicates if the center of the osculating circle
// is left or right (port or starboard) of the curve in the forward direction.
func RadCurvature(p0, p1, p2 fixed.Point26_6, dm fixed.Int52_12) fixed.Int26_6 {
	a, b := p2.Sub(p1), p1.Sub(p0)
	abdot, bbdot := pDot(a, b), pDot(b, b)
	h := a.Sub(b.Mul(fixed.Int26_6((abdot << 6) / bbdot))) // h is the vector rejection of a onto b
	if h.X == 0 && h.Y == 0 {                              // points are co-linear
		return 0
	}
	radCurve := fixed.Int26_6((fixed.Int52_12(a.X*a.X+a.Y*a.Y) * dm / fixed.Int52_12(pLen(h)<<6)) >> 6)
	if a.X*b.Y > b.X*a.Y { // xprod sign
		return radCurve
	}
	return -radCurve
}

// CircleCircleIntersection calculates the points of intersection of
// two circles or returns with intersects == false if no such points exist.
func CircleCircleIntersection(ct, cl fixed.Point26_6, rt, rl fixed.Int26_6) (xt1, xt2 fixed.Point26_6, intersects bool) {
	dc := cl.Sub(ct)
	d := pLen(dc)

	// Check for solvability.
	if d > (rt + rl) {
		return // No solution. Circles do not intersect.
	}
	// check if  d < abs(rt-rl)
	if da := rt - rl; (da > 0 && d < da) || (da < 0 && d < -da) {
		return // No solution. One circle is contained by the other.
	}

	rlf, rtf, df := float64(rl), float64(rt), float64(d)
	af := (rtf*rtf - rlf*rlf + df*df) / df / 2.0
	hfd := math.Sqrt(rtf*rtf-af*af) / df
	afd := af / df

	rOffx, rOffy := float64(-dc.Y)*hfd, float64(dc.X)*hfd
	p2x := float64(ct.X) + float64(dc.X)*afd
	p2y := float64(ct.Y) + float64(dc.Y)*afd
	xt1x, xt1y := p2x+rOffx, p2y+rOffy
	xt2x, xt2y := p2x-rOffx, p2y-rOffy
	return fixed.Point26_6{fixed.Int26_6(xt1x), fixed.Int26_6(xt1y)},
		fixed.Point26_6{fixed.Int26_6(xt2x), fixed.Int26_6(xt2y)}, true
}

// CalcIntersect calculates the points of intersection of two fixed point lines
// and panics if the determinate is zero. You have been warned.
func CalcIntersect(a1, a2, b1, b2 fixed.Point26_6) (x fixed.Point26_6) {
	da, db, ds := a2.Sub(a1), b2.Sub(b1), a1.Sub(b1)
	det := float32(da.X*db.Y - db.X*da.Y) // Determinate
	t := float32(ds.Y*db.X-ds.X*db.Y) / det
	x = a1.Add(fixed.Point26_6{fixed.Int26_6(float32(da.X) * t), fixed.Int26_6(float32(da.Y) * t)})
	return
}

// RayCircleIntersection calculates the points of intersection of
// a ray starting at s2 passing through s1 and a circle in fixed point.
// Returns intersects == false if no solution is possible. If two
// solutions are possible, the point closest to s2 is returned
func RayCircleIntersection(s1, s2, c fixed.Point26_6, r fixed.Int26_6) (x fixed.Point26_6, intersects bool) {
	n := float64(s2.X - c.X) // Calculating using 64* rather than divide
	m := float64(s2.Y - c.Y)

	e := float64(s2.X - s1.X)
	d := float64(s2.Y - s1.Y)

	f := float64(r)
	// Quadratic normal form coefficients
	A, B, C := e*e+d*d, -2*(e*n+m*d), n*n+m*m-f*f

	D := B*B - 4*A*C

	if D <= 0 {
		return // No intersection or is tangent
	}

	D = math.Sqrt(D)
	t1, t2 := (-B+D)/(2*A), (-B-D)/(2*A)
	p1OnSide := t1 > 0
	p2OnSide := t2 > 0

	switch {
	case p1OnSide && p2OnSide:
		if t2 < t1 { // both on ray, use closest to s2
			t1 = t2
		}
	case p2OnSide: // Only p2 on ray
		t1 = t2
	case p1OnSide: // only p1 on ray
	default: // Neither solution is on the ray
		return
	}
	return fixed.Point26_6{fixed.Int26_6((n - e*t1)) + c.X,
		fixed.Int26_6((m - d*t1)) + c.Y}, true

}
