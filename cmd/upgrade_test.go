package cmd

import (
	"runtime"
	"testing"
)

func TestExpectedAssetName(t *testing.T) {
	name := expectedAssetName()
	expected := "polymarket_" + runtime.GOOS + "_" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		expected += ".exe"
	}
	if name != expected {
		t.Errorf("expected %s, got %s", expected, name)
	}
}

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Error("expected non-empty version")
	}
}
