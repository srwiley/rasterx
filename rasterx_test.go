// Copyright 2018 by the rasterx Authors. All rights reserved.
// Created 2018 by S.R.Wiley
package rasterx_test

import (
	"bufio"
	"image"
	"image/color"
	"image/png"
	"os"

	"testing"

	. "github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

func toFixedP(x, y float64) (p fixed.Point26_6) {
	p.X = fixed.Int26_6(x * 64)
	p.Y = fixed.Int26_6(y * 64)
	return
}

func GetTestPath() (testPath Path) {
	//Path for Q
	//M210.08,222.97
	testPath.Start(toFixedP(210.08, 222.97))
	//L192.55,244.95
	testPath.Line(toFixedP(192.55, 244.95))
	//Q146.53,229.95,115.55,209.55
	testPath.QuadBezier(toFixedP(146.53, 229.95), toFixedP(115.55, 209.55))
	//Q102.50,211.00,95.38,211.00
	testPath.QuadBezier(toFixedP(102.50, 211.00), toFixedP(95.38, 211.00))
	//Q56.09,211.00,31.17,182.33
	testPath.QuadBezier(toFixedP(56.09, 211.00), toFixedP(31.17, 182.33))
	//Q6.27,153.66,6.27,108.44
	testPath.QuadBezier(toFixedP(6.27, 153.66), toFixedP(6.27, 108.44))
	//Q6.27,61.89,31.44,33.94
	testPath.QuadBezier(toFixedP(6.27, 61.89), toFixedP(31.44, 33.94))
	//Q56.62,6.00,98.55,6.00
	testPath.QuadBezier(toFixedP(56.62, 6.00), toFixedP(98.55, 6.00))
	//Q141.27,6.00,166.64,33.88
	testPath.QuadBezier(toFixedP(141.27, 6.00), toFixedP(166.64, 33.88))
	//Q192.02,61.77,192.02,108.70
	testPath.QuadBezier(toFixedP(192.02, 61.77), toFixedP(192.02, 108.70))
	//Q192.02,175.67,140.86,202.05
	testPath.QuadBezier(toFixedP(192.02, 175.67), toFixedP(140.86, 202.05))
	//Q173.42,216.66,210.08,222.97
	testPath.QuadBezier(toFixedP(173.42, 216.66), toFixedP(210.08, 222.97))
	//z
	testPath.Stop(false)
	//M162.22,109.69 M162.22,109.69
	testPath.Start(toFixedP(162.22, 109.69))
	//Q162.22,70.11,145.61,48.55
	testPath.QuadBezier(toFixedP(162.22, 70.11), toFixedP(145.61, 48.55))
	//Q129.00,27.00,98.42,27.00
	testPath.QuadBezier(toFixedP(129.00, 27.00), toFixedP(98.42, 27.00))
	//Q69.14,27.00,52.53,48.62
	testPath.QuadBezier(toFixedP(69.14, 27.00), toFixedP(52.53, 48.62))
	//Q35.92,70.25,35.92,108.50
	testPath.QuadBezier(toFixedP(35.92, 70.25), toFixedP(35.92, 108.50))
	//Q35.92,146.75,52.53,168.38
	testPath.QuadBezier(toFixedP(35.92, 146.75), toFixedP(52.53, 168.38))
	//Q69.14,190.00,98.42,190.00
	testPath.QuadBezier(toFixedP(69.14, 190.00), toFixedP(98.42, 190.00))
	//Q128.34,190.00,145.28,168.70
	testPath.QuadBezier(toFixedP(128.34, 190.00), toFixedP(145.28, 168.70))
	//Q162.22,147.41,162.22,109.69
	testPath.QuadBezier(toFixedP(162.22, 147.41), toFixedP(162.22, 109.69))
	//z
	testPath.Stop(false)

	return
}

var (
	p         = GetTestPath()
	wx, wy    = 512, 512
	img       = image.NewRGBA(image.Rect(0, 0, wx, wy))
	src       = image.NewUniform(color.NRGBA{0, 0, 255, 255})
	scannerGV = NewScannerGV(wx, wy, img, img.Bounds())
)

func BenchmarkScanGV(b *testing.B) {
	src.C = color.NRGBA{0, 0, 255, 255}
	f := NewFiller(wx, wy, scannerGV)
	p.AddTo(f)
	for i := 0; i < b.N; i++ {
		f.Draw()
	}
}

func BenchmarkFillGV(b *testing.B) {
	src.C = color.NRGBA{0, 0, 255, 255}
	f := NewFiller(wx, wy, scannerGV)
	for i := 0; i < b.N; i++ {
		p.AddTo(f)
		f.Draw()
		f.Clear()
	}
}

func BenchmarkDashGV(b *testing.B) {
	src.C = color.NRGBA{0, 0, 255, 255}
	d := NewDasher(wx, wy, scannerGV)
	d.SetStroke(10*64, 4*64, RoundCap, nil, RoundGap, ArcClip, []float64{33, 12}, 0)
	for i := 0; i < b.N; i++ {
		p.AddTo(d)
		d.Draw()
		d.Clear()
	}
}

func SaveToPngFile(filePath string, m image.Image) error {
	// Create the file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	// Create Writer from file
	b := bufio.NewWriter(f)
	// Write the image into the buffer
	err = png.Encode(b, m)
	if err != nil {
		return err
	}
	err = b.Flush()
	if err != nil {
		return err
	}
	return nil
}

// TestMultiFunction tests a Dasher's ability to function
// as a filler, stroker, and dasher by invoking the corresponding anonymous structs
func TestMultiFunctionGV(t *testing.T) {

	src.C = color.NRGBA{0, 0, 255, 255}
	d := NewDasher(wx, wy, scannerGV)
	d.SetStroke(10*64, 4*64, RoundCap, nil, RoundGap, ArcClip, []float64{33, 12}, 0)
	// p is in the shape of a capital Q
	p := GetTestPath()

	f := &d.Filler // This is the anon Filler in the Dasher. It also satisfies
	// the Rasterizer interface, and will only perform a fill on the path.

	p.AddTo(f)
	f.Draw()
	f.Clear()

	src.C = color.NRGBA{240, 124, 0, 255}
	s := &d.Stroker // This is the anon Stroke in the Dasher. It also satisfies
	// the Rasterizer interface, but will perform a fill on the path.
	p.AddTo(s)
	s.Draw()
	s.Clear()

	// Now lets use the Dasher itself; it will perform a dashed stroke if dashes are set
	// in the SetStroke method.
	src.C = color.NRGBA{255, 0, 0, 255}
	p.AddTo(d)
	d.Draw()
	d.Clear()

	err := SaveToPngFile("testdata/tmfGV.png", img)
	if err != nil {
		t.Error(err)
	}
}

// TestStandAloneFunction tests a filler, stroker and dasher, each
// initialized seperately.
//func TestStandAloneFunction(t *testing.T) {
//	wx, wy := 512, 512

//	// Create an RGBA image and a painter for the image.
//	img := image.NewRGBA(image.Rect(0, 0, wx, wy))
//	painter := NewRGBAPainter(img)

//	p := GetTestPath()

//	f := NewFiller(wx, wy) // A filler can only fill a path
//	p.AddTo(f)
//	painter.SetColor(color.NRGBA{0, 0, 255, 255})
//	f.Rasterize(painter)

//	s := NewStroker(wx, wy) // A stroker can fill or stroke a path
//	s.SetStroke(10*64, 4*64, RoundCap, nil, RoundGap, ArcClip)
//	p.AddTo(s)
//	painter.SetColor(color.NRGBA{255, 255, 0, 255})
//	s.Rasterize(painter)

//	// Now lets use the Dasher
//	d := NewDasher(wx, wy) // A dasher can fill or stroke a path or stroke a dashed path
//	// Set the values for the stroking and dashing.
//	d.SetStroke(10*64, 4*64, RoundCap, nil, RoundGap, ArcClip, []float64{33, 12}, 0)
//	p.AddTo(d)
//	painter.SetColor(color.NRGBA{0, 255, 0, 255})
//	d.Rasterize(painter)

//	err := SaveToPngFile("testdata/tsaf.png", img)
//	if err != nil {
//		t.Error(err)
//	}
//}

// TestShiftedMultiFunction tests a Dasher's ability to function
// as a filler, stroker, and dasher by invoking the corresponding anonymous structs
// and also test the TransAdder struct's ability to translate and scale the path.
//func TestShiftedMultiFunction(t *testing.T) {
//	wx, wy := 512, 512
//	img := image.NewRGBA(image.Rect(0, 0, wx, wy))
//	painter := NewRGBAPainter(img)

//	d := NewDasher(wx, wy)
//	p := GetTestPath()
//	p.AddTo(&d.Filler) // By passing in the filler rasterizer d will act as a filler
//	painter.SetColor(color.NRGBA{0, 0, 255, 255})
//	d.Rasterize(painter)
//	d.Clear()

//	// This will shift and scale the path points as they are passed to the Stroker
//	// It statisfies the Adder interface
//	sta := NewTransAdder(&d.Stroker)

//	painter.SetColor(color.NRGBA{255, 255, 0, 255})
//	d.SetStroke(12*64, 20*64, RoundCap, ButtCap, RoundGap, ArcClip, []float64{33, 12}, 0)
//	sta.TranslateTo(40, 40)
//	p.AddTo(sta)
//	d.Rasterize(painter)
//	d.Clear()

//	dta := NewTransAdder(d) // Now it will use the Dashes if any were set by the SetStroke method
//	dta.TranslateTo(150, 80)
//	dta.SetScale(.95, 1.3) // Squish it a little bit

//	painter.SetColor(color.NRGBA{0, 255, 0, 255})
//	p.AddTo(dta)
//	d.Rasterize(painter)

//	err := SaveToPngFile("testdata/tsmf.png", img)
//	if err != nil {
//		t.Error(err)
//	}

//}
