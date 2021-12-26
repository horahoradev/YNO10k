package protocol

import (
	"fmt"
	"testing"
)

type wow struct {
	MatchPrefix string `ynoproto:"m"`
	One         int    `ynoproto:"nonempty"`
	Two         int    `ynoproto:"nonnegative"`
}

func TestMarshal(t *testing.T) {
	testMsg := []byte(fmt.Sprintf("%s\uffff%s\uffff%s", "m", "1", "2"))

	w := wow{}

	matched, err := Marshal(testMsg, &w)
	if err != nil {
		t.Fatalf("Marshal error: %s", err)
	}

	if !matched {
		t.Fatalf("Failed to match")
	}

	if !(w.One == 1 && w.Two == 2) {
		t.Fatalf("Parse error")
	}

}
