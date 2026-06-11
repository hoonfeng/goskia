package skia

// #include "goskia.h"
import "C"

import "runtime"

// Path is a compound geometric path of lines, quadratics, conics, and cubics.
//
// The geometry-building methods return the receiver so calls can be chained:
//
//	p := skia.NewPath().MoveTo(0, 0).LineTo(10, 0).LineTo(10, 10).Close()
type Path struct{ ptr *C.sk_path_t }

// NewPath returns a new, empty path.
func NewPath() *Path {
	p := &Path{ptr: C.sk_path_new()}
	runtime.SetFinalizer(p, (*Path).release)
	return p
}

// MoveTo begins a new contour at (x, y).
func (p *Path) MoveTo(x, y float32) *Path {
	C.sk_path_move_to(p.ptr, C.float(x), C.float(y))
	return p
}

// LineTo adds a line from the last point to (x, y).
func (p *Path) LineTo(x, y float32) *Path {
	C.sk_path_line_to(p.ptr, C.float(x), C.float(y))
	return p
}

// QuadTo adds a quadratic from the last point, with control (x0,y0) to (x1,y1).
func (p *Path) QuadTo(x0, y0, x1, y1 float32) *Path {
	C.sk_path_quad_to(p.ptr, C.float(x0), C.float(y0), C.float(x1), C.float(y1))
	return p
}

// ConicTo adds a conic from the last point with the given control and weight.
func (p *Path) ConicTo(x0, y0, x1, y1, w float32) *Path {
	C.sk_path_conic_to(p.ptr, C.float(x0), C.float(y0), C.float(x1), C.float(y1), C.float(w))
	return p
}

// CubicTo adds a cubic from the last point with two control points to (x2,y2).
func (p *Path) CubicTo(x0, y0, x1, y1, x2, y2 float32) *Path {
	C.sk_path_cubic_to(p.ptr, C.float(x0), C.float(y0), C.float(x1), C.float(y1), C.float(x2), C.float(y2))
	return p
}

// Close closes the current contour, connecting back to its start point.
func (p *Path) Close() *Path {
	C.sk_path_close(p.ptr)
	return p
}

// AddRect adds a closed rectangle contour.
func (p *Path) AddRect(r Rect, dir PathDirection) *Path {
	cr := r.c()
	C.sk_path_add_rect(p.ptr, &cr, C.sk_path_direction_t(dir))
	return p
}

// AddOval adds a closed oval contour bounded by r.
func (p *Path) AddOval(r Rect, dir PathDirection) *Path {
	cr := r.c()
	C.sk_path_add_oval(p.ptr, &cr, C.sk_path_direction_t(dir))
	return p
}

// AddCircle adds a closed circle contour centered at (x, y).
func (p *Path) AddCircle(x, y, radius float32, dir PathDirection) *Path {
	C.sk_path_add_circle(p.ptr, C.float(x), C.float(y), C.float(radius), C.sk_path_direction_t(dir))
	return p
}

// AddRoundRect adds a closed rounded-rectangle contour.
func (p *Path) AddRoundRect(r Rect, rx, ry float32, dir PathDirection) *Path {
	cr := r.c()
	C.sk_path_add_rounded_rect(p.ptr, &cr, C.float(rx), C.float(ry), C.sk_path_direction_t(dir))
	return p
}

// AddPoly adds a polyline through the given points, optionally closing it.
func (p *Path) AddPoly(pts []Point, closed bool) *Path {
	if len(pts) == 0 {
		return p
	}
	cpts := make([]C.sk_point_t, len(pts))
	for i, pt := range pts {
		cpts[i] = pt.c()
	}
	C.sk_path_add_poly(p.ptr, &cpts[0], C.int(len(pts)), C.bool(closed))
	return p
}

// Reset clears the path, releasing any allocated memory.
func (p *Path) Reset() *Path {
	C.sk_path_reset(p.ptr)
	return p
}

// Bounds returns the bounds of the path's control points.
func (p *Path) Bounds() Rect {
	var r C.sk_rect_t
	C.sk_path_get_bounds(p.ptr, &r)
	return rectFromC(r)
}

// TightBounds returns the tight bounds of the path's geometry.
func (p *Path) TightBounds() Rect {
	var r C.sk_rect_t
	C.sk_path_compute_tight_bounds(p.ptr, &r)
	return rectFromC(r)
}

// SetFillType sets how the path's interior is determined.
func (p *Path) SetFillType(ft FillType) { C.sk_path_set_filltype(p.ptr, C.sk_path_filltype_t(ft)) }

// FillType returns the path's fill type.
func (p *Path) FillType() FillType { return FillType(C.sk_path_get_filltype(p.ptr)) }

// Transform applies m to all points in the path, in place.
func (p *Path) Transform(m Matrix) *Path {
	cm := m.c()
	C.sk_path_transform(p.ptr, &cm)
	return p
}

// Contains reports whether (x, y) is inside the (filled) path.
func (p *Path) Contains(x, y float32) bool {
	return bool(C.sk_path_contains(p.ptr, C.float(x), C.float(y)))
}

// IsConvex reports whether the path is convex.
func (p *Path) IsConvex() bool { return bool(C.sk_path_is_convex(p.ptr)) }

// CountPoints returns the number of points in the path.
func (p *Path) CountPoints() int { return int(C.sk_path_count_points(p.ptr)) }

func (p *Path) release() {
	if p != nil && p.ptr != nil {
		C.sk_path_delete(p.ptr)
		p.ptr = nil
		runtime.SetFinalizer(p, nil)
	}
}

// Release frees the path. It is safe to call multiple times.
func (p *Path) Release() { p.release() }
