package remote

import (
	"testing"
)

func TestIsRsyncUrl(t *testing.T) {
	var tests = []struct {
		input string
		valid bool
	}{
		{"", false},
		{"rsync:/", false},
		{"rsync://", true},
	}
	for i, test := range tests {
		if test.valid != isRsyncUrl(test.input) {
			t.Errorf("Test: %d should not result in error", i)
		}
	}
}
