package protocol

import (
	"errors"
	"fmt"
	"testing"
)

type wow struct {
	MatchPrefix string `ynoproto:"m"`
	X           int    `ynoproto:"nonempty"`
	Y           int    `ynoproto:"nonnegative"`
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		fmtString   string
		errExpected error
		Xval        int
		Yval        int
	}{
		{
			fmtString:   "%s\uffff%s\uffff%s",
			errExpected: nil,
			Xval:        1,
			Yval:        2,
		},
		{
			fmtString:   "%s\uffff-%s\uffff-%s",
			errExpected: errors.New("message value, -2, for field Y violates nonnegative annotation"),
			Xval:        -1,
			Yval:        -2,
		},
		{
			fmtString:   "%s\uffff-%s\uffff%s",
			errExpected: nil,
			Xval:        -1,
			Yval:        2,
		},
	}

	for _, test := range tests {

		testMsg := []byte(fmt.Sprintf(test.fmtString, "m", "1", "2"))

		w := wow{}

		matched, err := Marshal(testMsg, &w)
		if test.errExpected == nil && err != nil {
			t.Fatalf("Received unexpected error")
		}

		if err != nil && err.Error() != test.errExpected.Error() {
			t.Fatalf("Did not get expected error: %s", err)
		}

		if test.errExpected == nil && !matched {
			t.Fatalf("Failed to match for fmtstring %s", test.fmtString)
		}

		// No guarantees on contents if we returned an error
		if err == nil && !(w.X == test.Xval && w.Y == test.Yval) {
			t.Fatalf("Parse error")
		}

	}
}
