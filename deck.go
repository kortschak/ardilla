// Copyright ©2023 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ardilla

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"time"
	"unsafe"

	"golang.org/x/image/draw"

	"github.com/sstallion/go-hid"
)

// Deck is a Stream Deck device.
type Deck struct {
	desc   *device
	serial string // serial is the cached serial for reconnection.
	dev    hidDevice
	buf    []byte
}

type hidDevice interface {
	io.Reader
	io.Writer
	io.Closer
	GetFeatureReport([]byte) (int, error)
	SendFeatureReport([]byte) (int, error)
}

// NewDeck returns the first a Deck using the HID corresponding the the given
// Stream Deck pid and serial. If serial is empty the first matching pid is
// used.
func NewDeck(pid PID, serial string) (*Deck, error) {
	desc, ok := devices[pid]
	if !ok && pid != hid.ProductIDAny {
		return nil, fmt.Errorf("%s not a valid deck device identifier", pid)
	}
	if pid == hid.ProductIDAny {
		// Find the first El Gato device with matching serial.
		hid.Enumerate(vidElGato, uint16(pid), func(info *hid.DeviceInfo) error {
			if serial == "" || serial == info.SerialNbr {
				pid = PID(info.ProductID)
			}
			return io.EOF
		})
		desc, ok = devices[pid]
		if !ok {
			return nil, fmt.Errorf("%s not a known deck device identifier", pid)
		}
	}
	var (
		dev hidDevice
		err error
	)
	if serial != "" {
		dev, err = hid.Open(vidElGato, uint16(pid), serial)
	} else {
		dev, err = hid.OpenFirst(vidElGato, uint16(pid))
	}
	if err != nil {
		return nil, err
	}
	d := &Deck{desc: &desc, serial: serial, dev: dev, buf: make([]byte, desc.bufLen())}
	err = d.ResetKeyStream()
	if err != nil {
		d.dev.Close()
		return nil, err
	}
	if d.serial == "" {
		d.serial, err = d.Serial()
		if err != nil {
			d.dev.Close()
			return nil, err
		}
	}
	return d, nil
}

// ErrNotConnected indicates that the Deck is no longer connected.
var ErrNotConnected = errors.New("device not connected")

// Reconnect attempts to reconnect to the receiver's device each delay until
// successful or the context is cancelled. Reconnect returns the last error
// if ctx is cancelled.
func (d *Deck) Reconnect(ctx context.Context, delay time.Duration) error {
	var err error
	for {
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return err
		case <-timer.C:
		}
		var found bool
		hid.Enumerate(vidElGato, uint16(d.PID()), func(info *hid.DeviceInfo) error {
			if info.SerialNbr == d.serial {
				found = true
			}
			return nil
		})
		if !found {
			err = ErrNotConnected
			continue
		}
		var _d *Deck
		_d, err = NewDeck(d.PID(), d.serial)
		if err == nil {
			d.Close()
			*d = *_d
			return nil
		}
	}
	return err
}

func (d *Deck) checkConnected(err error) error {
	if err == nil {
		return nil
	}
	var found bool
	hid.Enumerate(vidElGato, uint16(d.PID()), func(info *hid.DeviceInfo) error {
		if info.SerialNbr == d.serial {
			found = true
		}
		return nil
	})
	if !found {
		return ErrNotConnected
	}
	return err
}

// Serials returns the list of El Gato device serial numbers matching the
// provided product ID.
func Serials(pid PID) ([]string, error) {
	_, ok := devices[pid]
	if !ok {
		return nil, fmt.Errorf("%s not a valid deck device identifier", pid)
	}
	var serials []string
	err := hid.Enumerate(vidElGato, uint16(pid), func(info *hid.DeviceInfo) error {
		serials = append(serials, info.SerialNbr)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return serials, nil
}

// ResetKeyStream sends a blank key report to the Stream Deck, resetting the
// key image streamer in the device. This prevents previously started partial
// writes from corrupting images sent later.
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

// KeyStates returns a slice of booleans indicating which buttons are pressed.
// The length of the returned slice is given by the Len method.
func (d *Deck) KeyStates() ([]bool, error) {
	buf := make([]byte, d.desc.keyStatesOffset+d.Len())
	_, err := d.dev.Read(buf)
	if err != nil {
		return nil, d.checkConnected(err)
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
	return d.checkConnected(err)
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
	return d.checkConnected(err)
}

// SetImage renders the provided image on the button at the given row and
// column. If img is a *RawImage the internal representation will be used
// directly.
func (d *Deck) SetImage(row, col int, img image.Image) error {
	if row < 0 || d.desc.rows < row {
		return fmt.Errorf("row out of bounds: %d", row)
	}
	if col < 0 || d.desc.cols < col {
		return fmt.Errorf("column out of bounds: %d", col)
	}
	key := row*d.desc.cols + col

	var (
		raw *RawImage
		err error
	)
	switch img := img.(type) {
	case *RawImage:
		if img.pid == d.desc.PID {
			raw = img
			break
		}
		// Unwrap the original and reprocess.
		raw, err = d.RawImage(img.rawImage.Image) //lint:ignore QF1008 rawImage included for clarity.
		if err != nil {
			return err
		}
	default:
		raw, err = d.RawImage(img)
		if err != nil {
			return err
		}
	}
	buf := bytes.NewReader(raw.data)

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
			return d.checkConnected(err)
		}
		page++
	}
	return nil
}

// RawImage returns an image.Image has had the internal image representation
// pre-computed after resizing to fit the Deck's button size. The original image
// is retained in the returned image.
func (d *Deck) RawImage(img image.Image) (*RawImage, error) {
	if !d.desc.visual {
		return nil, fmt.Errorf("images not supported by %s", d.desc)
	}
	if raw, ok := img.(*RawImage); ok {
		if raw.pid == d.desc.PID {
			return raw, nil
		}
		// Unwrap the original and reprocess.
		img = raw.Image
	}

	orig := img
	if img.Bounds() != d.desc.bounds() {
		dst := image.NewRGBA(d.desc.bounds())
		draw.BiLinear.Scale(dst, keepAspectRatio(dst, img), img, img.Bounds(), draw.Src, nil)
		img = dst
	}

	var buf bytes.Buffer
	err := d.desc.encode(&buf, d.desc.transform(img))
	if err != nil {
		return nil, err
	}
	return &RawImage{rawImage{
		Image: orig,
		data:  buf.Bytes(),
		pid:   d.desc.PID,
	}}, nil
}

// RawImage is an image.Image that holds pre-computed data in the raw format
// used by a specific El Gato Stream Deck device.
type RawImage struct {
	rawImage
}

type rawImage struct {
	image.Image
	data []byte
	pid  PID
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

// PID returns the effective PID of the receiver..
func (d *Deck) PID() PID {
	return d.desc.PID
}

// Serial returns the serial number of the device.
func (d *Deck) Serial() (string, error) {
	if d.serial != "" {
		return d.serial, nil
	}
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
	return string(buf[:idx]), d.checkConnected(err)
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
	return string(buf[:idx]), d.checkConnected(err)
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
