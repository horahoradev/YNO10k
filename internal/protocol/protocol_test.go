package protocol

import (
	"testing"
)

type wow struct {
	MatchPrefix uint8  `ynoproto:"FF"`
	X           uint16 `ynoproto:"nonempty"`
	Y           uint16 `ynoproto:"nonnegative"`
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		payload     []byte
		errExpected error
		match       bool
		Xval        uint16
		Yval        uint16
	}{
		{
			payload:     []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			errExpected: nil,
			match:       true,
			Xval:        0xFFFF,
			Yval:        0xFFFF,
		},
		{
			payload:     []byte{0xFE, 0xFF, 0xFF, 0xFF, 0xFF},
			errExpected: nil,
			match:       false,
			Xval:        0xFFFF,
			Yval:        0xFFFF,
		},
	}

	for i, test := range tests {

		testMsg := test.payload

		w := wow{}

		matched, err := Marshal(testMsg, &w)
		if test.errExpected == nil && err != nil {
			t.Fatalf("Received unexpected error %s", err)
		}

		if err != nil && err.Error() != test.errExpected.Error() {
			t.Fatalf("Did not get expected error: %s", err)
		}

		if test.errExpected == nil && matched != test.match {
			t.Fatalf("Failed to match for payload %d", i)
		}

		// No guarantees on contents if we returned an error
		if err == nil && matched && !(w.X == test.Xval && w.Y == test.Yval) {
			t.Fatalf("Parse error for payload %d", i)
		}

	}
}
