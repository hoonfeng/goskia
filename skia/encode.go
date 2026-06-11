package skia

// #include "goskia.h"
import "C"

import (
	"fmt"
	"unsafe"
)

// encodePixmap encodes the pixels described by pm into the requested format and
// returns a Go copy of the encoded bytes.
func encodePixmap(pm *C.sk_pixmap_t, format EncodedFormat, quality int) ([]byte, error) {
	if quality <= 0 || quality > 100 {
		quality = 100
	}

	ws := C.sk_dynamicmemorywstream_new()
	if ws == nil {
		return nil, fmt.Errorf("skia: encode: failed to create write stream")
	}
	defer C.sk_dynamicmemorywstream_destroy(ws)
	base := (*C.sk_wstream_t)(unsafe.Pointer(ws))

	var ok C.bool
	switch format {
	case EncodedFormatPNG:
		// NOTE: SkiaSharp's sk_pngencoder_encode dereferences the options
		// pointer unconditionally, so it must never be nil.
		opts := C.sk_pngencoder_options_t{
			fFilterFlags: C.sk_pngencoder_filterflags_t(C.ALL_SK_PNGENCODER_FILTER_FLAGS),
			fZLibLevel:   C.int(6),
		}
		ok = C.sk_pngencoder_encode(base, pm, &opts)
	case EncodedFormatJPEG:
		opts := C.sk_jpegencoder_options_t{
			fQuality:     C.int(quality),
			fDownsample:  C.sk_jpegencoder_downsample_t(C.DOWNSAMPLE_420_SK_JPEGENCODER_DOWNSAMPLE),
			fAlphaOption: C.sk_jpegencoder_alphaoption_t(C.IGNORE_SK_JPEGENCODER_ALPHA_OPTION),
		}
		ok = C.sk_jpegencoder_encode(base, pm, &opts)
	case EncodedFormatWEBP:
		opts := C.sk_webpencoder_options_t{
			fCompression: C.sk_webpencoder_compression_t(C.LOSSY_SK_WEBPENCODER_COMPTRESSION),
			fQuality:     C.float(quality),
		}
		ok = C.sk_webpencoder_encode(base, pm, &opts)
	default:
		return nil, fmt.Errorf("skia: encode: unsupported format %d (supported: PNG, JPEG, WEBP)", format)
	}
	if !bool(ok) {
		return nil, fmt.Errorf("skia: encode: encoder failed for format %d", format)
	}

	data := C.sk_dynamicmemorywstream_detach_as_data(ws)
	if data == nil {
		return nil, fmt.Errorf("skia: encode: failed to detach encoded data")
	}
	defer C.sk_data_unref(data)
	n := C.sk_data_get_size(data)
	if n == 0 {
		return nil, nil
	}
	return C.GoBytes(C.sk_data_get_data(data), C.int(n)), nil
}
