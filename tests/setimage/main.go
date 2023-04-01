// Copyright ©2023 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"io"
	"os"
	"time"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"

	"github.com/kortschak/ardilla"
	"golang.org/x/image/draw"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	pids := []ardilla.PID{
		ardilla.StreamDeckMini,
		ardilla.StreamDeckMiniV2,
		ardilla.StreamDeckOriginal,
		ardilla.StreamDeckOriginalV2,
		ardilla.StreamDeckMK2,
		ardilla.StreamDeckXL,
		ardilla.StreamDeckPedal,
	}

	dev := flag.String("device", "", fmt.Sprintf("device name from %s", pids))
	ser := flag.String("serial", "", "device serial number")
	path := flag.String("image", "", "filename of image (bmp, gif, jpeg, png or tiff)")
	row := flag.Int("row", 0, "row of target button")
	col := flag.Int("col", 0, "column of target button")
	flag.Parse()

	if *path == "" {
		flag.Usage()
		os.Exit(2)
	}

	pid := ardilla.PID(0xffff)
	for _, id := range pids {
		if *dev == "" {
			pid = 0
			break
		}
		if *dev == id.String() {
			pid = id
			break
		}
	}
	if pid == 0xffff {
		fmt.Fprintf(os.Stderr, "%q is not a known device", *dev)
		flag.Usage()
		return 2
	}

	f, err := os.Open(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open image data: %v\n", err)
		return 1
	}
	defer f.Close()

	var img image.Image
	// Work around the effective immutability of image.Decode type registration.
	r := asReaderPeaker(f)
	if hasMagic("GIF8?a", r) {
		img, err = decodeAllGIF(r)
	} else {
		img, _, err = image.Decode(r)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode image data: %v\n", err)
		return 1
	}

	d, err := ardilla.NewDeck(pid, *ser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open device: %v\n", err)
		return 1
	}
	defer d.Close()

	switch img := img.(type) {
	case aGIF:
		dst := image.NewRGBA(img.Bounds())
		err = img.animate(context.Background(), dst, func(img image.Image) error {
			return d.SetImage(*row, *col, img)
		})
	default:
		err = d.SetImage(*row, *col, img)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set image: %v\n", err)
		return 1
	}

	return 0
}

// hasMagic returns whether r starts with the provided magic bytes.
func hasMagic(magic string, r readPeaker) bool {
	b, err := r.Peek(len(magic))
	if err != nil || len(b) != len(magic) {
		return false
	}
	for i, c := range b {
		if magic[i] != c && magic[i] != '?' {
			return false
		}
	}
	return true
}

// readPeaker is an io.Reader that can also peek n bytes ahead.
type readPeaker interface {
	io.Reader
	Peek(n int) ([]byte, error)
}

// asReader converts an io.Reader to a readPeaker.
func asReaderPeaker(r io.Reader) readPeaker {
	if r, ok := r.(readPeaker); ok {
		return r
	}
	return bufio.NewReader(r)
}

// aGIF is an animated GIF.
type aGIF struct {
	*gif.GIF
}

// decodeAllGIF returns an aGIF or gif.GIF decoded from the provided io.Reader.
// If the GIF data encodes a single frame, the image returned is a gif.GIF,
// otherwise an aGIF is returned. When the result is an aGIF, GIF delay,
// disposal and global background index values are checked for validity.
func decodeAllGIF(r io.Reader) (image.Image, error) {
	g, err := gif.DecodeAll(r)
	if err != nil {
		return nil, err
	}
	if len(g.Image) == 1 {
		return g.Image[0], nil
	}
	if len(g.Image) != len(g.Delay) && g.Delay != nil {
		return nil, fmt.Errorf("mismatched image count and delay count: %d != %d", len(g.Image), len(g.Delay))
	}
	if len(g.Image) != len(g.Disposal) && g.Disposal != nil {
		return nil, fmt.Errorf("mismatched image count and disposal count: %d != %d", len(g.Image), len(g.Disposal))
	}
	pal, ok := g.Config.ColorModel.(color.Palette)
	if idx := int(g.BackgroundIndex); ok && idx >= len(pal) {
		return nil, fmt.Errorf("global background colour index not in palette: %d", idx)
	}
	return aGIF{g}, nil
}

func (img aGIF) ColorModel() color.Model {
	if img.Config.ColorModel != nil {
		return img.Config.ColorModel
	}
	return img.GIF.Image[0].ColorModel()
}

func (img aGIF) Bounds() image.Rectangle {
	return img.GIF.Image[0].Bounds()
}

func (img aGIF) At(x, y int) color.Color {
	return img.GIF.Image[0].At(x, y)
}

// animate renders the receiver's frames into dst and calls fn on each
// rendered frame.
func (img aGIF) animate(ctx context.Context, dst draw.Image, fn func(image.Image) error) error {
	const (
		restoreBackground = 2
		restorePrevious   = 3
	)
	var background image.Image
	pal, ok := img.Config.ColorModel.(color.Palette)
	if idx := int(img.BackgroundIndex); ok {
		background = &image.Uniform{pal[idx]}
	}

	loopCount := img.LoopCount
	if loopCount <= 0 {
		loopCount = -loopCount - 1
	}
	for i := 0; i <= loopCount || loopCount == -1; i++ {
		for f, frame := range img.Image {
			var restore *image.Paletted
			if img.Disposal != nil && img.Disposal[f] == restorePrevious {
				restore = image.NewPaletted(frame.Bounds(), frame.Palette)
				draw.Copy(restore, restore.Bounds().Min, dst, frame.Bounds(), draw.Over, nil)
			}
			draw.Copy(dst, frame.Bounds().Min, frame, frame.Bounds(), draw.Over, nil)
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			err := fn(dst)
			if err != nil {
				return err
			}
			if img.Delay != nil {
				delay := time.NewTimer(10 * time.Duration(img.Delay[f]) * time.Millisecond)
				select {
				case <-ctx.Done():
					delay.Stop()
					return nil
				case <-delay.C:
				}
			} else {
				select {
				case <-ctx.Done():
					return nil
				default:
				}
			}
			if img.Disposal != nil {
				switch img.Disposal[f] {
				case restoreBackground:
					if background == nil {
						if idx := int(img.BackgroundIndex); idx < len(frame.Palette) {
							background = &image.Uniform{frame.Palette[idx]}
						} else {
							// No available background, so make this
							// clear in the rendered image.
							background = &image.Uniform{color.RGBA{R: 0xff, A: 0xff}}
						}
					}
					draw.Copy(dst, frame.Bounds().Min, background, frame.Bounds(), draw.Over, nil)
				case restorePrevious:
					draw.Copy(dst, frame.Bounds().Min, restore, restore.Bounds(), draw.Over, nil)
				}
			}
			select {
			case <-ctx.Done():
				return nil
			default:
			}
		}
	}
	return fn(dst)
}
