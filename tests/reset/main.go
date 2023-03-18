// Copyright ©2023 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"

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
	ser := flag.String("serial", "", "device serial number")
	flag.Parse()

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

	d, err := ardilla.NewDeck(pid, *ser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open device: %v\n", err)
		return 1
	}
	defer d.Close()

	err = d.Reset()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to reset device: %v\n", err)
		return 1
	}

	return 0
}
