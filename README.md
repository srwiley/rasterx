# rasterx

Rasterx is a golang package derived from the raster package of the [golang translation of freetype](https://github.com/golang/freetype). It enhances the orginal by additional features including a path stroking function capable of SVG 2.0 compliant arc joins and explicit loop closing.

* Paths can be explicity closed loops or left open, resulting in a line join or end caps, respectively.
* Arc joins are supported, which causes the extending edge from a Bezier curve to follow the radius of curvature at the end point rather than a straight line miter, resulting in a more fliud looking join. 
* Not specified by SVG 2.0, but supported in rasterx, is the arc-clip join, which is the arc join analog of a miter-clip join, both of which invoke the gap fill function at a point in proportion to a distance from the join point and the line width.
* Several cap and gap functions in addition to those specified by SVG2.0 are implemented.

![rasterx example](/doc/TestShapes4.svg.png?raw=true "Rasterx Example")

The above image shows the effect of using different join modes for a stroked curving path. The top stroked path uses miter (green) or arc (red, yellow, orange) join functions with high miter limit. The middle and lower path shows the effect of using the miter-clip and arc-clip joins, repectively, with different miter-limit values. The black chevrons at the top show different cap and gap functions.

Rasterx has refactored the orignal raster package in addition to providing a new stroker. This is to try and reduce code redundancy and divide functionality. Rasterizer is no longer a struct, but is instead an interface. The antialiasing and bezier curve decomposition functions in the orginal raster package were refactored into two stucts, Scanner and Filler, which logically seperate those functions. The Scanner implements the grainless algorithm for line rasterizing. The Filler has an embeded Scanner, but adds Bezier curve handling and flattening. The Stroker embeds a Filler but adds path stroking, and the Dasher adds the ability to create dashed stroked curve and embedds a Stroker. 

Also the Path data format is changed from the orignal raster package to 1) eliminate the redundant token on both sides of a path segment, and 2) add a close command. The close command is added to the Adder interface, and every Rasterizer also implements Adder.

![rasterx Scheme](/doc/schematic.png?raw=true "Rasterx Scheme")

Any instance of the Filler, Dasher, and Stroker can function on their own and each implements the Rasterizer interface. So if you need just the curve filling but no stroking capability, you only need to instantiate a Filler. On the other hand if you have created a Dasher and want to use it to Fill, you can just do this:

```golang
filler := &dasher.Filler
```
Now filler is a Filler rasterizer. See ![rasterx_test.go](/rasterx_test.go) for examples.
