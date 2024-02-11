package virtualterm

import (
	"fmt"
	"testing"
)

func TestAddClearTabStop(t *testing.T) {
	term := NewTerminal(nil)

	tests := []struct {
		CurPos int32
	}{
		{
			CurPos: 1,
		},
		{
			CurPos: 2,
		},
		{
			CurPos: 9,
		},
		{
			CurPos: 23,
		},
		{
			CurPos: 2,
		},
		{
			CurPos: 50,
		},
		{
			CurPos: 2,
		},
	}

	expected := "[1 8 9 16 23 24 32 40 48 50 56 64 72]"

	for _, test := range tests {
		term.curPos.X = test.CurPos
		term.c1AddTabStop()
	}

	term.csiClearTabStop()

	if fmt.Sprintf("%v", term._tabStops) != expected {
		t.Errorf("Expected does not match actual in test:")
		t.Logf("  expected: %s", expected)
		t.Logf("  actual:   %v", term._tabStops)
	}
}

func TestNextTabStop(t *testing.T) {
	term := NewTerminal(nil)

	tests := []struct {
		CurPos   int32
		Expected int32
	}{
		{
			CurPos:   0,
			Expected: 8,
		},
		{
			CurPos:   2,
			Expected: 6,
		},
		{
			CurPos:   8,
			Expected: 8,
		},
		{
			CurPos:   23,
			Expected: 1,
		},
	}

	for i, test := range tests {
		term.curPos.X = test.CurPos
		actual := term.nextTabStop()

		if actual != test.Expected {
			t.Errorf("Expected does not match actual in test %d:", i)
			t.Logf("  curPos.X: %d", term.curPos.X)
			t.Logf("  tabStops: %v", term._tabStops)
			t.Logf("  Expected: %d", test.Expected)
			t.Logf("  Actual:   %d", actual)
		}
	}
}
