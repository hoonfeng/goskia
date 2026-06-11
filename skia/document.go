package skia

// #include "goskia.h"
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// PDFDocument records vector drawing commands into a multi-page PDF held in
// memory. Typical use:
//
//	doc := skia.NewPDFDocument()
//	c := doc.BeginPage(595, 842) // A4 in points
//	// ... draw on c ...
//	doc.EndPage()
//	pdf, _ := doc.Close()        // []byte of the PDF
type PDFDocument struct {
	doc *C.sk_document_t
	ws  *C.sk_wstream_dynamicmemorystream_t
}

// NewPDFDocument creates an empty in-memory PDF document.
func NewPDFDocument() *PDFDocument {
	ws := C.sk_dynamicmemorywstream_new()
	doc := C.sk_document_create_pdf_from_stream((*C.sk_wstream_t)(unsafe.Pointer(ws)))
	if doc == nil {
		C.sk_dynamicmemorywstream_destroy(ws)
		return nil
	}
	return &PDFDocument{doc: doc, ws: ws}
}

// BeginPage starts a new page of the given size (in points) and returns its
// canvas. The canvas is valid until EndPage; do not release it.
func (d *PDFDocument) BeginPage(width, height float32) *Canvas {
	cv := C.sk_document_begin_page(d.doc, C.float(width), C.float(height), nil)
	runtime.KeepAlive(d)
	return &Canvas{ptr: cv, owner: d}
}

// EndPage finishes the current page.
func (d *PDFDocument) EndPage() {
	C.sk_document_end_page(d.doc)
	runtime.KeepAlive(d)
}

// Close finalizes the document and returns the encoded PDF bytes. The document
// must not be used afterwards.
func (d *PDFDocument) Close() ([]byte, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("skia: PDF document already closed")
	}
	C.sk_document_close(d.doc)
	C.sk_document_unref(d.doc)
	d.doc = nil

	data := C.sk_dynamicmemorywstream_detach_as_data(d.ws)
	C.sk_dynamicmemorywstream_destroy(d.ws)
	d.ws = nil
	if data == nil {
		return nil, fmt.Errorf("skia: failed to finalize PDF")
	}
	defer C.sk_data_unref(data)
	n := C.sk_data_get_size(data)
	if n == 0 {
		return nil, nil
	}
	return C.GoBytes(C.sk_data_get_data(data), C.int(n)), nil
}

// Abort discards the document without producing output.
func (d *PDFDocument) Abort() {
	if d.doc != nil {
		C.sk_document_abort(d.doc)
		C.sk_document_unref(d.doc)
		d.doc = nil
	}
	if d.ws != nil {
		C.sk_dynamicmemorywstream_destroy(d.ws)
		d.ws = nil
	}
}
