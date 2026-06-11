package skia

// #include "goskia.h"
import "C"
import "io"

// EncodeTo encodes the image and writes it to w. For PNG quality is ignored;
// for JPEG/WEBP it is the 0..100 quality.
func (img *Image) EncodeTo(w io.Writer, format EncodedFormat, quality int) (int, error) {
	b, err := img.Encode(format, quality)
	if err != nil {
		return 0, err
	}
	return w.Write(b)
}

// EncodeTo snapshots the surface, encodes it, and writes it to w.
func (s *Surface) EncodeTo(w io.Writer, format EncodedFormat, quality int) (int, error) {
	b, err := s.EncodeToBytes(format, quality)
	if err != nil {
		return 0, err
	}
	return w.Write(b)
}

// WritePNGTo snapshots the surface and writes it to w as PNG.
func (s *Surface) WritePNGTo(w io.Writer) error {
	_, err := s.EncodeTo(w, EncodedFormatPNG, 100)
	return err
}
