package main

import (
	"math"
	"testing"
)

func TestAbs(t *testing.T) {
	got := int(math.Abs(-1))
	if got != 1 {
		t.Errorf("Abs(-1) = %d; want 1", got)
	}
}
