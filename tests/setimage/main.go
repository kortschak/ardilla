// Copyright ©2023 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"image"
	"os"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"

	"github.com/kortschak/ardilla"
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
	path := flag.String("image", "", "filename of image (bmp, gif, jpeg, png or tiff)")
	row := flag.Int("row", 0, "row of target button")
	col := flag.Int("col", 0, "column of target button")
	flag.Parse()

	if *path == "" {
		flag.Usage()
		os.Exit(2)
	}

	var pid ardilla.PID
	for _, id := range pids {
		if *dev == id.String() {
			pid = id
			break
		}
	}
	if pid == 0 {
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

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode image data: %v\n", err)
		return 1
	}

	d, err := ardilla.NewDeck(pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open device: %v\n", err)
		return 1
	}
	defer d.Close()

	err = d.SetImage(*row, *col, img)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set image: %v\n", err)
		return 1
	}

	return 0
}
