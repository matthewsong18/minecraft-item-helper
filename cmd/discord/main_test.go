package main

import "testing"

func TestIsPlanCommandPath(t *testing.T) {
	if !isPlanCommandPath("/plan") {
		t.Fatal("/plan command path should be handled")
	}
	if isPlanCommandPath("/other") {
		t.Fatal("unregistered command path should not be handled")
	}
}
