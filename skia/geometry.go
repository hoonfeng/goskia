package skia

// #include "goskia.h"
import "C"

// Point is a 2D coordinate.
type Point struct{ X, Y float32 }

// Vector is a 2D displacement. It shares Point's representation.
type Vector = Point

func (p Point) c() C.sk_point_t { return C.sk_point_t{x: C.float(p.X), y: C.float(p.Y)} }

func pointFromC(c C.sk_point_t) Point { return Point{X: float32(c.x), Y: float32(c.y)} }

// IPoint is a 2D integer coordinate.
type IPoint struct{ X, Y int32 }

func (p IPoint) c() C.sk_ipoint_t { return C.sk_ipoint_t{x: C.int32_t(p.X), y: C.int32_t(p.Y)} }

// Point3 is a 3D coordinate.
type Point3 struct{ X, Y, Z float32 }

func (p Point3) c() C.sk_point3_t {
	return C.sk_point3_t{x: C.float(p.X), y: C.float(p.Y), z: C.float(p.Z)}
}

// Size is a 2D floating-point size.
type Size struct{ Width, Height float32 }

// ISize is a 2D integer size.
type ISize struct{ Width, Height int32 }

func (s ISize) c() C.sk_isize_t { return C.sk_isize_t{w: C.int32_t(s.Width), h: C.int32_t(s.Height)} }

// Rect is an axis-aligned rectangle with float32 edges.
type Rect struct{ Left, Top, Right, Bottom float32 }

// RectWH returns a rectangle at the origin with the given width and height.
func RectWH(w, h float32) Rect { return Rect{0, 0, w, h} }

// RectXYWH returns a rectangle at (x,y) with the given width and height.
func RectXYWH(x, y, w, h float32) Rect { return Rect{x, y, x + w, y + h} }

// Width returns the rectangle's width.
func (r Rect) Width() float32 { return r.Right - r.Left }

// Height returns the rectangle's height.
func (r Rect) Height() float32 { return r.Bottom - r.Top }

// IsEmpty reports whether the rectangle has no area.
func (r Rect) IsEmpty() bool { return r.Right <= r.Left || r.Bottom <= r.Top }

// Inset returns the rectangle shrunk by dx on the left/right and dy on top/bottom.
func (r Rect) Inset(dx, dy float32) Rect {
	return Rect{r.Left + dx, r.Top + dy, r.Right - dx, r.Bottom - dy}
}

// Offset returns the rectangle translated by (dx, dy).
func (r Rect) Offset(dx, dy float32) Rect {
	return Rect{r.Left + dx, r.Top + dy, r.Right + dx, r.Bottom + dy}
}

func (r Rect) c() C.sk_rect_t {
	return C.sk_rect_t{
		left:   C.float(r.Left),
		top:    C.float(r.Top),
		right:  C.float(r.Right),
		bottom: C.float(r.Bottom),
	}
}

func rectFromC(c C.sk_rect_t) Rect {
	return Rect{float32(c.left), float32(c.top), float32(c.right), float32(c.bottom)}
}

// IRect is an axis-aligned rectangle with int32 edges.
type IRect struct{ Left, Top, Right, Bottom int32 }

// IRectWH returns an integer rectangle at the origin with the given size.
func IRectWH(w, h int32) IRect { return IRect{0, 0, w, h} }

// Width returns the rectangle's width.
func (r IRect) Width() int32 { return r.Right - r.Left }

// Height returns the rectangle's height.
func (r IRect) Height() int32 { return r.Bottom - r.Top }

func (r IRect) c() C.sk_irect_t {
	return C.sk_irect_t{
		left:   C.int32_t(r.Left),
		top:    C.int32_t(r.Top),
		right:  C.int32_t(r.Right),
		bottom: C.int32_t(r.Bottom),
	}
}

func irectFromC(c C.sk_irect_t) IRect {
	return IRect{int32(c.left), int32(c.top), int32(c.right), int32(c.bottom)}
}

// RSXform is a compressed rotation+scale+translate transform used by DrawAtlas.
type RSXform struct{ SCos, SSin, TX, TY float32 }
