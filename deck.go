// Copyright ©2023 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ardilla

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"unsafe"

	"golang.org/x/image/draw"

	"github.com/sstallion/go-hid"
)

// Deck is a Stream Deck device.
type Deck struct {
	desc *device
	dev  hidDevice
	buf  []byte
}

type hidDevice interface {
	io.Reader
	io.Writer
	io.Closer
	GetFeatureReport([]byte) (int, error)
	SendFeatureReport([]byte) (int, error)
}

// NewDeck returns the first HID corresponding the the given Stream Deck pid.
func NewDeck(pid PID) (*Deck, error) {
	desc, ok := devices[pid]
	if !ok {
		return nil, fmt.Errorf("%s not a valid deck device identifier", pid)
	}
	dev, err := hid.OpenFirst(vidElGato, uint16(pid))
	if err != nil {
		return nil, err
	}
	d := &Deck{desc: &desc, dev: dev, buf: make([]byte, desc.bufLen())}
	err = d.ResetKeyStream()
	if err != nil {
		d.dev.Close()
		return nil, err
	}
	return d, nil
}

// Sends a blank key report to the Stream Deck, resetting the key image
// streamer in the device. This prevents previously started partial key
// writes that were not completed from corrupting images sent from this
// application.
func (d *Deck) ResetKeyStream() error {
	if !d.desc.visual {
		return nil
	}
	buf := d.buf[:d.desc.payloadLen]
	zero(buf)
	copy(buf, d.desc.resetKeyStream)
	_, err := d.dev.SendFeatureReport(buf)
	return err
}

// Close closes the device.
func (d *Deck) Close() error {
	return d.dev.Close()
}

// Layout returns the number of rows and columns of buttons on the device.
func (d *Deck) Layout() (rows, cols int) {
	return d.desc.rows, d.desc.cols
}

// Key returns the key number corresponding to the given row and column.
// It panics if row or col are out of bounds.
func (d *Deck) Key(row, col int) int {
	if row < 0 || d.desc.rows < row {
		panic(fmt.Sprintf("row out of bounds: %d", row))
	}
	if col < 0 || d.desc.cols < col {
		panic(fmt.Sprintf("column out of bounds: %d", col))
	}
	return row*d.desc.cols + col
}

// Len returns the number of buttons on the device.
func (d *Deck) Len() int {
	return d.desc.rows * d.desc.cols
}

// KeyStates returns a slice of booleans indicating which buttons are pressed
// The length of the returned slice is given by the Len method.
func (d *Deck) KeyStates() ([]bool, error) {
	buf := make([]byte, d.desc.keyStatesOffset+d.Len())
	_, err := d.dev.Read(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[d.desc.keyStatesOffset:]
	return *(*[]bool)(unsafe.Pointer(&buf)), nil
}

// Resets the Stream Deck, clearing all button images and showing the standby
// image.
func (d *Deck) Reset() error {
	if !d.desc.visual {
		return nil
	}
	buf := d.buf[:d.desc.payloadLen]
	zero(buf)
	copy(buf, d.desc.reset)
	_, err := d.dev.SendFeatureReport(buf)
	return err
}

// SetBrightness sets the global screen brightness of the Stream Deck, across
// all the device's buttons.
func (d *Deck) SetBrightness(percent int) error {
	if !d.desc.visual {
		return nil
	}
	if percent < 0 || 100 < percent {
		return fmt.Errorf("brightness out of range: %d", percent)
	}
	buf := d.buf[:d.desc.payloadLen]
	zero(buf)
	copy(buf, d.desc.brightness)
	buf[len(d.desc.brightness)] = byte(percent)
	_, err := d.dev.SendFeatureReport(buf)
	return err
}

// SetImage renders the provided image on the button at the given row and
// column.
func (d *Deck) SetImage(row, col int, img image.Image) error {
	if !d.desc.visual {
		return fmt.Errorf("images not supported by %s", d.desc)
	}
	if row < 0 || d.desc.rows < row {
		return fmt.Errorf("row out of bounds: %d", row)
	}
	if col < 0 || d.desc.cols < col {
		return fmt.Errorf("column out of bounds: %d", col)
	}
	key := row*d.desc.cols + col

	if img.Bounds() != d.desc.bounds() {
		dst := image.NewRGBA(d.desc.bounds())
		draw.BiLinear.Scale(dst, keepAspectRatio(dst, img), img, img.Bounds(), draw.Src, nil)
		img = dst
	}

	var buf bytes.Buffer
	err := d.desc.encode(&buf, d.desc.transform(img))
	if err != nil {
		return err
	}
	pkt := make([]byte, d.desc.imgReportLen)
	copy(pkt, d.desc.imageHeader)
	var page int
	for buf.Len() != 0 {
		n, err := buf.Read(pkt[len(d.desc.imageHeader):])
		if err != nil && err != io.EOF {
			return err
		}
		done := buf.Len() == 0 || n < d.desc.imgReportLen-len(d.desc.imageHeader)
		d.desc.fillHeader(pkt[:len(d.desc.imageHeader)], key, page, n, done)
		_, err = d.dev.Write(pkt)
		if err != nil {
			return err
		}
		page++
	}
	return nil
}

func keepAspectRatio(dst, src image.Image) image.Rectangle {
	b := dst.Bounds()
	dx, dy := src.Bounds().Dx(), src.Bounds().Dy()
	switch {
	case dx < dy:
		dx, dy = dx*b.Max.X/dy, b.Max.Y
	case dx > dy:
		dx, dy = b.Max.X, dy*b.Max.Y/dx
	default:
		return b
	}
	offset := image.Point{X: (b.Dx() - dx) / 2, Y: (b.Dy() - dy) / 2}
	return image.Rectangle{Max: image.Point{X: dx, Y: dy}}.Add(offset)
}

// Bounds returns the image bounds for buttons on the device. If the device
// is not visual an error is returned.
func (d *Deck) Bounds() (image.Rectangle, error) {
	if !d.desc.visual {
		return image.Rectangle{}, fmt.Errorf("images not supported by %s", d.desc)
	}
	return d.desc.bounds(), nil
}

// Serial returns the serial number of the device.
func (d *Deck) Serial() (string, error) {
	payloadLen := d.desc.serialPayloadLen
	if payloadLen == 0 {
		payloadLen = d.desc.payloadLen
	}
	buf := d.buf[:payloadLen]
	zero(buf)
	copy(buf, d.desc.serial)
	buf[len(d.desc.serial)] = byte(payloadLen)
	_, err := d.dev.GetFeatureReport(buf)
	buf = buf[d.desc.serialOffset:]
	idx := bytes.IndexByte(buf, 0)
	if idx < 0 {
		return string(buf), nil
	}
	return string(buf[:idx]), err
}

// Firmware returns the firmware version number of the device.
func (d *Deck) Firmware() (string, error) {
	buf := d.buf[:d.desc.payloadLen]
	zero(buf)
	copy(buf, d.desc.firmware)
	buf[len(d.desc.firmware)] = byte(d.desc.payloadLen)
	_, err := d.dev.GetFeatureReport(buf)
	buf = buf[d.desc.firmwareOffset:]
	idx := bytes.IndexByte(buf, 0)
	if idx < 0 {
		return string(buf), nil
	}
	return string(buf[:idx]), err
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
