package tools

import "testing"

func TestHumanize(t *testing.T) {
	h := HumanFloat64(123416789.123)
	if h != "123.4M" {
		t.Errorf("Expected '123.4M' got '%v'", h)
	}
	h = HumanFloat64(123416.123)
	if h != "123.4K" {
		t.Errorf("Expected '123.4K' got '%v'", h)
	}
	h = HumanFloat64(123.123)
	if h != "123.1" {
		t.Errorf("Expected '123.1' got '%v'", h)
	}
	h = HumanFloat64(0.123)
	if h != "0.123" {
		t.Errorf("Expected '0.123' got '%v'", h)
	}
}
