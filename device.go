// Copyright ©2023 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ardilla

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"io"

	"golang.org/x/image/bmp"
)

const vidElGato = 0x0fd9

// PID is an El Gato HID product ID.
//
//go:generate stringer -type PID
type PID uint16

const (
	StreamDeckMini       PID = 0x0063
	StreamDeckMiniV2     PID = 0x0090
	StreamDeckOriginal   PID = 0x0060
	StreamDeckOriginalV2 PID = 0x006d
	StreamDeckMK2        PID = 0x0080
	StreamDeckXL         PID = 0x006c
	StreamDeckPedal      PID = 0x0086
)

// device is an El Gato Stream Deck device description.
type device struct {
	PID

	cols int
	rows int

	visual    bool
	keySize   image.Point
	transform func(image.Image) image.Image
	encode    func(io.Writer, image.Image) error

	imgReportLen int
	imageHeader  []byte
	fillHeader   func(dst []byte, key, page, len int, done bool)

	payloadLen       int
	serialPayloadLen int

	// prefixes
	resetKeyStream []byte
	reset          []byte
	brightness     []byte
	serial         []byte
	firmware       []byte

	// offsets
	keyStatesOffset int
	serialOffset    int
	firmwareOffset  int
}

func (d *device) bufLen() int {
	if d.serialPayloadLen != 0 {
		return d.serialPayloadLen
	}
	return d.payloadLen
}

func (d *device) bounds() image.Rectangle {
	return image.Rectangle{Max: d.keySize}
}

func transpose(img image.Image) image.Image {
	return t{img}
}

type t struct{ image.Image }

func (i t) At(x, y int) color.Color {
	b := i.Bounds()
	return i.Image.At(y-b.Min.Y+b.Min.X, x-b.Min.X+b.Min.Y)
}

func rotate180(img image.Image) image.Image {
	return r180{img}
}

type r180 struct{ image.Image }

func (i r180) At(x, y int) color.Color {
	b := i.Bounds()
	return i.Image.At(b.Dx()-x+2*b.Min.X, b.Dy()-y+2*b.Min.Y)
}

func jpegEncode(w io.Writer, img image.Image) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: 95})
}

var devices = map[PID]device{
	StreamDeckMini: {
		PID: StreamDeckMini,

		cols: 3, rows: 2,

		visual:    true,
		keySize:   image.Point{80, 80},
		transform: transpose,
		encode:    bmp.Encode,

		imgReportLen: 1024,
		imageHeader:  []byte{0x02, 0x01, 0xff /*page*/, 0x00, 0xff /*done*/, 0xff /*key+1*/, 15: 0},
		fillHeader:   writeHeaderV1,

		payloadLen: 17,

		resetKeyStream: []byte{0x02},
		reset:          []byte{0x0b, 0x63},
		brightness:     []byte{0x05, 0x55, 0xaa, 0xd1, 0x01},
		serial:         []byte{0x03},
		serialOffset:   5,
		firmware:       []byte{0x04},
		firmwareOffset: 5,

		keyStatesOffset: 1,
	},

	StreamDeckMiniV2: {
		PID: StreamDeckMiniV2,

		cols: 3, rows: 2,

		visual:    true,
		keySize:   image.Point{80, 80},
		transform: transpose,
		encode:    bmp.Encode,

		imgReportLen: 1024,
		imageHeader:  []byte{0x02, 0x01, 0xff /*page*/, 0x00, 0xff /*done*/, 0xff /*key+1*/, 15: 0},
		fillHeader:   writeHeaderV1,

		payloadLen:       17,
		serialPayloadLen: 32,

		resetKeyStream: []byte{0x02},
		reset:          []byte{0x0b, 0x63},
		brightness:     []byte{0x05, 0x55, 0xaa, 0xd1, 0x01},
		serial:         []byte{0x03},
		serialOffset:   5,
		firmware:       []byte{0x04},
		firmwareOffset: 5,

		keyStatesOffset: 1,
	},

	StreamDeckOriginal: {
		PID: StreamDeckOriginal,

		cols: 5, rows: 3,

		visual:    true,
		keySize:   image.Point{72, 72},
		transform: rotate180,
		encode:    bmp.Encode,

		imgReportLen: 8191,
		imageHeader:  []byte{0x02, 0x01, 0xff /*page*/, 0x00, 0xff /*done*/, 0xff /*key+1*/, 15: 0},
		fillHeader:   writeHeaderV1,

		payloadLen: 17,

		resetKeyStream: []byte{0x02},
		reset:          []byte{0x0b, 0x63},
		brightness:     []byte{0x05, 0x55, 0xaa, 0xd1, 0x01},
		serial:         []byte{0x03},
		serialOffset:   5,
		firmware:       []byte{0x04},
		firmwareOffset: 5,

		keyStatesOffset: 1,
	},

	StreamDeckOriginalV2: {
		PID: StreamDeckOriginalV2,

		cols: 5, rows: 3,

		visual:    true,
		keySize:   image.Point{72, 72},
		transform: rotate180,
		encode:    jpegEncode,

		imgReportLen: 1024,
		imageHeader:  []byte{0x02, 0x07, 0xff /*key*/, 0xff /*done*/, 0xff, 0xff /*length le*/, 0xff, 0xff /*page le*/},
		fillHeader:   writeHeaderV2,

		payloadLen: 32,

		resetKeyStream: []byte{0x02},
		reset:          []byte{0x03, 0x02},
		brightness:     []byte{0x03, 0x08},
		serial:         []byte{0x06},
		serialOffset:   2,
		firmware:       []byte{0x05},
		firmwareOffset: 6,

		keyStatesOffset: 4,
	},

	StreamDeckMK2: {
		PID: StreamDeckMK2,

		cols: 5, rows: 3,

		visual:    true,
		keySize:   image.Point{72, 72},
		transform: rotate180,
		encode:    jpegEncode,

		imgReportLen: 1024,
		imageHeader:  []byte{0x02, 0x07, 0xff /*key*/, 0xff /*done*/, 0xff, 0xff /*length le*/, 0xff, 0xff /*page le*/},
		fillHeader:   writeHeaderV2,

		payloadLen: 32,

		resetKeyStream: []byte{0x02},
		reset:          []byte{0x03, 0x02},
		brightness:     []byte{0x03, 0x08},
		serial:         []byte{0x06},
		serialOffset:   2,
		firmware:       []byte{0x05},
		firmwareOffset: 6,

		keyStatesOffset: 4,
	},

	StreamDeckXL: {
		PID: StreamDeckXL,

		cols: 8, rows: 4,

		visual:    true,
		keySize:   image.Point{96, 96},
		transform: rotate180,
		encode:    jpegEncode,

		imgReportLen: 1024,
		imageHeader:  []byte{0x02, 0x07, 0xff /*key*/, 0xff /*done*/, 0xff, 0xff /*length le*/, 0xff, 0xff /*page le*/},
		fillHeader:   writeHeaderV2,

		payloadLen: 32,

		resetKeyStream: []byte{0x02},
		reset:          []byte{0x03, 0x02},
		brightness:     []byte{0x03, 0x08},
		serial:         []byte{0x06},
		serialOffset:   2,
		firmware:       []byte{0x05},
		firmwareOffset: 6,

		keyStatesOffset: 4,
	},

	StreamDeckPedal: {
		PID: StreamDeckPedal,

		cols: 3, rows: 1,

		payloadLen: 32,

		serial:         []byte{0x06},
		serialOffset:   2,
		firmware:       []byte{0x05},
		firmwareOffset: 6,

		keyStatesOffset: 4,
	},
}

func writeHeaderV1(dst []byte, key, page, len int, done bool) {
	dst[2] = byte(page)
	dst[4] = boolByte(done)
	dst[5] = byte(key + 1)
}

func writeHeaderV2(dst []byte, key, page, len int, done bool) {
	dst[2] = byte(key)
	dst[3] = boolByte(done)
	binary.LittleEndian.PutUint16(dst[4:], uint16(len))
	binary.LittleEndian.PutUint16(dst[6:], uint16(page))
}

func boolByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
