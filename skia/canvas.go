package skia

// #include "goskia.h"
import "C"

import "runtime"

// Canvas provides the drawing surface API: shapes, images, text, plus a
// matrix/clip stack. A Canvas obtained from a Surface is owned by that surface.
type Canvas struct {
	ptr   *C.sk_canvas_t
	owner any // keeps the owning Surface (or Bitmap) alive; nil if standalone
}

// Clear fills the entire canvas with a single color, ignoring the clip.
func (c *Canvas) Clear(color Color) {
	C.sk_canvas_clear(c.ptr, color.c())
	runtime.KeepAlive(c)
}

// DrawColor fills the clip with color using the given blend mode.
func (c *Canvas) DrawColor(color Color, mode BlendMode) {
	C.sk_canvas_draw_color(c.ptr, color.c(), C.sk_blendmode_t(mode))
	runtime.KeepAlive(c)
}

// DrawPaint fills the clip with the paint (e.g. a shader).
func (c *Canvas) DrawPaint(p *Paint) {
	C.sk_canvas_draw_paint(c.ptr, paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(p)
}

// DrawRect draws a rectangle.
func (c *Canvas) DrawRect(r Rect, p *Paint) {
	cr := r.c()
	C.sk_canvas_draw_rect(c.ptr, &cr, paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(p)
}

// DrawOval draws an oval bounded by r.
func (c *Canvas) DrawOval(r Rect, p *Paint) {
	cr := r.c()
	C.sk_canvas_draw_oval(c.ptr, &cr, paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(p)
}

// DrawCircle draws a circle centered at (cx, cy).
func (c *Canvas) DrawCircle(cx, cy, radius float32, p *Paint) {
	C.sk_canvas_draw_circle(c.ptr, C.float(cx), C.float(cy), C.float(radius), paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(p)
}

// DrawRoundRect draws a rectangle with rounded corners of radii (rx, ry).
func (c *Canvas) DrawRoundRect(r Rect, rx, ry float32, p *Paint) {
	cr := r.c()
	C.sk_canvas_draw_round_rect(c.ptr, &cr, C.float(rx), C.float(ry), paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(p)
}

// DrawPath draws a path.
func (c *Canvas) DrawPath(path *Path, p *Paint) {
	C.sk_canvas_draw_path(c.ptr, path.ptr, paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(path)
	runtime.KeepAlive(p)
}

// DrawLine draws a line segment from (x0, y0) to (x1, y1).
func (c *Canvas) DrawLine(x0, y0, x1, y1 float32, p *Paint) {
	C.sk_canvas_draw_line(c.ptr, C.float(x0), C.float(y0), C.float(x1), C.float(y1), paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(p)
}

// DrawArc draws an arc of the oval bounded by oval. If useCenter is true the
// arc is closed back to the center (a pie slice).
func (c *Canvas) DrawArc(oval Rect, startAngle, sweepAngle float32, useCenter bool, p *Paint) {
	co := oval.c()
	C.sk_canvas_draw_arc(c.ptr, &co, C.float(startAngle), C.float(sweepAngle), C.bool(useCenter), paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(p)
}

// DrawPoints draws the given points according to mode (points, lines, polygon).
func (c *Canvas) DrawPoints(mode PointMode, pts []Point, p *Paint) {
	if len(pts) == 0 {
		return
	}
	cpts := make([]C.sk_point_t, len(pts))
	for i, pt := range pts {
		cpts[i] = pt.c()
	}
	C.sk_canvas_draw_points(c.ptr, C.sk_point_mode_t(mode), C.size_t(len(pts)), &cpts[0], paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(p)
}

// DrawImage draws img with its top-left at (x, y).
func (c *Canvas) DrawImage(img *Image, x, y float32, sampling SamplingOptions, p *Paint) {
	cs := sampling.c()
	C.sk_canvas_draw_image(c.ptr, img.ptr, C.float(x), C.float(y), &cs, paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(img)
	runtime.KeepAlive(p)
}

// DrawImageRect draws the src rectangle of img into the dst rectangle.
func (c *Canvas) DrawImageRect(img *Image, src, dst Rect, sampling SamplingOptions, p *Paint) {
	csrc, cdst, cs := src.c(), dst.c(), sampling.c()
	C.sk_canvas_draw_image_rect(c.ptr, img.ptr, &csrc, &cdst, &cs, paintPtr(p))
	runtime.KeepAlive(c)
	runtime.KeepAlive(img)
	runtime.KeepAlive(p)
}

// Save pushes the current matrix and clip onto the save stack and returns the
// stack depth that RestoreToCount can target.
func (c *Canvas) Save() int {
	n := int(C.sk_canvas_save(c.ptr))
	runtime.KeepAlive(c)
	return n
}

// SaveLayer allocates an offscreen layer for subsequent drawing, composited
// with p when restored. bounds may be nil for an unbounded layer.
func (c *Canvas) SaveLayer(bounds *Rect, p *Paint) int {
	var cb *C.sk_rect_t
	if bounds != nil {
		b := bounds.c()
		cb = &b
	}
	n := int(C.sk_canvas_save_layer(c.ptr, cb, paintPtr(p)))
	runtime.KeepAlive(c)
	runtime.KeepAlive(p)
	return n
}

// Restore pops the most recent save.
func (c *Canvas) Restore() {
	C.sk_canvas_restore(c.ptr)
	runtime.KeepAlive(c)
}

// RestoreToCount restores the save stack to the given depth.
func (c *Canvas) RestoreToCount(count int) {
	C.sk_canvas_restore_to_count(c.ptr, C.int(count))
	runtime.KeepAlive(c)
}

// SaveCount returns the current save-stack depth.
func (c *Canvas) SaveCount() int {
	n := int(C.sk_canvas_get_save_count(c.ptr))
	runtime.KeepAlive(c)
	return n
}

// Translate post-translates the current matrix.
func (c *Canvas) Translate(dx, dy float32) {
	C.sk_canvas_translate(c.ptr, C.float(dx), C.float(dy))
	runtime.KeepAlive(c)
}

// Scale post-scales the current matrix.
func (c *Canvas) Scale(sx, sy float32) {
	C.sk_canvas_scale(c.ptr, C.float(sx), C.float(sy))
	runtime.KeepAlive(c)
}

// Rotate post-rotates the current matrix by degrees (clockwise).
func (c *Canvas) Rotate(degrees float32) {
	C.sk_canvas_rotate_degrees(c.ptr, C.float(degrees))
	runtime.KeepAlive(c)
}

// RotateRadians post-rotates the current matrix by radians (clockwise).
func (c *Canvas) RotateRadians(radians float32) {
	C.sk_canvas_rotate_radians(c.ptr, C.float(radians))
	runtime.KeepAlive(c)
}

// Skew post-skews the current matrix.
func (c *Canvas) Skew(sx, sy float32) {
	C.sk_canvas_skew(c.ptr, C.float(sx), C.float(sy))
	runtime.KeepAlive(c)
}

// Concat post-multiplies the current matrix by m.
func (c *Canvas) Concat(m Matrix) {
	cm := m.c44()
	C.sk_canvas_concat(c.ptr, &cm)
	runtime.KeepAlive(c)
}

// SetMatrix replaces the current matrix.
func (c *Canvas) SetMatrix(m Matrix) {
	cm := m.c44()
	C.sk_canvas_set_matrix(c.ptr, &cm)
	runtime.KeepAlive(c)
}

// ResetMatrix sets the current matrix to the identity.
func (c *Canvas) ResetMatrix() {
	C.sk_canvas_reset_matrix(c.ptr)
	runtime.KeepAlive(c)
}

// GetMatrix returns the current total matrix.
func (c *Canvas) GetMatrix() Matrix {
	var cm C.sk_matrix44_t
	C.sk_canvas_get_matrix(c.ptr, &cm)
	runtime.KeepAlive(c)
	return matrixFromC44(cm)
}

// ClipRect intersects (or otherwise combines) the clip with r.
func (c *Canvas) ClipRect(r Rect, op ClipOp, doAA bool) {
	cr := r.c()
	C.sk_canvas_clip_rect_with_operation(c.ptr, &cr, C.sk_clipop_t(op), C.bool(doAA))
	runtime.KeepAlive(c)
}

// ClipPath combines the clip with path.
func (c *Canvas) ClipPath(path *Path, op ClipOp, doAA bool) {
	C.sk_canvas_clip_path_with_operation(c.ptr, path.ptr, C.sk_clipop_t(op), C.bool(doAA))
	runtime.KeepAlive(c)
	runtime.KeepAlive(path)
}
