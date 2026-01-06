package container

import "testing"

func TestMust(t *testing.T) {
	// Tests that the helper doesn't panic on nil
	must(nil)
}