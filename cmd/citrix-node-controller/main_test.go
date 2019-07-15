package main

import (
	"testing"
	//"fmt"
)

func TestMain(t *testing.T) {
	error := InitCitrixNodeController()
	if error != nil {
		t.Error("Expected non error case")
	}
}
