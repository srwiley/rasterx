// Copyright 2018 by the rasterx Authors. All rights reserved.
// Created 2018 by S.R.Wiley
package rasterx_test

import (
	"bufio"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"testing"

	. "github.com/srwiley/rasterx"
	"golang.org/x/image/colornames"
)

func getPartPath() (testPath Path) {
	//M210.08,222.97
	testPath.Start(ToFixedP(210.08, 222.97))
	//L192.55,244.95
	testPath.Line(ToFixedP(192.55, 244.95))
	//Q146.53,229.95,115.55,209.55
	testPath.QuadBezier(ToFixedP(146.53, 229.95), ToFixedP(115.55, 209.55))
	//Q102.50,211.00,95.38,211.00
	testPath.QuadBezier(ToFixedP(102.50, 211.00), ToFixedP(95.38, 211.00))
	//Q56.09,211.00,31.17,182.33
	testPath.QuadBezier(ToFixedP(56.09, 211.00), ToFixedP(31.17, 182.33))
	//Q6.27,153.66,6.27,108.44
	testPath.QuadBezier(ToFixedP(6.27, 153.66), ToFixedP(6.27, 108.44))
	//Q6.27,61.89,31.44,33.94
	testPath.QuadBezier(ToFixedP(6.27, 61.89), ToFixedP(31.44, 33.94))
	//Q56.62,6.00,98.55,6.00
	testPath.QuadBezier(ToFixedP(56.62, 6.00), ToFixedP(98.55, 6.00))
	//Q141.27,6.00,166.64,33.88
	testPath.QuadBezier(ToFixedP(141.27, 6.00), ToFixedP(166.64, 33.88))
	//Q192.02,61.77,192.02,108.70
	testPath.QuadBezier(ToFixedP(192.02, 61.77), ToFixedP(192.02, 108.70))
	//Q192.02,175.67,140.86,202.05
	testPath.QuadBezier(ToFixedP(192.02, 175.67), ToFixedP(140.86, 202.05))
	//Q173.42,216.66,210.08,222.97
	testPath.QuadBezier(ToFixedP(173.42, 216.66), ToFixedP(210.08, 222.97))
	//z
	testPath.Stop(true)
	return
}

func GetTestPath() (testPath Path) {
	//Path for Q
	testPath = getPartPath()

	//M162.22,109.69 M162.22,109.69
	testPath.Start(ToFixedP(162.22, 109.69))
	//Q162.22,70.11,145.61,48.55
	testPath.QuadBezier(ToFixedP(162.22, 70.11), ToFixedP(145.61, 48.55))
	//Q129.00,27.00,98.42,27.00
	testPath.QuadBezier(ToFixedP(129.00, 27.00), ToFixedP(98.42, 27.00))
	//Q69.14,27.00,52.53,48.62
	testPath.QuadBezier(ToFixedP(69.14, 27.00), ToFixedP(52.53, 48.62))
	//Q35.92,70.25,35.92,108.50
	testPath.QuadBezier(ToFixedP(35.92, 70.25), ToFixedP(35.92, 108.50))
	//Q35.92,146.75,52.53,168.38
	testPath.QuadBezier(ToFixedP(35.92, 146.75), ToFixedP(52.53, 168.38))
	//Q69.14,190.00,98.42,190.00
	testPath.QuadBezier(ToFixedP(69.14, 190.00), ToFixedP(98.42, 190.00))
	//Q128.34,190.00,145.28,168.70
	testPath.QuadBezier(ToFixedP(128.34, 190.00), ToFixedP(145.28, 168.70))
	//Q162.22,147.41,162.22,109.69
	testPath.QuadBezier(ToFixedP(162.22, 147.41), ToFixedP(162.22, 109.69))
	//z
	testPath.Stop(true)

	return
}

func BenchmarkScanGV(b *testing.B) {
	var (
		p         = GetTestPath()
		wx, wy    = 512, 512
		img       = image.NewRGBA(image.Rect(0, 0, wx, wy))
		scannerGV = NewScannerGV(wx, wy, img, img.Bounds())
	)
	f := NewFiller(wx, wy, scannerGV)
	p.AddTo(f)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Draw()
	}
}

func BenchmarkFillGV(b *testing.B) {
	var (
		p         = GetTestPath()
		wx, wy    = 512, 512
		img       = image.NewRGBA(image.Rect(0, 0, wx, wy))
		scannerGV = NewScannerGV(wx, wy, img, img.Bounds())
	)
	f := NewFiller(wx, wy, scannerGV)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.AddTo(f)
		f.Draw()
		f.Clear()
	}
}

func BenchmarkDashGV(b *testing.B) {
	var (
		p         = GetTestPath()
		wx, wy    = 512, 512
		img       = image.NewRGBA(image.Rect(0, 0, wx, wy))
		scannerGV = NewScannerGV(wx, wy, img, img.Bounds())
	)
	b.ResetTimer()
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

func TestRoundRect(t *testing.T) {
	var (
		wx, wy    = 512, 512
		img       = image.NewRGBA(image.Rect(0, 0, wx, wy))
		scannerGV = NewScannerGV(wx, wy, img, img.Bounds())
		f         = NewFiller(wx, wy, scannerGV)
	)

	scannerGV.SetColor(colornames.Cadetblue)
	AddRoundRect(30, 30, 130, 130, 40, 40, 0, RoundGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Burlywood)
	AddRoundRect(140, 30, 240, 130, 10, 40, 0, RoundGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Yellowgreen)
	AddRoundRect(250, 30, 350, 130, 40, 10, 0, RoundGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Lightgreen)
	AddRoundRect(370, 30, 470, 130, 20, 20, 45, RoundGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Cadetblue)
	AddRoundRect(30, 140, 130, 240, 40, 40, 0, QuadraticGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Burlywood)
	AddRoundRect(140, 140, 240, 240, 10, 40, 0, QuadraticGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Yellowgreen)
	AddRoundRect(250, 140, 350, 240, 40, 10, 0, QuadraticGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Blueviolet)
	AddRoundRect(370, 140, 470, 240, 20, 20, 45, QuadraticGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Cadetblue)
	AddRoundRect(30, 250, 130, 350, 40, 40, 0, CubicGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Burlywood)
	AddRoundRect(140, 250, 240, 350, 10, 40, 0, CubicGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Yellowgreen)
	AddRoundRect(250, 250, 350, 350, 40, 10, 0, CubicGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Lightgreen)
	AddRoundRect(370, 250, 470, 350, 20, 20, 45, CubicGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Cadetblue)
	AddRoundRect(30, 360, 130, 460, 40, 40, 0, FlatGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Burlywood)
	AddRoundRect(140, 360, 240, 460, 10, 40, 0, FlatGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Yellowgreen)
	AddRoundRect(250, 360, 350, 460, 40, 10, 0, FlatGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Blueviolet)
	AddRoundRect(370, 360, 470, 460, 20, 20, 45, FlatGap, f)
	f.Draw()
	f.Clear()

	err := SaveToPngFile("testdata/roundRectGV.png", img)
	if err != nil {
		t.Error(err)
	}

}

func TestShapes(t *testing.T) {
	var (
		wx, wy    = 512, 512
		img       = image.NewRGBA(image.Rect(0, 0, wx, wy))
		scannerGV = NewScannerGV(wx, wy, img, img.Bounds())
		f         = NewFiller(wx, wy, scannerGV)
	)

	scannerGV.SetColor(colornames.Blueviolet)
	AddEllipse(240, 200, 140, 180, 0, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Darkseagreen)
	AddEllipse(240, 200, 40, 180, 45, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Darkgoldenrod)
	AddCircle(300, 300, 80, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Forestgreen)
	AddRoundRect(30, 30, 130, 130, 10, 20, 45, RoundGap, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(ApplyOpacity(colornames.Lightgoldenrodyellow, 0.6))
	AddCircle(80, 80, 50, f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Firebrick)
	AddRect(370, 370, 400, 500, 15, f)
	f.Draw()
	f.Clear()

	err := SaveToPngFile("testdata/shapeGV.png", img)
	if err != nil {
		t.Error(err)
	}
}

// TestGradient tests a Dasher's ability to function
// as a filler, stroker, and dasher by invoking the corresponding anonymous structs
func TestGradient(t *testing.T) {
	var (
		wx, wy    = 512, 512
		img       = image.NewRGBA(image.Rect(0, 0, wx, wy))
		scannerGV = NewScannerGV(wx, wy, img, img.Bounds())
	)

	linearGradient := &Gradient{Points: [5]float64{0, 0, 1, 0, 0},
		IsRadial: false, Bounds: struct{ X, Y, W, H float64 }{
			X: 50, Y: 50, W: 100, H: 100}, Matrix: Identity}

	linearGradient.Stops = []GradStop{
		GradStop{StopColor: colornames.Aquamarine, Offset: 0.3, Opacity: 1.0},
		GradStop{StopColor: colornames.Skyblue, Offset: 0.6, Opacity: 1},
		GradStop{StopColor: colornames.Darksalmon, Offset: 1.0, Opacity: .75},
	}

	radialGradient := &Gradient{Points: [5]float64{0.5, 0.5, 0.5, 0.5, 0.5},
		IsRadial: true, Bounds: struct{ X, Y, W, H float64 }{
			X: 230, Y: 230, W: 100, H: 100},
		Matrix: Identity, Spread: ReflectSpread}

	radialGradient.Stops = []GradStop{
		GradStop{StopColor: colornames.Orchid, Offset: 0.3, Opacity: 1},
		GradStop{StopColor: colornames.Bisque, Offset: 0.6, Opacity: 1},
		GradStop{StopColor: colornames.Chartreuse, Offset: 1.0, Opacity: 0.4},
	}

	d := NewDasher(wx, wy, scannerGV)
	d.SetStroke(10*64, 4*64, RoundCap, nil, RoundGap, ArcClip, []float64{33, 12}, 0)
	// p is in the shape of a capital Q
	p := getPartPath()

	f := &d.Filler // This is the anon Filler in the Dasher. It also satisfies
	// the Rasterizer interface, and will only perform a fill on the path.

	scannerGV.SetColor(radialGradient.GetColorFunction(1))
	offsetPath := &MatrixAdder{Adder: f, M: Identity.Translate(180, 180)}

	p.AddTo(offsetPath)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(linearGradient.GetColorFunction(1.0))
	p.AddTo(f)
	f.Draw()
	f.Clear()

	// Let try a sinusoidal grid pattern.
	var colF ColorFunc = func(x, y int) color.Color {
		sinx, siny, sinxy := math.Sin(float64(x)*math.Pi/20), math.Sin(float64(y)*math.Pi/10),
			math.Sin(float64(y+x)*math.Pi/30)
		r := (1 + sinx) * 120
		g := (1 + siny) * 120
		b := (1 + sinxy) * 120
		return &color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	}

	scannerGV.SetColor(colF)
	AddRect(20, 300, 150, 450, 0, f)

	f.Draw()
	f.Clear()

	err := SaveToPngFile("testdata/gradGV.png", img)
	if err != nil {
		t.Error(err)
	}
}

// TestMultiFunction tests a Dasher's ability to function
// as a filler, stroker, and dasher by invoking the corresponding anonymous structs
func TestMultiFunctionGV(t *testing.T) {

	var (
		wx, wy    = 512, 512
		img       = image.NewRGBA(image.Rect(0, 0, wx, wy))
		scannerGV = NewScannerGV(wx, wy, img, img.Bounds())
	)

	scannerGV.SetColor(colornames.Cornflowerblue)
	d := NewDasher(wx, wy, scannerGV)
	d.SetStroke(10*64, 4*64, RoundCap, nil, RoundGap, ArcClip, []float64{33, 12}, 0)
	// p is in the shape of a capital Q
	p := GetTestPath()

	f := &d.Filler // This is the anon Filler in the Dasher. It also satisfies
	// the Rasterizer interface, and will only perform a fill on the path.

	p.AddTo(f)
	f.Draw()
	f.Clear()

	scannerGV.SetColor(colornames.Cornsilk)

	s := &d.Stroker // This is the anon Stroke in the Dasher. It also satisfies
	// the Rasterizer interface, but will perform a fill on the path.
	p.AddTo(s)
	s.Draw()
	s.Clear()

	scannerGV.SetColor(colornames.Darkolivegreen)

	// Now lets use the Dasher itself; it will perform a dashed stroke if dashes are set
	// in the SetStroke method.
	p.AddTo(d)
	d.Draw()
	d.Clear()

	err := SaveToPngFile("testdata/tmfGV.png", img)
	if err != nil {
		t.Error(err)
	}
}
