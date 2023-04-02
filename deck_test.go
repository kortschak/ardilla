// Copyright ©2023 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ardilla

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "regenerate golden images")

var resetKeyStreamTests = []struct {
	pid    PID
	visual bool
	want   string
}{
	{
		pid:    StreamDeckMini,
		visual: true,
		want:   "SendFeatureReport([]byte{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:    StreamDeckMiniV2,
		visual: true,
		want:   "SendFeatureReport([]byte{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:    StreamDeckOriginal,
		visual: true,
		want:   "SendFeatureReport([]byte{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:    StreamDeckOriginalV2,
		visual: true,
		want:   "SendFeatureReport([]byte{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:    StreamDeckMK2,
		visual: true,
		want:   "SendFeatureReport([]byte{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:    StreamDeckXL,
		visual: true,
		want:   "SendFeatureReport([]byte{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:    StreamDeckPedal,
		visual: false,
	},
}

func TestDeckResetKeyStream(t *testing.T) {
	for _, test := range resetKeyStreamTests {
		t.Run(fmt.Sprint(test.pid), func(t *testing.T) {
			d, err := newTestDeck(test.pid)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dev := &virtDev{Writer: io.Discard}
			d.setDev(dev)

			err = d.ResetKeyStream()
			if err != nil {
				t.Errorf("unexpected error for ResetKeyStream: %v", err)
			}
			var wantActions int
			if test.visual {
				wantActions = 1
			}
			if len(dev.actions) != wantActions {
				t.Errorf("unexpected number of actions for ResetKeyStream: %v", err)
			}
			if !test.visual {
				return
			}
			got := dev.actions[0]
			if got != test.want {
				t.Errorf("unexpected action for ResetKeyStream:\ngot: %s\nwant:%s", got, test.want)
			}
		})
	}
}

var resetTests = []struct {
	pid    PID
	visual bool
	want   string
}{
	{
		pid:    StreamDeckMini,
		visual: true,
		want:   "SendFeatureReport([]byte{0xb, 0x63, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:    StreamDeckMiniV2,
		visual: true,
		want:   "SendFeatureReport([]byte{0xb, 0x63, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:    StreamDeckOriginal,
		visual: true,
		want:   "SendFeatureReport([]byte{0xb, 0x63, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:    StreamDeckOriginalV2,
		visual: true,
		want:   "SendFeatureReport([]byte{0x3, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:    StreamDeckMK2,
		visual: true,
		want:   "SendFeatureReport([]byte{0x3, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:    StreamDeckXL,
		visual: true,
		want:   "SendFeatureReport([]byte{0x3, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:    StreamDeckPedal,
		visual: false,
	},
}

func TestDeckReset(t *testing.T) {
	for _, test := range resetTests {
		t.Run(fmt.Sprint(test.pid), func(t *testing.T) {
			d, err := newTestDeck(test.pid)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dev := &virtDev{Writer: io.Discard}
			d.setDev(dev)

			err = d.Reset()
			if err != nil {
				t.Errorf("unexpected error for Reset: %v", err)
			}
			var wantActions int
			if test.visual {
				wantActions = 1
			}
			if len(dev.actions) != wantActions {
				t.Errorf("unexpected number of actions for Reset: %v", err)
			}
			if !test.visual {
				return
			}
			got := dev.actions[0]
			if got != test.want {
				t.Errorf("unexpected action for Reset:\ngot: %s\nwant:%s", got, test.want)
			}
		})
	}
}

var setBrightnessTests = []struct {
	pid     PID
	visual  bool
	percent int
	want    string
	wantErr error
}{
	{
		pid:     StreamDeckMini,
		visual:  true,
		percent: 1,
		want:    "SendFeatureReport([]byte{0x5, 0x55, 0xaa, 0xd1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:     StreamDeckMini,
		visual:  true,
		percent: -1,
		wantErr: errors.New("brightness out of range: -1"),
	},
	{
		pid:     StreamDeckMini,
		visual:  true,
		percent: 101,
		wantErr: errors.New("brightness out of range: 101"),
	},
	{
		pid:     StreamDeckMiniV2,
		visual:  true,
		percent: 1,
		want:    "SendFeatureReport([]byte{0x5, 0x55, 0xaa, 0xd1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:     StreamDeckOriginal,
		visual:  true,
		percent: 1,
		want:    "SendFeatureReport([]byte{0x5, 0x55, 0xaa, 0xd1, 0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:     StreamDeckOriginalV2,
		visual:  true,
		percent: 1,
		want:    "SendFeatureReport([]byte{0x3, 0x8, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:     StreamDeckMK2,
		visual:  true,
		percent: 1,
		want:    "SendFeatureReport([]byte{0x3, 0x8, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:     StreamDeckXL,
		visual:  true,
		percent: 1,
		want:    "SendFeatureReport([]byte{0x3, 0x8, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:     StreamDeckPedal,
		visual:  false,
		percent: 1,
	},
}

func TestDeckSetBrightness(t *testing.T) {
	for _, test := range setBrightnessTests {
		t.Run(fmt.Sprint(test.pid), func(t *testing.T) {
			d, err := newTestDeck(test.pid)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dev := &virtDev{Writer: io.Discard}
			d.setDev(dev)

			err = d.SetBrightness(test.percent)
			if !sameError(err, test.wantErr) {
				t.Errorf("unexpected error for SetImage: got:%v want:%v", err, test.wantErr)
			}
			if err != nil {
				return
			}

			var wantActions int
			if test.visual {
				wantActions = 1
			}
			if len(dev.actions) != wantActions {
				t.Errorf("unexpected number of actions for Reset: %v", err)
			}
			if !test.visual {
				return
			}
			got := dev.actions[0]
			if got != test.want {
				t.Errorf("unexpected action for Reset:\ngot: %s\nwant:%s", got, test.want)
			}
		})
	}
}

var serialTests = []struct {
	pid        PID
	data       string
	want       string
	wantAction string
}{
	{
		pid:        StreamDeckMini,
		data:       padZero("xxxxx0123456789", 17),
		want:       "0123456789",
		wantAction: "GetFeatureReport([]byte{0x3, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:        StreamDeckMiniV2,
		data:       padZero("xxxxx01234567890123456789", 32),
		want:       "01234567890123456789",
		wantAction: "GetFeatureReport([]byte{0x3, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:        StreamDeckOriginal,
		data:       padZero("xxxxx0123456789", 17),
		want:       "0123456789",
		wantAction: "GetFeatureReport([]byte{0x3, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:        StreamDeckOriginalV2,
		data:       padZero("xx01234567890123456789", 32),
		want:       "01234567890123456789",
		wantAction: "GetFeatureReport([]byte{0x6, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:        StreamDeckMK2,
		data:       padZero("xx01234567890123456789", 32),
		want:       "01234567890123456789",
		wantAction: "GetFeatureReport([]byte{0x6, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:        StreamDeckXL,
		data:       padZero("xx01234567890123456789", 32),
		want:       "01234567890123456789",
		wantAction: "GetFeatureReport([]byte{0x6, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:        StreamDeckPedal,
		data:       padZero("xx01234567890123456789", 32),
		want:       "01234567890123456789",
		wantAction: "GetFeatureReport([]byte{0x6, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
}

func TestDeckSerial(t *testing.T) {
	for _, test := range serialTests {
		t.Run(fmt.Sprint(test.pid), func(t *testing.T) {
			d, err := newTestDeck(test.pid)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dev := &virtDev{Reader: strings.NewReader(test.data)}
			d.setDev(dev)

			got, err := d.Serial()
			if err != nil {
				t.Errorf("unexpected error for Serial: %v", err)
			}
			if got != test.want {
				t.Errorf("unexpected result for Serial: got:%s want:%s", got, test.want)
			}
			if len(dev.actions) != 1 {
				t.Errorf("unexpected number of actions for Serial: %v", err)
			}
			gotAction := dev.actions[0]
			if gotAction != test.wantAction {
				t.Errorf("unexpected action for Serial:\ngot: %s\nwant:%s", gotAction, test.wantAction)
			}
		})
	}
}

var firmwareTests = []struct {
	pid        PID
	data       string
	want       string
	wantAction string
}{
	{
		pid:        StreamDeckMini,
		data:       padZero("xxxxx0123456789", 17),
		want:       "0123456789",
		wantAction: "GetFeatureReport([]byte{0x4, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:        StreamDeckMiniV2,
		data:       padZero("xxxxx0123456789", 17),
		want:       "0123456789",
		wantAction: "GetFeatureReport([]byte{0x4, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:        StreamDeckOriginal,
		data:       padZero("xxxxx0123456789", 17),
		want:       "0123456789",
		wantAction: "GetFeatureReport([]byte{0x4, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (17, <nil>)",
	},
	{
		pid:        StreamDeckOriginalV2,
		data:       padZero("xxxxxx0123456789", 32),
		want:       "0123456789",
		wantAction: "GetFeatureReport([]byte{0x5, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:        StreamDeckMK2,
		data:       padZero("xxxxxx0123456789", 32),
		want:       "0123456789",
		wantAction: "GetFeatureReport([]byte{0x5, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:        StreamDeckXL,
		data:       padZero("xxxxxx0123456789", 32),
		want:       "0123456789",
		wantAction: "GetFeatureReport([]byte{0x5, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
	{
		pid:        StreamDeckPedal,
		data:       padZero("xxxxxx0123456789", 32),
		want:       "0123456789",
		wantAction: "GetFeatureReport([]byte{0x5, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}) -> (32, <nil>)",
	},
}

func TestDeckFirmware(t *testing.T) {
	for _, test := range firmwareTests {
		t.Run(fmt.Sprint(test.pid), func(t *testing.T) {
			d, err := newTestDeck(test.pid)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dev := &virtDev{Reader: strings.NewReader(test.data)}
			d.setDev(dev)

			got, err := d.Firmware()
			if err != nil {
				t.Errorf("unexpected error for Serial: %v", err)
			}
			if got != test.want {
				t.Errorf("unexpected result for Serial: got:%s want:%s", got, test.want)
			}
			if len(dev.actions) != 1 {
				t.Errorf("unexpected number of actions for Serial: %v", err)
			}
			gotAction := dev.actions[0]
			if gotAction != test.wantAction {
				t.Errorf("unexpected action for Serial:\ngot: %s\nwant:%s", gotAction, test.wantAction)
			}
		})
	}
}

var keyStateTests = []struct {
	pid        PID
	data       []byte
	want       []bool
	wantAction string
}{
	{
		pid:        StreamDeckMini,
		data:       prependZero(1, []byte{2: 1, 5: 1}),
		want:       []bool{2: true, 5: true},
		wantAction: "Read(7 bytes) -> (7, <nil>)",
	},
	{
		pid:        StreamDeckMiniV2,
		data:       prependZero(1, []byte{2: 1, 5: 1}),
		want:       []bool{2: true, 5: true},
		wantAction: "Read(7 bytes) -> (7, <nil>)",
	},
	{
		pid:        StreamDeckOriginal,
		data:       prependZero(1, []byte{2: 1, 5: 1, 14: 0}),
		want:       []bool{2: true, 5: true, 14: false},
		wantAction: "Read(16 bytes) -> (16, <nil>)",
	},
	{
		pid:        StreamDeckOriginalV2,
		data:       prependZero(4, []byte{2: 1, 5: 1, 14: 0}),
		want:       []bool{2: true, 5: true, 14: false},
		wantAction: "Read(19 bytes) -> (19, <nil>)",
	},
	{
		pid:        StreamDeckMK2,
		data:       prependZero(4, []byte{2: 1, 5: 1, 14: 0}),
		want:       []bool{2: true, 5: true, 14: false},
		wantAction: "Read(19 bytes) -> (19, <nil>)",
	},
	{
		pid:        StreamDeckXL,
		data:       prependZero(4, []byte{2: 1, 5: 1, 31: 0}),
		want:       []bool{2: true, 5: true, 31: false},
		wantAction: "Read(36 bytes) -> (36, <nil>)",
	},
	{
		pid:        StreamDeckPedal,
		data:       prependZero(4, []byte{0: 1, 2: 1}),
		want:       []bool{0: true, 2: true},
		wantAction: "Read(7 bytes) -> (7, <nil>)",
	},
}

func TestDeckKeyStates(t *testing.T) {
	for _, test := range keyStateTests {
		t.Run(fmt.Sprint(test.pid), func(t *testing.T) {
			d, err := newTestDeck(test.pid)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dev := &virtDev{Reader: bytes.NewReader(test.data)}
			d.setDev(dev)

			got, err := d.KeyStates()
			if err != nil {
				t.Errorf("unexpected error for KeyStates: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("unexpected result for KeyStates:\ngot: %v\nwant:%v", got, test.want)
			}
			if len(dev.actions) != 1 {
				t.Errorf("unexpected number of actions for KeyStates: %v", err)
			}
			gotAction := dev.actions[0]
			if gotAction != test.wantAction {
				t.Errorf("unexpected action for KeyStates:\ngot: %s\nwant:%s", gotAction, test.wantAction)
			}
		})
	}
}

var setImageTests = []struct {
	pid         PID
	row         int
	col         int
	format      string
	headerLen   int
	wantHeaders [][]byte
	wantErr     error
}{
	{
		pid: StreamDeckMini, headerLen: 16,
		row: 1, col: 2,
		format: "bmp",
		wantHeaders: [][]byte{
			{0x2, 0x1, 0x0, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x1, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x2, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x3, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x4, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x5, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x6, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x7, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x8, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x9, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xa, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xb, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xc, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xd, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xe, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xf, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x10, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x11, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x12, 0x0, 0x0, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x13, 0x0, 0x1, devices[StreamDeckMini].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
	},
	{
		pid: StreamDeckMini, headerLen: 16,
		row: 3, col: 2,
		wantErr: errors.New("row out of bounds: 3"),
	},
	{
		pid: StreamDeckMiniV2, headerLen: 16,
		row: 1, col: 2,
		format: "bmp",
		wantHeaders: [][]byte{
			{0x2, 0x1, 0x0, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x1, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x2, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x3, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x4, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x5, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x6, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x7, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x8, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x9, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xa, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xb, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xc, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xd, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xe, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0xf, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x10, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x11, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x12, 0x0, 0x0, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x13, 0x0, 0x1, devices[StreamDeckMiniV2].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
	},
	{
		pid: StreamDeckOriginal, headerLen: 16,
		row: 1, col: 2,
		format: "bmp",
		wantHeaders: [][]byte{
			{0x2, 0x1, 0x0, 0x0, 0x0, devices[StreamDeckOriginal].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			{0x2, 0x1, 0x1, 0x0, 0x1, devices[StreamDeckOriginal].key(1, 2) + 1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
	},
	{
		pid: StreamDeckOriginalV2, headerLen: 8,
		row: 1, col: 2,
		format: "jpeg",
		wantHeaders: [][]byte{
			{0x2, 0x7, devices[StreamDeckOriginalV2].key(1, 2), 0x0, 0xf8, 0x3, 0x0, 0x0},
			{0x2, 0x7, devices[StreamDeckOriginalV2].key(1, 2), 0x0, 0xf8, 0x3, 0x1, 0x0},
			{0x2, 0x7, devices[StreamDeckOriginalV2].key(1, 2), 0x1, 0xae, 0x3, 0x2, 0x0},
		},
	},
	{
		pid: StreamDeckMK2, headerLen: 8,
		row: 1, col: 2,
		format: "jpeg",
		wantHeaders: [][]byte{
			{0x2, 0x7, devices[StreamDeckMK2].key(1, 2), 0x0, 0xf8, 0x3, 0x0, 0x0},
			{0x2, 0x7, devices[StreamDeckMK2].key(1, 2), 0x0, 0xf8, 0x3, 0x1, 0x0},
			{0x2, 0x7, devices[StreamDeckMK2].key(1, 2), 0x1, 0xae, 0x3, 0x2, 0x0},
		},
	},
	{
		pid: StreamDeckXL, headerLen: 8,
		row: 1, col: 2,
		format: "jpeg",
		wantHeaders: [][]byte{
			{0x2, 0x7, devices[StreamDeckXL].key(1, 2), 0x0, 0xf8, 0x3, 0x0, 0x0},
			{0x2, 0x7, devices[StreamDeckXL].key(1, 2), 0x0, 0xf8, 0x3, 0x1, 0x0},
			{0x2, 0x7, devices[StreamDeckXL].key(1, 2), 0x0, 0xf8, 0x3, 0x2, 0x0},
			{0x2, 0x7, devices[StreamDeckXL].key(1, 2), 0x0, 0xf8, 0x3, 0x3, 0x0},
			{0x2, 0x7, devices[StreamDeckXL].key(1, 2), 0x1, 0x1a, 0x0, 0x4, 0x0}},
	},
	{
		pid: StreamDeckPedal,
		row: 0, col: 2,
		wantErr: errors.New("images not supported by StreamDeckPedal"),
	},
}

func (d device) key(row, col int) byte {
	return byte(row*d.cols + col)
}

func TestDeckSetImage(t *testing.T) {
	f, err := os.Open("testdata/gopher.png")
	if err != nil {
		t.Fatalf("unable to open test image: %v", err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("unable to open decode image: %v", err)
	}
	for _, test := range setImageTests {
		t.Run(fmt.Sprint(test.pid), func(t *testing.T) {
			for _, precompute := range []bool{false, true} {
				name := "direct"
				if precompute {
					name = "precompute"
				}
				t.Run(name, func(t *testing.T) {
					d, err := newTestDeck(test.pid)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					buf := &imageCapture{headerLen: test.headerLen}
					dev := &virtDev{Writer: buf}
					d.setDev(dev)

					img := img
					if precompute && test.wantErr == nil {
						img, err = d.RawImage(img)
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
					}
					err = d.SetImage(test.row, test.col, img)
					if !sameError(err, test.wantErr) {
						t.Errorf("unexpected error for SetImage: got:%v want:%v", err, test.wantErr)
					}
					if err != nil {
						return
					}

					name := fmt.Sprintf("%s-%d-%d.%s", test.pid, test.row, test.col, test.format)
					path := filepath.Join("testdata", name)
					if *update {
						err = os.WriteFile(path, buf.image, 0o644)
						if err != nil {
							t.Fatalf("unexpected error writing golden file: %v", err)
						}
					}

					want, err := os.ReadFile(path)
					if err != nil {
						t.Fatalf("unexpected error reading golden file: %v", err)
					}

					if !bytes.Equal(buf.image, want) {
						err = os.WriteFile(filepath.Join("testdata", "failed-"+name), buf.image, 0o644)
						if err != nil {
							t.Fatalf("unexpected error writing failed file: %v", err)
						}
						t.Errorf("image mismatch: %s", name)
					}

					if !reflect.DeepEqual(buf.headers, test.wantHeaders) {
						t.Errorf("unexpected header:\ngot: %#v\nwant:%#v", buf.headers, test.wantHeaders)
					}
				})
			}
		})
	}
}

func BenchmarkSetImage(b *testing.B) {
	f, err := os.Open("testdata/gopher.png")
	if err != nil {
		b.Fatalf("unable to open test image: %v", err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		b.Fatalf("unable to open decode image: %v", err)
	}
	for _, pid := range []PID{StreamDeckOriginal, StreamDeckOriginalV2} {
		b.Run(pid.String(), func(b *testing.B) {
			d, err := newTestDeck(pid)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
			d.setDev(&virtDev{Writer: io.Discard})
			b.Run("direct", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					err = d.SetImage(0, 0, img)
					if err != nil {
						b.Errorf("unexpected error for SetImage: %v", err)
					}
				}
			})

			// rawImage contains the resized image and the raw data.
			rawImage, err := d.RawImage(img)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}

			b.Run("resized", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					err = d.SetImage(0, 0, rawImage.Image)
					if err != nil {
						b.Errorf("unexpected error for SetImage: %v", err)
					}
				}
			})

			b.Run("raw", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					err = d.SetImage(0, 0, rawImage)
					if err != nil {
						b.Errorf("unexpected error for SetImage: %v", err)
					}
				}
			})
		})
	}
}

type imageCapture struct {
	headerLen int
	headers   [][]byte
	image     []byte
}

func (w *imageCapture) Write(b []byte) (int, error) {
	if len(b) < w.headerLen {
		w.headers = append(w.headers, append(b[:0:0], b...))
		return len(b), io.ErrShortWrite
	}
	w.headers = append(w.headers, append(b[:0:0], b[:w.headerLen]...))
	w.image = append(w.image, b[w.headerLen:]...)
	return len(b), nil
}

func sameError(a, b error) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil, b == nil, a.Error() != b.Error():
		return false
	default:
		return true
	}
}

func padZero(s string, n int) string {
	if len(s) > n {
		panic("string is too long")
	}
	b := make([]byte, n)
	copy(b, s)
	return string(b)
}

func prependZero(n int, b []byte) []byte {
	return append(make([]byte, n), b...)
}

func newTestDeck(pid PID) (*Deck, error) {
	desc, ok := devices[pid]
	if !ok {
		return nil, fmt.Errorf("%s not a valid deck device identifier", pid)
	}
	d := &Deck{desc: &desc, buf: make([]byte, desc.bufLen())}
	return d, nil
}

func (d *Deck) setDev(dev *virtDev) {
	d.dev = dev
}

type virtDev struct {
	io.Reader
	io.Writer
	io.Closer
	actions []string
}

func (d *virtDev) Read(b []byte) (int, error) {
	n, err := d.Reader.Read(b)
	d.actions = append(d.actions, fmt.Sprintf("Read(%d bytes) -> (%d, %v)", len(b), n, err))
	return n, err
}

func (d *virtDev) Write(b []byte) (int, error) {
	n, err := d.Writer.Write(b)
	d.actions = append(d.actions, fmt.Sprintf("Write(%#v) -> (%d, %v)", b, n, err))
	return n, err
}

func (d *virtDev) Close() error {
	err := d.Closer.Close()
	d.actions = append(d.actions, fmt.Sprintf("Close() -> %v", err))
	return err
}

func (d *virtDev) GetFeatureReport(b []byte) (int, error) {
	s := make([]byte, len(b))
	copy(s, b)
	n, err := d.Reader.Read(b)
	d.actions = append(d.actions, fmt.Sprintf("GetFeatureReport(%#v) -> (%d, %v)", s, n, err))
	return n, err
}

func (d *virtDev) SendFeatureReport(b []byte) (int, error) {
	n, err := d.Writer.Write(b)
	d.actions = append(d.actions, fmt.Sprintf("SendFeatureReport(%#v) -> (%d, %v)", b, n, err))
	return n, err
}
